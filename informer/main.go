package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// 加载配置文件
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 创建informer factory
	// 这里使用带参数的，指定了ns
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace("default"))

	// 创建一个pod资源的informer
	informer := factory.Core().V1().Pods().Informer()

	// 创建event handler
	// 需要提供3个函数对应不同的时间
	// type ResourceEventHandlerFuncs struct {
	// 	AddFunc    func(obj interface{})
	// 	UpdateFunc func(oldObj, newObj interface{})
	// 	DeleteFunc func(obj interface{})
	// }

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("get ADD Event")
		},
		UpdateFunc: func(newobj, obj interface{}) {
			fmt.Println("get update Event")
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("get delete Event")
		},
	})

	// 启动informer
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)
	<-stopCh

}
