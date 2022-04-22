/*
 * Copyright (C) 2022 Rohith Jayawardene <gambol99@gmail.com>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package v1alpha1

import (
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alphav1 "github.com/appvia/terraform-controller/pkg/apis/core/v1alpha1"
)

// PolicyKind is the kind for a Policy
const PolicyKind = "Policy"

// PolicyGVK is the GVK for a Policy
var PolicyGVK = schema.GroupVersionKind{
	Group:   GroupVersion.Group,
	Version: GroupVersion.Version,
	Kind:    PolicyKind,
}

const (
	// DefaultVariablesAnnotation is the annotation applied when default variables are set
	DefaultVariablesAnnotation = "terraform.appvia.io/defaults"
	// SkipDefaultsValidationCheck is the annotation indicating to skip the check
	SkipDefaultsValidationCheck = "terraform.appvia.io/skip-defaults-check"
)

// ModuleConstraints provides a collection of constraints on modules
type ModuleConstraints struct {
	// Allowed is a list of regexes which are permitted as module sources
	// +kubebuilder:validation:Optional
	Allowed []string `json:"allowed,omitempty"`
}

// Matches returns true if the module matches the regex
func (m *ModuleConstraints) Matches(module string) (bool, error) {
	for _, m := range m.Allowed {
		re, err := regexp.Compile(m)
		if err != nil {
			return false, err
		}

		if re.MatchString(module) {
			return true, nil
		}
	}

	return false, nil
}

// Constraints defined a collection of constraints which can be applied against
// the terraform configurations
type Constraints struct {
	// Modules is a list of regexes which are permitted as module sources
	// +kubebuilder:validation:Optional
	Modules *ModuleConstraints `json:"modules,omitempty"`
}

// VerifyConstraints verifies the constraints
type VerifyConstraints struct {
	// URL is the respository url to checkout for checks
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// DefaultVariablesSelector is used to determine which configurations the variables
// should be injected into - this can take into account the namespace labels and the
// modules themselvesA
type DefaultVariablesSelector struct {
	// Namespace selectors all configurations under one or more namespaces, determined by the
	// labeling on the namespace.
	// +kubebuilder:validation:Optional
	Namespace *metav1.LabelSelector `json:"namespace,omitempty"`
	// Modules provides a collection of regexes which are used to match against the
	// configuration module
	// +kubebuilder:validation:Optional
	Modules []string `json:"modules,omitempty"`
}

// IsLabelsMatch returns if the selector matches the namespace label selector
func (d DefaultVariablesSelector) IsLabelsMatch(object client.Object) (bool, error) {
	m, err := metav1.LabelSelectorAsSelector(d.Namespace)
	if err != nil {
		return false, err
	}

	return m.Matches(labels.Set(object.GetLabels())), nil
}

// IsModulesMatch returns true of the module matches the regex
func (d DefaultVariablesSelector) IsModulesMatch(config *Configuration) (bool, error) {
	if len(d.Modules) == 0 {
		return false, nil
	}

	for _, x := range d.Modules {
		re, err := regexp.Compile(x)
		if err != nil {
			return false, err
		}

		if re.MatchString(config.Spec.Module) {
			return true, nil
		}
	}

	return false, nil
}

// DefaultVariables provides platform administrators the ability to inject
// default variables into a configuration
type DefaultVariables struct {
	// Selector is used to determine which configurations the variables should be injected into
	// +kubebuilder:validation:Required
	Selector DefaultVariablesSelector `json:"selector"`
	// Variables is a collection of variables to inject into the configuration
	// +kubebuilder:validation:Required
	// +kubebuilder:pruning:PreserveUnknownFields
	Variables runtime.RawExtension `json:"variables"`
}

// PolicySpec defines the desired state of a provider
// +k8s:openapi-gen=true
type PolicySpec struct {
	// Summary provides a short description of the policy
	// +kubebuilder:validation:Optional
	Summary string `json:"summary,omitempty"`
	// Constraints defines the constraints which can be applied to the terraform configurations
	// +kubebuilder:validation:Optional
	Constraints *Constraints `json:"constraints,omitempty"`
	// Defaults provides the default variables to inject into the configurations
	// +kubebuilder:validation:Optional
	Defaults []DefaultVariables `json:"defaults,omitempty"`
}

// +kubebuilder:webhook:name=policies.terraform.appvia.io,mutating=false,path=/validate/terraform.appvia.io/policies,verbs=delete,groups="terraform.appvia.io",resources=policies,versions=v1alpha1,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Policy is the schema for provider definitions in terraform controller
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=policies,scope=Cluster,categories={terraform}
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// PolicyStatus defines the observed state of a provider
// +k8s:openapi-gen=true
type PolicyStatus struct {
	corev1alphav1.CommonStatus `json:",inline"`
}

// GetCommonStatus returns the common status
func (p *Policy) GetCommonStatus() *corev1alphav1.CommonStatus {
	return &p.Status.CommonStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList contains a list of providers
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
