/*
Copyright 2023 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha1

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeModelSpec struct {
	// Provider specifies the provider implementing this node config.
	Provider string `json:"provider" yaml:"provider"`
	// Interfaces defines the interfaces belonging to the node model
	Interfaces []NodeModelInterface `json:"interfaces" yaml:"interfaces"`
	// ParametersRef points to the vendor or implementation specific params for the
	// node.
	// +optional
	ParametersRef *corev1.ObjectReference `json:"parametersRef,omitempty" yaml:"parametersRef,omitempty"`

	// TODO
	// indicate if the router works with global vlans or not
}

type NodeModelInterface struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Speed       *string `json:"speed" yaml:"speed"`
	VLANTagging bool    `json:"vlanTagging,omitempty" yaml:"vlanTagging,omitempty"`
}

//+kubebuilder:object:root=true

// NodeModel is the Schema for the srlinux node model API.
type NodeModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NodeModelSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// NodeModelList contains a list of srlinux NodeModels.
type NodeModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeModel `json:"items"`
}

// Node type metadata.
var (
	NodeModelKind = reflect.TypeOf(NodeModel{}).Name()
)
