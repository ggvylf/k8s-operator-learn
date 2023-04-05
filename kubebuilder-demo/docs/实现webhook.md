[TOC]
# 引用
https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/
https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/


# 资源的校验和准入控制
简单的资源显示可以通过注释的方式来指定Scheme
```shell
//+kubebuilder:default:enable_ingress=false
```

复杂的资源控制就需要通过准入控制和资源验证来实现
MutatingAdmissionWebhook 和ValidatingAdmissionWebhook 


## 初始化webhook
```shell
# webhook
kubebuilder create webhook --group ingress --version v1 --kind App --defaulting --programmatic-validation 
```

## webhook相关文件
api/v1beta1/app_webhook.go webhook对应的handler，我们添加业务逻辑的地方

api/v1beta1/webhook_suite_test.go 测试

config/certmanager 自动生成自签名的证书，用于webhook server提供https服务

config/webhook 用于注册webhook到k8s中

config/crd/patches 为conversion自动注入caBoundle

config/default/manager_webhook_patch.yaml 让manager的deployment支持webhook请求

config/default/webhookcainjection_patch.yaml 为webhook server注入caBoundle

注入caBoundle由cert-manager的ca-injector 组件实现


## 修改配置文件启用webhook
config/default/kustomization.yaml
```yaml
# Adds namespace to all resources.
namespace: kubebuilder-demo-system

# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: kubebuilder-demo-

# Labels to add to all resources and selectors.
#labels:
#- includeSelectors: true
#  pairs:
#    someName: someValue

resources:
- ../crd
- ../rbac
- ../manager
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in
# crd/kustomization.yaml
- ../webhook
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'. 'WEBHOOK' components are required.
- ../certmanager
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
#- ../prometheus

patchesStrategicMerge:
# Protect the /metrics endpoint by putting it behind auth.
# If you want your controller-manager to expose the /metrics
# endpoint w/o any authn/z, please comment the following line.
- manager_auth_proxy_patch.yaml



# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in
# crd/kustomization.yaml
- manager_webhook_patch.yaml

# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'.
# Uncomment 'CERTMANAGER' sections in crd/kustomization.yaml to enable the CA injection in the admission webhooks.
# 'CERTMANAGER' needs to be enabled to use ca injection
- webhookcainjection_patch.yaml

# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER' prefix.
# Uncomment the following replacements to add the cert-manager CA injection annotations
replacements:
 - source: # Add cert-manager annotation to ValidatingWebhookConfiguration, MutatingWebhookConfiguration and CRDs
     kind: Certificate
     group: cert-manager.io
     version: v1
     name: serving-cert # this name should match the one in certificate.yaml
     fieldPath: .metadata.namespace # namespace of the certificate CR
   targets:
     - select:
         kind: ValidatingWebhookConfiguration
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 0
         create: true
     - select:
         kind: MutatingWebhookConfiguration
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 0
         create: true
     - select:
         kind: CustomResourceDefinition
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 0
         create: true
 - source:
     kind: Certificate
     group: cert-manager.io
     version: v1
     name: serving-cert # this name should match the one in certificate.yaml
     fieldPath: .metadata.name
   targets:
     - select:
         kind: ValidatingWebhookConfiguration
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 1
         create: true
     - select:
         kind: MutatingWebhookConfiguration
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 1
         create: true
     - select:
         kind: CustomResourceDefinition
       fieldPaths:
         - .metadata.annotations.[cert-manager.io/inject-ca-from]
       options:
         delimiter: '/'
         index: 1
         create: true
 - source: # Add cert-manager annotation to the webhook Service
     kind: Service
     version: v1
     name: webhook-service
     fieldPath: .metadata.name # namespace of the service
   targets:
     - select:
         kind: Certificate
         group: cert-manager.io
         version: v1
       fieldPaths:
         - .spec.dnsNames.0
         - .spec.dnsNames.1
       options:
         delimiter: '.'
         index: 0
         create: true
 - source:
     kind: Service
     version: v1
     name: webhook-service
     fieldPath: .metadata.namespace # namespace of the service
   targets:
     - select:
         kind: Certificate
         group: cert-manager.io
         version: v1
       fieldPaths:
         - .spec.dnsNames.0
         - .spec.dnsNames.1
       options:
         delimiter: '.'
         index: 1
         create: true

```
config/crd/kustomization.yaml
```yaml
# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
- bases/ingress.example.com_apps.yaml
#+kubebuilder:scaffold:crdkustomizeresource

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
- patches/webhook_in_apps.yaml
#+kubebuilder:scaffold:crdkustomizewebhookpatch

# [CERTMANAGER] To enable cert-manager, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
- patches/cainjection_in_apps.yaml
#+kubebuilder:scaffold:crdkustomizecainjectionpatch

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml


```


## 配置EnableIngress字段的默认值
api/v1/app_webhook.go
```go
func (r *App) Default() {
	applog.Info("default", "name", r.Name)


	// 覆盖EnableIngress的默认值，改成相反的
	r.Spec.EnableIngress = !r.Spec.EnableIngress
}
```

## 对触发的WebHook时间进行校验
这里只对create和update做了校验，detete不做变更
```go
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateCreate() error {
	applog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	// return nil

	// 调用自己的校验逻辑
	return r.validateApp()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateUpdate(old runtime.Object) error {
	applog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	// return nil

	// 调用自己的校验逻辑
	return r.validateApp()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateDelete() error {
	applog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// 针对App资源做校验
func (r *App) validateApp() error {

    // EnableIngress为true的时候EnableService必须是true
	// EnableService是false，EnableIngress是true，报错创建失败，
	if !r.Spec.EnableService && r.Spec.EnableIngress {
		return apierrors.NewInvalid(GroupVersion.WithKind("App").GroupKind(), r.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("enable_service"),
					r.Spec.EnableService,
					"enable_service should be true when enable_ingress is true"),
			},
		)
	}
	return nil
}


```