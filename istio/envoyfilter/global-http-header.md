# 使用 Envoyfilter 给 istio-ingresstaeway 添加全局的 HTTP Header

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: global-http-header
  namespace: istio-system
spec:
  workloadSelector:
    labels:
      istio: ingressgateway
  configPatches:
  - applyTo: ROUTE_CONFIGURATION
    match:
      # context: SIDECAR_OUTBOUND
      context: GATEWAY
    patch:
      operation: MERGE
      value:
        response_headers_to_add:
          - header:
              key: referrer-policy
              value: no-referrer
            append: false
          - header:
              key: x-frame-options
              value: deny
            append: false
          - header:
              key: strict-transport-security
              value: max-age=31536000
            append: false
          - header:
              key: x-content-type-options
              value: nosniff
            append: false
          - header:
              key: x-download-options
              value: noopen
            append: false
          - header:
              key: x-permitted-cross-domain-policies
              value: none
            append: false
          - header:
              key: x-xss-protection
              value: 1; mode=block
            append: false
          - header:
              key: x-frame-options
              value: SAMEORIGIN
            append: false
```
