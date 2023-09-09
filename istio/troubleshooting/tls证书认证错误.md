# SSL routines: OPENSSL_internal:CERTIFICATE_VERIFY_FAILED

> upstream connect error or disconnect/reset before headers. retried and the latest reset reason: connection failure transport failure reason: TLS error: 268435581: SSL routines: OPENSSL_internal:CERTIFICATE_VERIFY_FAILED

## 原因

可能是集群时间同步问题。

## 解决方案

检查各个节点的时间是否一致；重启 istio 和注入了 `Sidecar` 的 Pod。
