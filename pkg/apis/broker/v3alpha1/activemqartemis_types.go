package v3alpha1

import (
	"github.com/RHsyseng/operator-utils/pkg/olm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ActiveMQArtemisSpec defines the desired state of ActiveMQArtemis
// +k8s:openapi-gen=true
type ActiveMQArtemisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	AdminUser      string                  `json:"adminUser,omitempty"`
	AdminPassword  string                  `json:"adminPassword,omitempty"`
	DeploymentPlan DeploymentPlanType      `json:"deploymentPlan,omitempty"`
	Acceptors      []AcceptorType          `json:"acceptors,omitempty"`
	Connectors     []ConnectorType         `json:"connectors,omitempty"`
	Console        ConsoleType             `json:"console,omitempty"`
	Version        string                  `json:"version,omitempty"`
	Upgrades       ActiveMQArtemisUpgrades `json:"upgrades,omitempty"`
    //below are v3alpha1 types
	Name                                string                          `json:"name,omitempty"`
	SystemPropertyPrefix                string                          `json:"systemPropertyPrefix,omitempty"`
	InternalNamingPrefix                string                          `json:"internalNamingPrefix,omitempty"`
	AmqpUseCoreSubscriptionNaming       bool                            `json:"amqpUseCoreSubscriptionNaming,omitempty"`
	ResolveProtocols                    bool                            `json:"resolveProtocols,omitempty"`
	JournalDatasync                     bool                            `json:"journalDatasync,omitempty"`
	PersistenceEnabled                  bool                            `json:"persistenceEnabled,omitempty"`
	ScheduledThreadPoolMaxSize          int32                           `json:"scheduledThreadPoolMaxSize,omitempty"`
	ThreadPoolMaxSize                   int32                           `json:"threadPoolMaxSize,omitempty"`
	GracefulShutdownEnabled             bool                            `json:"gracefulShutdownEnabled,omitempty"`
	GracefulShutdownTimeout             int64                           `json:"gracefulShutdownTimeout,omitempty"`
	SecurityEnabled                     bool                            `json:"securityEnabled,omitempty"`
	SecurityInvalidationInterval        int64                           `json:"securityInvalidationInterval,omitempty"`
	JournalLockAcquisitionTimeout       int64                           `json:"journalLockAcquisitionTimeout,omitempty"`
	WildCardRoutingEnabled              bool                            `json:"wildCardRoutingEnabled,omitempty"`
	ManagementAddress                   string                          `json:"managementAddress,omitempty"`
	ManagementNotificationAddress       string                          `json:"managementNotificationAddress,omitempty"`
	ClusterUser                         string                          `json:"clusterUser,omitempty"`
	ClusterPassword                     string                          `json:"clusterPassword,omitempty"`
	PasswordCodec                       string                          `json:"passwordCodec,omitempty"`
	MaskPassword                        bool                            `json:"maskPassword,omitempty"`
	LogDelegateFactoryClassName         string                          `json:"logDelegateFactoryClassName,omitempty"`
	JmxManagementEnabled                bool                            `json:"jmxManagementEnabled,omitempty"`
	JmxDomain                           string                          `json:"jmxDomain,omitempty"`
	JmxUseBrokerName                    bool                            `json:"jmxUseBrokerName,omitempty"`
	MessageCounterEnabled               bool                            `json:"messageCounterEnabled,omitempty"`
	MessageCounterSamplePeriod          int64                           `json:"messageCounterSamplePeriod,omitempty"`
	MessageCounterMaxDayHistory         int32                           `json:"messageCounterMaxDayHistory,omitempty"`
	ConnectionTtlOverride               int64                           `json:"connectionTtlOverride,omitempty"`
	ConnectionTtlCheckInterval          int64                           `json:"connectionTtlCheckInterval,omitempty"`
	ConfigurationFileRefreshPeriod      int64                           `json:"configurationFileRefreshPeriod,omitempty"`
	AsyncConnectionExecutionEnabled     bool                            `json:"asyncConnectionExecutionEnabled,omitempty"`
	TransactionTimeout                  int64                           `json:"transactionTimeout,omitempty"`
	TransactionTimeoutScanPeriod        int64                           `json:"transactionTimeoutScanPeriod,omitempty"`
	MessageExpiryScanPeriod             int64                           `json:"messageExpiryScanPeriod,omitempty"`
	MessageExpiryThreadPriority         int32                           `json:"messageExpiryThreadPriority,omitempty"`
	AddressQueueScanPeriod              int64                           `json:"addressQueueScanPeriod,omitempty"`
	IdCacheSize                         int32                           `json:"idCacheSize,omitempty"`
	PersistIdCache                      bool                            `json:"persistIdCache,omitempty"`
	PersistDeliveryCountBeforeDelivery  bool                            `json:"persistDeliveryCountBeforeDelivery,omitempty"`
	PopulateValidatedUser               bool                            `json:"populateValidatedUser,omitempty"`
	RejectEmptyValidatedUser            bool                            `json:"rejectEmptyValidatedUser,omitempty"`
	PagingDirectory                     string                          `json:"pagingDirectory,omitempty"`
	BindingsDirectory                   string                          `json:"bindingsDirectory,omitempty"`
	CreateBindingsDir                   bool                            `json:"createBindingsDir,omitempty"`
	PageMaxConcurrentIo                 int32                           `json:"pageMaxConcurrentIo,omitempty"`
	ReadWholePage                       bool                            `json:"readWholePage,omitempty"`
	JournalDirectory                    string                          `json:"journalDirectory,omitempty"`
	NodeManagerLockDirectory            string                          `json:"nodeManagerLockDirectory,omitempty"`
	CreateJournalDir                    bool                            `json:"createJournalDir,omitempty"`
	JournalBufferTimeout                int64                           `json:"journalBufferTimeout,omitempty"`
	JournalDeviceBlockSize              int64                           `json:"journalDeviceBlockSize,omitempty"`
	JournalBufferSize                   string                          `json:"journalBufferSize,omitempty"`
	JournalSyncTransactional            bool                            `json:"journalSyncTransactional,omitempty"`
	JournalSyncNonTransactional         bool                            `json:"journalSyncNonTransactional,omitempty"`
	LogJournalWriteRate                 bool                            `json:"logJournalWriteRate,omitempty"`
	JournalFileSize                     string                          `json:"journalFileSize,omitempty"`
	JournalMinFiles                     int32                           `json:"journalMinFiles,omitempty"`
	JournalPoolFiles                    int32                           `json:"journalPoolFiles,omitempty"`
	JournalCompactPercentage            int32                           `json:"journalCompactPercentage,omitempty"`
	JournalCompactMinFiles              int32                           `json:"journalCompactMinFiles,omitempty"`
	JournalMaxIo                        int32                           `json:"journalMaxIo,omitempty"`
	JournalFileOpenTimeout              int32                           `json:"journalFileOpenTimeout,omitempty"`
	ServerDumpInterval                  int64                           `json:"serverDumpInterval,omitempty"`
	GlobalMaxSize                       string                          `json:"globalMaxSize,omitempty"`
	MaxDiskUsage                        int32                           `json:"maxDiskUsage,omitempty"`
	DiskScanPeriod                      int64                           `json:"diskScanPeriod,omitempty"`
	MemoryWarningThreshold              int32                           `json:"memoryWarningThreshold,omitempty"`
	MemoryMeasureInterval               int64                           `json:"memoryMeasureInterval,omitempty"`
	LargeMessagesDirectory              string                          `json:"largeMessagesDirectory,omitempty"`
	CriticalAnalyzer                    bool                            `json:"criticalAnalyzer,omitempty"`
	CriticalAnalyzerTimeout             int64                           `json:"criticalAnalyzerTimeout,omitempty"`
	CriticalAnalyzerCheckPeriod         int64                           `json:"criticalAnalyzerCheckPeriod,omitempty"`
	PageSyncTimeout                     int32                           `json:"pageSyncTimeout,omitempty"`
	NetworkCheckList                    string                          `json:"networkCheckList,omitempty"`
	NetworkCheckURLList                 string                          `json:"networkCheckURLList,omitempty"`
	NetworkCheckPeriod                  int64                           `json:"networkCheckPeriod,omitempty"`
	NetworkCheckTimeout                 int64                           `json:"networkCheckTimeout,omitempty"`
	NetworkCheckNIC                     string                          `json:"networkCheckNIC,omitempty"`
	NetworkCheckPingCommand             string                          `json:"networkCheckPingCommand,omitempty"`
	NetworkCheckPing6Command            string                          `json:"networkCheckPing6Command,omitempty"`
	RemotingIncomingInterceptors        []RemotingIncomingInterceptorType `json:"remotingIncomingInterceptors,omitempty"`
	RemotingOutgoingInterceptors        []RemotingOutgoingInterceptorType `json:"remotingOutgoingInterceptors,omitempty"`
	BroadcastGroups                     BroadcastGroupsType             `json:"broadcastGroups,omitempty"`
	DiscoveryGroups                     DiscoveryGroupsType             `json:"discoveryGroups,omitempty"`
	Diverts                             DivertsType                     `json:"diverts,omitempty"`
	Queues                              QueuesType                      `json:"queues,omitempty"`
	Bridges                             BridgesType                     `json:"bridges,omitempty"`
	Federations                         FederationsType                 `json:"federations,omitempty"`
	HaPolicy                            HaPolicyType                    `json:"haPolicy,omitempty"`
	ClusterConnections                  ClusterConnectionsType          `json:"clusterConnections,omitempty"`
	GroupingHandler                     GroupingHandlerType             `json:"groupingHandler,omitempty"`
	Store                               StoreType                       `json:"store,omitempty"`
	CriticalAnalyzerPolicy              string                          `json:"criticalAnalyzerPolicy,omitempty"`
	SecuritySettings                    SecuritySettingsType            `json:"securitySettings,omitempty"`
	BrokerPlugins                       BrokerPluginsType               `json:"brokerPlugins,omitempty"`
	MetricsPluginDeprecated             MetricsPluginDeprecatedType     `json:"metricsPluginDeprecated,omitempty"`
	Metrics                             MetricsType                     `json:"metrics,omitempty"`
	AddressSettings                     AddressSettingsType             `json:"addressSettings,omitempty"`
	ResourceLimitSettings               ResourceLimitSettingsType       `json:"resourceLimitSettings,omitempty"`
	ConnectorServices                   ConnectorServicesType           `json:"connectorServices,omitempty"`
	Addresses                           AddressesType                   `json:"addresses,omitempty"`
	WildcardAddresses                   WildcardAddressesType           `json:"wildcardAddresses,omitempty"`
	Jms                                 JmsType                         `json:"jms,omitempty"`
	LoggingProperties                   LoggingPropertiesType           `json:"loggingProperties,omitempty"`
	ManagementContext                   ManagementContextType           `json:"managementContext,omitempty"`
	LoginConfig                         LoginConfigType                 `json:"loginConfig,omitempty"`
	JolokiaAccess                       JolokiaAccessType               `json:"jolokiaAccess,omitempty"`
	Bootstrap                           BootstrapType                   `json:"bootstrap,omitempty"`
	ArtemisProfile                      ArtemisProfileType              `json:"artemisProfile,omitempty"`
	ArtemisRoles                        []ArtemisRoleType               `json:"artemisRoles,omitempty"`
	ArtemisUsers                        []ArtemisUserType               `json:"artemisUsers,omitempty"`
}

