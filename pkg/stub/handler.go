package stub

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-sdk/minikube/pkg/apis/alexellis/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Minikube:

		err := sdk.Create(newMinikubePod(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("failed to create pod : %v", err)
			return err
		}

		// err2 := sdk.Create(newProxyDaemonset(o))
		// if err2 != nil && !errors.IsAlreadyExists(err2) {
		// 	logrus.Errorf("failed to create pod : %v", err2)
		// 	return err2
		// }
	}

	return nil
}

func newProxyDaemonset(cr *v1alpha1.Minikube) *v1beta1.DaemonSet {
	ds := v1beta1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.ObjectMeta.Name + "-proxy",
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Minikube",
				}),
			},
		},
		Spec: v1beta1.DaemonSetSpec{

			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"daemonset": cr.Name + "-daemonset"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"daemonset": cr.Name + "-daemonset"},
				},
				Spec: corev1.PodSpec{
					// NodeSelector: map[string]string{"daemon": cr.Spec.Label},
					Containers: []corev1.Container{
						{
							Name:  "proxy",
							Image: "alexellis2/squid-proxy:0.1",
							Ports: []corev1.ContainerPort{
								{
									Name:          "squid",
									ContainerPort: 3128,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	return &ds
}

func newMinikubePod(cr *v1alpha1.Minikube) *corev1.Pod {
	labels := map[string]string{
		"app": cr.ObjectMeta.Name + "-minikube",
	}
	privileged := true
	propagate := corev1.MountPropagationHostToContainer
	// user := int64(0)
	pathTypeHostDir := corev1.HostPathDirectory
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.ObjectMeta.Name + "-minikube",
			Namespace: cr.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cr, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Minikube",
				}),
			},
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			// SecurityContext: &corev1.PodSecurityContext{
			// 	RunAsUser: &user,
			// },
			Volumes: []corev1.Volume{
				{
					Name: "kvm",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/dev/kvm",
						},
					},
				},
				{
					Name: "virsh-lib",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/libvirt",
							Type: &pathTypeHostDir,
						},
					},
				},
				{
					Name: "virsh-run",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/run/libvirt",
							Type: &pathTypeHostDir,
						},
					},
				},
				{
					Name: "root-minikube",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/root/.minikube",
							Type: &pathTypeHostDir,
						},
					},
				},
				{
					Name: "root-kube",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/root/.kube",
							Type: &pathTypeHostDir,
						},
					},
				},
				{
					Name: "var-mkaas",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/mkaas",
							Type: &pathTypeHostDir,
						},
					},
				},
			},
			HostNetwork: true,
			Containers: []corev1.Container{
				{
					Name:  "libvirt",
					Image: "alexellis2/libvirt-xenial-minikube:0.3",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "CLUSTER_CPUS",
							Value: fmt.Sprintf("%d", cr.Spec.CPUCount),
						},
						{
							Name:  "CLUSTER_MEMORY",
							Value: fmt.Sprintf("%d", cr.Spec.MemoryMB),
						},
						{
							Name:  "CLUSTER_NAME",
							Value: cr.Spec.ClusterName,
						},
					},
					Ports: []corev1.ContainerPort{},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath:        "/var/run/libvirt",
							Name:             "virsh-run",
							MountPropagation: &propagate,
						},
						{
							MountPath:        "/var/lib/libvirt",
							Name:             "virsh-lib",
							MountPropagation: &propagate,
						},
						{
							MountPath:        "/root/.minikube",
							Name:             "root-minikube",
							MountPropagation: &propagate,
						},
						{
							MountPath:        "/root/.kube",
							Name:             "root-kube",
							MountPropagation: &propagate,
						},
						{
							MountPath:        "/var/mkaas",
							Name:             "var-mkaas",
							MountPropagation: &propagate,
						},
					},
				},
			},
		},
	}
}
