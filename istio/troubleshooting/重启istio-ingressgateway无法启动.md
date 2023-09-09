# 重启 istio-ingressgateway 无法启动

可能是 `RequestAuthentication` 的 jwksUri 地址无法访问，导致 istiod 无法下发配置给 istio-ingressgateway

issue: <https://github.com/istio/istio/pull/39341>

## 解决方案

删除出现问题的 `RequestAuthentication` 资源，重启 istio 相关的 Pod，或者将 istio 升级至 1.15 以上版本。