type DeploymentPlanType struct {
	Image              string `json:"image,omitempty"`
	Size               int32  `json:"size,omitempty"`
	RequireLogin       bool   `json:"requireLogin,omitempty"`
	PersistenceEnabled bool   `json:"persistenceEnabled,omitempty"`
	JournalType        string `json:"journalType,omitempty"`
	MessageMigration   *bool  `json:"messageMigration,omitempty"`
}

type AcceptorType struct {
	Name                string `json:"name"`
	Port                int32  `json:"port,omitempty"`
	Protocols           string `json:"protocols,omitempty"`
	SSLEnabled          bool   `json:"sslEnabled,omitempty"`
	SSLSecret           string `json:"sslSecret,omitempty"`
	EnabledCipherSuites string `json:"enabledCipherSuites,omitempty"`
	EnabledProtocols    string `json:"enabledProtocols,omitempty"`
	NeedClientAuth      bool   `json:"needClientAuth,omitempty"`
	WantClientAuth      bool   `json:"wantClientAuth,omitempty"`
	VerifyHost          bool   `json:"verifyHost,omitempty"`
	SSLProvider         string `json:"sslProvider,omitempty"`
	SNIHost             string `json:"sniHost,omitempty"`
	Expose              bool   `json:"expose,omitempty"`
	AnycastPrefix       string `json:"anycastPrefix,omitempty"`
	MulticastPrefix     string `json:"multicastPrefix,omitempty"`
	ConnectionsAllowed  int    `json:"connectionsAllowed,omitempty"`
}

type ConnectorType struct {
	Name                string `json:"name"`
	Type                string `json:"type,omitempty"`
	Host                string `json:"host"`
	Port                int32  `json:"port"`
	SSLEnabled          bool   `json:"sslEnabled,omitempty"`
	SSLSecret           string `json:"sslSecret,omitempty"`
	EnabledCipherSuites string `json:"enabledCipherSuites,omitempty"`
	EnabledProtocols    string `json:"enabledProtocols,omitempty"`
	NeedClientAuth      bool   `json:"needClientAuth,omitempty"`
	WantClientAuth      bool   `json:"wantClientAuth,omitempty"`
	VerifyHost          bool   `json:"verifyHost,omitempty"`
	SSLProvider         string `json:"sslProvider,omitempty"`
	SNIHost             string `json:"sniHost,omitempty"`
	Expose              bool   `json:"expose,omitempty"`
}

