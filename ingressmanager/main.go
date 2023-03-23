package main

import (
	"log"

	"k8s-operator-learn/ingressmanager/controller"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// 获取配置文件
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln("get kubeconfig failed")
		}
		config = inClusterConfig
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("clientset init failed")
	}

	// 创建InformerFactory
	factory := informers.NewSharedInformerFactory(clientset, 0)
	serviceInformer := factory.Core().V1().Services()
	ingressInformer := factory.Networking().V1().Ingresses()

	// 创建Controller
	controller := controller.NewController(clientset, serviceInformer, ingressInformer)

	// 启动InformerFactory
	// 这里要通过先启动Informer，才能初始化controller的List和Watch
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// 启动Controller
	controller.Run(stopCh)

}
