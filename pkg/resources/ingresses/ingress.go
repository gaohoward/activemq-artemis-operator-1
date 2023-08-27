package ingresses

import (
	//"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/selectors"

	"fmt"

	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/certutil"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultIngressDomain  = "apps.artemiscloud.io"
	TLS_PASSTHROUGH       = "passthrough"
	TLS_EDGE              = "edge"
	CM_ANN_ISSUER         = "cert-manager.io/issuer"
	CM_ANN_CLUSTER_ISSUER = "cert-manager.io/cluster-issuer"
)

func NewIngressForCRWithSSL(existing *netv1.Ingress, namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string, sslEnabled bool, domain string, brokerHost string, tlsTermType *string, annotations map[string]string, client rtclient.Client) (*netv1.Ingress, error) {

	pathType := netv1.PathTypePrefix

	var desired *netv1.Ingress
	if existing == nil {
		desired = &netv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels:    labels,
				Name:      targetServiceName + "-ing",
				Namespace: namespacedName.Namespace,
			},
			Spec: netv1.IngressSpec{},
		}
	} else {
		desired = existing
	}

	desired.Spec.Rules = []netv1.IngressRule{
		{
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: []netv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: netv1.IngressBackend{
								Service: &netv1.IngressServiceBackend{
									Name: targetServiceName,
								},
							},
						},
					},
				},
			},
		},
	}

	portName := ""
	portNumber := -1
	portValue := intstr.FromString(targetPortName)
	if portNumber = portValue.IntValue(); portNumber == 0 {
		portName = portValue.String()
	}

	if portName == "" {
		desired.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port = netv1.ServiceBackendPort{
			Number: int32(portNumber),
		}
	} else {
		desired.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port = netv1.ServiceBackendPort{
			Name: portName,
		}
	}

	if domain == "" {
		domain = defaultIngressDomain
	}

	host := desired.GetObjectMeta().GetName() + "." + domain

	if brokerHost != "" {
		host = brokerHost
	}

	desired.Spec.Rules[0].Host = host

	tlsTerminationType := TLS_PASSTHROUGH
	if tlsTermType != nil {
		if err := validateTlsTermType(*tlsTermType); err != nil {
			return nil, err
		}
		tlsTerminationType = *tlsTermType
	}

	if desired.Annotations == nil {
		desired.Annotations = make(map[string]string)
	}

	for k, v := range annotations {
		desired.Annotations[k] = v
	}

	desired.Spec.TLS = nil

	switch tlsTerminationType {
	case TLS_PASSTHROUGH:
		if sslEnabled {
			desired.Annotations["nginx.ingress.kubernetes.io/ssl-passthrough"] = "true"
			if _, ok := desired.Annotations[CM_ANN_CLUSTER_ISSUER]; ok {
				return nil, fmt.Errorf("no cert-manager annotation is allowed for ssl passthough type ingress: %s", CM_ANN_CLUSTER_ISSUER)
			}
			if _, ok := desired.Annotations[CM_ANN_ISSUER]; ok {
				return nil, fmt.Errorf("no cert-manager annotation is allowed for ssl passthough type ingress: %s", CM_ANN_ISSUER)
			}
			desired.Spec.TLS = []netv1.IngressTLS{{Hosts: []string{host}}}
		}
	case TLS_EDGE:
		if sslEnabled {
			return nil, fmt.Errorf("edge type tls termination is not supported when console is sslEnabled")
		}
		//either the ingress controller provides its own cert or cert-manager provides it via annotations
		//if cert-manager annotations are provided, check their validity
		if issuer, ok := annotations[CM_ANN_ISSUER]; ok {
			if err := certutil.CheckIssuerExists(issuer, &desired.Namespace, client); err != nil {
				return nil, err
			}
		}
		if clusterIssuer, ok := annotations[CM_ANN_CLUSTER_ISSUER]; ok {
			if err := certutil.CheckIssuerExists(clusterIssuer, nil, client); err != nil {
				return nil, err
			}
		}
		// without cert-manager annotations this secret must be provided by the user.
		ingressTlsSecretName := host + "-secret"
		desired.Spec.TLS = []netv1.IngressTLS{
			{
				Hosts:      []string{host},
				SecretName: ingressTlsSecretName,
			},
		}
	}

	return desired, nil
}

func validateTlsTermType(value string) error {
	if value != TLS_PASSTHROUGH && value != TLS_EDGE {
		return fmt.Errorf("invalid route termination type: %s", value)
	}
	return nil
}
