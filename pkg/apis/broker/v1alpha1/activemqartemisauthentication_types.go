package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ActiveMQArtemisAuthenticationSpec defines the desired state of ActiveMQArtemisAuthentication
// +k8s:openapi-gen=true
type ActiveMQArtemisAuthenticationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Domain     DomainType     `json:"domain"`
	CertDomain CertDomainType `json:"certDomain"`
}

type DomainType struct {
	Name         *string           `json:"name,omitempty"`
	LoginModules []LoginModuleType `json:"loginModules,omitempty"`
}

type LoginModuleType struct {
	ClassName string       `json:"className,omitempty"`
	Flag      string       `json:"flag,omitempty"`
	Options   []OptionType `json:"options,omitempty"`
}

type OptionType struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type CertDomainType struct {
	Name  *string        `json:"name,omitempty"`
	Debug *bool          `json:"debug,omitempty"`
	Users []CertUserType `json:"users,omitempty"`
}

type CertUserType struct {
	Name  string   `json:"name,omitempty"`
	Dn    string   `json:"dn,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// ActiveMQArtemisAuthenticationStatus defines the observed state of ActiveMQArtemisAuthentication
// +k8s:openapi-gen=true
type ActiveMQArtemisAuthenticationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisAuthentication is the Schema for the activemqartemisauthentications API
// +k8s:openapi-gen=true
// +genclient
type ActiveMQArtemisAuthentication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveMQArtemisAuthenticationSpec   `json:"spec,omitempty"`
	Status ActiveMQArtemisAuthenticationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisAuthenticationList contains a list of ActiveMQArtemisAuthentication
type ActiveMQArtemisAuthenticationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveMQArtemisAuthentication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveMQArtemisAuthentication{}, &ActiveMQArtemisAuthenticationList{})
}
