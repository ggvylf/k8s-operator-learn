package main

import (
	"context"
	"fmt"
	"log"

	v1 "controller-tool-example/pkg/apis/crd.example.com/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalln(err)
	}

	// 这里使用RESTClient
	// 不能使用内置资源的clientset
	config.APIPath = "/apis/"
	config.NegotiatedSerializer = v1.Codecs.WithoutConversion()
	config.GroupVersion = &v1.GroupVersion

	client, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatalln(err)
	}

	// 获取cr
	foo := v1.Foo{}
	err = client.Get().Namespace("default").Resource("foos").Name("example-foo").Do(context.TODO()).Into(&foo)
	if err != nil {
		log.Fatalln(err)
	}

	// 对获取到的cr做DeepCopy，修改cr中字段的值
	newObj := foo.DeepCopy()
	newObj.Spec.Name = "test2"

	// 打印cr
	fmt.Println(foo.Spec.Name)
	fmt.Println(foo.Spec.Replicas)

	// 打印DeepCopy后的newObj
	fmt.Println(newObj.Spec.Name)

}
