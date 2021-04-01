package v2alpha2activemqartemisaddress

import (
	"context"
	"strconv"

	mgmt "github.com/artemiscloud/activemq-artemis-management"
	brokerv2alpha2 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v2alpha2"
	v2alpha4 "github.com/artemiscloud/activemq-artemis-operator/pkg/controller/broker/v2alpha4/activemqartemis"

	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/secrets"
	ss "github.com/artemiscloud/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/channels"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_v2alpha2activemqartemisaddress")

type AddressDeployment struct {
	AddressResource brokerv2alpha2.ActiveMQArtemisAddress
	//a 0-len array means all statefulsets
	SsTargetNameBuilders []namer.NamerData
}

var namespacedNameToAddressName = make(map[types.NamespacedName]AddressDeployment)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ActiveMQArtemisAddress Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	go setupAddressObserver(mgr, channels.AddressListeningCh)
	return &ReconcileActiveMQArtemisAddress{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func setupAddressObserver(mgr manager.Manager, c chan types.NamespacedName) {
	log.Info("Setting up address observer")
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Error(err, "Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "Error building kubernetes clientset: %s", err.Error())
	}

	namespace, err := k8sutil.GetWatchNamespace()

	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		return
	}

	observer := NewAddressObserver(kubeClient, namespace, mgr.GetClient())

	if err = observer.Run(channels.AddressListeningCh); err != nil {
		log.Error(err, "Error running controller: %s", err.Error())
	}

	log.Info("Finish setup address observer")
	return
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("v2alpha2activemqartemisaddress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemisAddress
	err = c.Watch(&source.Kind{Type: &brokerv2alpha2.ActiveMQArtemisAddress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemisAddress
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv2alpha2.ActiveMQArtemisAddress{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileActiveMQArtemisAddress{}

// ReconcileActiveMQArtemisAddress reconciles a ActiveMQArtemisAddress object
type ReconcileActiveMQArtemisAddress struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ActiveMQArtemisAddress object and makes changes based on the state read
// and what is in the ActiveMQArtemisAddress.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileActiveMQArtemisAddress) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemisAddress")

	// Fetch the ActiveMQArtemisAddress instance
	instance := &brokerv2alpha2.ActiveMQArtemisAddress{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		// Delete action
		addressInstance, lookupSucceeded := namespacedNameToAddressName[request.NamespacedName]

		if lookupSucceeded {
			if addressInstance.AddressResource.Spec.RemoveFromBrokerOnDelete {
				err = deleteQueue(&addressInstance, request, r.client)
			} else {
				log.Info("Not to delete address", "address", addressInstance)
			}
			delete(namespacedNameToAddressName, request.NamespacedName)
		}
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		log.Error(err, "Requeue the request for error")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	} else {
		addressDeployment := AddressDeployment{
			AddressResource:      *instance,
			SsTargetNameBuilders: createNameBuilders(instance),
		}
		err = createQueue(&addressDeployment, request, r.client)
		if nil == err {
			namespacedNameToAddressName[request.NamespacedName] = addressDeployment
		}
	}

	return reconcile.Result{}, nil
}

func createNameBuilders(instance *brokerv2alpha2.ActiveMQArtemisAddress) []namer.NamerData {
	var nameBuilders []namer.NamerData = nil
	for _, crName := range instance.Spec.TargetBrokerCrNames {
		if crName != "*" {
			builder := namer.CreateStatefulSetNameBuilder(crName)
			log.Info("created a new name builder", "builder", builder, "buldername", builder.Name())
			nameBuilders = append(nameBuilders, builder)
			log.Info("added one builder for "+crName, "builders", nameBuilders, "len", len(nameBuilders))
		} else {
			return nil
		}
	}
	log.Info("Created ss name builder for addr", "instance", instance, "builders", nameBuilders)
	return nameBuilders
}

func createQueue(instance *AddressDeployment, request reconcile.Request, client client.Client) error {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Creating ActiveMQArtemisAddress")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for _, a := range artemisArray {
			if nil == a {
				reqLogger.Info("Creating ActiveMQArtemisAddress artemisArray had a nil!")
				continue
			}
			_, err := a.CreateQueue(instance.AddressResource.Spec.AddressName, instance.AddressResource.Spec.QueueName, instance.AddressResource.Spec.RoutingType)
			if nil != err {
				reqLogger.Info("Creating ActiveMQArtemisAddress error for " + instance.AddressResource.Spec.QueueName)
				break
			} else {
				reqLogger.Info("Created ActiveMQArtemisAddress for " + instance.AddressResource.Spec.QueueName)
			}
		}
	}

	return err
}

func deleteQueue(instance *AddressDeployment, request reconcile.Request, client client.Client) error {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Deleting ActiveMQArtemisAddress")

	var err error = nil
	artemisArray := getPodBrokers(instance, request, client)
	if nil != artemisArray {
		for _, a := range artemisArray {
			_, err := a.DeleteQueue(instance.AddressResource.Spec.QueueName)
			if nil != err {
				reqLogger.Info("Deleting ActiveMQArtemisAddress error for " + instance.AddressResource.Spec.QueueName)
				break
			} else {
				reqLogger.Info("Deleted ActiveMQArtemisAddress for " + instance.AddressResource.Spec.QueueName)
				reqLogger.Info("Checking parent address for bindings " + instance.AddressResource.Spec.AddressName)
				bindingsData, err := a.ListBindingsForAddress(instance.AddressResource.Spec.AddressName)
				if nil == err {
					if "" == bindingsData.Value {
						reqLogger.Info("No bindings found removing " + instance.AddressResource.Spec.AddressName)
						a.DeleteAddress(instance.AddressResource.Spec.AddressName)
					} else {
						reqLogger.Info("Bindings found, not removing " + instance.AddressResource.Spec.AddressName)
					}
				}
			}
		}
	}

	return err
}

func getPodBrokers(instance *AddressDeployment, request reconcile.Request, client client.Client) []*mgmt.Artemis {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Getting Pod Brokers", "instance", instance)

	var artemisArray []*mgmt.Artemis = nil

	targetCrNamespacedNames := createTargetCrNamespacedNames(request.Namespace, instance.AddressResource.Spec.TargetBrokerCrNames)

	log.Info("target Cr names", "result", targetCrNamespacedNames)

	var ssNames []types.NamespacedName = v2alpha4.GetDeployedStatefuleSetNames(targetCrNamespacedNames)

	log.Info("got taget ssNames from broker controller", "ssNames", ssNames)

	for n, ssNamespacedName := range ssNames {
		log.Info("Now retrieve ss", "order", n, "ssName", ssNamespacedName)
		statefulset, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, client)
		if nil != err {
			reqLogger.Error(err, "error retriving ss")
			reqLogger.Info("Statefulset: " + ssNamespacedName.Name + " not found")
		} else {
			reqLogger.Info("Statefulset: " + ssNamespacedName.Name + " found")
			pod := &corev1.Pod{}
			podNamespacedName := types.NamespacedName{
				Name:      statefulset.Name + "-0",
				Namespace: request.Namespace,
			}

			// For each of the replicas
			var i int = 0
			var replicas int = int(*statefulset.Spec.Replicas)
			log.Info("finding pods in ss", "replicas", replicas)
			for i = 0; i < replicas; i++ {
				s := statefulset.Name + "-" + strconv.Itoa(i)
				podNamespacedName.Name = s
				log.Info("Trying finding pod " + s)
				if err = client.Get(context.TODO(), podNamespacedName, pod); err != nil {
					if errors.IsNotFound(err) {
						reqLogger.Error(err, "Pod IsNotFound", "Namespace", request.Namespace, "Name", request.Name)
					} else {
						reqLogger.Error(err, "Pod lookup error", "Namespace", request.Namespace, "Name", request.Name)
					}
				} else {
					reqLogger.Info("Pod found", "Namespace", request.Namespace, "Name", request.Name)
					containers := pod.Spec.Containers //get env from this
					var jolokiaUser string
					var jolokiaPassword string
					var jolokiaProtocol string
					if len(containers) == 1 {
						envVars := containers[0].Env
						for _, oneVar := range envVars {
							if "AMQ_USER" == oneVar.Name {
								jolokiaUser = getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client)
							}
							if "AMQ_PASSWORD" == oneVar.Name {
								jolokiaPassword = getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client)
							}
							if "AMQ_CONSOLE_ARGS" == oneVar.Name {
								jolokiaProtocol = getEnvVarValue(&oneVar, &podNamespacedName, statefulset, client)
							}
							if jolokiaUser != "" && jolokiaPassword != "" && jolokiaProtocol != "" {
								break
							}
						}
					}

					if jolokiaProtocol == "" {
						jolokiaProtocol = "http"
					} else {
						jolokiaProtocol = "https"
					}

					reqLogger.Info("New Jolokia with ", "User: ", jolokiaUser, "Protocol: ", jolokiaProtocol)
					artemis := mgmt.GetArtemis(pod.Status.PodIP, "8161", "amq-broker", jolokiaUser, jolokiaPassword, jolokiaProtocol)
					artemisArray = append(artemisArray, artemis)
				}
			}
		}
	}

	log.Info("Finally we gathered some mgmt arry", "size", len(artemisArray))
	return artemisArray
}

