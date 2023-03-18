package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

	// 初始化restclinet
	config.GroupVersion = &metav1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"
	restclient, err := rest.RESTClientFor(config)

	if err != nil {
		fmt.Println("create restclient failed,err=", err)
	}

	// 查询pod信息
	// 需要提供ns和pod的名字
	//
	pods := corev1.PodList{}
	err = restclient.Get().
		Namespace("kube-system").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&pods)

	if err != nil {
		fmt.Println("get pod failed,err=", err)
	}
	for _, pod := range pods.Items {
		fmt.Println(pod.Name, pod.Namespace)

	}
}
