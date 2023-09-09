# Sidecar 启动顺序问题

一些服务在往 istio 上迁移过渡的过程中，有时可能会遇到 Pod 启动失败，然后一直重启，排查原因是业务启动时需要调用其它服务（比如从配置中心拉取配置），如果失败就退出，没有重试逻辑。调用失败的原因是 envoy 还没就绪（envoy也需要从控制面拉取配置，需要一点时间），导致业务发出的流量无法被处理，从而调用失败。

## 最佳实践

目前这类问题的最佳实践是让应用更加健壮一点，增加一下重试逻辑，不要一上来调用失败就立马退出，如果嫌改动麻烦，也可以在启动命令前加下 sleep，等待几秒 (可能不太优雅)。

如果不想对应用做任何改动，也可以参考下面的规避方案。

## 规避方案: 调整 sidecar 注入顺序

**修改 istio 的 configmap 全局配置:**

```bash
kubectl -n istio-system edit cm istio
```

在 `defaultConfig` 下加入 `holdApplicationUntilProxyStarts: true`

```yaml
apiVersion: v1
data:
  mesh: |-
    defaultConfig:
      holdApplicationUntilProxyStarts: true
kind: ConfigMap
```

**修改业务容器:**

为需要打开此开关的 Pod 加上 `proxy.istio.io/config` 注解，将 `holdApplicationUntilProxyStarts` 置为 `true`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      annotations:
        proxy.istio.io/config: |
          holdApplicationUntilProxyStarts: true
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
```
