package utils

import (
	"bytes"
	"html/template"
	myappv1 "kubebuilder-demo/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// 根据App的内容和模板名称，填充模板并返回
func parseTemplate(tmplName string, app *myappv1.App) []byte {
	// 文件中读取模板
	tmpl, err := template.ParseFiles("internal/controller/template/" + tmplName + ".yml")
	if err != nil {
		panic(err)
	}

	// 模板字段赋值
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()

}

// 根据模板生成Deployment
func NewDeployment(app *myappv1.App) *appsv1.Deployment {
	d := &appsv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", app), d)
	if err != nil {

		panic(err)

	}
	return d
}

// 根据模板生成Service
func NewService(app *myappv1.App) *corev1.Service {
	s := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", app), s)
	if err != nil {

		panic(err)

	}
	return s
}

// 根据模板生成Ingress
func NewIngress(app *myappv1.App) *netv1.Ingress {
	i := &netv1.Ingress{}
	err := yaml.Unmarshal(parseTemplate("service", app), i)
	if err != nil {

		panic(err)

	}
	return i
}
