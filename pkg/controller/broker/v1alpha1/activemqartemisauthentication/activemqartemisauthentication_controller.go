package v1alpha1activemqartemisauthentication

import (
	"context"

	brokerv1alpha1 "github.com/artemiscloud/activemq-artemis-operator/pkg/apis/broker/v1alpha1"
	v2alpha5 "github.com/artemiscloud/activemq-artemis-operator/pkg/controller/broker/v2alpha5/activemqartemis"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/environments"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_v1alpha1activemqartemisauthentication")
var namespacedNameToAddressName = make(map[types.NamespacedName]brokerv1alpha1.ActiveMQArtemisAuthentication)

// Add creates a new ActiveMQArtemisAuthentication Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileActiveMQArtemisAuthentication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("v1alpha1activemqartemisauthentication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemisAuthentication
	err = c.Watch(&source.Kind{Type: &brokerv1alpha1.ActiveMQArtemisAuthentication{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileActiveMQArtemisAuthentication{}

type ReconcileActiveMQArtemisAuthentication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileActiveMQArtemisAuthentication) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemisAuthentication")

	instance := &brokerv1alpha1.ActiveMQArtemisAuthentication{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)

	reqLogger.Info("Fetched instance", "the instance", instance)
	v2alpha5.AddBrokerConfigHandler(request.NamespacedName, &ActiveMQArtemisAuthenticationConfigHandler{
		instance,
	})
	return reconcile.Result{}, err
}

type ActiveMQArtemisAuthenticationConfigHandler struct {
	AuthenticationCR *brokerv1alpha1.ActiveMQArtemisAuthentication
}

func (r *ActiveMQArtemisAuthenticationConfigHandler) Config(initContainers []corev1.Container, outputDir string) (value []string) {
	log.Info("Reconciling authentication", "cr", r.AuthenticationCR)
	var configCmds = []string{"echo \"making dir " + outputDir + "\"", "mkdir -p " + outputDir}
	filePath := outputDir + "/security-config.yaml"
	cmdPersistCRAsYaml, err := r.persistCR(filePath)
	if err != nil {
		log.Error(err, "Error marshalling authentication CR", "cr", r.AuthenticationCR)
		return nil
	}
	log.Info("get the command", "value", cmdPersistCRAsYaml)
	configCmds = append(configCmds, cmdPersistCRAsYaml)
	configCmds = append(configCmds, "/opt/amq-broker/script/cfg/authentication.sh")
	//export env var SECURITY_CFG_YAML
	envVarName := "SECURITY_CFG_YAML"
	envVar := corev1.EnvVar{
		envVarName,
		filePath,
		nil,
	}
	environments.Create(initContainers, &envVar)

	return configCmds
}

func (r *ActiveMQArtemisAuthenticationConfigHandler) persistCR(filePath string) (value string, err error) {

	data, err := yaml.Marshal(r.AuthenticationCR)
	if err != nil {
		return "", err
	}
	return "echo \"" + string(data) + "\" > " + filePath, nil
}
