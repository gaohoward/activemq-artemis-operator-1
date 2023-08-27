/*
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
// +kubebuilder:docs-gen:collapse=Apache License

package controllers

import (
	"crypto/x509"
	"encoding/json"
	"os"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/routes"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/certutil"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	brokerCrName = "broker-cert-mgr"

	selfsignedIssuerName        = "selfsigned-issuer"
	selfsignedClusterIssuerName = "cluster-selfsigned-issuer"
)

var (
	serverCertNoKeystoreName              = "server-cert-secret-no-keystore"
	serverCertWithPkcs12Name              = "server-cert-pkcs12"
	serverCertWithJksName                 = "server-cert-jks"
	serverCertClusterIssuerNoKeystoreName = "cluster-server-cert-secret-no-keystore"
	serverCertClusterIssuerWithPkcs12Name = "cluster-server-cert-pkcs12"
	serverCertClusterIssuerWithJksName    = "cluster-server-cert-jks"
	cmCommonSecretName                    = "cm-common-secret-for-test"
	pkcsPasswordKey                       = "pkcs12-password"
	jksPasswordKey                        = "jks-password"
	pkcsPassword                          = "pkcs12-password"
	jksPassword                           = "jks-password"
	adminUser                             = "testuser"
	adminPassword                         = "testpassword"
	ingressTypePassThrough                = "passthrough"
	ingressTypeEdge                       = "edge"
	routeTypeReencrypt                    = "reencrypt"
	issuerForIngress                      = "selfsigned-issuer-ingress"
	issuerKind                            = "Issuer"
	ingressCommonName                     = "arkamq.io"
	userKeystoreSecretName                = "user-common-tls-secret"
)

var _ = Describe("artemis controller with cert manager test", Label("controller-cert-mgr-test"), func() {
	var installedCertManager bool = false

	BeforeEach(func() {
		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
			//if cert manager is not installed, install it
			if !CertManagerInstalled() {
				if isOpenshift {
					//openshift need install cert-manager from operatorHUB
					Fail("Please install cert-manager from operatorHub and also manually install openshift-route")
				}
				Expect(InstallCertManager()).To(Succeed())
				installedCertManager = true
			}

			InstallSelfSignIssuer(selfsignedIssuerName)
			InstallClusteredSelfSignIssuer(selfsignedClusterIssuerName)
		}
	})

	AfterEach(func() {
		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
			UninstallClusteredSelfSignIssuer(selfsignedClusterIssuerName)
			UninstallSelfSignIssuer(selfsignedIssuerName)

			if installedCertManager {
				Expect(UninstallCertManager()).To(Succeed())
				installedCertManager = false
			}
		}
	})

	Describe("cert manager integration test", func() {
		Context("tls exposure with cert manager and local issuer", func() {
			BeforeEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					InstallSecret(cmCommonSecretName, defaultNamespace, func(candidate *corev1.Secret) {
						candidate.StringData[pkcsPasswordKey] = pkcsPassword
						candidate.StringData[jksPasswordKey] = jksPassword
					})
					InstallSelfSignedCert(serverCertNoKeystoreName, defaultNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedIssuerName,
							Kind: "Issuer",
						}
					})
					InstallSelfSignedCert(serverCertWithPkcs12Name, defaultNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.Keystores = &cmv1.CertificateKeystores{
							PKCS12: &cmv1.PKCS12Keystore{
								Create: true,
								PasswordSecretRef: cmmetav1.SecretKeySelector{
									LocalObjectReference: cmmetav1.LocalObjectReference{
										Name: cmCommonSecretName,
									},
									Key: pkcsPasswordKey,
								},
							},
						}
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedIssuerName,
							Kind: "Issuer",
						}
					})
					InstallSelfSignedCert(serverCertWithJksName, defaultNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.Keystores = &cmv1.CertificateKeystores{
							JKS: &cmv1.JKSKeystore{
								Create: true,
								PasswordSecretRef: cmmetav1.SecretKeySelector{
									LocalObjectReference: cmmetav1.LocalObjectReference{
										Name: cmCommonSecretName,
									},
									Key: jksPasswordKey,
								},
							},
						}
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedIssuerName,
							Kind: "Issuer",
						}
					})
				}
			})
			AfterEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					UninstallSelfSignedCert(serverCertNoKeystoreName, defaultNamespace)
					UninstallSelfSignedCert(serverCertWithPkcs12Name, defaultNamespace)
					UninstallSelfSignedCert(serverCertWithJksName, defaultNamespace)
					UninstallSecret(cmCommonSecretName, defaultNamespace)
				}
			})
			It("cert has no keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					testCertWithNoKeystoreConfigured(serverCertNoKeystoreName)
				}
			})

			It("cert has pkcs12 keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					testCertWithKeystoreConfigured(serverCertWithPkcs12Name, pkcsPassword)
				}
			})
			It("cert has jks keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					testCertWithKeystoreConfigured(serverCertWithJksName, jksPassword)
				}
			})
		})
		Context("console tls exposure with cert manager using cluster issuer", func() {
			clusterIssuerCertificateNamespace := "cert-namespace"

			BeforeEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

					createNamespace(clusterIssuerCertificateNamespace)

					// the doc https://cert-manager.io/docs/reference/api-docs/#meta.cert-manager.io/v1.LocalObjectReference
					// seems not accurate for cluster issuer cert where the keystore password secret
					// should be in the same namespace as the cert's, not the cluster resource namespace.
					InstallSecret(cmCommonSecretName, clusterIssuerCertificateNamespace, func(candidate *corev1.Secret) {
						candidate.StringData[pkcsPasswordKey] = pkcsPassword
						candidate.StringData[jksPasswordKey] = jksPassword
					})
					InstallSelfSignedCert(serverCertClusterIssuerNoKeystoreName, clusterIssuerCertificateNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedClusterIssuerName,
							Kind: "ClusterIssuer",
						}
					})
					InstallSelfSignedCert(serverCertClusterIssuerWithPkcs12Name, clusterIssuerCertificateNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.Keystores = &cmv1.CertificateKeystores{
							PKCS12: &cmv1.PKCS12Keystore{
								Create: true,
								PasswordSecretRef: cmmetav1.SecretKeySelector{
									LocalObjectReference: cmmetav1.LocalObjectReference{
										Name: cmCommonSecretName,
									},
									Key: pkcsPasswordKey,
								},
							},
						}
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedClusterIssuerName,
							Kind: "ClusterIssuer",
						}
					})
					InstallSelfSignedCert(serverCertClusterIssuerWithJksName, clusterIssuerCertificateNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.Keystores = &cmv1.CertificateKeystores{
							JKS: &cmv1.JKSKeystore{
								Create: true,
								PasswordSecretRef: cmmetav1.SecretKeySelector{
									LocalObjectReference: cmmetav1.LocalObjectReference{
										Name: cmCommonSecretName,
									},
									Key: jksPasswordKey,
								},
							},
						}
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedClusterIssuerName,
							Kind: "ClusterIssuer",
						}
					})
				}
			})
			AfterEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					UninstallSelfSignedCert(serverCertClusterIssuerNoKeystoreName, clusterIssuerCertificateNamespace)
					UninstallSelfSignedCert(serverCertClusterIssuerWithPkcs12Name, clusterIssuerCertificateNamespace)
					UninstallSelfSignedCert(serverCertClusterIssuerWithJksName, clusterIssuerCertificateNamespace)
					UninstallSecret(cmCommonSecretName, clusterIssuerCertificateNamespace)

					deleteNamespace(clusterIssuerCertificateNamespace, Default)
				}
			})
			It("cert has no keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					certLocation := serverCertClusterIssuerNoKeystoreName + ":" + clusterIssuerCertificateNamespace
					testCertWithNoKeystoreConfigured(certLocation)
				}
			})

			It("cert has pkcs12 keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					certLocation := serverCertClusterIssuerWithPkcs12Name + ":" + clusterIssuerCertificateNamespace
					testCertWithKeystoreConfigured(certLocation, pkcsPassword)
				}
			})
			It("cert has jks keystore and truststore configured", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					certLocation := serverCertClusterIssuerWithJksName + ":" + clusterIssuerCertificateNamespace
					testCertWithKeystoreConfigured(certLocation, jksPassword)
				}
			})
		})

		Context("tls ingress exposure with cert manager and local issuer", func() {
			BeforeEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					InstallSelfSignIssuer(issuerForIngress)
					InstallSecret(cmCommonSecretName, defaultNamespace, func(candidate *corev1.Secret) {
						candidate.StringData[pkcsPasswordKey] = pkcsPassword
						candidate.StringData[jksPasswordKey] = jksPassword
					})
					InstallSelfSignedCert(serverCertWithPkcs12Name, defaultNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.Keystores = &cmv1.CertificateKeystores{
							PKCS12: &cmv1.PKCS12Keystore{
								Create: true,
								PasswordSecretRef: cmmetav1.SecretKeySelector{
									LocalObjectReference: cmmetav1.LocalObjectReference{
										Name: cmCommonSecretName,
									},
									Key: pkcsPasswordKey,
								},
							},
						}
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: selfsignedIssuerName,
							Kind: "Issuer",
						}
					})
				}
			})
			AfterEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					UninstallSelfSignedCert(serverCertWithPkcs12Name, defaultNamespace)
					UninstallSecret(cmCommonSecretName, defaultNamespace)
					UninstallSelfSignIssuer(issuerForIngress)
				}
			})
			It("test console exposure with cert-manager", Label("console-expose-cert-manager"), func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					if isOpenshift {
						testRouteTlsTerminateWithCertManager()
					} else {
						testIngressTlsTerminateWithCertManager()
					}
				}
			})
			It("test console exposure without cert-manager", Label("console-expose-no-cert-manager"), func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					if isOpenshift {
						testRouteTlsTerminateWithoutCertManager()
					} else {
						testIngressTlsTerminateWithoutCertManager()
					}
				}
			})

		})
	})
})

func testIngressTlsTerminateWithoutCertManager() {
	By("creating user secret")
	//each pod need a separate secret where the cert's common name matches host name
	//otherwise ingress refuses to use it
	hostName := brokerCrName + "-wconsj-0-svc-ing.apps.artemiscloud.io"
	tlsKeyPemBytes, tlsCrtPemBytes, err := GeneratePemCertificate(func(cert *x509.Certificate) {
		cert.Subject.CommonName = hostName
	})
	Expect(err).To(BeNil())

	userIngressTlsSecretName := hostName + "-secret"
	userIngressTlsSecret := InstallSecret(userIngressTlsSecretName, defaultNamespace, func(candidate *corev1.Secret) {
		candidate.StringData["tls.key"] = string(tlsKeyPemBytes)
		candidate.StringData["tls.crt"] = string(tlsCrtPemBytes)
		candidate.Type = corev1.SecretTypeTLS
	})

	By("Deploying the broker cr exposing console with edge terminaion")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = false
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypeEdge,
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

		ingName := brokerCr.Name + "-wconsj-0-svc-ing"
		ingress := netv1.Ingress{}
		ingKey := types.NamespacedName{Name: ingName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, ingKey, &ingress)).Should(Succeed())

		g.Expect(len(ingress.Spec.TLS)).To(BeEquivalentTo(1))
		tls0 := ingress.Spec.TLS[0]
		g.Expect(tls0.SecretName).To(Equal(userIngressTlsSecret.Name))
		g.Expect(tls0.Hosts[0]).To(Equal(hostName))

		ingSecretKey := types.NamespacedName{Name: hostName + "-secret", Namespace: defaultNamespace}
		ingSecret := corev1.Secret{}
		g.Expect(k8sClient.Get(ctx, ingSecretKey, &ingSecret)).Should(Succeed())
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	CleanResource(userIngressTlsSecret, userIngressTlsSecretName, userIngressTlsSecret.Namespace)
}

func testRouteTlsTerminateWithoutCertManager() {
	By("creating user secret")
	userSecret, err := CreateTlsSecret(userKeystoreSecretName, defaultNamespace, "password", nil)
	Expect(err).To(Succeed())
	Expect(k8sClient.Create(ctx, userSecret)).Should(Succeed())
	secretKey := types.NamespacedName{Name: userKeystoreSecretName, Namespace: defaultNamespace}
	theSecret := &corev1.Secret{}
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, secretKey, theSecret)).Should(Succeed())
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("deploying broker exposing console reencrypt with user sslsecret")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.SSLSecret = userKeystoreSecretName
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &routeTypeReencrypt,
		}
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue), condition.Message)

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationReencrypt))
		g.Expect(route.Spec.TLS.DestinationCACertificate).NotTo(BeEmpty())
		g.Expect(route.Spec.TLS.DestinationCACertificate).To(ContainSubstring("BEGIN CERTIFICATE"))
		cert, err := certutil.ParsePemCertificate(&route.Spec.TLS.DestinationCACertificate, &pkcsPassword)
		g.Expect(err).To(Succeed())
		g.Expect(cert.Subject.CommonName).To(Equal("ArtemisCloud Broker"))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("deploying broker exposing console passthrough with user sslsecret")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {
		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.SSLSecret = userKeystoreSecretName
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypePassThrough,
		}
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue), condition.Message)

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationPassthrough))

		By("checking jolokia access")
		pod0Name := createdBrokerCr.Name + "-ss-0"
		Eventually(func(g Gomega) {
			checkReadPodStatus(pod0Name, createdBrokerCr.Name, g)
		}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	CleanResource(theSecret, theSecret.Name, theSecret.Namespace)

	By("Deploying broker exposing console reencrypt no cert manager")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.BrokerCert = &serverCertWithPkcs12Name
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &routeTypeReencrypt,
		}
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue), condition.Message)

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationReencrypt))
		g.Expect(route.Spec.TLS.DestinationCACertificate).NotTo(BeEmpty())
		g.Expect(route.Spec.TLS.DestinationCACertificate).To(ContainSubstring("BEGIN CERTIFICATE"))
		cert, err := certutil.ParsePemCertificate(&route.Spec.TLS.DestinationCACertificate, &pkcsPassword)
		g.Expect(err).To(Succeed())
		g.Expect(cert.Subject.CommonName).To(Equal("arkmq.org"))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
}

func testRouteTlsTerminateWithCertManager() {

	By("Deploying the broker cr exposing console sslEnabled with invalid cert-manager annotation")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.BrokerCert = &serverCertWithPkcs12Name
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypePassThrough,
			Annotations: []brokerv1beta1.KeyValueType{
				{
					Key:   routes.CM_ANN_ISSUER,
					Value: &issuerForIngress,
				},
			},
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionFalse))
		g.Expect(condition.Message).Should(ContainSubstring("no cert-manager annotation is allowed for ssl passthough type route"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing console sslEnabled without cert-manager annotations")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.BrokerCert = &serverCertWithPkcs12Name
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypePassThrough,
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue), condition.Message)

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationPassthrough))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing console on edge route without cert-mgr and user secret")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = false
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypeEdge,
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationEdge))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing console on edge route with user secret")
	tlsKeyPemBytes, tlsCrtPemBytes, err := GeneratePemCertificate(nil)
	Expect(err).To(BeNil())
	_, caCrtPemBytes, err := GeneratePemCertificate(func(ca *x509.Certificate) {
		ca.IsCA = true
	})
	Expect(err).To(BeNil())

	userRouteTlsSecretName := brokerCrName + "-wconsj-0-svc-rte.apps.artemiscloud.io-secret"
	userRouteTlsSecret := InstallSecret(userRouteTlsSecretName, defaultNamespace, func(candidate *corev1.Secret) {
		candidate.StringData["tls.key"] = string(tlsKeyPemBytes)
		candidate.StringData["tls.crt"] = string(tlsCrtPemBytes)
		candidate.StringData["ca.crt"] = string(caCrtPemBytes)
	})

	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = false
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypeEdge,
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationEdge))
		g.Expect(route.Spec.TLS.Key).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.Key).Should(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		g.Expect(route.Spec.TLS.Key).Should(ContainSubstring("END RSA PRIVATE KEY"))
		g.Expect(route.Spec.TLS.Certificate).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.Certificate).Should(ContainSubstring("BEGIN CERTIFICATE"))
		g.Expect(route.Spec.TLS.Certificate).Should(ContainSubstring("END CERTIFICATE"))
		g.Expect(route.Spec.TLS.CACertificate).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.CACertificate).Should(ContainSubstring("BEGIN CERTIFICATE"))
		g.Expect(route.Spec.TLS.CACertificate).Should(ContainSubstring("END CERTIFICATE"))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	CleanResource(userRouteTlsSecret, userRouteTlsSecret.Name, userRouteTlsSecret.Namespace)
	/*
	   By("Deploying the broker cr exposing console with cert-manager on edge route")
	   brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

	   		candidate.Name = brokerCrName
	   		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
	   		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
	   			InitialDelaySeconds: 1,
	   			PeriodSeconds:       1,
	   			TimeoutSeconds:      5,
	   		}
	   		candidate.Spec.Console.Expose = true
	   		candidate.Spec.Console.SSLEnabled = false
	   		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
	   			Type: &ingressTypeEdge,
	   			Annotations: []brokerv1beta1.KeyValueType{
	   				{
	   					Key:   "cert-manager.io/issuer-name",
	   					Value: &issuerForIngress,
	   				},
	   				{
	   					Key:   "cert-manager.io/issuer-kind",
	   					Value: &issuerKind,
	   				},
	   				{
	   					Key:   "cert-manager.io/common-name",
	   					Value: &ingressCommonName,
	   				},
	   			},
	   		}
	   	})

	   By("Checking the broker status reflect the truth")

	   	Eventually(func(g Gomega) {
	   		crdRef := types.NamespacedName{
	   			Namespace: brokerCr.Namespace,
	   			Name:      brokerCr.Name,
	   		}
	   		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

	   		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
	   		g.Expect(condition).NotTo(BeNil())
	   		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

	   		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
	   		route := routev1.Route{}
	   		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
	   		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

	   		issuerAnnotation, ok := route.Annotations[routes.CM_ANN_ISSUER]
	   		g.Expect(ok).To(BeTrue())
	   		g.Expect(issuerAnnotation).To(Equal(issuerForIngress))
	   		issuerKindAnnotation, ok := route.Annotations[routes.CM_ANN_ISSUER_KIND]
	   		g.Expect(ok).To(BeTrue())
	   		g.Expect(issuerKindAnnotation).To(Equal(issuerKind))
	   		commonNameAnnotation, ok := route.Annotations["cert-manager.io/common-name"]
	   		g.Expect(ok).To(BeTrue())
	   		g.Expect(commonNameAnnotation).To(Equal(ingressCommonName))
	   		g.Expect(route.Spec.TLS.Key).To(Equal("xyz"))
	   	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	   CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	*/

	By("Deploying the broker cr exposing console on reencrypt route with user secret")
	_, caCrtPemBytes, err = GeneratePemCertificate(func(ca *x509.Certificate) {
		ca.IsCA = true
	})
	Expect(err).To(BeNil())

	userRouteTlsSecretName = brokerCrName + "-wconsj-0-svc-rte.apps.artemiscloud.io-secret"
	userRouteTlsSecret = InstallSecret(userRouteTlsSecretName, defaultNamespace, func(candidate *corev1.Secret) {
		candidate.StringData["tls.key"] = string(tlsKeyPemBytes)
		candidate.StringData["tls.crt"] = string(tlsCrtPemBytes)
		candidate.StringData["ca.crt"] = string(caCrtPemBytes)
	})

	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.BrokerCert = &serverCertWithPkcs12Name
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &routeTypeReencrypt,
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

		rteName := brokerCr.Name + "-wconsj-0-svc-rte"
		route := routev1.Route{}
		rteKey := types.NamespacedName{Name: rteName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, rteKey, &route)).Should(Succeed())

		g.Expect(route.Spec.TLS.Termination).Should(Equal(routev1.TLSTerminationReencrypt))
		g.Expect(route.Spec.TLS.Key).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.Key).Should(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		g.Expect(route.Spec.TLS.Key).Should(ContainSubstring("END RSA PRIVATE KEY"))
		g.Expect(route.Spec.TLS.Certificate).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.Certificate).Should(ContainSubstring("BEGIN CERTIFICATE"))
		g.Expect(route.Spec.TLS.Certificate).Should(ContainSubstring("END CERTIFICATE"))
		g.Expect(route.Spec.TLS.DestinationCACertificate).ShouldNot(BeEmpty())
		g.Expect(route.Spec.TLS.DestinationCACertificate).Should(ContainSubstring("BEGIN CERTIFICATE"))
		g.Expect(route.Spec.TLS.DestinationCACertificate).Should(ContainSubstring("END CERTIFICATE"))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	CleanResource(userRouteTlsSecret, userRouteTlsSecret.Name, userRouteTlsSecret.Namespace)
}

