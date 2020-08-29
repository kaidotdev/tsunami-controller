// Package v1 contains API Schema definitions for the tsunami v1 API group
// +kubebuilder:object:generate=true
// +groupName=tsunami.kaidotorg.github.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "tsunami.kaidotorg.github.io", Version: "v1"} // nolint:gochecknoglobals

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion} // nolint:gochecknoglobals

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme // nolint:gochecknoglobals
)
