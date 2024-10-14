// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MountpointClaimSpec defines the desired state of MountpointClaim
type MountpointClaimSpec struct {
	NodeID   string `json:"nodeId,omitempty"`
	PodID    string `json:"podId,omitempty"`
	VolumeID string `json:"volumeId,omitempty"`
}

// MountpointClaimStatus defines the observed state of MountpointClaim
type MountpointClaimStatus struct {
	Status          string  `json:"status,omitempty"`
	MountpointPodID *string `json:"mountpointPodId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=mpc
// +kubebuilder:subresource:status

// MountpointClaim is the Schema for the mountpointclaims API
type MountpointClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MountpointClaimSpec   `json:"spec,omitempty"`
	Status MountpointClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MountpointClaimList contains a list of MountpointClaim
type MountpointClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MountpointClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MountpointClaim{}, &MountpointClaimList{})
}