func testIngressTlsTerminateWithCertManager() {
	By("Deploying the broker cr exposing console sslEnabled with invalid cert-manager annotation")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.BrokerCert = &serverCertWithPkcs12Name
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypePassThrough,
			Annotations: []brokerv1beta1.KeyValueType{
				{
					Key:   "cert-manager.io/issuer",
					Value: &issuerForIngress,
				},
			},
		}
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionFalse))
		g.Expect(condition.Message).Should(ContainSubstring("no cert-manager annotation is allowed for ssl passthough type ingress"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing console with cert-manager")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = false
		candidate.Spec.Console.TlsTermination = brokerv1beta1.TlsTerminationType{
			Type: &ingressTypeEdge,
			Annotations: []brokerv1beta1.KeyValueType{
				{
					Key:   "cert-manager.io/issuer",
					Value: &issuerForIngress,
				},
				{
					Key:   "cert-manager.io/common-name",
					Value: &ingressCommonName,
				},
			},
		}
	})
	By("Checking the broker status reflect the truth")
	ingSecret := corev1.Secret{}
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))

		ingName := brokerCr.Name + "-wconsj-0-svc-ing"
		ingress := netv1.Ingress{}
		ingKey := types.NamespacedName{Name: ingName, Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, ingKey, &ingress)).Should(Succeed())

		issuerAnnotation, ok := ingress.Annotations["cert-manager.io/issuer"]
		g.Expect(ok).To(BeTrue())
		g.Expect(issuerAnnotation).To(Equal(issuerForIngress))
		commonNameAnnotation, ok := ingress.Annotations["cert-manager.io/common-name"]
		g.Expect(ok).To(BeTrue())
		g.Expect(commonNameAnnotation).To(Equal(ingressCommonName))
		hostName := brokerCr.Name + "-wconsj-0-svc-ing.apps.artemiscloud.io"
		ingSecretKey := types.NamespacedName{Name: hostName + "-secret", Namespace: defaultNamespace}
		g.Expect(k8sClient.Get(ctx, ingSecretKey, &ingSecret)).Should(Succeed())
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
	CleanResource(&ingSecret, ingSecret.Name, ingSecret.Namespace)
}

