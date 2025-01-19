# Secret Syncer

## 概述
Secret Syncer 是一个用于在不同命名空间之间同步Secret的工具。

它将指定命名空间的Secret，同步到其他的命名空间中，包括TLS Secret、Opaque Secret、Docker Registry Secret等。

## 安装



## 示例

以下是一个示例 SecretSync 资源的定义，用于将名为 `qcloudregistrykey` 的 Secret 同步到 `default` 命名空间中：
```yaml
apiVersion: sync.92ac.cn/v1
kind: SecretSync
metadata:
  name: secretsync-sample
spec:
  secret_name: "qcloudregistrykey"
  namespaces: "default"
```

## 许可证
此项目根据 MIT 许可证授权。有关详细信息，请参阅 LICENSE 文件。