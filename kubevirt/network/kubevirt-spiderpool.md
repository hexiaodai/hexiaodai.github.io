# KubeVirt + SpiderPool 使用指南

## 安装 SpiderPool

安装文档：<https://spidernet-io.github.io/spiderpool/v0.7/usage/install/overlay/get-started-calico-zh_cn>

注意：

- SpiderPool 版本必须大于等于 v0.8
- 您需指定 enableKubevirtStaticIP: true（在 v0.8 版本中，这是默认值）

  例如，使用 helm

  ```bash
  helm install spiderpool --set global.ipam.enableKubevirtStaticIP=true
  ```

  或者安装完 SpiderPool 后，使用 kubectl，补充 `enableKubevirtStaticIP: true`

  ```bash
  kubectl -n kube-system edit cm spiderpool-conf
  ```

## KubeVirt + SpiderPool 网络方案

> SpiderPool 版本：v0.8
> 
> KubeVirt 版本：v1.0.1

| 网络模式 | CNI | 是否安装 SpiderPool | VirtualMachine IP | 固定 IP | Live Migration |
| ---- | ---- | ---- | ---- | ---- | ---- |
| Masquerade | Calico | 否 | 私网 IP | ❌ | ✅ |
|  | Cilium | 否 | 私网 IP | ❌ | ✅ |
|  | Flannel | 否 | 私网 IP | ❌ | ✅ |
| Passt | macvlan | 是 | Pod IP | ✅ | ❌ |
|  | ipvlan | 是 | Pod IP | ✅ | ❌ |
| Bridge | ovs | 是 | Pod IP | ✅ | ❌ |

> Masquerade 模式：KubeVirt 通过使用 NAT 技术将私有 IP 地址范围 10.0.2/24 网段分配给虚拟机，并将其隐藏在 Pod IP 地址之后。
>
> Bridge 模式：Pod 的 IP 由 Kubernetes 集群内部的网络插件分配，用于每个容器组，IP 地址来自私有 IP 地址范围，只有当 Pod 重启时，IP 地址才会发生并发。同时无法支持直接从集群外部直接访问虚拟机。

## Masquerade + Calico + SpiderPool

SpiderPool 默认会创建一个 适用于 Calico 的 spidermultusconfigs 资源。使用 kubectl 查看：

```bash
[root@virtnest-rook-ceph-1 ~]# kubectl get spidermultusconfigs -n kube-system
NAME                      AGE
calico                    34d
```

> 低于 v0.8 版本，名字是 k8s-pod-network，这里没有本质区别。

### KubeVirt 如何使用？

三步走：

> 编辑 VirtualMachine CR 资源

- 添加 `annotations[v1.multus-cni.io/default-network] = kube-system/k8s-pod-network` 到 `spec.template.metadata.annotations`。
  
  ```yaml
  annotations:
    v1.multus-cni.io/default-network: kube-system/k8s-pod-network
  ```

- 修改 `spec.template.spec.networks`。
  
  ```yaml
  networks:
  - name: default
    pod: {}
  ```

- 修改 `spec.template.spec.domain.devices.interfaces`。

  ```yaml
  interfaces:
  - masquerade: {}
    name: default
  ```

🙋🏻问：步骤一中，为什么不是 `kube-system/calico` 而是 `kube-system/k8s-pod-network`？

要回答这个问题，我们使用 kubectl：

```bash
[root@virtnest-rook-ceph-1 ~]# k get spidermultusconfigs -n kube-system calico -o yaml
apiVersion: spiderpool.spidernet.io/v2beta1
kind: SpiderMultusConfig
metadata:
  annotations:
    multus.spidernet.io/cr-name: k8s-pod-network
  name: calico
  namespace: kube-system
spec:
  cniType: custom
  enableCoordinator: false
```

> 其中 `multus.spidernet.io/cr-name: k8s-pod-network` 是 multus CR 真实的名字，你可以理解 `name: calico` 是给用户看的。

下面是一份完整的 VirtualMachine CR 示例 YAML：