func getEnvVarValue(envVar *corev1.EnvVar, namespace *types.NamespacedName, statefulset *appsv1.StatefulSet, client client.Client) string {
	var result string
	if envVar.Value == "" {
		result = getEnvVarValueFromSecret(envVar.Name, envVar.ValueFrom, namespace, statefulset, client)
	} else {
		result = envVar.Value
	}
	return result
}

func getEnvVarValueFromSecret(envName string, varSource *corev1.EnvVarSource, namespace *types.NamespacedName, statefulset *appsv1.StatefulSet, client client.Client) string {

	reqLogger := log.WithValues("Namespace", namespace.Name, "StatefulSet", statefulset.Name)

	var result string = ""

	secretName := varSource.SecretKeyRef.LocalObjectReference.Name
	secretKey := varSource.SecretKeyRef.Key

	namespacedName := types.NamespacedName{
		Name:      secretName,
		Namespace: statefulset.Namespace,
	}
	// Attempt to retrieve the secret
	stringDataMap := map[string]string{
		envName: "",
	}
	theSecret := secrets.NewSecret(namespacedName, secretName, stringDataMap)
	var err error = nil
	if err = resources.Retrieve(namespacedName, client, theSecret); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Secret IsNotFound.", "Secret Name", secretName, "Key", secretKey)
		}
	} else {
		elem, ok := theSecret.Data[envName]
		if ok {
			result = string(elem)
		}
	}
	return result
}

