package main

import (
	"context"
	"log"

	// 使用项目内生成的代码
	clientset "crd-operator/pkg/generated/clientset/versioned"
	"crd-operator/pkg/generated/informers/externalversions"

	// informer "crd-operator/pkg/generated/clientset/versioned"
	// lister "crd-operator/pkg/generated/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalln(err)
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	list, err := clientset.CrdV1().Foos("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	for _, foo := range list.Items {
		println(foo.Name, foo.Namespace, foo.Kind)
	}

	factory := externalversions.NewSharedInformerFactory(clientset, 0)
	factory.Crd().V1().Foos().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//todo
		},
	})
}
