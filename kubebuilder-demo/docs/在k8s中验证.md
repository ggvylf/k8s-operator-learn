[TOC]
# 引用

# 验证controller
## 安装cert-manager
image使用maslennikovyv//cert-manager-webhook:v1.11.0

注意替换相关的image
```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml

```

## 如有必要指定bin文件的版本，修改makefile
```shell
## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.1
CONTROLLER_TOOLS_VERSION ?= v0.11.1
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
需要先生成yaml文件，最后再构建image

修改makefile，把原本的install和deploy的操作替换成导出yaml文件

```shell
mkdir -p outputyaml

vim Makefile
.PHONY: myinstall
myinstall: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd > outputyaml/makeinstall.yaml

.PHONY: mydeploy
mydeploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default > outputyaml/makedeploy.yaml

.PHONY: mydocker-build
mydocker-build: test ## Build docker image with the manager.

```

这里要替换下kube-rbac-proxy的image
config/default/manager_auth_proxy_patch.yaml
```shell
image: hubimage/kube-rbac-proxy:v0.14.0
```

生成crd和operator的yaml文件
```shell
# crd
make myinstall

# operator
IMG=app-controller:v0.0.1 make mydeploy
```

修改yaml文件
```shell
# null这里应该是null值而不是string
creationTimestamp: "null"

# 注释掉
#runAsNonRoot: true

```


构建镜像前的准备
```shell
make mydocker-build
```

修改Dockerfile 配置镜像
```shell
RUN  GOPROXY=https://goproxy.cn go mod download

FROM alpine:3.17
WORKDIR /
COPY --from=builder /workspace/manager .
ENTRYPOINT ["/manager"]
```


拷贝代码到有docker或者是有containerd的环境下
构建镜像
```shell
nerdctl -n k8s.io image ls|grep app
nerdctl -n k8s.io rmi app-controller:v0.0.1
nerdctl --namespace k8s.io build --no-cache -t app-controller:v0.0.1 . 
```


部署
```shell
# 创建crd
kubectl apply -f outputyaml/makeinstall.yaml

# 创建operator
kubectl apply -f outputyaml/makedeploy.yaml
```





# 验证WebHook
```shell
kubectl get apps.ingress.example.com -A
```


## k8s在本地
修改cr并验证
```shell
# 都是false 创建失败
apiVersion: ingress.example.com/v1
kind: App
metadata:
  labels:
    app.kubernetes.io/name: app
    app.kubernetes.io/instance: app-sample
    app.kubernetes.io/part-of: kubebuilder-demo
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebuilder-demo
  name: app-sample
spec:
  image: nginx:latest
  replicas: 1
  enable_ingress: false #会被修改为true
  enable_service: false #将会失败

# 都是true 创建成功
apiVersion: ingress.example.com/v1
kind: App
metadata:
  labels:
    app.kubernetes.io/name: app
    app.kubernetes.io/instance: app-sample
    app.kubernetes.io/part-of: kubebuilder-demo
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebuilder-demo
  name: app-sample
spec:
  image: nginx:latest
  replicas: 1
  enable_ingress: false #会被修改为true
  enable_service: true #成功


# 都是false 创建成功
apiVersion: ingress.example.com/v1
kind: App
metadata:
  labels:
    app.kubernetes.io/name: app
    app.kubernetes.io/instance: app-sample
    app.kubernetes.io/part-of: kubebuilder-demo
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubebuilder-demo
  name: app-sample
spec:
  image: nginx:latest
  replicas: 1
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