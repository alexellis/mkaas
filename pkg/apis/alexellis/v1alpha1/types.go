package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MinikubeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Minikube `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Minikube struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MinikubeSpec   `json:"spec"`
	Status            MinikubeStatus `json:"status,omitempty"`
}

type MinikubeSpec struct {
	MemoryMB    int    `json:"memoryMB"`
	CPUCount    int    `json:"cpuCount"`
	ClusterName string `json:"clusterName"`
}

type MinikubeStatus struct {
	// Fill me
}
