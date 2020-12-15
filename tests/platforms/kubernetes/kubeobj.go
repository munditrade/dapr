// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package kubernetes

import (
	"fmt"
	"os"

	v1alpha1 "github.com/dapr/dapr/pkg/apis/components/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// TestAppLabelKey is the label key for Kubernetes label selector
	TestAppLabelKey = "testapp"
	// DaprSideCarName is the Pod name of Dapr side car
	DaprSideCarName = "daprd"

	// DefaultContainerPort is the default container port exposed from test app
	DefaultContainerPort = 3000
	// DefaultExternalPort is the default external port exposed by load balancer ingress
	DefaultExternalPort = 3000

	// DaprComponentsKind is component kind
	DaprComponentsKind = "components.dapr.io"

	// DaprTestNamespaceEnvVar is the environment variable for setting the Kubernetes namespace for e2e tests
	DaprTestNamespaceEnvVar = "DAPR_TEST_NAMESPACE"

	// Environment variable for setting Kubernetes node affinity OS
	TargetOsEnvVar = "TARGET_OS"

	// Environment variable for setting Kubernetes node affinity ARCH
	TargetArchEnvVar = "TARGET_ARCH"
)

var (
	// DaprTestNamespace is the default Kubernetes namespace for e2e tests
	DaprTestNamespace = "dapr-tests"

	// TargetOs is default os affinity for Kubernetes nodes
	TargetOs = "linux"

	// TargetArch is the default architecture affinity for Kubernetes nodes
	TargetArch = "amd64"
)

// buildDeploymentObject creates the Kubernetes Deployment object for dapr test app
func buildDeploymentObject(namespace string, appDesc AppDescription) *appsv1.Deployment {
	annotationObject := map[string]string{}

	if appDesc.AppPort == 0 { // If AppPort is negative, assume this has been set explicitly
		appDesc.AppPort = DefaultContainerPort
	}

	if appDesc.DaprEnabled {
		annotationObject = map[string]string{
			"dapr.io/enabled":                           "true",
			"dapr.io/app-id":                            appDesc.AppName,
			"dapr.io/app-port":                          fmt.Sprintf("%d", appDesc.AppPort),
			"dapr.io/sidecar-cpu-limit":                 appDesc.DaprCPULimit,
			"dapr.io/sidecar-cpu-request":               appDesc.DaprCPURequest,
			"dapr.io/sidecar-memory-limit":              appDesc.DaprMemoryLimit,
			"dapr.io/sidecar-memory-request":            appDesc.DaprMemoryRequest,
			"dapr.io/sidecar-readiness-probe-threshold": "15",
			"dapr.io/sidecar-liveness-probe-threshold":  "15",
		}
	}
	if appDesc.AppProtocol != "" {
		annotationObject["dapr.io/app-protocol"] = appDesc.AppProtocol
	}
	if appDesc.AppHost != "" {
		annotationObject["dapr.io/app-host"] = appDesc.AppHost
	}
	if appDesc.MetricsPort != "" {
		annotationObject["dapr.io/metrics-port"] = appDesc.MetricsPort
	}
	if appDesc.Config != "" {
		annotationObject["dapr.io/config"] = appDesc.Config
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appDesc.AppName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(appDesc.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					TestAppLabelKey: appDesc.AppName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						TestAppLabelKey: appDesc.AppName,
					},
					Annotations: annotationObject,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            appDesc.AppName,
							Image:           fmt.Sprintf("%s/%s", appDesc.RegistryName, appDesc.ImageName),
							ImagePullPolicy: apiv1.PullAlways,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: DefaultContainerPort,
								},
							},
						},
					},
					Affinity: &apiv1.Affinity{
						NodeAffinity: &apiv1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
								NodeSelectorTerms: []apiv1.NodeSelectorTerm{
									{
										MatchExpressions: []apiv1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: "In",
												Values:   []string{TargetOs},
											},
											{
												Key:      "kubernetes.io/arch",
												Operator: "In",
												Values:   []string{TargetArch},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildServiceObject creates the Kubernetes Service Object for dapr test app
func buildServiceObject(namespace string, appDesc AppDescription) *apiv1.Service {
	serviceType := apiv1.ServiceTypeClusterIP

	if appDesc.IngressEnabled {
		serviceType = apiv1.ServiceTypeLoadBalancer
	}

	targetPort := DefaultContainerPort
	if appDesc.AppPort > 0 {
		targetPort = appDesc.AppPort
	}

	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appDesc.AppName,
			Namespace: namespace,
			Labels: map[string]string{
				TestAppLabelKey: appDesc.AppName,
			},
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				TestAppLabelKey: appDesc.AppName,
			},
			Ports: []apiv1.ServicePort{
				{
					Protocol:   apiv1.ProtocolTCP,
					Port:       DefaultExternalPort,
					TargetPort: intstr.IntOrString{IntVal: int32(targetPort)},
				},
			},
			Type: serviceType,
		},
	}
}

// buildDaprComponentObject creates dapr component object
func buildDaprComponentObject(componentName string, typeName string, metaData []v1alpha1.MetadataItem) *v1alpha1.Component {
	return &v1alpha1.Component{
		TypeMeta: metav1.TypeMeta{
			Kind: DaprComponentsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentName,
		},
		Spec: v1alpha1.ComponentSpec{
			Type:     typeName,
			Metadata: metaData,
		},
	}
}

// buildNamespaceObject creates the Kubernetes Namespace object
func buildNamespaceObject(namespace string) *apiv1.Namespace {
	return &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func init() {
	if ns, ok := os.LookupEnv(DaprTestNamespaceEnvVar); ok {
		DaprTestNamespace = ns
	}
	if os, ok := os.LookupEnv(TargetOsEnvVar); ok {
		TargetOs = os
	}
	if arch, ok := os.LookupEnv(TargetArchEnvVar); ok {
		TargetArch = arch
	}
}
