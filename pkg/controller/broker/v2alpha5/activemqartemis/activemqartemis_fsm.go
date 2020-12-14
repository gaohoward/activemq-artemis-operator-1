package v2alpha5activemqartemis

import (
	brokerv2alpha5 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v2alpha5"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/config"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/fsm"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Names of states
const (
	CreatingK8sResources   = "creating_k8s_resources"
	ConfiguringEnvironment = "configuring_broker_environment"
	CreatingContainer      = "creating_container"
	ContainerRunning       = "running"
	Scaling                = "scaling"
)

// IDs of states
const (
	NotCreatedID           = -1
	CreatingK8sResourcesID = 0
	//ConfiguringEnvironment = "configuring_broker_environment"
	//CreatingContainer      = "creating_container"
	ContainerRunningID          = 1
	ScalingID                   = 2
	NumActiveMQArtemisFSMStates = 3
)

// Completion of CreatingK8sResources state
const (
	None                   = 0
	CreatedHeadlessService = 1 << 0
	//CreatedPersistentVolumeClaim = 1 << 1
	CreatedStatefulSet           = 1 << 1
	CreatedConsoleJolokiaService = 1 << 2
	CreatedMuxProtocolService    = 1 << 3
	CreatedPingService           = 1 << 4
	CreatedRouteOrIngress        = 1 << 5
	CreatedCredentialsSecret     = 1 << 6
	CreatedNettySecret           = 1 << 7

	Complete = CreatedHeadlessService |
		//CreatedPersistentVolumeClaim |
		CreatedConsoleJolokiaService |
		CreatedMuxProtocolService |
		CreatedStatefulSet |
		CreatedPingService |
		CreatedRouteOrIngress |
		CreatedCredentialsSecret |
		CreatedNettySecret
)

// Machine id
const (
	ActiveMQArtemisFSMID = 0
)

type ActiveMQArtemisFSM struct {
	m                  fsm.IMachine
	namespacedName     types.NamespacedName
	customResource     *brokerv2alpha5.ActiveMQArtemis
	prevCustomResource *brokerv2alpha5.ActiveMQArtemis
	lastResourceID     int64
	r                  *ReconcileActiveMQArtemis
}

// Need to deep-copy the instance?
func MakeActiveMQArtemisFSM(instance *brokerv2alpha5.ActiveMQArtemis, _namespacedName types.NamespacedName, r *ReconcileActiveMQArtemis) ActiveMQArtemisFSM {

	var creatingK8sResourceIState fsm.IState
	var containerRunningIState fsm.IState
	var scalingIState fsm.IState

	amqbfsm := ActiveMQArtemisFSM{
		m: fsm.NewMachine(),
	}

	amqbfsm.namespacedName = _namespacedName
	amqbfsm.customResource = instance
	amqbfsm.prevCustomResource = &brokerv2alpha5.ActiveMQArtemis{}
	amqbfsm.r = r

	// TODO: Fix disconnect here between passing the parent and being added later as adding implies parenthood
	creatingK8sResourceState := MakeCreatingK8sResourcesState(&amqbfsm, _namespacedName)
	creatingK8sResourceIState = &creatingK8sResourceState
	amqbfsm.Add(&creatingK8sResourceIState)

	containerRunningState := MakeContainerRunningState(&amqbfsm, _namespacedName)
	containerRunningIState = &containerRunningState
	amqbfsm.Add(&containerRunningIState)

	scalingState := MakeScalingState(&amqbfsm, _namespacedName)
	scalingIState = &scalingState
	amqbfsm.Add(&scalingIState)

	return amqbfsm
}

func NewActiveMQArtemisFSM(instance *brokerv2alpha5.ActiveMQArtemis, _namespacedName types.NamespacedName, r *ReconcileActiveMQArtemis) *ActiveMQArtemisFSM {

	amqbfsm := MakeActiveMQArtemisFSM(instance, _namespacedName, r)

	return &amqbfsm
}
func (amqbfsm *ActiveMQArtemisFSM) Add(s *fsm.IState) {

	amqbfsm.m.Add(s)
}

func (amqbfsm *ActiveMQArtemisFSM) Remove(s *fsm.IState) {

	amqbfsm.m.Remove(s)
}

func ID() int {

	return ActiveMQArtemisFSMID
}

func (amqbfsm *ActiveMQArtemisFSM) Enter(startStateID int) error {

	var err error = nil

	// For the moment sequentially set stuff up
	// k8s resource creation and broker environment configuration can probably be done concurrently later
	amqbfsm.r.result = reconcile.Result{}
	if err = amqbfsm.m.Enter(CreatingK8sResourcesID); nil != err {
		err, _ = amqbfsm.m.Update()
	}

	return err
}

func (amqbfsm *ActiveMQArtemisFSM) UpdateCustomResource(newRc *brokerv2alpha5.ActiveMQArtemis) {
	log.Info("Updating current res in fsm", "current", *amqbfsm.customResource, "new coming", *newRc)
	*amqbfsm.customResource = *newRc
	if amqbfsm.prevCustomResource.Name == "" {
		//first time
		*amqbfsm.prevCustomResource = *newRc
		log.Info("Update prev initial: ", "prev", *amqbfsm.prevCustomResource)
	}
}

func (amqbfsm *ActiveMQArtemisFSM) Update() (error, int) {

	// Was the current state complete?
	amqbfsm.r.result = reconcile.Result{}
	err, nextStateID := amqbfsm.m.Update()
	ssNamespacedName := types.NamespacedName{Name: statefulsets.NameBuilder.Name(), Namespace: amqbfsm.customResource.Namespace}
	UpdatePodStatus(amqbfsm.customResource, amqbfsm.r.client, ssNamespacedName)

	return err, nextStateID
}

func (amqbfsm *ActiveMQArtemisFSM) Exit() error {

	amqbfsm.r.result = reconcile.Result{}
	err := amqbfsm.m.Exit()

	return err
}

func (amqbfsm *ActiveMQArtemisFSM) ProcessCustomResourceForUpdate() bool {
	resID := amqbfsm.customResource.GetObjectMeta().GetGeneration()
	log.Info("processing resource for update", "current id", amqbfsm.lastResourceID, " new res id ", resID)
	if amqbfsm.lastResourceID == resID {
		log.Info("already processed res ", "id", resID)
		return false
	}

	result := amqbfsm.ProcessLogging() || amqbfsm.ProcessAddressSettings()

	if result {
		log.Info("There are new logging config need merge")
		amqbfsm.MergeCustomResource()
		log.Info("Merged", "result", amqbfsm.prevCustomResource)
	}
	amqbfsm.lastResourceID = amqbfsm.customResource.GetObjectMeta().GetGeneration()
	log.Info("Setting last resource id", "id", amqbfsm.lastResourceID)

	return result
}

func (amqbfsm *ActiveMQArtemisFSM) ProcessAddressSettings() bool {
	log.Info("Process addresssettings")

	if len(amqbfsm.customResource.Spec.AddressSettings.AddressSetting) == 0 {
		return false
	}

	//we need to compare old with new and update if they are different.
	return amqbfsm.compareAddressSettings(&amqbfsm.prevCustomResource.Spec.AddressSettings, &amqbfsm.customResource.Spec.AddressSettings)
}

func (amqbfsm *ActiveMQArtemisFSM) ProcessLogging() bool {
	log.Info("Process Logging configuration")

	if len(amqbfsm.customResource.Spec.Logging.Logger) == 0 && len(amqbfsm.customResource.Spec.Logging.Handler) == 0 && len(amqbfsm.customResource.Spec.Logging.Formatter) == 0 {
		log.Info("No new broker specific config, no need to update")
		return false
	}

	return amqbfsm.compareLogging(&amqbfsm.prevCustomResource.Spec.Logging, &amqbfsm.customResource.Spec.Logging)
}

func (amqbfsm *ActiveMQArtemisFSM) MergeCustomResource() {
	amqbfsm.mergeLogging(&amqbfsm.prevCustomResource.Spec.Logging, &amqbfsm.customResource.Spec.Logging)
	amqbfsm.mergeAddressSettings(&amqbfsm.prevCustomResource.Spec.AddressSettings, &amqbfsm.customResource.Spec.AddressSettings)
}

//unlike logging we just replace the whole settings with new one
func (amqbfsm *ActiveMQArtemisFSM) mergeAddressSettings(existingSettings *brokerv2alpha5.AddressSettingsType, newSettings *brokerv2alpha5.AddressSettingsType) {
	*existingSettings = *newSettings
}

func (amqbfsm *ActiveMQArtemisFSM) mergeLogging(existingLogging *brokerv2alpha5.LoggingType, newLogging *brokerv2alpha5.LoggingType) {
	log.Info("Merge logging", "old", existingLogging, "new", newLogging)
	loggerMap := make(map[string]brokerv2alpha5.LoggerType)
	for _, logger := range existingLogging.Logger {
		loggerMap[logger.Name] = logger
	}

	mergeLoggers := make([]brokerv2alpha5.LoggerType, 0)
	//remember updated loggers
	updated := make([]string, 0)
	for _, nlogger := range newLogging.Logger {
		if oldlogger, ok := loggerMap[nlogger.Name]; ok {
			if nlogger.Level != nil {
				if oldlogger.Level == nil {
					oldlogger.Level = new(string)
				}
				*oldlogger.Level = *nlogger.Level
			}
			if nlogger.Handlers != nil {
				if oldlogger.Handlers == nil {
					oldlogger.Handlers = new(string)
				}
				*oldlogger.Handlers = *nlogger.Handlers
			}
			if nlogger.UseParentHandlers != nil {
				if oldlogger.UseParentHandlers == nil {
					oldlogger.UseParentHandlers = new(bool)
				}
				*oldlogger.UseParentHandlers = *nlogger.UseParentHandlers
			}
			mergeLoggers = append(mergeLoggers, oldlogger)
			updated = append(updated, nlogger.Name)
		} else {
			mergeLoggers = append(mergeLoggers, nlogger)
		}
	}
	for _, ln := range updated {
		delete(loggerMap, ln)
	}
	//now added remaining loggers
	for _, l := range loggerMap {
		mergeLoggers = append(mergeLoggers, l)
	}
	existingLogging.Logger = mergeLoggers

	//handlers
	log.Info("Merging handlers....")
	handlerMap := make(map[string]brokerv2alpha5.HandlerType)
	for _, handler := range existingLogging.Handler {
		handlerMap[handler.Name] = handler
	}
	mergeHandlers := make([]brokerv2alpha5.HandlerType, 0)
	//remember updated loggers
	updatedHandlers := make([]string, 0)

	for _, nhandler := range newLogging.Handler {
		if oldhandler, ok := handlerMap[nhandler.Name]; ok {
			if nhandler.ClassName != nil {
				if oldhandler.ClassName == nil {
					oldhandler.ClassName = new(string)
				}
				*oldhandler.ClassName = *nhandler.ClassName
			}
			if len(nhandler.Properties) > 0 {
				handlerPropMap := make(map[string]string)
				oldHandlerPropRef := &oldhandler.Properties
				for _, v := range oldhandler.Properties {
					handlerPropMap[v.Name] = v.Value
				}
				for _, v := range nhandler.Properties {
					handlerPropMap[v.Name] = v.Value
				}
				//empty old properties and add all new
				mergeHandlerProperties := make([]brokerv2alpha5.PropertyValueType, 0)
				for k, v := range handlerPropMap {
					mergeHandlerProperties = append(mergeHandlerProperties, brokerv2alpha5.PropertyValueType{k, v})
				}
				*oldHandlerPropRef = mergeHandlerProperties
			}
			mergeHandlers = append(mergeHandlers, oldhandler)
			updatedHandlers = append(updatedHandlers, nhandler.Name)
		} else {
			mergeHandlers = append(mergeHandlers, nhandler)
		}
	}
	for _, hn := range updatedHandlers {
		delete(handlerMap, hn)
	}
	//now added remaining handlers
	for _, h := range handlerMap {
		mergeHandlers = append(mergeHandlers, h)
	}
	existingLogging.Handler = mergeHandlers

	log.Info("Now merging formatters...")
	//formatters
	formatterMap := make(map[string]brokerv2alpha5.HandlerType)
	for _, formatter := range existingLogging.Formatter {
		formatterMap[formatter.Name] = formatter
	}
	mergeFormatters := make([]brokerv2alpha5.HandlerType, 0)
	//remember updated loggers
	updatedFormatters := make([]string, 0)

	for _, nformatter := range newLogging.Formatter {
		if oldformatter, ok := formatterMap[nformatter.Name]; ok {
			if nformatter.ClassName != nil {
				if oldformatter.ClassName == nil {
					oldformatter.ClassName = new(string)
				}
				*oldformatter.ClassName = *nformatter.ClassName
			}
			if len(nformatter.Properties) > 0 {
				formatterPropMap := make(map[string]string)
				oldFormatterPropRef := &oldformatter.Properties
				for _, v := range oldformatter.Properties {
					formatterPropMap[v.Name] = v.Value
				}
				for _, v := range nformatter.Properties {
					formatterPropMap[v.Name] = v.Value
				}
				//update the old formatter
				mergeFormatterProperties := make([]brokerv2alpha5.PropertyValueType, 0)
				for k, v := range formatterPropMap {
					mergeFormatterProperties = append(mergeFormatterProperties, brokerv2alpha5.PropertyValueType{k, v})
				}
				*oldFormatterPropRef = mergeFormatterProperties
			}
			mergeFormatters = append(mergeFormatters, oldformatter)
			updatedFormatters = append(updatedFormatters, nformatter.Name)
		} else {
			mergeFormatters = append(mergeFormatters, nformatter)
		}
	}
	for _, fn := range updatedFormatters {
		delete(formatterMap, fn)
	}
	//now added remaining formatter
	for _, f := range formatterMap {
		mergeFormatters = append(mergeFormatters, f)
	}
	existingLogging.Formatter = mergeFormatters
}

//returns true if currentAddressSettings need update
func (amqbfsm *ActiveMQArtemisFSM) compareAddressSettings(currentAddressSettings *brokerv2alpha5.AddressSettingsType, newAddressSettings *brokerv2alpha5.AddressSettingsType) bool {

	if (*currentAddressSettings).ApplyRule == nil {
		if (*newAddressSettings).ApplyRule != nil {
			return true
		}
	} else {
		if (*newAddressSettings).ApplyRule != nil {
			if *(*currentAddressSettings).ApplyRule != *(*newAddressSettings).ApplyRule {
				return true
			}
		} else {
			return true
		}
	}
	if len((*currentAddressSettings).AddressSetting) != len((*newAddressSettings).AddressSetting) || !config.IsEqualV2Alpha5((*currentAddressSettings).AddressSetting, (*newAddressSettings).AddressSetting) {
		return true
	}
	return false
}

//return true if need update
func (amqbfsm *ActiveMQArtemisFSM) compareLogging(currentLogging *brokerv2alpha5.LoggingType, newLogging *brokerv2alpha5.LoggingType) bool {
	if len(currentLogging.Logger) != len(newLogging.Logger) || len(currentLogging.Handler) != len(newLogging.Handler) || len(currentLogging.Formatter) != len(newLogging.Formatter) {
		return true
	}
	if !amqbfsm.compareLogger(currentLogging.Logger, newLogging.Logger) {
		if !amqbfsm.compareHandlerOrFormatter(currentLogging.Handler, newLogging.Handler) {
			if !amqbfsm.compareHandlerOrFormatter(currentLogging.Formatter, newLogging.Formatter) {
				return false
			}
		}
	}
	return true
}

func (amqbfsm *ActiveMQArtemisFSM) compareLogger(currentLogger []brokerv2alpha5.LoggerType, newLogger []brokerv2alpha5.LoggerType) bool {
	currentLoggerMap := make(map[string]brokerv2alpha5.LoggerType)
	for _, logger := range currentLogger {
		currentLoggerMap[logger.Name] = logger
	}
	newLoggerMap := make(map[string]brokerv2alpha5.LoggerType)
	for _, logger := range newLogger {
		newLoggerMap[logger.Name] = logger
	}
	for n, logger := range currentLoggerMap {
		if nlogger, ok := newLoggerMap[n]; ok {
			if logger.Level == nil {
				if nlogger.Level != nil {
					return true
				}
			} else {
				if nlogger.Level != nil {
					if *logger.Level != *nlogger.Level {
						return true
					}
				}
			}
			if logger.Handlers == nil {
				if nlogger.Handlers != nil {
					return true
				}
			} else {
				if nlogger.Handlers != nil {
					if *logger.Handlers != *nlogger.Handlers {
						return true
					}
				}
			}
			if logger.UseParentHandlers == nil {
				if nlogger.UseParentHandlers != nil {
					return true
				}
			} else {
				if nlogger.UseParentHandlers != nil {
					if *logger.UseParentHandlers != *nlogger.UseParentHandlers {
						return true
					}
				}
			}
		} else {
			//existing logger not in new, continue, as at this point
			//we have equal size of loggers, so we can be sure there must
			//be some new loggers
			return true
		}
	}
	return false
}

func (amqbfsm *ActiveMQArtemisFSM) compareHandlerOrFormatter(currentOne []brokerv2alpha5.HandlerType, newOne []brokerv2alpha5.HandlerType) bool {
	log.Info("comparing handler for formatter", "current", currentOne, "new", newOne)
	currentOneMap := make(map[string]brokerv2alpha5.HandlerType)
	for _, curVal := range currentOne {
		currentOneMap[curVal.Name] = curVal
	}
	newOneMap := make(map[string]brokerv2alpha5.HandlerType)
	for _, newVal := range newOne {
		newOneMap[newVal.Name] = newVal
	}
	for n, cur := range currentOneMap {
		if nOne, ok := newOneMap[n]; ok {
			if cur.ClassName == nil {
				if nOne.ClassName != nil {
					return true
				}
			} else {
				if nOne.ClassName != nil {
					if *cur.ClassName != *nOne.ClassName {
						return true
					}
				}
			}
			if len(nOne.Properties) > 0 {
				if len(cur.Properties) == 0 || len(cur.Properties) != len(nOne.Properties) {
					return true
				}
				//now the two have same length
				curPropMap := make(map[string]string)
				for _, v := range cur.Properties {
					curPropMap[v.Name] = v.Value
				}
				newPropMap := make(map[string]string)
				for _, v := range nOne.Properties {
					newPropMap[v.Name] = v.Value
				}
				for name, value := range curPropMap {
					if nValue, ok := newPropMap[name]; ok {
						if value != nValue {
							return true
						}
					} else {
						return true
					}
				}
			}
		} else {
			//existing logger not in new, continue, as at this point
			//we have equal size of loggers, so we can be sure there must
			//be some new loggers
			return true
		}
	}
	return false
}
