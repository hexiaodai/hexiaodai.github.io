<!-- date: 2023-09-01 -->

# 自定义 proxy 日志级别

本文介绍在 istio 中如何自定义数据面（proxy）的日志级别，方便我们排查问题时进行调试。

> 调低 proxy 日志级别进行 debug 有助于排查问题，但输出内容较多且耗资源，不建议在生产环境一直开启低级别的日志，istio 默认使用 warning 级别。

**使用 istioctl 动态调整 proxy 日志级别**

```bash
# istio-proxy
istioctl -n default proxy-config log productpage-v1-8b588bf6d-2w86d --level debug

# istio-ingressgateway
istioctl -n istio-system proxy-config log istio-ingressgateway-d94b4444b-dzpk9 --level debug
```

**通过 annotation 指定 proxy 日志级别**

```yaml
  template:
    metadata:
      annotations:
        "sidecar.istio.io/logLevel": debug # 可选: trace, debug, info, warning, error, critical, off
```

**全局配置**

> 如果是测试集群，你也可以全局配置 proxy 日志级别

**使用 kubectl 全局调整 proxy 日志级别**

```bash
kubectl -n istio-system edit configmap istio-sidecar-injector
```

修改 `values` 里面的 `global.proxy.logLevel` 字段即可。

**使用 istioctl 全局调整 proxy 日志级别**

```bash
istioctl install --set values.global.proxy.logLevel=debug
```

**配置 envoy componentLogLevel**

```yaml
  template:
    metadata:
      annotations:
        "sidecar.istio.io/componentLogLevel": "ext_authz:trace,filter:debug"
```
