package controller

import (
	"reflect"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// 新增svc
func (c *controller) addService(obj interface{}) {
	c.addqueue(obj)

}

// 更新svc
func (c *controller) updateService(oldobj interface{}, newobj interface{}) {

	if reflect.DeepEqual(oldobj, newobj) {
		// TODO 比较之后的操作
		return

	}
	c.addqueue(newobj)

}

// 删除ingress
func (c *controller) deleteIngress(obj interface{}) {
	// 类型断言
	ingress := obj.(*networkingv1.Ingress)

	//增加关联关系
	ownerReference := metav1.GetControllerOf(ingress)

	// 判断关联关系
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	c.queue.Add(ingress.Namespace + "/" + ingress.Name)

}

// 添加到workqueue
func (c *controller) addqueue(obj interface{}) {
	// 从obj的meta中获取namespace和name，作为workqueue的key
	// <namespace>/<name>
	// <namespace>
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)

}
