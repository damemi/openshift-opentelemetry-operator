package v1alpha1

import (
	operatorv1 "github.com/openshift/api/operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenTelemetry is the schema for the OpenTelemetry Operator API
// +k8s:openapi-gen=true
// +genclient
type OpenTelemetry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenTelemetrySpec   `json:"spec"`
	Status OpenTelemetryStatus `json:"status"`
}

// OpenTelemetrySpec defines the desired state of KubeDescheduler
type OpenTelemetrySpec struct {
	operatorv1.OperatorSpec `json:",inline"`

	Image  string `json:"image,omitempty"`
	Config string `json:"config"`
}

// OpenTelemetryStatus defines the observed state of KubeDescheduler
type OpenTelemetryStatus struct {
	operatorv1.OperatorStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenTelemetryList contains a list of OpenTelemetry
type OpenTelemetryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenTelemetry `json:"items"`
}