func checkReadPodStatus(podName string, crName string, g Gomega) {
	curlUrl := "https://" + podName + ":8161/console/jolokia/read/org.apache.activemq.artemis:broker=\"amq-broker\"/Status"
	command := []string{"curl", "-k", "-u", "testuser:testpassword", curlUrl}

	result := ExecOnPod(podName, crName, defaultNamespace, command, g)
	var rootMap map[string]any
	g.Expect(json.Unmarshal([]byte(result), &rootMap)).To(Succeed())
	value := rootMap["value"].(string)
	var valueMap map[string]any
	g.Expect(json.Unmarshal([]byte(value), &valueMap)).To(Succeed())
	serverInfo := valueMap["server"].(map[string]any)
	serverState := serverInfo["state"].(string)
	g.Expect(serverState).To(Equal("STARTED"))
}

func checkMessagingInPod(podName string, crName string, portNumber string, trustStoreLoc string, trustStorePass string, g Gomega) {
	tcpUrl := "tcp://" + podName + ":" + portNumber + "?sslEnabled=true&trustStorePath=" + trustStoreLoc + "&trustStorePassword=" + trustStorePass
	sendCommand := []string{"amq-broker/bin/artemis", "producer", "--user", "testuser", "--password", "testpassword", "--url", tcpUrl, "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}
	result := ExecOnPod(podName, crName, defaultNamespace, sendCommand, g)
	g.Expect(result).To(ContainSubstring("Produced: 1 messages"))
	receiveCommand := []string{"amq-broker/bin/artemis", "consumer", "--user", "testuser", "--password", "testpassword", "--url", tcpUrl, "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}
	result = ExecOnPod(podName, crName, defaultNamespace, receiveCommand, g)
	g.Expect(result).To(ContainSubstring("Consumed: 1 messages"))
}

func testCertWithNoKeystoreConfigured(certLoc string) {
	By("Deploying the broker cr")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.UseClientAuth = false
		candidate.Spec.Console.BrokerCert = &certLoc
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionFalse))
		g.Expect(condition.Message).Should(ContainSubstring("doesn't have keystore options configured"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing acceptor ssl")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Acceptors = []brokerv1beta1.AcceptorType{{
			Name:       "new-acceptor",
			Port:       62666,
			Protocols:  "all",
			SSLSecret:  "cert-secret",
			VerifyHost: true,
			SNIHost:    candidate.Name + "-ss-0",
			Expose:     true,
			SSLEnabled: true,
			BrokerCert: &certLoc,
		}}
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionFalse))
		g.Expect(condition.Message).Should(ContainSubstring("doesn't have keystore options configured"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
}

func testCertWithKeystoreConfigured(certLoc string, storePassword string) {
	By("Deploying the broker cr")
	brokerCr, createdBrokerCr := DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.RequireLogin = true
		candidate.Spec.AdminUser = adminUser
		candidate.Spec.AdminPassword = adminPassword
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.UseClientAuth = false
		candidate.Spec.Console.BrokerCert = &certLoc
	})
	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdKey := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdKey, createdBrokerCr)).Should(Succeed())
		g.Expect(len(createdBrokerCr.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("checking jolokia access")
	pod0Name := createdBrokerCr.Name + "-ss-0"
	Eventually(func(g Gomega) {
		checkReadPodStatus(pod0Name, createdBrokerCr.Name, g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing acceptor ssl")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.RequireLogin = true
		candidate.Spec.AdminUser = adminUser
		candidate.Spec.AdminPassword = adminPassword
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Acceptors = []brokerv1beta1.AcceptorType{{
			Name:       "new-acceptor",
			Port:       62666,
			Protocols:  "all",
			SSLSecret:  "acceptor-ssl-secret",
			VerifyHost: true,
			SNIHost:    candidate.Name + "-ss-0",
			Expose:     true,
			SSLEnabled: true,
			BrokerCert: &certLoc,
		}}
		candidate.Spec.Connectors = []brokerv1beta1.ConnectorType{
			{
				Name:             "new-connector",
				Host:             candidate.Name + "-ss-0",
				Port:             62666,
				EnabledProtocols: "all",
				SSLEnabled:       true,
				Expose:           true,
				SSLSecret:        "connector-ssl-secret",
				BrokerCert:       &certLoc,
			},
		}

		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.UseClientAuth = false
		candidate.Spec.Console.BrokerCert = &certLoc
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())
		g.Expect(len(createdBrokerCr.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("Checking acceptor handling request")
	Eventually(func(g Gomega) {
		checkMessagingInPod(pod0Name, createdBrokerCr.Name, "62666", "/etc/acceptor-ssl-secret-volume/client.ts", storePassword, g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("Checking connector having correct ssl parameters")
	Eventually(func(g Gomega) {
		command := []string{"sh", "-c", "echo $AMQ_CONNECTORS"}

		result := ExecOnPod(pod0Name, createdBrokerCr.Name, defaultNamespace, command, g)
		g.Expect(result).To(ContainSubstring("new-connector"))
		g.Expect(result).To(ContainSubstring("keyStorePassword=" + storePassword))
		g.Expect(result).To(ContainSubstring("trustStorePassword=" + storePassword))
		g.Expect(result).To(ContainSubstring("sslEnabled=true"))
		g.Expect(result).To(ContainSubstring("keyStorePath=\\/etc\\/connector-ssl-secret-volume\\/broker.ks"))
		g.Expect(result).To(ContainSubstring("trustStorePath=\\/etc\\/connector-ssl-secret-volume\\/client.ts"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
}
