package routes

import (
	"fmt"

	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/certutil"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultRouteDomain = "apps.artemiscloud.io"
	TLS_PASSTHROUGH    = "passthrough"
	TLS_EDGE           = "edge"
	TLS_REENCRYPT      = "reencrypt"
	CM_ANN_ISSUER      = "cert-manager.io/issuer-name"
	CM_ANN_ISSUER_KIND = "cert-manager.io/issuer-kind"
)

func NewRouteDefinitionForCR(existing *routev1.Route, namespacedName types.NamespacedName, labels map[string]string, targetServiceName string, targetPortName string, domain string, brokerHost string, sslEnabled bool, tlsTermType *string, annotations map[string]string, client rtclient.Client, targetCA []byte) (*routev1.Route, error) {

	var desired *routev1.Route = nil
	if existing == nil {
		desired = &routev1.Route{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "route.openshift.io/v1",
				Kind:       "Route",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels:    labels,
				Name:      targetServiceName + "-rte",
				Namespace: namespacedName.Namespace,
			},
			Spec: routev1.RouteSpec{},
		}

	} else {
		desired = existing.DeepCopy()
		desired.Labels = labels
	}

	if domain == "" {
		domain = defaultRouteDomain
	}

	if brokerHost != "" {
		desired.Spec.Host = brokerHost
	} else if domain != "" {
		desired.Spec.Host = desired.GetObjectMeta().GetName() + "." + domain
	}

	desired.Spec.Port = &routev1.RoutePort{
		TargetPort: intstr.FromString(targetPortName),
	}

	desired.Spec.To = routev1.RouteTargetReference{
		Kind: "Service",
		Name: targetServiceName,
	}
	tlsTerminationType := TLS_PASSTHROUGH
	if tlsTermType != nil {
		if err := validateTlsTermType(*tlsTermType); err != nil {
			return nil, err
		}
		tlsTerminationType = *tlsTermType
	}

	if len(annotations) > 0 && desired.Annotations == nil {
		desired.Annotations = make(map[string]string)
	}

	for k, v := range annotations {
		desired.Annotations[k] = v
	}

	desired.Spec.TLS = nil

	switch tlsTerminationType {
	case TLS_PASSTHROUGH:
		if sslEnabled {
			if _, ok := desired.Annotations[CM_ANN_ISSUER]; ok {
				return nil, fmt.Errorf("no cert-manager annotation is allowed for ssl passthough type route: %s", CM_ANN_ISSUER)
			}
		}
		desired.Spec.TLS = &routev1.TLSConfig{
			Termination: routev1.TLSTerminationPassthrough,
			// empty or redirect
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
		}
	case TLS_EDGE:
		desired.Spec.TLS = &routev1.TLSConfig{
			Termination:                   routev1.TLSTerminationEdge,
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
		}
		if issuer, ok := desired.Annotations[CM_ANN_ISSUER]; ok {
			issuerNs := &desired.Namespace
			if kind, ok := desired.Annotations[CM_ANN_ISSUER_KIND]; ok {
				if kind == "ClusterIssuer" {
					issuerNs = nil
				}
			}
			if err := certutil.CheckIssuerExists(issuer, issuerNs, client); err != nil {
				return nil, err
			}
		} else {
			//user can optionally supply their own certs in a secret. otherwise let the route have its own
			routeSecretName := desired.Spec.Host + "-secret"
			secKey := types.NamespacedName{Name: routeSecretName, Namespace: desired.Namespace}
			userRouteSecret := corev1.Secret{}
			if err := resources.Retrieve(secKey, client, &userRouteSecret); err == nil {
				fmt.Println("yes secret exists " + userRouteSecret.Name)
				//user route secret should be a standard tls secret
				if tlsKey, ok := userRouteSecret.Data["tls.key"]; ok {
					desired.Spec.TLS.Key = string(tlsKey)
				} else {
					return nil, fmt.Errorf("user tls secret doesn't have tls.key")
				}
				if tlsCrt, ok := userRouteSecret.Data["tls.crt"]; ok {
					desired.Spec.TLS.Certificate = string(tlsCrt)
				} else {
					return nil, fmt.Errorf("user tls secret doesn't have tls.crt")
				}
				if caCrt, ok := userRouteSecret.Data["ca.crt"]; ok {
					desired.Spec.TLS.CACertificate = string(caCrt)
				}
			}
		}
	case TLS_REENCRYPT:
		desired.Spec.TLS = &routev1.TLSConfig{
			Termination:                   routev1.TLSTerminationReencrypt,
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
		}
		if issuer, ok := desired.Annotations[CM_ANN_ISSUER]; ok {
			issuerNs := &desired.Namespace
			if kind, ok := desired.Annotations[CM_ANN_ISSUER_KIND]; ok {
				if kind == "ClusterIssuer" {
					issuerNs = nil
				}
			}
			if err := certutil.CheckIssuerExists(issuer, issuerNs, client); err != nil {
				return nil, err
			}
		} else {
			routeSecretName := desired.Spec.Host + "-secret"
			secKey := types.NamespacedName{Name: routeSecretName, Namespace: desired.Namespace}
			userRouteSecret := corev1.Secret{}
			if err := resources.Retrieve(secKey, client, &userRouteSecret); err == nil {
				//user route secret should be a standard tls secret
				if tlsKey, ok := userRouteSecret.Data["tls.key"]; ok {
					desired.Spec.TLS.Key = string(tlsKey)
				} else {
					return nil, fmt.Errorf("user tls secret doesn't have tls.key")
				}
				if tlsCrt, ok := userRouteSecret.Data["tls.crt"]; ok {
					desired.Spec.TLS.Certificate = string(tlsCrt)
				} else {
					return nil, fmt.Errorf("user tls secret doesn't have tls.crt")
				}
				if caCrt, ok := userRouteSecret.Data["ca.crt"]; ok {
					desired.Spec.TLS.CACertificate = string(caCrt)
				}
				if destCrt, ok := userRouteSecret.Data["destCa.crt"]; ok {
					desired.Spec.TLS.DestinationCACertificate = string(destCrt)
				} else {
					//need to fetch the CA from target console's secret
					if targetCA != nil {
						desired.Spec.TLS.DestinationCACertificate = string(targetCA)
					}
				}
			} else {
				// route use its own default cert.
				if targetCA != nil {
					desired.Spec.TLS.DestinationCACertificate = string(targetCA)
				}
			}
		}
	}
	return desired, nil
}

func validateTlsTermType(value string) error {
	if value != TLS_PASSTHROUGH && value != TLS_REENCRYPT && value != TLS_EDGE {
		return fmt.Errorf("invalid route termination type: %s", value)
	}
	return nil
}
