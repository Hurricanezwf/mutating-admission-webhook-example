# 简介

这是 mutating admission webhook 的例子;

# 环境准备

## kubernetes 集群准备

推荐使用 kind

step1: 在本地准备 kind 配置文件 kind-config.yaml:

```yaml
cat <<EOF >kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  apiServer:
      extraArgs:
        enable-admission-plugins: NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
EOF
```

step2: 使用 kind 创建 kubernetes

```bash
kind create cluster --name=k121 --image kindest/node:v1.21.10 --kubeconfig=$HOME/.kube/config.k121 --config=kind-config.yaml
```

## 编译 webhook 服务程序, 上传至镜像仓库.

step1: 编译

```bash
go build
```

step2: 打包上传

```
docker build -f dockerfile -t ${YOUR IMAGE}:${YOUR IMAGE TAG}
docker push ${YOUR IMAGE}:${YOUR IMAGE TAG}
```



# 快速上手

step1: 运行 sign.sh

```bash
bash ./scripts/sign.sh
```

step2: 将控制台输出的最后一行证书内容粘贴至 ./manifests/hook.yaml 的 caBundle 字段里

step3: 部署 mutating admission webhook 服务以及 hook 配置

```bash
kubectl apply -f ./manifests/deployment.yaml
kubectl apply -f ./manifests/service.yaml
kubectl apply -f ./manifests/hook.yaml
```

step4: 创建一个 deployment 测试效果

```bash
kubectl apply -f ./manifests/test.yaml
kubectl get po -o yaml

```


# 问题排查

1. 如果 pod 未启动, 则使用 `kubectl get events -w` 查看事件，根据事件提示判断问题.

2. 如果报错： `no kind "admission.k8s.io/v1" is registered for version "AdmissionReview" `, 说明集群未开启 MutatingAdmissionWebhook 的功能
