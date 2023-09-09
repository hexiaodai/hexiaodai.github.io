<!-- date: 2023-09-01 -->

# A/B 测试

除了加权路由之外，Istio 还可以配置为根据 HTTP 匹配条件将流量路由到金丝雀。在 A/B 测试场景中，您将使用 HTTP 标头或 cookie 来定位特定的用户群体。这对于需要会话关联的前端应用程序特别有用。

> 前置条件：部署 [Bookinfo Application](https://raw.githubusercontent.com/istio/istio/release-1.18/samples/bookinfo/platform/kube/bookinfo.yaml)

```bash
❯ kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.18/samples/bookinfo/platform/kube/bookinfo.yaml
❯ kubectl get pod
NAME                             READY   STATUS    RESTARTS   AGE
details-v1-5ffd6b64f7-vzdhg      2/2     Running   0          6d19h
productpage-v1-8b588bf6d-2w86d   2/2     Running   0          6d19h
ratings-v1-5f9699cfdf-7dcxv      2/2     Running   0          6d19h
reviews-v1-569db879f5-7bszt      2/2     Running   0          6d19h
reviews-v2-65c4dc6fdc-6wn9q      2/2     Running   0          6d19h
reviews-v3-c9c4fb987-nd9rr       2/2     Running   0          6d19h
```

定义 reviews 服务的网关规则

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: reviews-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - '*'
    port:
      name: http
      number: 8080
      protocol: HTTP
```

定义 reviews 服务的流量目标规则

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: reviews
spec:
  host: reviews
  subsets:
  - labels:
      version: v1
    name: v1
  - labels:
      version: v2
    name: v2
  - labels:
      version: v3
    name: v3
```

下面的配置将针对 Firefox 用户和拥有内部 cookie 的用户运行 A/B 测试。这些用户的流量会路由到金丝雀服务。

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reviews
spec:
  gateways:
  - reviews-gateway.istio-system.svc.cluster.local
  hosts:
  - '*'
  http:
  - match:
    - headers:
        user-agent:
          regex: .*Firefox.*
    - headers:
        cookie:
          regex: ^(.*?;)?(type=insider)(;.*)?$
    route:
    - destination:
        host: reviews
        subset: v1
      weight: 0
    - destination:
        host: reviews
        subset: v2
      weight: 100
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 100
```

查看遥测数据分析金丝雀期间生成流量。在金丝雀分析期间，失败检查的数量达到金丝雀分析阈值时，流量将路由回主节点，金丝雀将缩放为零。

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reviews
spec:
  gateways:
  - reviews-gateway.istio-system.svc.cluster.local
  hosts:
  - '*'
  http:
  - match:
    - headers:
        user-agent:
          regex: .*Firefox.*
    - headers:
        cookie:
          regex: ^(.*?;)?(type=insider)(;.*)?$
    route:
    - destination:
        host: reviews
        subset: v1
      weight: 100
    - destination:
        host: reviews
        subset: v2
      weight: 0
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 100
```

在金丝雀分析期间，确认符合预期时，流量将路由到金丝雀。

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
  name: reviews
spec:
  gateways:
  - reviews-gateway.istio-system.svc.cluster.local
  hosts:
  - '*'
  http:
  - match:
    - headers:
        user-agent:
          regex: .*Firefox.*
    - headers:
        cookie:
          regex: ^(.*?;)?(type=insider)(;.*)?$
    route:
    - destination:
        host: reviews
        subset: v2
      weight: 100
    - destination:
        host: reviews
        subset: v1
      weight: 0
  - route:
    - destination:
        host: reviews
        subset: v2
      weight: 100
```
