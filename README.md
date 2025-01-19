# Secret Syncer

## 概述
Secret Syncer 是一个用于在不同命名空间之间同步Secret的工具。

它将指定命名空间的Secret，同步到其他的命名空间中，包括TLS Secret、Opaque Secret、Docker Registry Secret等。


## 功能

拷贝指定的Secret到其他的命名空间中

适用场景：
- docker registry secret: 用于在不同命名空间中使用相同的docker registry secret, 用来避免拉取凭据的手动创建
- tls secret: 只需要在一个命名空间中创建tls secret，其他命名空间中的pod就可以使用这个tls secret

## 安装

```shell
git pull https://github.com/deepwzh/secret-sync-operator
cd chart/secret-sync-operator
helm install secret-sync-operator . -n secret-sync-operator --create-namespace
```

## 示例

以下是一个示例 SecretSync 资源的定义，用于将名为 `qcloudregistrykey` 的 Secret 同步到 `default` 命名空间中：
```yaml
apiVersion: sync.92ac.cn/v1
kind: SecretSync
metadata:
  name: secretsync-sample
spec:
  # 要同步的secret name
  secret_name: "qcloudregistrykey"
  # 同步的目标命名空间，*表示所有命名空间（包括后续新增的）。或者指定具体的命名空间(,分隔）
  namespaces: "*"
```

## 许可证
此项目根据 MIT 许可证授权。有关详细信息，请参阅 LICENSE 文件。