type ConsoleType struct {
	Expose        bool   `json:"expose,omitempty"`
	SSLEnabled    bool   `json:"sslEnabled,omitempty"`
	SSLSecret     string `json:"sslSecret,omitempty"`
	UseClientAuth bool   `json:"useClientAuth,omitempty"`
}

type RemotingIncomingInterceptorType struct {
	ClassName                           string                          `json:"className,omitempty"`
}

type RemotingOutgoingInterceptorType struct {
	ClassName                           string                          `json:"className,omitempty"`
}

type BroadcastGroupsType struct {
	BroadcastGroup                      []BroadcastGroupType            `json:"broadcastGroup,omitempty"`
}

type DiscoveryGroupsType struct {
	DiscoveryGroup                      []DiscoveryGroupType            `json:"discoveryGroup,omitempty"`
}

type DivertsType struct {
	Divert                              []DivertType                    `json:"divert,omitempty"`
}

type QueuesType struct {
	Queue                               []CoreQueueType                     `json:"queue,omitempty"`
}
type BridgesType struct {
	Bridge                              []BridgeType                    `json:"bridge,omitempty"`
}
type FederationsType struct {
	Federation                          []FederationType                `json:"federation,omitempty"`
}
type HaPolicyType struct {
	LiveOnly                            LiveOnlyType                    `json:"liveOnly,omitempty"`
	Replication                         ReplicationType                 `json:"replication,omitempty"`
	SharedStore                         SharedStoreType                 `json:"sharedStore,omitempty"`
}
type ClusterConnectionsType struct {
	ClusterConnectionUri                []ClusterConnectionUriType      `json:"clusterConnectionUri,omitempty"`
	ClusterConnection                   []ClusterConnectionType         `json:"clusterConnection,omitempty"`
}
type GroupingHandlerType struct {
	Type                                string                          `json:"type,omitempty"`
	Address                             string                          `json:"address,omitempty"`
	Timeout                             int32                           `json:"timeout,omitempty"`
	GroupTimeout                        int32                           `json:"groupTimeout,omitempty"`
	ReaperPeriod                        int32                           `json:"reaperPeriod,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type StoreType struct {
	FileStore                           string                          `json:"fileStore,omitempty"`
	DatabaseStore                       DatabaseStoreType               `json:"databaseStore,omitempty"`
}
type SecuritySettingsType struct {
	SecuritySetting                     []SecuritySettingType           `json:"securitySetting,omitempty"`
	SecuritySettingPlugin               SecuritySettingPluginType       `json:"securitySettingPlugin,omitempty"`
	RoleMapping                         []RoleMappingType               `json:"roleMapping,omitempty"`
}
type BrokerPluginsType struct {
	BrokerPlugin                        []BrokerPluginType              `json:"brokerPlugin,omitempty"`
}
type MetricsPluginDeprecatedType struct {
	Property                            []PropertyType                  `json:"property,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
}
type MetricsType struct {
	JvmMemory                           bool                            `json:"jvmMemory,omitempty"`
	JvmThreads                          bool                            `json:"jvmThreads,omitempty"`
	JvmGc                               bool                            `json:"jvmGc,omitempty"`
	Plugin                              PluginType                      `json:"plugin,omitempty"`
}
type AddressSettingsType struct {
	AddressSetting                      []AddressSettingType            `json:"addressSetting,omitempty"`
}
type ResourceLimitSettingsType struct {
	ResourceLimitSetting                []ResourceLimitSettingType      `json:"resourceLimitSetting,omitempty"`
}
type ConnectorServicesType struct {
	ConnectorService                    []ConnectorServiceType          `json:"connectorService,omitempty"`
}
type AddressesType struct {
	Address                             []AddresType                    `json:"address,omitempty"`
}
type WildcardAddressesType struct {
	Enabled                             bool                            `json:"enabled,omitempty"`
	RoutingEnabled                      bool                            `json:"routingEnabled,omitempty"`
	Delimiter                           string                          `json:"delimiter,omitempty"`
	AnyWords                            string                          `json:"anyWords,omitempty"`
	SingleWord                          string                          `json:"singleWord,omitempty"`
}
type JmsType struct {
	JmxDomain                           string                          `json:"jmxDomain,omitempty"`
	Queue                               []JmsQueueType                     `json:"queue,omitempty"`
	Topic                               []JmsTopicType                     `json:"topic,omitempty"`
}
type LoggingPropertiesType struct {
	Loggers                             []string                        `json:"loggers,omitempty"`
	LoggerLevel                         string                          `json:"loggerLevel,omitempty"`
	LoggerProperties                    []string                        `json:"loggerProperties,omitempty"`
	Handlers                            []HandlerType                   `json:"handlers,omitempty"`
	Formatters                          []FormatterType                 `json:"formatters,omitempty"`
}
type ManagementContextType struct {
	Connector                           JmxConnectorType                   `json:"connector,omitempty"`
	Authorisation                       AuthorisationType               `json:"authorisation,omitempty"`
}
type LoginConfigType struct {
	Entries                             []LoginEntryType                     `json:"entries,omitempty"`
}
type JolokiaAccessType struct {
	AllowOrigins                        []string                        `json:"allowOrigins,omitempty"`
	StrictChecking                      bool                            `json:"strictChecking,omitempty"`
}
type BootstrapType struct {
	JaasSecurity                        JaasSecurityType                `json:"jaasSecurity,omitempty"`
	ServerConfiguration                 string                          `json:"serverConfiguration,omitempty"`
	Web                                 WebType                         `json:"web,omitempty"`
}
type ArtemisProfileType struct {
	Home                                string                          `json:"home,omitempty"`
	Instance                            string                          `json:"instance,omitempty"`
	DataDir                             string                          `json:"dataDir,omitempty"`
	InstanceUri                         string                          `json:"instanceUri,omitempty"`
	JavaArgs                            []string                        `json:"javaArgs,omitempty"`
	JavaArgsRun                         []string                        `json:"javaArgsRun,omitempty"`
}
type ArtemisRoleType struct {
	RoleName                            string                          `json:"roleName,omitempty"`
	Users                               string                          `json:"users,omitempty"`
}
type ArtemisUserType struct {
	Name                                string                          `json:"name,omitempty"`
	Password                            string                          `json:"password,omitempty"`
}

