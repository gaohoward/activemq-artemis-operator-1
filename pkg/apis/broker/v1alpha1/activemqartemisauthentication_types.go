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
	LoginModules     LoginModulesType     `json:"loginModules"`
	SecurityDomains  SecurityDomainsType  `json:"securityDomains"`
	SecuritySettings SecuritySettingsType `json:"securitySettings"`
}

type LoginModulesType struct {
	PropertiesLoginModules  []PropertiesLoginModuleType  `json:"propertiesLoginModules,omitempty"`
	GuestLoginModules       []GuestLoginModuleType       `json:"guestLoginModules,omitempty"`
	KeycloakLoginModules    []KeycloakLoginModuleType    `json:"keycloakLoginModules,omitempty"`
	CertificateLoginModules []CertificateLoginModuleType `json:"certificateLoginModule,omitempty"`
	LdapLoginModules        []LdapLoginModuleType        `json:"ldapLoginModules,omitempty"`
	Krb5LoginModules        []Krb5LoginModuleType        `json:"krb5LoginModules,omitempty"`
	GenericLoginModules     []GenericLoginModuleType     `json:"genericLoginModules,omitempty"`
}

type PropertiesLoginModuleType struct {
	Name  string     `json:"name,omitempty"`
	Users []UserType `json:"users,omitempty"`
}

