package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Iscsi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IscsiSpec   `json:"spec,omitempty"`
	Status            IscsiStatus `json:"status,omitempty"`
}

type IscsiSpec struct {
	Target     string   `json:"target"`
	Port       string   `json:"port"`
	Iqn        string   `json:"iqn"`
	Initiators []string `json:"initiators"`
}

type IscsiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Iscsi `json:"items"`
}

type IscsiStatus struct {
	Phase    StatusPhase `json:"phase,omitempty"`
	Message  string      `json:"message,omitempty"`
	Capacity `json:"capacity,omitempty"`
}