type BroadcastGroupType struct {
	LocalBindAddress                    string                          `json:"localBindAddress,omitempty"`
	LocalBindPort                       int32                           `json:"localBindPort,omitempty"`
	GroupAddress                        string                          `json:"groupAddress,omitempty"`
	GroupPort                           int32                           `json:"groupPort,omitempty"`
	BroadcastPeriod                     int32                           `json:"broadcastPeriod,omitempty"`
	JgroupsFile                         string                          `json:"jgroupsFile,omitempty"`
	JgroupsChannel                      string                          `json:"jgroupsChannel,omitempty"`
	ConnectorRef                        []string                        `json:"connectorRef,omitempty"`
}
type DiscoveryGroupType struct {
	GroupAddress                        string                          `json:"groupAddress,omitempty"`
	GroupPort                           int32                           `json:"groupPort,omitempty"`
	JgroupsFile                         string                          `json:"jgroupsFile,omitempty"`
	JgroupsChannel                      string                          `json:"jgroupsChannel,omitempty"`
	RefreshTimeout                      int32                           `json:"refreshTimeout,omitempty"`
	LocalBindAddress                    string                          `json:"localBindAddress,omitempty"`
	LocalBindPort                       int32                           `json:"localBindPort,omitempty"`
	InitialWaitTimeout                  int32                           `json:"initialWaitTimeout,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type DivertType struct {
	TransformerClassName                string                          `json:"transformerClassName,omitempty"`
	Transformer                         TransformerType                 `json:"transformer,omitempty"`
	Exclusive                           bool                            `json:"exclusive,omitempty"`
	RoutingName                         string                          `json:"routingName,omitempty"`
	Address                             string                          `json:"address,omitempty"`
	ForwardingAddress                   string                          `json:"forwardingAddress,omitempty"`
	Filter                              FilterType                      `json:"filter,omitempty"`
	RoutingType                         string                          `json:"routingType,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type CoreQueueType struct {
	Address                             string                          `json:"address,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Filter                              FilterType                      `json:"filter,omitempty"`
	Durable                             bool                            `json:"durable,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	MaxConsumers                        int32                           `json:"maxConsumers,omitempty"`
	PurgeOnNoConsumers                  bool                            `json:"purgeOnNoConsumers,omitempty"`
	Exclusive                           bool                            `json:"exclusive,omitempty"`
	GroupRebalance                      bool                            `json:"groupRebalance,omitempty"`
	GroupBuckets                        int32                           `json:"groupBuckets,omitempty"`
	GroupFirstKey                       string                          `json:"groupFirstKey,omitempty"`
	LastValue                           bool                            `json:"lastValue,omitempty"`
	LastValueKey                        string                          `json:"lastValueKey,omitempty"`
	NonDestructive                      bool                            `json:"nonDestructive,omitempty"`
	ConsumersBeforeDispatch             int32                           `json:"consumersBeforeDispatch,omitempty"`
	DelayBeforeDispatch                 int32                           `json:"delayBeforeDispatch,omitempty"`
}
type BridgeType struct {
	QueueName                           string                          `json:"queueName,omitempty"`
	ForwardingAddress                   string                          `json:"forwardingAddress,omitempty"`
	Ha                                  bool                            `json:"ha,omitempty"`
	Filter                              FilterType                      `json:"filter,omitempty"`
	TransformerClassName                string                          `json:"transformerClassName,omitempty"`
	Transformer                         TransformerType                 `json:"transformer,omitempty"`
	MinLargeMessageSize                 string                          `json:"minLargeMessageSize,omitempty"`
	CheckPeriod                         int32                           `json:"checkPeriod,omitempty"`
	ConnectionTtl                       int32                           `json:"connectionTtl,omitempty"`
	RetryInterval                       int32                           `json:"retryInterval,omitempty"`
	RetryIntervalMultiplier             int32                           `json:"retryIntervalMultiplier,omitempty"`
	MaxRetryInterval                    int32                           `json:"maxRetryInterval,omitempty"`
	InitialConnectAttempts              int32                           `json:"initialConnectAttempts,omitempty"`
	ReconnectAttempts                   int32                           `json:"reconnectAttempts,omitempty"`
	FailoverOnServerShutdown            bool                            `json:"failoverOnServerShutdown,omitempty"`
	UseDuplicateDetection               bool                            `json:"useDuplicateDetection,omitempty"`
	ConfirmationWindowSize              string                          `json:"confirmationWindowSize,omitempty"`
	ProducerWindowSize                  string                          `json:"producerWindowSize,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Password                            string                          `json:"password,omitempty"`
	ReconnectAttemptsSameNode           int32                           `json:"reconnectAttemptsSameNode,omitempty"`
	RoutingType                         string                          `json:"routingType,omitempty"`
	StaticConnectors                    StaticConnectorsType            `json:"staticConnectors,omitempty"`
	DiscoveryGroupRef                   DiscoveryGroupRefType           `json:"discoveryGroupRef,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type FederationType struct {
	Upstream                            []UpstreamType                  `json:"upstream,omitempty"`
	Downstream                          []DownstreamType                `json:"downstream,omitempty"`
	PolicySet                           []PolicySetType                 `json:"policySet,omitempty"`
	QueuePolicy                         []QueuePolicyType               `json:"queuePolicy,omitempty"`
	AddressPolicy                       []AddressPolicyType             `json:"addressPolicy,omitempty"`
	Transformer                         []FederationTransformerType               `json:"transformer,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Password                            string                          `json:"password,omitempty"`
}
type LiveOnlyType struct {
	ScaleDown                           ScaleDownType                   `json:"scaleDown,omitempty"`
}
type ReplicationType struct {
	Master                              ReplicationMasterType                      `json:"master,omitempty"`
	Slave                               ReplicationSlaveType                       `json:"slave,omitempty"`
	Colocated                           ReplicationColocatedType                   `json:"colocated,omitempty"`
}
type SharedStoreType struct {
	Master                              SharedStoreMasterType                      `json:"master,omitempty"`
	Slave                               SharedStoreSlaveType                       `json:"slave,omitempty"`
	Colocated                           SharedStoreColocatedType                   `json:"colocated,omitempty"`
}
type ClusterConnectionUriType struct {
	Address                             string                          `json:"address,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type ClusterConnectionType struct {
	Address                             string                          `json:"address,omitempty"`
	ConnectorRef                        string                          `json:"connectorRef,omitempty"`
	CheckPeriod                         int32                           `json:"checkPeriod,omitempty"`
	ConnectionTtl                       int32                           `json:"connectionTtl,omitempty"`
	MinLargeMessageSize                 int32                           `json:"minLargeMessageSize,omitempty"`
	CallTimeout                         int32                           `json:"callTimeout,omitempty"`
	RetryInterval                       int32                           `json:"retryInterval,omitempty"`
	RetryIntervalMultiplier             int32                           `json:"retryIntervalMultiplier,omitempty"`
	MaxRetryInterval                    int32                           `json:"maxRetryInterval,omitempty"`
	InitialConnectAttempts              int32                           `json:"initialConnectAttempts,omitempty"`
	ReconnectAttempts                   int32                           `json:"reconnectAttempts,omitempty"`
	UseDuplicateDetection               bool                            `json:"useDuplicateDetection,omitempty"`
	ForwardWhenNoConsumers              bool                            `json:"forwardWhenNoConsumers,omitempty"`
	MessageLoadBalancing                string                          `json:"messageLoadBalancing,omitempty"`
	MaxHops                             int32                           `json:"maxHops,omitempty"`
	ConfirmationWindowSize              int32                           `json:"confirmationWindowSize,omitempty"`
	ProducerWindowSize                  int32                           `json:"producerWindowSize,omitempty"`
	CallFailoverTimeout                 int32                           `json:"callFailoverTimeout,omitempty"`
	NotificationInterval                int32                           `json:"notificationInterval,omitempty"`
	NotificationAttempts                int32                           `json:"notificationAttempts,omitempty"`
	ScaleDownConnector                  string                          `json:"scaleDownConnector,omitempty"`
	StaticConnectors                    CCStaticConnectorsType            `json:"staticConnectors,omitempty"`
	DiscoveryGroupRef                   DiscoveryGroupRefType           `json:"discoveryGroupRef,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	Uri                                 string                          `json:"uri,omitempty"`
}
type DatabaseStoreType struct {
	JdbcDriverClassName                 string                          `json:"jdbcDriverClassName,omitempty"`
	JdbcConnectionUrl                   string                          `json:"jdbcConnectionUrl,omitempty"`
	JdbcUser                            string                          `json:"jdbcUser,omitempty"`
	JdbcPassword                        string                          `json:"jdbcPassword,omitempty"`
	MessageTableName                    string                          `json:"messageTableName,omitempty"`
	BindingsTableName                   string                          `json:"bindingsTableName,omitempty"`
	LargeMessageTableName               string                          `json:"largeMessageTableName,omitempty"`
	PageStoreTableName                  string                          `json:"pageStoreTableName,omitempty"`
	NodeManagerStoreTableName           string                          `json:"nodeManagerStoreTableName,omitempty"`
	JdbcNetworkTimeout                  int32                           `json:"jdbcNetworkTimeout,omitempty"`
	JdbcLockRenewPeriod                 int32                           `json:"jdbcLockRenewPeriod,omitempty"`
	JdbcLockExpiration                  int32                           `json:"jdbcLockExpiration,omitempty"`
	JdbcJournalSyncPeriod               string                          `json:"jdbcJournalSyncPeriod,omitempty"`
}
type SecuritySettingType struct {
	Permission                          []PermissionType                `json:"permission,omitempty"`
	Match                               string                          `json:"match,omitempty"`
}
type SecuritySettingPluginType struct {
	Setting                             []SettingType                   `json:"setting,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
}
type RoleMappingType struct {
	From                                string                          `json:"from,omitempty"`
	To                                  string                          `json:"to,omitempty"`
}
type BrokerPluginType struct {
	Property                            []PropertyType                  `json:"property,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
}
type PropertyType struct {
	Key                                 string                          `json:"key,omitempty"`
	Value                               string                          `json:"value,omitempty"`
}
type PluginType struct {
	Property                            []PropertyType                  `json:"property,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
}
type AddressSettingType struct {
	DeadLetterAddress                   string                          `json:"deadLetterAddress,omitempty"`
	AutoCreateDeadLetterResources       bool                            `json:"autoCreateDeadLetterResources,omitempty"`
	DeadLetterQueuePrefix               string                          `json:"deadLetterQueuePrefix,omitempty"`
	DeadLetterQueueSuffix               string                          `json:"deadLetterQueueSuffix,omitempty"`
	ExpiryAddress                       string                          `json:"expiryAddress,omitempty"`
	AutoCreateExpiryResources           bool                            `json:"autoCreateExpiryResources,omitempty"`
	ExpiryQueuePrefix                   string                          `json:"expiryQueuePrefix,omitempty"`
	ExpiryQueueSuffix                   string                          `json:"expiryQueueSuffix,omitempty"`
	ExpiryDelay                         int32                           `json:"expiryDelay,omitempty"`
	MinExpiryDelay                      int32                           `json:"minExpiryDelay,omitempty"`
	MaxExpiryDelay                      int32                           `json:"maxExpiryDelay,omitempty"`
	RedeliveryDelay                     int32                           `json:"redeliveryDelay,omitempty"`
	RedeliveryDelayMultiplier           int32                           `json:"redeliveryDelayMultiplier,omitempty"`
	RedeliveryCollisionAvoidanceFactor  int32                           `json:"redeliveryCollisionAvoidanceFactor,omitempty"`
	MaxRedeliveryDelay                  int32                           `json:"maxRedeliveryDelay,omitempty"`
	MaxDeliveryAttempts                 int32                           `json:"maxDeliveryAttempts,omitempty"`
	MaxSizeBytes                        string                          `json:"maxSizeBytes,omitempty"`
	MaxSizeBytesRejectThreshold         int32                           `json:"maxSizeBytesRejectThreshold,omitempty"`
	PageSizeBytes                       string                          `json:"pageSizeBytes,omitempty"`
	PageMaxCacheSize                    int32                           `json:"pageMaxCacheSize,omitempty"`
	AddressFullPolicy                   string                          `json:"addressFullPolicy,omitempty"`
	MessageCounterHistoryDayLimit       int32                           `json:"messageCounterHistoryDayLimit,omitempty"`
	LastValueQueue                      bool                            `json:"lastValueQueue,omitempty"`
	DefaultLastValueQueue               bool                            `json:"defaultLastValueQueue,omitempty"`
	DefaultLastValueKey                 string                          `json:"defaultLastValueKey,omitempty"`
	DefaultNonDestructive               bool                            `json:"defaultNonDestructive,omitempty"`
	DefaultExclusiveQueue               bool                            `json:"defaultExclusiveQueue,omitempty"`
	DefaultGroupRebalance               bool                            `json:"defaultGroupRebalance,omitempty"`
	DefaultGroupBuckets                 int32                           `json:"defaultGroupBuckets,omitempty"`
	DefaultGroupFirstKey                string                          `json:"defaultGroupFirstKey,omitempty"`
	DefaultConsumersBeforeDispatch      int32                           `json:"defaultConsumersBeforeDispatch,omitempty"`
	DefaultDelayBeforeDispatch          int32                           `json:"defaultDelayBeforeDispatch,omitempty"`
	RedistributionDelay                 int32                           `json:"redistributionDelay,omitempty"`
	SendToDlaOnNoRoute                  bool                            `json:"sendToDlaOnNoRoute,omitempty"`
	SlowConsumerThreshold               int32                           `json:"slowConsumerThreshold,omitempty"`
	SlowConsumerPolicy                  string                          `json:"slowConsumerPolicy,omitempty"`
	SlowConsumerCheckPeriod             int32                           `json:"slowConsumerCheckPeriod,omitempty"`
	AutoCreateJmsQueues                 bool                            `json:"autoCreateJmsQueues,omitempty"`
	AutoDeleteJmsQueues                 bool                            `json:"autoDeleteJmsQueues,omitempty"`
	AutoCreateJmsTopics                 bool                            `json:"autoCreateJmsTopics,omitempty"`
	AutoDeleteJmsTopics                 bool                            `json:"autoDeleteJmsTopics,omitempty"`
	AutoCreateQueues                    bool                            `json:"autoCreateQueues,omitempty"`
	AutoDeleteQueues                    bool                            `json:"autoDeleteQueues,omitempty"`
	AutoDeleteCreatedQueues             bool                            `json:"autoDeleteCreatedQueues,omitempty"`
	AutoDeleteQueuesDelay               int32                           `json:"autoDeleteQueuesDelay,omitempty"`
	AutoDeleteQueuesMessageCount        int32                           `json:"autoDeleteQueuesMessageCount,omitempty"`
	ConfigDeleteQueues                  string                          `json:"configDeleteQueues,omitempty"`
	AutoCreateAddresses                 bool                            `json:"autoCreateAddresses,omitempty"`
	AutoDeleteAddresses                 bool                            `json:"autoDeleteAddresses,omitempty"`
	AutoDeleteAddressesDelay            int32                           `json:"autoDeleteAddressesDelay,omitempty"`
	ConfigDeleteAddresses               string                          `json:"configDeleteAddresses,omitempty"`
	ManagementBrowsePageSize            int32                           `json:"managementBrowsePageSize,omitempty"`
	DefaultPurgeOnNoConsumers           bool                            `json:"defaultPurgeOnNoConsumers,omitempty"`
	DefaultMaxConsumers                 int32                           `json:"defaultMaxConsumers,omitempty"`
	DefaultQueueRoutingType             string                          `json:"defaultQueueRoutingType,omitempty"`
	DefaultAddressRoutingType           string                          `json:"defaultAddressRoutingType,omitempty"`
	DefaultConsumerWindowSize           int32                           `json:"defaultConsumerWindowSize,omitempty"`
	DefaultRingSize                     int32                           `json:"defaultRingSize,omitempty"`
	RetroactiveMessageCount             int32                           `json:"retroactiveMessageCount,omitempty"`
	EnableMetrics                       bool                            `json:"enableMetrics,omitempty"`
	Match                               string                          `json:"match,omitempty"`
}
type ResourceLimitSettingType struct {
	MaxConnections                      int32                           `json:"maxConnections,omitempty"`
	MaxQueues                           int32                           `json:"maxQueues,omitempty"`
	Match                               string                          `json:"match,omitempty"`
}
type ConnectorServiceType struct {
	FactoryClass                        string                          `json:"factoryClass,omitempty"`
	Param                               []ParamType                     `json:"param,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type AddresType struct {
	Anycast                             AnycastType                     `json:"anycast,omitempty"`
	Multicast                           MulticastType                   `json:"multicast,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type JmsQueueType struct {
	Selector                            SelectorType                    `json:"selector,omitempty"`
	Durable                             bool                            `json:"durable,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type JmsTopicType struct {
	Name                                string                          `json:"name,omitempty"`
}
type HandlerType struct {
	Name                                string                          `json:"name,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
	Level                               string                          `json:"level,omitempty"`
	Formatter                           string                          `json:"formatter,omitempty"`
	Properties                          []PropertyType                  `json:"properties,omitempty"`
}
type FormatterType struct {
	Name                                string                          `json:"name,omitempty"`
	ClassName                           string                          `json:"className,omitempty"`
	Properties                          []PropertyType                  `json:"properties,omitempty"`
}
type JmxConnectorType struct {
	ConnectorHost                       string                          `json:"connectorHost,omitempty"`
	ConnectorPort                       int32                           `json:"connectorPort,omitempty"`
	RmiRegistryPort                     int32                           `json:"rmiRegistryPort,omitempty"`
	JmxRealm                            string                          `json:"jmxRealm,omitempty"`
	ObjectName                          string                          `json:"objectName,omitempty"`
	AuthenticatorType                   string                          `json:"authenticatorType,omitempty"`
	Secured                             bool                            `json:"secured,omitempty"`
	KeyStoreProvider                    string                          `json:"keyStoreProvider,omitempty"`
	KeyStorePath                        string                          `json:"keyStorePath,omitempty"`
	KeyStorePassword                    string                          `json:"keyStorePassword,omitempty"`
	TrustStoreProvider                  string                          `json:"trustStoreProvider,omitempty"`
	TrustStorePath                      string                          `json:"trustStorePath,omitempty"`
	TrustStorePassword                  string                          `json:"trustStorePassword,omitempty"`
	PasswordCodec                       string                          `json:"passwordCodec,omitempty"`
}
type AuthorisationType struct {
	Whitelist                           WhitelistType                   `json:"whitelist,omitempty"`
	DefaultAccess                       DefaultAccessType               `json:"defaultAccess,omitempty"`
	RoleAccess                          RoleAccessType                  `json:"roleAccess,omitempty"`
}
type LoginEntryType struct {
	Name                                string                          `json:"name,omitempty"`
	ModuleConfig                        []ModuleConfigType              `json:"moduleConfig,omitempty"`
}
type JaasSecurityType struct {
	Domain                              string                          `json:"domain,omitempty"`
	CertificateDomain                   string                          `json:"certificateDomain,omitempty"`
}
type WebType struct {
	Bind                                string                          `json:"bind,omitempty"`
	Path                                string                          `json:"path,omitempty"`
	Apps                                []AppType                       `json:"apps,omitempty"`
}

