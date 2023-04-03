[TOC]
# 引用

# 验证controller
需要提前安装cert-manager
image使用maslennikovyv//cert-manager-webhook:v1.11.0
注意替换相关的image
```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml

```

## k8s在本地的情况
```shell
# 初始化crd
make install

# 构建镜像 
IMG=app-controller:v0.0.1 make docker-build

# 推送到仓库
IMG=app-controller:v0.0.1 make docker-push

# 部署
IMG=app-controller:v0.0.1 make deploy
```

创建cr验证
```shell
kubectl apply -f config/samples
```

## k8s不在本地的情况
如有必要制定版本，修改makefile
```shell
## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.1
CONTROLLER_TOOLS_VERSION ?= v0.11.1
```

初始化crd
```shell
make install
```

这里实际上执行的是，由于本地没有kubectl，这里把kustomize调用kubectl apply改成导出到文件，后续手工执行
```shell
/home/ggvylf/go/src/github.com/ggvylf/k8s-operator-learn/kubebuilder-demo/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/ggvylf/go/src/github.com/ggvylf/k8s-operator-learn/kubebuilder-demo/bin/kustomize build config/crd |tee  config/crd/bases/a.yaml
```

构建镜像
```shell
IMG=app-controller:v0.0.1 make docker-build
```

实际执行的是
```shell
/home/ggvylf/go/src/github.com/ggvylf/k8s-operator-learn/kubebuilder-demo/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/ggvylf/go/src/github.com/ggvylf/k8s-operator-learn/kubebuilder-demo/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
docker build -t app-controller:v0.0.1 .
```

拷贝代码到k8s环境下制作镜像
修改Dockerfile 配置镜像
```shell
FROM alpine:3.17
WORKDIR /
COPY --from=builder /workspace/manager .
ENTRYPOINT ["/manager"]
```

```shell
nerdctl --namespace k8s.io build -t app-controller:v0.0.1 . -f 
```


部署
```shell

IMG=app-controller:v0.0.1 make deploy
```
实际上执行的是
```shell

```



# 验证WebHook
## k8s在本地

修改cr并验证
```shell
# 都是false 创建失败
apiVersion: ingress.baiding.tech/v1beta1
kind: App
metadata:
  name: app-sample
spec:
  image: nginx:latest
  replicas: 3
  enable_ingress: false #会被修改为true
  enable_service: false #将会失败

# 都是true 创建成功
apiVersion: ingress.baiding.tech/v1beta1
kind: App
metadata:
  name: app-sample
spec:
  image: nginx:latest
  replicas: 3
  enable_ingress: false #会被修改为true
  enable_service: true #成功


# 都是false 创建成功
apiVersion: ingress.baiding.tech/v1beta1
kind: App
metadata:
  name: app-sample
spec:
  image: nginx:latest
  replicas: 3
  enable_ingress: true #会被修改为false
  enable_service: false #成功

```



## k8s不在本地的方法



## 本地测试的方法
修改makefile，增加dev的配置
```shell
.PHONY: dev
dev: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/dev | kubectl apply -f -
.PHONY: undev
undev: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/dev | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

```

获取cert-manager生成的证书文件
```shell
kubectl get secrets webhook-server-cert -n  kubebuilder-demo-system -o jsonpath='{..tls\.crt}' |base64 -d > certs/tls.crt
kubectl get secrets webhook-server-cert -n  kubebuilder-demo-system -o jsonpath='{..tls\.key}' |base64 -d > certs/tls.key
```

修改main.go，指定本地证书
main.go
```go
	if os.Getenv("ENVIRONMENT") == "DEV" {
		path, err := os.Getwd()
		if err != nil {
			setupLog.Error(err, "unable to get work dir")
			os.Exit(1)
		}
		options.CertDir = path + "/certs"
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

```

部署和清理
```shell
make dev

make undev

```