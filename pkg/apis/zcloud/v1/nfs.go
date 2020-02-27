package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Nfs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NfsSpec   `json:"spec,omitempty"`
	Status            NfsStatus `json:"status,omitempty"`
}

type NfsSpec struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

type NfsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nfs `json:"items"`
}

type NfsStatus struct {
	Phase    StatusPhase `json:"phase,omitempty"`
	Message  string      `json:"message,omitempty"`
	Capacity `json:"capacity,omitempty"`
}
