/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ingressv1 "kubebuilder-demo/api/v1"
	"kubebuilder-demo/internal/controller/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ingress.example.com,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.example.com,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.example.com,resources=apps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
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

	// Deployment的处理
	// 根据App中的内容初始化Deployment，这里使用模板文件来创建资源，暂时不用构造appsv1.Deployment
	deployment := utils.NewDeployment(app)
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