> 虚拟机用户名和密码：root / dangerous

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: masquerade-calico
spec:
  dataVolumeTemplates:
  - metadata:
      name: masquerade-calico
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: rook-ceph-block
      source:
        registry:
          url: docker://release-ci.daocloud.io/virtnest/system-images/centos-7.9-x86_64:v1
  runStrategy: Always
  template:
    metadata:
      annotations:
        v1.multus-cni.io/default-network: kube-system/k8s-pod-network
    spec:
      architecture: amd64
      domain:
        cpu:
          cores: 1
          model: host-model
          sockets: 2
          threads: 1
        devices:
          disks:
          - disk:
              bus: virtio
            name: masquerade-calico
          - disk:
              bus: virtio
            name: cloudinitdisk
          interfaces:
          - masquerade: {}
            name: default
        features:
          acpi:
            enabled: true
        machine:
          type: q35
        resources:
          requests:
            memory: 1Gi
      networks:
      - name: default
        pod: {}
      volumes:
      - dataVolume:
          name: masquerade-calico
        name: masquerade-calico
      - cloudInitNoCloud:
          userData: |
            #cloud-config
            ssh_pwauth: true
            disable_root: false
            chpasswd: {"list": "root:dangerous", expire: False}
            runcmd:
              - sed -i "/#\?PermitRootLogin/s/^.*$/PermitRootLogin yes/g" /etc/ssh/sshd_config 
        name: cloudinitdisk