type FilterType struct {
	String                              string                          `json:"string,omitempty"`
}
type TransformerType struct {
	ClassName                           string                          `json:"className,omitempty"`
	Property                            []PropertyType                  `json:"property,omitempty"`
}
type DiscoveryGroupRefType struct {
	DiscoveryGroupName                  string                          `json:"discoveryGroupName,omitempty"`
}
type UpstreamType struct {
	Ha                                  bool                            `json:"ha,omitempty"`
	CircuitBreakerTimeout               int32                           `json:"circuitBreakerTimeout,omitempty"`
	ShareConnection                     bool                            `json:"shareConnection,omitempty"`
	ConnectionTtl                       int32                           `json:"connectionTtl,omitempty"`
	CallTimeout                         int32                           `json:"callTimeout,omitempty"`
	RetryInterval                       int32                           `json:"retryInterval,omitempty"`
	RetryIntervalMultiplier             int32                           `json:"retryIntervalMultiplier,omitempty"`
	MaxRetryInterval                    int32                           `json:"maxRetryInterval,omitempty"`
	InitialConnectAttempts              int32                           `json:"initialConnectAttempts,omitempty"`
	ReconnectAttempts                   int32                           `json:"reconnectAttempts,omitempty"`
	CheckPeriod                         int32                           `json:"checkPeriod,omitempty"`
	CallFailoverTimeout                 int32                           `json:"callFailoverTimeout,omitempty"`
	StaticConnectors                    StaticConnectorsType            `json:"staticConnectors,omitempty"`
	DiscoveryGroupRef                   DiscoveryGroupRefType           `json:"discoveryGroupRef,omitempty"`
	Policy                              []PolicyType                    `json:"policy,omitempty"`
	PriorityAdjustment                  int32                           `json:"priorityAdjustment,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Password                            string                          `json:"password,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type DownstreamType struct {
	Ha                                  bool                            `json:"ha,omitempty"`
	CircuitBreakerTimeout               int32                           `json:"circuitBreakerTimeout,omitempty"`
	ShareConnection                     bool                            `json:"shareConnection,omitempty"`
	ConnectionTtl                       int32                           `json:"connectionTtl,omitempty"`
	CallTimeout                         int32                           `json:"callTimeout,omitempty"`
	RetryInterval                       int32                           `json:"retryInterval,omitempty"`
	RetryIntervalMultiplier             int32                           `json:"retryIntervalMultiplier,omitempty"`
	MaxRetryInterval                    int32                           `json:"maxRetryInterval,omitempty"`
	InitialConnectAttempts              int32                           `json:"initialConnectAttempts,omitempty"`
	ReconnectAttempts                   int32                           `json:"reconnectAttempts,omitempty"`
	CheckPeriod                         int32                           `json:"checkPeriod,omitempty"`
	CallFailoverTimeout                 int32                           `json:"callFailoverTimeout,omitempty"`
	StaticConnectors                    StaticConnectorsType            `json:"staticConnectors,omitempty"`
	DiscoveryGroupRef                   DiscoveryGroupRefType           `json:"discoveryGroupRef,omitempty"`
	Policy                              []PolicyType                    `json:"policy,omitempty"`
	PriorityAdjustment                  int32                           `json:"priorityAdjustment,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Password                            string                          `json:"password,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	UpstreamConnectorRef                string                          `json:"upstreamConnectorRef,omitempty"`
}
type PolicySetType struct {
	Policy                              []PolicyType                    `json:"policy,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type QueuePolicyType struct {
	Include                             []ComplexIncludeType                   `json:"include,omitempty"`
	Exclude                             []ComplexExcludeType                   `json:"exclude,omitempty"`
	TransformerRef                      string                          `json:"transformerRef,omitempty"`
	PriorityAdjustment                  int32                           `json:"priorityAdjustment,omitempty"`
	IncludeFederated                    bool                            `json:"includeFederated,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type AddressPolicyType struct {
	Include                             []IncludeType                   `json:"include,omitempty"`
	Exclude                             []ExcludeType                   `json:"exclude,omitempty"`
	TransformerRef                      string                          `json:"transformerRef,omitempty"`
	AutoDelete                          bool                            `json:"autoDelete,omitempty"`
	AutoDeleteDelay                     int32                           `json:"autoDeleteDelay,omitempty"`
	AutoDeleteMessageCount              int32                           `json:"autoDeleteMessageCount,omitempty"`
	MaxHops                             int32                           `json:"maxHops,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	EnableDivertBindings                bool                            `json:"enableDivertBindings,omitempty"`
}
type FederationTransformerType struct {
	ClassName                           string                          `json:"className,omitempty"`
	Property                            []PropertyType                  `json:"property,omitempty"`
	Name                                string                          `json:"name,omitempty"`
}
type ScaleDownType struct {
	Enabled                             bool                            `json:"enabled,omitempty"`
	GroupName                           string                          `json:"groupName,omitempty"`
	DiscoveryGroupRef                   DiscoveryGroupRefType           `json:"discoveryGroupRef,omitempty"`
	Connectors                          ConnectorsType                  `json:"connectors,omitempty"`
}
type ReplicationMasterType struct {
	GroupName                           string                          `json:"groupName,omitempty"`
	ClusterName                         string                          `json:"clusterName,omitempty"`
	CheckForLiveServer                  bool                            `json:"checkForLiveServer,omitempty"`
	InitialReplicationSyncTimeout       int32                           `json:"initialReplicationSyncTimeout,omitempty"`
	VoteOnReplicationFailure            bool                            `json:"voteOnReplicationFailure,omitempty"`
	QuorumSize                          int32                           `json:"quorumSize,omitempty"`
	VoteRetries                         int32                           `json:"voteRetries,omitempty"`
	VoteRetryWait                       int32                           `json:"voteRetryWait,omitempty"`
	QuorumVoteWait                      int32                           `json:"quorumVoteWait,omitempty"`
	RetryReplicationWait                int32                           `json:"retryReplicationWait,omitempty"`
}

type ReplicationColocatedType struct {
	RequestBackup                       bool                            `json:"requestBackup,omitempty"`
	BackupRequestRetries                int32                           `json:"backupRequestRetries,omitempty"`
	BackupRequestRetryInterval          int32                           `json:"backupRequestRetryInterval,omitempty"`
	MaxBackups                          int32                           `json:"maxBackups,omitempty"`
	BackupPortOffset                    int32                           `json:"backupPortOffset,omitempty"`
	Excludes                            ExcludesType                    `json:"excludes,omitempty"`
	Master                              ReplicationMasterType                      `json:"master,omitempty"`
	Slave                               ReplicationSlaveType                       `json:"slave,omitempty"`
}
type SharedStoreMasterType struct {
	FailbackDelay                       int32                           `json:"failbackDelay,omitempty"`
	FailoverOnShutdown                  bool                            `json:"failoverOnShutdown,omitempty"`
	WaitForActivation                   bool                            `json:"waitForActivation,omitempty"`
}
type SharedStoreSlaveType struct {
	AllowFailback                       bool                            `json:"allowFailback,omitempty"`
	FailbackDelay                       int32                           `json:"failbackDelay,omitempty"`
	FailoverOnShutdown                  bool                            `json:"failoverOnShutdown,omitempty"`
	ScaleDown                           ScaleDownType                   `json:"scaleDown,omitempty"`
	RestartBackup                       bool                            `json:"restartBackup,omitempty"`
}
type SharedStoreColocatedType struct {
	RequestBackup                       bool                            `json:"requestBackup,omitempty"`
	BackupRequestRetries                int32                           `json:"backupRequestRetries,omitempty"`
	BackupRequestRetryInterval          int32                           `json:"backupRequestRetryInterval,omitempty"`
	MaxBackups                          int32                           `json:"maxBackups,omitempty"`
	BackupPortOffset                    int32                           `json:"backupPortOffset,omitempty"`
	Master                              SharedStoreMasterType                      `json:"master,omitempty"`
	Slave                               SharedStoreSlaveType                       `json:"slave,omitempty"`
}
type CCStaticConnectorsType struct {
	ConnectorRef                        []string                        `json:"connectorRef,omitempty"`
	AllowDirectConnectionsOnly          bool                            `json:"allowDirectConnectionsOnly,omitempty"`
}
type PermissionType struct {
	Type                                string                          `json:"type,omitempty"`
	Roles                               string                          `json:"roles,omitempty"`
}
type SettingType struct {
	Name                                string                          `json:"name,omitempty"`
	Value                               string                          `json:"value,omitempty"`
}
type ParamType struct {
	Key                                 string                          `json:"key,omitempty"`
	Value                               string                          `json:"value,omitempty"`
}
type AnycastType struct {
	Queue                               []QueueType                     `json:"queue,omitempty"`
}
type MulticastType struct {
	Queue                               []QueueType                     `json:"queue,omitempty"`
}
type SelectorType struct {
	String                              string                          `json:"string,omitempty"`
}
type WhitelistType struct {
	Entry                               []EntryType                     `json:"entry,omitempty"`
}
type DefaultAccessType struct {
	Access                              []AccesType                     `json:"access,omitempty"`
}
type RoleAccessType struct {
	Match                               []MatchType                     `json:"match,omitempty"`
}
type ModuleConfigType struct {
	ClassName                           string                          `json:"className,omitempty"`
	Flag                                string                          `json:"flag,omitempty"`
	Properties                          []PropertyType                  `json:"properties,omitempty"`
}
type AppType struct {
	Url                                 string                          `json:"url,omitempty"`
	War                                 string                          `json:"war,omitempty"`
}
type PolicyType struct {
	Ref                                 string                          `json:"ref,omitempty"`
}
type StaticConnectorsType struct {
	ConnectorRef                        []string                        `json:"connectorRef,omitempty"`
}
type ComplexIncludeType struct {
	QueueMatch                          string                          `json:"queueMatch,omitempty"`
	AddressMatch                        string                          `json:"addressMatch,omitempty"`
}
type ComplexExcludeType struct {
	QueueMatch                          string                          `json:"queueMatch,omitempty"`
	AddressMatch                        string                          `json:"addressMatch,omitempty"`
}
type IncludeType struct {
	AddressMatch                        string                          `json:"addressMatch,omitempty"`
}
type ExcludeType struct {
	AddressMatch                        string                          `json:"addressMatch,omitempty"`
}
type ConnectorsType struct {
	ConnectorRef                        []string                        `json:"connectorRef,omitempty"`
}
type ExcludesType struct {
	ConnectorRef                        []string                        `json:"connectorRef,omitempty"`
}
type ReplicationSlaveType struct {
	GroupName                           string                          `json:"groupName,omitempty"`
	ClusterName                         string                          `json:"clusterName,omitempty"`
	MaxSavedReplicatedJournalsSize      int32                           `json:"maxSavedReplicatedJournalsSize,omitempty"`
	ScaleDown                           ScaleDownType                   `json:"scaleDown,omitempty"`
	RestartBackup                       bool                            `json:"restartBackup,omitempty"`
	AllowFailback                       bool                            `json:"allowFailback,omitempty"`
	FailbackDelay                       int32                           `json:"failbackDelay,omitempty"`
	InitialReplicationSyncTimeout       int32                           `json:"initialReplicationSyncTimeout,omitempty"`
	VoteOnReplicationFailure            bool                            `json:"voteOnReplicationFailure,omitempty"`
	QuorumSize                          int32                           `json:"quorumSize,omitempty"`
	VoteRetries                         int32                           `json:"voteRetries,omitempty"`
	VoteRetryWait                       int32                           `json:"voteRetryWait,omitempty"`
	RetryReplicationWait                int32                           `json:"retryReplicationWait,omitempty"`
	QuorumVoteWait                      int32                           `json:"quorumVoteWait,omitempty"`
}
type QueueType struct {
	Filter                              FilterType                      `json:"filter,omitempty"`
	Durable                             bool                            `json:"durable,omitempty"`
	User                                string                          `json:"user,omitempty"`
	Name                                string                          `json:"name,omitempty"`
	MaxConsumers                        int32                           `json:"maxConsumers,omitempty"`
	PurgeOnNoConsumers                  bool                            `json:"purgeOnNoConsumers,omitempty"`
	Exclusive                           bool                            `json:"exclusive,omitempty"`
	GroupRebalance                      bool                            `json:"groupRebalance,omitempty"`
	GroupBuckets                        int32                           `json:"groupBuckets,omitempty"`
	GroupFirstKey                       string                          `json:"groupFirstKey,omitempty"`
	LastValue                           bool                            `json:"lastValue,omitempty"`
	LastValueKey                        string                          `json:"lastValueKey,omitempty"`
	NonDestructive                      bool                            `json:"nonDestructive,omitempty"`
	ConsumersBeforeDispatch             int32                           `json:"consumersBeforeDispatch,omitempty"`
	DelayBeforeDispatch                 int32                           `json:"delayBeforeDispatch,omitempty"`
	RingSize                            int32                           `json:"ringSize,omitempty"`
	Enabled                             bool                            `json:"enabled,omitempty"`
}
type EntryType struct {
	Domain                              string                          `json:"domain,omitempty"`
	Key                                 string                          `json:"key,omitempty"`
}
type AccesType struct {
	Method                              string                          `json:"method,omitempty"`
	Role                                string                          `json:"role,omitempty"`
}
type MatchType struct {
	Domain                              string                          `json:"domain,omitempty"`
	Key                                 string                          `json:"key,omitempty"`
	Access                              AccessType                      `json:"access,omitempty"`
}
type AccessType struct {
	Method                              string                          `json:"method,omitempty"`
	Roles                               string                          `json:"roles,omitempty"`
}
// ActiveMQArtemis App product upgrade flags
type ActiveMQArtemisUpgrades struct {
	Enabled bool `json:"enabled"`
	Minor   bool `json:"minor"`
}

// ActiveMQArtemisStatus defines the observed state of ActiveMQArtemis
// +k8s:openapi-gen=true
type ActiveMQArtemisStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	PodStatus olm.DeploymentStatus `json:"podStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemis is the Schema for the activemqartemis API
// +k8s:openapi-gen=true
// +genclient
type ActiveMQArtemis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveMQArtemisSpec   `json:"spec,omitempty"`
	Status ActiveMQArtemisStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ActiveMQArtemisList contains a list of ActiveMQArtemis
type ActiveMQArtemisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveMQArtemis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveMQArtemis{}, &ActiveMQArtemisList{})
}
