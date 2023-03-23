[TOC]
# k8s的yaml

## 需要手工给deploy增加sa
kubectl create deploy ingress-manager  --dry-run=client --image=ingress-manager:v0.0.1 -o yaml

## 需要手工给sa里添加ns
kubectl create sa ingress-manager-sa --dry-run=client -o yaml
kubectl create clusterrole ingress-manager-clusterrole --dry-run=client --resource=service,ingress 
kubectl create clusterrolebinding ingress-manager-rbac  --clusterrole=ingress-manager-clusterrole --serviceaccount=default:ingress-manager-sa --dry-run=client -o yaml 


## 构建镜像
nerdctl build -t ingress-manager:v0.0.1 .



## 验证代码功能
```shell
apiVersion: v1
kind: Service
metadata:
  name: mynginx-svc
  annotations:
	ingress/http: "true"
spec:
  selector:
    app: mynginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: mynginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mynginx
  template:
    metadata:
      labels:
        app: mynginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
```


获取ingress对外端口
```shell
kubectl get svc -n ingress-nginx
NAME                                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx-controller             NodePort    10.233.56.88   <none>        80:30497/TCP,443:31996/TCP   5h55m
ingress-nginx-controller-admission   ClusterIP   10.233.30.99   <none>        443/TCP                      5h55m
```
可以看到 对外30497对应的80端口 31996对应443端口

```shell
curl -v http://10.0.56.31:30497  -H "Host: example.com"
```