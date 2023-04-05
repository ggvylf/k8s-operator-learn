[TOC]
## 初始化项目
webhook根据需求生成
```shell
# 项目初始化
kubebuilder init --domain example.com --repo kubebuilder-demo --plugins=go/v4-alpha

# controller
kubebuilder create api --group ingress --version v1 --kind App

# webhook
kubebuilder create webhook --group ingress --version v1 --kind App --defaulting --programmatic-validation

```
## 增加crd需要的字段
修改app_types.go文件，
api/v1/app_types.go
```go
type AppSpec struct {
	// 镜像
	Image string `json:"image"`
	// 副本属
	Replicas int64 `json:"replicas"`

	// 是否自动创建tag
	EnableIngress bool `json:"enable_ingress,omitempty"`
	EnableService bool `json:"enable_service"`
}
```

重新生成crd
```shell
make manifests

```

可以在config目录下查看生成的crd文件和rbac文件以及webhook文件


## 自己实现各种资源的期望状态
实现Reconcile逻辑

修改controller文件
controllers/app_controller.go
```go
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// App的处理

	// 初始化空的App
	app := &ingressv1.App{}
	// 从指定ns中获取App的期望状态
	err := r.Get(ctx, req.NamespacedName, app)
	// 获取不到就返回错误
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil

	// Deployment的处理
	// 根据App中的内容初始化Deployment，这里使用模板文件来创建资源，暂时不用构造appsv1.Deployment
	deployment := utils.NewDepoyment(app)
	// 设置资源的OwnerReference
	err = controllerutil.SetControllerReference(app, deployment, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 查看同名资源是否存在，没有就create，有就update
	d := &appsv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, d)
	// 资源不存在，create
	if errors.IsNotFound(err) {
		err := r.Create(ctx, deployment)
		if err != nil {
			logger.Error(err, "create deployment failed")
			return ctrl.Result{}, err
		}
		// 资源已存在，update
	} else {
		err := r.Update(ctx, deployment)
		if err != nil {

			logger.Error(err, "update deployment failed")
			return ctrl.Result{}, err
		}
	}

	//Service的处理
	service := utils.NewService(app)
	err = controllerutil.SetControllerReference(app, service, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 查看同名资源是否存在，没有就create，有就update，同时还需要考虑flag的问题
	s := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, s)
	// 资源不存在同时EnableService为true
	if errors.IsNotFound(err) && app.Spec.EnableService {
		err := r.Create(ctx, service)
		if err != nil {
			logger.Error(err, "create service failed")
			return ctrl.Result{}, err
		}
	} else {
		// 资源已存在,同时EnableService为true 跳过
		if app.Spec.EnableService {
			logger.Info("skip update service")

			// 资源已存在,同时EnableService为false，删除资源
		} else {
			err := r.Delete(ctx, s)
			if err != nil {
				logger.Error(err, "delete service failed")
				return ctrl.Result{}, err
			}
		}

	}

	// ingress的处理
	ingress := utils.NewIngress(app)
	err = controllerutil.SetControllerReference(app, ingress, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 查看同名资源是否存在，没有就create，有就update，，同时还需要考虑flag的问题
	i := &netv1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, i)
	// 资源不存在同时EnableIngress为true
	if errors.IsNotFound(err) && app.Spec.EnableIngress {
		err := r.Create(ctx, ingress)
		if err != nil {
			logger.Error(err, "create ingress failed")
			return ctrl.Result{}, err
		}
	} else {
		// 资源已存在,同时EnableIngress为true 跳过
		if app.Spec.EnableIngress {
			logger.Info("skip update ingress")

			//资源已存在,同时EnableIngress为false 删除资源
		} else {
			err := r.Delete(ctx, i)
			if err != nil {
				logger.Error(err, "delete ingress failed")
				return ctrl.Result{}, err
			}
		}

	}

	return ctrl.Result{}, nil
}
```

把对应资源添加到Manager中
controllers/app_controller.go
```go
// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1.App{}).
		// 添加要管理的资源
		Owns(&appsv1.Deployment{}).
		Owns(&netv1.Ingress{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
```

## 内建资源使用模板生成
相关代码在controllers/utils/resource.go
模板文件在controllers/template下

## 内建资源增加rbac
controllers/app_controller.go
```go
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

```

## 构建docker镜像
```shell
nercictl build -t app-controller:v0.0.1 

```

## 部署到k8s中
```shell
IMG=xxxxx make deploy

```


## k8s中验证
cr
```yaml
apiVersion: ingress.example.com/v1
kind: App
metadata:
  name: app-sample
spec:
  image: nginx
  replicas: 1
  enable_ingress: false #默认值为false，需求为：设置为反向值;为true时，enable_service必须为true
  enable_service: false

```
