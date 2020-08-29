package v1

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScanSpec struct {
	// matchLabels is a map of {key,value} pairs.
	MatchLabels          map[string]string    `json:"matchLabels"`
	Template             Template             `json:"template,omitempty"`
	ScannerContainerSpec ScannerContainerSpec `json:"scannerContainerSpec,omitempty"`
}

type Template struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +kubebuilder:pruning:PreserveUnknownFields
	metaV1.ObjectMeta `json:"metadata,omitempty"`
}

// Additional Spec for scanner container.
type ScannerContainerSpec struct {
	// Compute Resources required by this container.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	Resources v1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type ScanStatus struct{}

// +kubebuilder:object:root=true

type Scan struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScanSpec   `json:"spec,omitempty"`
	Status ScanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ScanList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []Scan `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&Scan{}, &ScanList{})
}
