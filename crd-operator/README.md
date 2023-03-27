## 使用code-generator自动生成代码
```shell
cd crd-operator
bash /home/ggvylf/go/pkg/mod/k8s.io/code-generator@v0.26.3/generate-groups.sh all \
crd-operator/pkg/generated \
crd-operator/pkg/apis \
crd.example.com:v1  \
--go-header-file=/home/ggvylf/go/pkg/mod/k8s.io/code-generator@v0.26.3/hack/boilerplate.go.txt \
--output-base ../
```

## 创建crd和cr
```shell
kubectl get crd foos.crd.example.com
NAME                   CREATED AT
foos.crd.example.com   2023-03-25T09:59:24Z

kubectl get foo
NAME          AGE
example-foo   24s
```

## 运行程序
```shell
go run main.go

# output
example-foo default Foo

```