```

VirtualMachine 创建完成并且启动成功后，使用如下命令查看 VirtualMachineInstance 的 IP 地址：

> 可以看到 VirtualMachineInstance 的 IP 地址就是 Pod CIDR 的地址。这里没有问题。

```bash
[root@virtnest-rook-ceph-1 ~]# k get vmi -n develop
NAME                       AGE     PHASE     IP               NODENAME               READY
masquerade-calico          19s     Running   10.233.107.215   virtnest-rook-ceph-1   True
[root@virtnest-rook-ceph-1 ~]# kubectl cluster-info dump | grep -i podcidr
                "podCIDR": "10.233.64.0/24",
                "podCIDRs": [
                "podCIDR": "10.233.65.0/24",
                "podCIDRs": [
                "podCIDR": "10.233.66.0/24",
                "podCIDRs": [
```

进入 VirtualMachine，查看虚拟机真实的 IP 地址。使用如下命令：

> 可以看到 KubeVirt 通过使用 NAT 技术将私有 IP 地址范围 10.0.2/24 网段分配给虚拟机。

```bash
[root@virtnest-rook-ceph-1 ~]# k get vmi -n develop
NAME                       AGE     PHASE     IP               NODENAME               READY
masquerade-calico          19s     Running   10.233.107.215   virtnest-rook-ceph-1   True
[root@virtnest-rook-ceph-1 ~]# ssh root@10.233.107.215
[root@masquerade-calico ~]# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1480 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 52:54:00:87:1f:11 brd ff:ff:ff:ff:ff:ff
    inet 10.0.2.2/24 brd 10.0.2.255 scope global dynamic eth0
       valid_lft 86313136sec preferred_lft 86313136sec
    inet6 fe80::5054:ff:fe87:1f11/64 scope link
       valid_lft forever preferred_lft forever
[root@masquerade-calico ~]#
```

🙋🏻问：如果我没有安装 SpiderPool，KubeVirt 如何使用 Masquerade 模式的网络？

很简单，只需要去除 “三步走” 中，第一步即可。KubeVirt 可以直接使用 Masquerade 模式。

### Passt + macvlan + SpiderPool

七步走：

- 开启 KubeVirt featureGates。使用 kubectl：

  ```bash
  kubectl edit kubevirt kubevirt -n kubevirt
  ```
  ```yaml
  configuration:
    developerConfiguration:
      featureGates:
      - Passt
  ```

- 创建 SpiderIPPool

  > 文档：<https://spidernet-io.github.io/spiderpool/v0.8/usage/install/overlay/get-started-calico-zh_cn/#spiderpool>
  
  本文集群节点网卡: ens192 所在子网为 10.7.0.0/16, 以该子网创建 SpiderIPPool：
  
  ```yaml
  apiVersion: spiderpool.spidernet.io/v2beta1
  kind: SpiderIPPool
  metadata:
    name: 10-7-v4
  spec:
    default: false
    disable: false
    # 节点网关地址
    gateway: 10.7.0.1
    ipVersion: 4
    ips:
    # 分配的 IP 范围
    - 10.7.120.100-10.7.120.200
    # 节点子网
    subnet: 10.7.120.11/16
    vlan: 0
  ```
  
  > subnet 应该与节点网卡 ens192 的子网保持一致，并且 ips 不与现有任何 IP 冲突。
  
  在节点上执行：

  ```bash
  [root@virtnest-rook-ceph-1 ~]# ip a
  2: ens192: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
      link/ether 00:50:56:b4:37:2d brd ff:ff:ff:ff:ff:ff
      inet 10.7.120.11/16 brd 10.7.255.255 scope global noprefixroute ens192
        valid_lft forever preferred_lft forever
      inet6 fe80::61d2:6f30:1e9:108/64 scope link noprefixroute
        valid_lft forever preferred_lft forever
  
  [root@virtnest-rook-ceph-1 ~]# ip route show default
  default via 10.7.0.1 dev ens192 proto static metric 100
  ```

  可得，ens192 网卡的 `gateway=10.7.0.1`, `subnet=10.7.120.11/16`

- 创建 SpiderMultusConfig
  
  > 文档：<https://spidernet-io.github.io/spiderpool/v0.8/usage/install/overlay/get-started-calico-zh_cn/#spidermultusconfig>

  ```yaml
  apiVersion: spiderpool.spidernet.io/v2beta1
  kind: SpiderMultusConfig
  metadata:
    name: macvlan-ens192
    namespace: kube-system
  spec:
    cniType: macvlan
    enableCoordinator: true
    macvlan:
      ippools:
        ipv4:
        # 选择 “步骤二” 创建的 SpiderIPPool CR
        - 10-7-v4
      master:
      # 节点的网卡名称
      - ens192
      vlanID: 0
  ```

  > `spec.macvlan.master` 设置为 ens192, ens192 必须存在于主机上。
  > `spec.macvlan.ippools.ipv4` 设置的子网和 ens192 保持一致。

  创建成功后，查看 Multus NAD 是否成功创建。使用 kubectl：
  
  ```bash
  kubectl get network-attachment-definitions.k8s.cni.cncf.io macvlan-ens192 -o yaml
  ```

- 添加 `annotations[v1.multus-cni.io/default-network] = kube-system/macvlan-ens192` 到 `spec.template.metadata.annotations`。

  ```yaml
  annotations:
    v1.multus-cni.io/default-network: kube-system/macvlan-ens192
  ```

  > 选择上文中创建的 SpiderMultusConfig CR

- （可选，指定 IP 池）添加 `ipam.spidernet.io/ippools = '[{"interface":"eth0","ipv4":["10-7-v4"]}]'` 到 `spec.template.metadata.annotations`。
  
  ```yaml
  annotations:
    ipam.spidernet.io/ippools: '[{"interface":"eth0","ipv4":["10-7-v4"]}]'
  ```

  > 选择上文中创建的 SpiderIPPool CR

- 修改 `spec.template.spec.networks`。
  
  ```yaml
  networks:
  - name: default
    pod: {}
  ```

- 修改 `spec.template.spec.domain.devices.interfaces`。

  ```yaml
  interfaces:
  - passt: {}
    name: default
  ```

下面是一份完整的 VirtualMachine CR 示例 YAML：

> 虚拟机用户名和密码：root / dangerous

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: passt-macvlan-spiderpool
spec:
  dataVolumeTemplates:
  - metadata:
      name: systemdisk-passt-macvlan-spiderpool
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: rook-ceph-block
      source:
        registry:
          url: docker://release-ci.daocloud.io/virtnest/system-images/centos-7.9-x86_64:v1
  runStrategy: Always
  template:
    metadata:
      annotations:
        v1.multus-cni.io/default-network: default/macvlan-ens192
        ipam.spidernet.io/ippools: '[{"interface":"eth0","ipv4":["10-7-v4"]}]'
    spec:
      architecture: amd64
      domain:
        cpu:
          cores: 1
          model: host-model
          sockets: 2
          threads: 1
        devices:
          disks:
          - disk:
              bus: virtio
            name: systemdisk-passt-macvlan-spiderpool
          - disk:
              bus: virtio
            name: cloudinitdisk
          interfaces:
          - passt: {}
            name: default
        features:
          acpi:
            enabled: true
        machine:
          type: q35
        resources:
          requests:
            memory: 1Gi
      networks:
      - name: default
        pod: {}
      volumes:
      - dataVolume:
          name: systemdisk-passt-macvlan-spiderpool
        name: systemdisk-passt-macvlan-spiderpool
      - cloudInitNoCloud:
          userData: |
            #cloud-config
            ssh_pwauth: true
            disable_root: false
            chpasswd: {"list": "root:dangerous", expire: False}
            runcmd:
              - sed -i "/#\?PermitRootLogin/s/^.*$/PermitRootLogin yes/g" /etc/ssh/sshd_config 
        name: cloudinitdisk
```

### Bridge + ovs + SpiderPool

待补充...

- 安装 Open vSwitch
  
  > 文档：<https://spidernet-io.github.io/spiderpool/v0.8/usage/install/underlay/get-started-ovs-zh_CN/#_1>
  > 
  > 注意，每个节点都需安装 Open vSwitch

  安装完成后，使用 `ls /opt/cni/bin`：

  ```bash
  [root@virtnest-rook-ceph-1 ~]# ls /opt/cni/bin | grep ovs
  ovs
  ```

- 创建 SpiderIPPool
- 创建 SpiderMultusConfig
- 修改 `spec.template.spec.networks`。
  
  ```yaml
  networks:
  - multus:
      default: true
      networkName: kube-system/ovs-vlan30
    name: ovs-bridge1
  - multus:
      networkName: kube-system/ovs-vlan40
    name: ovs-bridge2
  ```

- （可选，指定 IP 池）添加 `ipam.spidernet.io/ippools = '[{"interface":"eth0","ipv4":["10-7-v4"]}]'` 到 `spec.template.metadata.annotations`。
  
  > 如果有多张网卡，IP 池的顺序和 `spec.template.spec.networks` 一致。（暂定）
  
  ```yaml
  annotations:
    ipam.spidernet.io/ippools: '[{"interface":"eth0","ipv4":["vlan40-v4"]}]'
  ```

- 修改 `spec.template.spec.domain.devices.interfaces`。

  ```yaml
  interfaces:
  - bridge: {}
    name: ovs-bridge1
  - bridge: {}
    name: ovs-bridge2
  ```

- 修改 `spec.template.spec.volumes.cloudInitNoCloud.networkData`。
  
  > 注意，此配置仅适用于 CentOS

  ```yaml
  networkData: |
    version: 2
    ethernets:
      enp1s0:
        dhcp4: true
      enp2s0:
        dhcp4: true
  ```

下面是一份完整的（CentOS）VirtualMachine CR 示例 YAML：

> 虚拟机用户名和密码：root / dangerous

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: bridge-ovs-spiderpool
spec:
  dataVolumeTemplates:
  - metadata:
      name: systemdisk-bridge-ovs-spiderpool
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: rook-ceph-block
      source:
        registry:
          url: docker://release-ci.daocloud.io/virtnest/system-images/centos-7.9-x86_64:v1
  runStrategy: Always
  template:
    metadata:
      annotations:
        ipam.spidernet.io/ippools: '[{"interface":"eth0","ipv4":["vlan40-v4"]}]'
    spec:
      architecture: amd64
      domain:
        cpu:
          cores: 1
          model: host-model
          sockets: 2
          threads: 1
        devices:
          disks:
          - disk:
              bus: virtio
            name: systemdisk-bridge-ovs-spiderpool
          - disk:
              bus: virtio
            name: cloudinitdisk
          interfaces:
          - bridge: {}
            name: ovs-bridge1
          - bridge: {}
            name: ovs-bridge2
        features:
          acpi:
            enabled: true
        machine:
          type: q35
        resources:
          requests:
            memory: 1Gi
      networks:
      - multus:
          default: true
          networkName: kube-system/ovs-vlan30
        name: ovs-bridge1
      - multus:
          networkName: kube-system/ovs-vlan40
        name: ovs-bridge2
      volumes:
      - dataVolume:
          name: systemdisk-bridge-ovs-spiderpool
        name: systemdisk-bridge-ovs-spiderpool
      - cloudInitNoCloud:
          networkData: |
            version: 2
            ethernets:
              enp1s0:
                dhcp4: true
              enp2s0:
                dhcp4: true
          userData: |
            #cloud-config
            ssh_pwauth: true
            disable_root: false
            chpasswd: {"list": "root:dangerous", expire: False}
            runcmd:
              - sed -i "/#\?PermitRootLogin/s/^.*$/PermitRootLogin yes/g" /etc/ssh/sshd_config 
        name: cloudinitdisk
```