type UserType struct {
	Name     string   `json:"name,omitempty"`
	Password *string  `json:"password,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

type GuestLoginModuleType struct {
	Name      string  `json:"name,omitempty"`
	GuestUser *string `json:"guestUser,omitempty"`
	GuestRole *string `json:"guestRole,omitempty"`
}

type KeycloakLoginModuleType struct {
	Name          string                          `json:"name,omitempty"`
	ModuleType    *string                         `json:"moduleType,omitempty"`
	Configuration KeycloakModuleConfigurationType `json:"configuration,omitempty"`
}

type KeycloakModuleConfigurationType struct {
	Realm                         *string            `json:"realm,omitempty"`
	RealmPublicKey                *string            `json:"realmPublicKey,omitempty"`
	AuthServerUrl                 *string            `json:"authServerUrl,omitempty"`
	SslRequired                   *string            `json:"sslRequired,omitempty"`
	Resource                      *string            `json:"resource,omitempty"`
	PublicClient                  *bool              `json:"publicClient,omitempty"`
	Credentials                   []KeyValueType     `json:"credentials,omitempty"`
	UseResourceRoleMappings       *bool              `json:"useResourceRoleMappings,omitempty"`
	EnableCors                    *bool              `json:"enableCors,omitempty"`
	CorsMaxAge                    *int64             `json:"corsMaxAge,omitempty"`
	CorsAllowedMethods            *string            `json:"corsAllowedMethods,omitempty"`
	CorsExposedHeaders            *string            `json:"corsExposedHeaders,omitempty"`
	ExposeToken                   *bool              `json:"exposeToken,omitempty"`
	BearerOnly                    *bool              `json:"bearerOnly,omitempty"`
	AutoDetectBearerOnly          *bool              `json:"autoDetectBearerOnly,omitempty"`
	ConnectionPoolSize            *int64             `json:"connectionPoolSize,omitempty"`
	AllowAnyHostName              *bool              `json:"allowAnyHostName,omitempty"`
	DisableTrustManager           *bool              `json:"disableTrustManager,omitempty"`
	TrustStore                    *string            `json:"trustStore,omitempty"`
	TrustStorePassword            *string            `json:"trustStorePassword,omitempty"`
	ClientKeyStore                *string            `json:"clientKeyStore,omitempty"`
	ClientKeyStorePassword        *string            `json:"clientKeyStorePassword,omitempty"`
	ClientKeyPassword             *string            `json:"clientKeyPassword,omitempty"`
	AlwaysRefreshToken            *bool              `json:"alwaysRefreshToken,omitempty"`
	RegisterNodeAtStartup         *bool              `json:"registerNodeAtStartup,omitempty"`
	RegisterNodePeriod            *int64             `json:"registerNodePeriod,omitempty"`
	TokenStore                    *string            `json:"tokenStore,omitempty"`
	AdapterStateCookiePath        *string            `json:"adapterStateCookiePath,omitempty"`
	PrincipalAttribute            *string            `json:"principalAttribute,omitempty"`
	ProxyUrl                      *string            `json:"proxyUrl,omitempty"`
	TurnOffChangeSessionIdOnLogin *bool              `json:"turnOffChangeSessionIdOnLogin,omitempty"`
	TokenMinimumTimeToLive        *int64             `json:"tokenMinimumTimeToLive,omitempty"`
	MinTimeBetweenJwksRequests    *int64             `json:"minTimeBetweenJwksRequests,omitempty"`
	PublicKeyCacheTtl             *int64             `json:"publicKeyCacheTtl,omitempty"`
	PolicyEnforcer                PolicyEnforcerType `json:"policyEnforcer,omitempty"`
	IgnoreOauthQueryParameter     *bool              `json:"ignoreOauthQueryParameter,omitempty"`
	VerifyTokenAudience           *bool              `json:"verifyTokenAudience,omitempty"`
	EnableBasicAuth               *bool              `json:"enableBasicAuth"`
	ConfidentialPort              *int32             `json:"confidentialPort,omitempty"`
	RedirectRewriteRules          []KeyValueType     `json:"redirectRewriteRules,omitempty"`
	EnablePkce                    *bool              `json:"enablePkce,omitempty"`
}

type KeyValueType struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type PolicyEnforcerType struct {
	EnforcementMode       *string                     `json:"policyEnforcerMode,omitempty"`
	Paths                 []PathConfigType            `json:"paths,omitempty"`
	PathCache             PathCacheType               `json:"pathCache,omitempty"`
	LazyLoadPaths         *bool                       `json:"lazyLoadPaths,omitempty"`
	OnDenyRedirectTo      *string                     `json:"onDenyRedirectTo,omitempty"`
	UserManagedAccess     *string                     `json:"userManagedAccess,omitempty"`
	ClaimInformationPoint []ClaimInformationPointType `json:"claimInformationPoint,omitempty"`
	HttpMethodAsScope     *bool                       `json:"httpMethodAsScope,omitempty"`
}

type PathConfigType struct {
	Name                  *string                     `json:"name,omitempty"`
	Type                  *string                     `json:"type,omitempty"`
	Path                  *string                     `json:"path,omitempty"`
	Methods               []MethodConfigType          `json:"methods,omitempty"`
	Scopes                []string                    `json:"scopes,omitempty"`
	Id                    *string                     `json:"id,omitempty"`
	EnforcementMode       *string                     `json:"enforcementMode,omitempty"`
	ClaimInformationPoint []ClaimInformationPointType `json:"claimInformationPoint,omitempty"`
}

type MethodConfigType struct {
	Method               *string  `json:"method,omitempty"`
	Scopes               []string `json:"scopes,omitempty"`
	ScopeEnforcementMode *string  `json:"scopeEnforcementMode,omitempty"`
}

type PathCacheType struct {
	MaxEntries *int64 `json:"maxEntries,omitempty"`
	Lifespan   *int64 `json:"lifespan,omitempty"`
}

type ClaimInformationPointType struct {
	Key    *string        `json:"key,omitempty"`
	Values []KeyValueType `json:"values,omitempty"`
}

type CertificateLoginModuleType struct {
	Name      *string        `json:"name,omitempty"`
	CertUsers []CertUserType `json:"certUsers,omitempty"`
}

type CertUserType struct {
	Name   *string  `json:"name,omitempty"`
	DnName *string  `json:"dnName,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

type LdapLoginModuleType struct {
	Name                         *string `json:"name,omitempty"`
	ConnectionUrl                *string `json:"connectionUrl,omitempty"`
	Authentication               *string `json:"authentication,omitempty"`
	ConnectionUserName           *string `json:"connectionUserName,omitempty"`
	ConnectionPassword           *string `json:"connectionPassword,omitempty"`
	SaslLoginConfigScope         *string `json:"saslLoginConfigScope,omitempty"`
	ConnectionProtocol           *string `json:"connectionProtocol,omitempty"`
	ConnectionPool               *bool   `json:"connectionPool,omitempty"`
	ConnectionTimeout            *int64  `json:"connectionTimeout,omitempty"`
	ReadTimeout                  *int64  `json:"readTimeout,omitempty"`
	UserBase                     *string `json:"UserBase,omitempty"`
	UserSearchMatching           *string `json:"userSearchMatching,omitempty"`
	UserSearchSubtree            *string `json:"userSearchSubtree,omitempty"`
	UserRoleName                 *string `json:"userRoleName,omitempty"`
	RoleBase                     *string `json:"roleBase,omitempty"`
	RoleName                     *string `json:"roleName,omitempty"`
	RoleSearchMatching           *string `json:"roleSearchMatching,omitempty"`
	RoleSearchSubtree            *string `json:"roleSearchSubtree,omitempty"`
	AuthenticateUser             *bool   `json:"authenticateUser,omitempty"`
	Referral                     *string `json:"referral,omitempty"`
	IgnorePartialResultException *bool   `json:"ignorePartialResultException,omitempty"`
	ExpandRoles                  *bool   `json:"expandRoles,omitempty"`
	ExpandRolesMatching          *string `json:"expandRolesMatching,omitempty"`
}

type Krb5LoginModuleType struct {
	Name *string `json:"name,omitempty"`
}

type GenericLoginModuleType struct {
	Name      *string        `json:"name,omitempty"`
	ClassName *string        `json:"className,omitempty"`
	Options   []KeyValueType `json:"options,omitempty"`
}

type SecurityDomainsType struct {
	BrokerDomain  BrokerDomainType `json:"brokerDomain,omitempty"`
	certDomain    CertDomainType   `json:"certDomain,omitempty"`
	consoleDomain BrokerDomainType `json:"consoleDomain,omitempty"`
}

type BrokerDomainType struct {
	Name         *string                    `json:"name,omitempty"`
	LoginModules []LoginModuleReferenceType `json:"loginModules,omitempty"`
}

type LoginModuleReferenceType struct {
	Name   *string `json:"name,omitempty"`
	Flag   *string `json:"flag,omitempty"`
	Debug  *bool   `json:"debug,omitempty"`
	Reload *bool   `json:"reload,omitempty"`
}

type CertDomainType struct {
	Name        *string                  `json:"name,omitempty"`
	LoginModule LoginModuleReferenceType `json:"loginModule"`
}

type SecuritySettingsType struct {
	Broker     []BrokerSecuritySettingType    `json:"broker,omitempty"`
	Management ManagementSecuritySettingsType `json:"management,omitempty"`
}

type BrokerSecuritySettingType struct {
	Match       string           `json:"match,omitempty"`
	Permissions []PermissionType `json:"permissions,omitempty"`
}

type PermissionType struct {
	OperationType string   `json:"operationType"`
	Roles         []string `json:"roles,omitempty"`
}

type ManagementSecuritySettingsType struct {
	Connector     ConnectorConfigType     `json:"connector,omitempty"`
	Authorisation AuthorisationConfigType `json:"authorisation,omitempty"`
}

type ConnectorConfigType struct {
	Host               *string `json:"host,omitempty"`
	Port               *int32  `json:"port,omitempty"`
	RmiRegistryPort    *int32  `json:"rmiRegistryPort,omitempty"`
	JmxRealm           *string `json:"jmxRealm,omitempty"`
	ObjectName         *string `json:"objectName,omitempty"`
	AuthenticatorType  *string `json:"authenticatorType,omitempty"`
	Secured            *bool   `json:"secured,omitempty"`
	KeyStoreProvider   *string `json:"keyStoreProvider,omitempty"`
	KeyStorePath       *string `json:"keyStorePath,omitempty"`
	KeyStorePassword   *string `json:"keyStorePassword,omitempty"`
	TrustStoreProvider *string `json:"trustStoreProvider,omitempty"`
	TrustStorePath     *string `json:"trustStorePath,omitempty"`
	TrustStorePassword *string `json:"trustStorePassword,omitempty"`
	PasswordCodec      *string `json:"passwordCodec,omitempty"`
}

type AuthorisationConfigType struct {
	Whitelist     []WhitelistEntryType `json:"whitelist,omitempty"`
	DefaultAccess []DefaultAccessType  `json:"defaultAccess,omitempty"`
	RoleAccess    []RoleAccessType     `json:"roleAccess,omitempty"`
}

type WhitelistEntryType struct {
	Domain *string `json:"domain,omitempty"`
	Key    *string `json:"key,omitempty"`
}

type DefaultAccessType struct {
	Method *string  `json:"method,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}

type RoleAccessType struct {
	Domain     *string             `json:"domain,omitempty"`
	Key        *string             `json:"key,omitempty"`
	AccessList []DefaultAccessType `json:"accessList,omitempty"`
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
