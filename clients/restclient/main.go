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

	// 填充config
	// 需要填充 GV APIPath NegotiatedSerializer
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"
	restclient, err := rest.RESTClientFor(config)

	if err != nil {
		fmt.Println("create restclient failed,err=", err)
	}

	// Read
	// GET /api/v1/namespaces/{namespace}/pods/{name}
	pod := &corev1.Pod{}
	err = restclient.Get().Namespace("kube-system").Resource("pods").Name("calico-node-9bc67").Do(context.TODO()).Into(pod)
	if err != nil {
		fmt.Println("get pod failed,err=", err)
	}
	fmt.Println(pod.Name, pod.Namespace)

	// List
	// GET /api/v1/namespaces/{namespace}/pods
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