func createTargetCrNamespacedNames(namespace string, targetCrNames []string) []types.NamespacedName {
	var result []types.NamespacedName = nil
	for _, crName := range targetCrNames {
		result = append(result, types.NamespacedName{
			Namespace: namespace,
			Name:      crName,
		})
	}
	return result
}

func GetStatefulSetNameForPod(pod *types.NamespacedName) (string, int) {
	for crName, addressDeployment := range namespacedNameToAddressName {
		if crName.Namespace != pod.Namespace {
			return "", -1
		}
		if len(addressDeployment.SsTargetNameBuilders) == 0 {
			//deploy to all sts, need get from broker controller
			ssNames := v2alpha4.GetDeployedStatefuleSetNames(nil)
			if len(ssNames) == 0 {
				return "", -1
			}
			for _, ssName := range ssNames {
				if ok, podSerial := podBelongsToStatefulset(pod, &ssName.Name); ok {
					return ssName.Name, podSerial
				}
			}
			return "", -1
		}
		//iterate and check the ss name
		for _, ssNameBuilder := range addressDeployment.SsTargetNameBuilders {
			ssName := ssNameBuilder.Name()
			if ok, podSerial := podBelongsToStatefulset(pod, &ssName); ok {
				return ssName, podSerial
			}
		}
	}
	return "", -1
}

func podBelongsToStatefulset(pod *types.NamespacedName, ssName *string) (bool, int) {
	//pod name must be <ss>-<n>
	if len(*ssName) > len(pod.Name)-2 {
		log.Info("Original pod name too short", "pod name", pod.Name, "base", ssName)
		return false, -1
	}

	podSerial := pod.Name[len(*ssName)+1:]

	//convert to int
	i, err := strconv.Atoi(podSerial)
	if err != nil || i < 1 {
		log.Error(err, "failed to convert pod name", "pod", pod)
		return false, -1
	}

	return true, i
}
