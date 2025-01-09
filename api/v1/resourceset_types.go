// Copyright 2024 Stefan Prodan.
// SPDX-License-Identifier: AGPL-3.0

package v1

import (
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluxcd/pkg/apis/meta"
)

const (
	ResourceSetKind = "ResourceSet"
)

// ResourceSetSpec defines the desired state of ResourceSet
type ResourceSetSpec struct {
	// CommonMetadata specifies the common labels and annotations that are
	// applied to all resources. Any existing label or annotation will be
	// overridden if its key matches a common one.
	// +optional
	CommonMetadata *CommonMetadata `json:"commonMetadata,omitempty"`

	// Inputs contains the list of resource group inputs.
	// +optional
	Inputs []ResourceSetInput `json:"inputs,omitempty"`

	// Resources contains the list of Kubernetes resources to reconcile.
	// +optional
	Resources []*apiextensionsv1.JSON `json:"resources,omitempty"`

	// DependsOn specifies the list of Kubernetes resources that must
	// exist on the cluster before the reconciliation process starts.
	// +optional
	DependsOn []Dependency `json:"dependsOn,omitempty"`

	// The name of the Kubernetes service account to impersonate
	// when reconciling the generated resources.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Wait instructs the controller to check the health
	// of all the reconciled resources.
	// +optional
	Wait bool `json:"wait,omitempty"`
}

// Dependency defines a ResourceSet dependency on a Kubernetes resource.
type Dependency struct {
	// APIVersion of the resource to depend on.
	// +required
	APIVersion string `json:"apiVersion"`

	// Kind of the resource to depend on.
	// +required
	Kind string `json:"kind"`

	// Name of the resource to depend on.
	// +required
	Name string `json:"name"`

	// Namespace of the resource to depend on.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Ready checks if the resource Ready status condition is true.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// ReadyExpr checks if the resource satisfies the given CEL expression.
	// The expression replaces the default readiness check and
	// is only evaluated if Ready is set to 'true'.
	// +optional
	ReadyExpr string `json:"readyExpr,omitempty"`
}

// ResourceSetInput defines the key-value pairs of the resource group input.
type ResourceSetInput map[string]*apiextensionsv1.JSON

// ResourceSetStatus defines the observed state of ResourceSet
type ResourceSetStatus struct {
	meta.ReconcileRequestStatus `json:",inline"`

	// Conditions contains the readiness conditions of the object.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Inventory contains a list of Kubernetes resource object references
	// last applied on the cluster.
	// +optional
	Inventory *ResourceInventory `json:"inventory,omitempty"`

	// LastAppliedRevision is the digest of the
	// generated resources that were last reconcile.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (in *ResourceSet) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *ResourceSet) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// IsDisabled returns true if the object has the reconcile annotation set to 'disabled'.
func (in *ResourceSet) IsDisabled() bool {
	val, ok := in.GetAnnotations()[ReconcileAnnotation]
	return ok && strings.ToLower(val) == DisabledValue
}

// GetInterval returns the interval at which the object should be reconciled.
// If no interval is set, the default is 60 minutes.
func (in *ResourceSet) GetInterval() time.Duration {
	val, ok := in.GetAnnotations()[ReconcileAnnotation]
	if ok && strings.ToLower(val) == DisabledValue {
		return 0
	}
	defaultInterval := 60 * time.Minute
	val, ok = in.GetAnnotations()[ReconcileEveryAnnotation]
	if !ok {
		return defaultInterval
	}
	interval, err := time.ParseDuration(val)
	if err != nil {
		return defaultInterval
	}
	return interval
}

// GetTimeout returns the timeout for the reconciliation process.
// If no timeout is set, the default is 5 minutes.
func (in *ResourceSet) GetTimeout() time.Duration {
	defaultTimeout := 5 * time.Minute
	val, ok := in.GetAnnotations()[ReconcileTimeoutAnnotation]
	if !ok {
		return defaultTimeout
	}
	timeout, err := time.ParseDuration(val)
	if err != nil {
		return defaultTimeout
	}
	return timeout
}

// GetInputs returns the resource group inputs.
func (in *ResourceSet) GetInputs() []ResourceSetInput {
	var inputs = make([]ResourceSetInput, len(in.Spec.Inputs))
	for i, input := range in.Spec.Inputs {
		inputs[i] = make(ResourceSetInput)
		for k, v := range input {
			inputs[i][k] = v
		}
	}
	return inputs
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rset
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// ResourceSet is the Schema for the ResourceSets API
type ResourceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceSetSpec   `json:"spec,omitempty"`
	Status ResourceSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceSetList contains a list of ResourceSet
type ResourceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceSet{}, &ResourceSetList{})
}
