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
	"encoding/json"
	"os"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
)

var _ = Describe("artemis controller with cert manager test", Label("controller-cert-mgr-test"), func() {
	var installedCertManager bool = false

	BeforeEach(func() {
		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
			//if cert manager is not installed, install it
			if !CertManagerInstalled() {
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
	})
})

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
		g.Expect(condition.Status).Should(Equal(metav1.ConditionUnknown))
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
		g.Expect(condition.Status).Should(Equal(metav1.ConditionUnknown))
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
