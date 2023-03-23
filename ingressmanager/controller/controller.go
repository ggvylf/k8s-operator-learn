package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	informersnetworkingv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	listernetworkingv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// worker的数量
	workerNum = 5

	// 最大重试次数
	maxRetry = 10
)

// 定义controller结构体
type controller struct {
	// k8s的client
	client kubernetes.Interface

	// service的Lister
	serviceLister listercorev1.ServiceLister

	// ingress的Lister
	ingressLister listernetworkingv1.IngressLister

	// 限速队列
	queue workqueue.RateLimitingInterface
}

// 初始化controller，接收client和相关的informer
// 初始化eventHandler
func NewController(client kubernetes.Interface, serviceInformer informerscorev1.ServiceInformer, ingressInformer informersnetworkingv1.IngressInformer) controller {
	c := controller{
		client:        client,
		serviceLister: serviceInformer.Lister(),
		ingressLister: ingressInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}

	// svc只关注添加和更新事件
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
	})

	// ingress只关注删除事件
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})

	return c
}

// 运行controller
func (c *controller) Run(stopCh chan struct{}) {
	for i := 0; i < workerNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

// controller的worker方法
func (c *controller) worker() {
	for c.processNextItem() {

	}
}

// 处理queue中的item
func (c *controller) processNextItem() bool {

	// 从队列里获取item
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	// 类型断言
	key := item.(string)

	// 处理key
	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

// 处理从queue获取的key
func (c *controller) syncService(key string) error {

	// 从key中，分离namespace和name
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// 根据namesapace和name 获取对应的svc
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// 从service.meta.Annotations中是否能获取到约定的字段，这里是ingress/http
	// metadata:
	//   annotations:
	//     ingress/http: "true"

	_, ok := service.GetAnnotations()["ingress/http"]

	// 查找是否有对应svc的ingress
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)

	// 获取ingress数据失败，并且报错内容不是 ingress不存在
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// svc有字段并且ingress不存在，就创建ingress
	if ok && apierrors.IsNotFound(err) {

		// 填充ingress对象
		ig := c.buildIngress(service)

		// 创建ingress
		_, err := c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		// svc没有字段 并且ingress存在，灸删除ingress
	} else if !ok && ingress != nil {
		// 删除ingress
		err := c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// 失败的key的处理
func (c *controller) handlerError(key string, err error) {

	// 超过最大重试次数
	if c.queue.NumRequeues(key) <= maxRetry {

		// 添加到限速队列
		c.queue.AddRateLimited(key)
		return
	}

	// 输出日志
	runtime.HandleError(err)

	// 停止重试
	c.queue.Forget(key)
}

// 根据svc的内容，填充ingress对象
func (c *controller) buildIngress(service *corev1.Service) *networkingv1.Ingress {

	// 初始化一个空的实例
	ingress := networkingv1.Ingress{}

	// 添加svc和ingress的OwnerReference 关联字段
	ingress.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(service, corev1.SchemeGroupVersion.WithKind("Service")),
	}

	// 填充ingress
	// ingress样例
	// 	apiVersion: networking.k8s.io/v1
	// kind: Ingress
	// metadata:
	//   name: ingress-myservicea
	// spec:
	//   rules:
	//   - host: myservicea.foo.org
	//     http:
	//       paths:
	//       - path: /
	//         pathType: Prefix
	//         backend:
	//           service:
	//             name: myservicea
	//             port:
	//               number: 80
	//   ingressClassName: nginx

	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := networkingv1.PathTypePrefix
	ingressclassname := "nginx"
	ingress.Spec = networkingv1.IngressSpec{
		IngressClassName: &ingressclassname,
		Rules: []networkingv1.IngressRule{
			{
				// 这里先写死 后续通过crd来获取
				Host: "example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: service.Name,
										Port: networkingv1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &ingress
}
