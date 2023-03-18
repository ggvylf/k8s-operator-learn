package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	// 初始化clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("create restclient failed,err=", err)
	}

	// 查询pod信息
	// 需要提供ns和pod的名字

	pod, err := clientset.CoreV1().Pods("kube-system").Get(context.TODO(), "calico-node-9bc67", v1.GetOptions{})
	if err != nil {
		fmt.Println("get pod failed,err=", err)
	}
	fmt.Println(pod.Name, pod.Namespace)
}
