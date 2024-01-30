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
	"fmt"
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

	rootIssuerName       = "root-issuer"
	rootCertName         = "root-cert"
	rootCertNamespce     = "cert-manager"
	rootCertSecretName   = "artemis-root-cert-secret"
	caIssuerName         = "broker-ca-issuer"
	caPemTrustStoreName  = "ca-truststore.pem"
	caTrustStorePassword = "changeit"
)

var (
	adminUser     = "testuser"
	adminPassword = "testpassword"

	serverCert   = "server-cert"
	rootIssuer   = &cmv1.ClusterIssuer{}
	rootCert     = &cmv1.Certificate{}
	caIssuer     = &cmv1.ClusterIssuer{}
	caBundleName = "ca-bundle"
)

var _ = Describe("artemis controller with cert manager test", Label("controller-cert-mgr-test"), func() {
	var installedCertManager bool = false

	BeforeEach(func() {
		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
			//if cert manager/trust manager is not installed, install it
			if !CertManagerInstalled() {
				Expect(InstallCertManager()).To(Succeed())
				installedCertManager = true
			}

			rootIssuer = InstallClusteredIssuer(rootIssuerName, nil)

			rootCert = InstallCert(rootCertName, rootCertNamespce, func(candidate *cmv1.Certificate) {
				candidate.Spec.IsCA = true
				candidate.Spec.CommonName = "artemis.root.ca"
				candidate.Spec.SecretName = rootCertSecretName
				candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
					Name: rootIssuer.Name,
					Kind: "ClusterIssuer",
				}
			})

			caIssuer = InstallClusteredIssuer(caIssuerName, func(candidate *cmv1.ClusterIssuer) {
				candidate.Spec.SelfSigned = nil
				candidate.Spec.CA = &cmv1.CAIssuer{
					SecretName: rootCertSecretName,
				}
			})
			InstallCaBundle(caBundleName, rootCertSecretName, caPemTrustStoreName)
		}
	})

	AfterEach(func() {
		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
			UnInstallCaBundle(caBundleName)
			UninstallClusteredIssuer(caIssuerName)
			UninstallCert(rootCert.Name, rootCert.Namespace)
			UninstallClusteredIssuer(rootIssuerName)

			if installedCertManager {
				Expect(UninstallCertManager()).To(Succeed())
				installedCertManager = false
			}
		}
	})

	Describe("cert manager integration test", func() {
		Context("tls exposure with cert manager", func() {
			BeforeEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					InstallCert(serverCert, defaultNamespace, func(candidate *cmv1.Certificate) {
						candidate.Spec.DNSNames = []string{brokerCrName + "-ss-0"}
						candidate.Spec.IssuerRef = cmmetav1.ObjectReference{
							Name: caIssuer.Name,
							Kind: "ClusterIssuer",
						}
					})
				}
			})
			AfterEach(func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					UninstallCert(serverCert, defaultNamespace)
				}
			})
			It("test configured with cert and ca bundle", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					testConfiguredWithCertAndBundle(serverCert+"-secret", caBundleName)
				}
			})
			It("test configured with cert and no ca bundle", func() {
				if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
					fmt.Println("Not implemented")
					//testConfiguredWithCertNoBundle(serverCertNoKeystoreName + "-secret")
				}
			})
		})
	})
})

type ConnectorConfig struct {
	Name    string
	Factory string
	Params  map[string]string
}

func getConnectorConfig(podName string, crName string, connectorName string, g Gomega) *ConnectorConfig {
	curlUrl := "https://" + podName + ":8161/console/jolokia/read/org.apache.activemq.artemis:broker=\"amq-broker\"/Connectors"
	command := []string{"curl", "-k", "-s", "-u", "testuser:testpassword", curlUrl}

	result := ExecOnPod(podName, crName, defaultNamespace, command, g)
	var rootMap map[string]any
	g.Expect(json.Unmarshal([]byte(result), &rootMap)).To(Succeed())
	connectors := rootMap["value"].([]ConnectorConfig)
	for _, v := range connectors {
		if v.Name == connectorName {
			return &v
		}
	}
	return nil
}

func checkReadPodStatus(podName string, crName string, g Gomega) {
	curlUrl := "https://" + podName + ":8161/console/jolokia/read/org.apache.activemq.artemis:broker=\"amq-broker\"/Status"
	command := []string{"curl", "-k", "-s", "-u", "testuser:testpassword", curlUrl}

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

func checkMessagingInPod(podName string, crName string, portNumber string, trustStoreLoc string, g Gomega) {
	tcpUrl := "tcp://" + podName + ":" + portNumber + "?sslEnabled=true&trustStorePath=" + trustStoreLoc + "&trustStoreType=PEM"
	sendCommand := []string{"amq-broker/bin/artemis", "producer", "--user", "testuser", "--password", "testpassword", "--url", tcpUrl, "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}
	result := ExecOnPod(podName, crName, defaultNamespace, sendCommand, g)
	g.Expect(result).To(ContainSubstring("Produced: 1 messages"))
	receiveCommand := []string{"amq-broker/bin/artemis", "consumer", "--user", "testuser", "--password", "testpassword", "--url", tcpUrl, "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}
	result = ExecOnPod(podName, crName, defaultNamespace, receiveCommand, g)
	g.Expect(result).To(ContainSubstring("Consumed: 1 messages"))
}

func testConfiguredWithCertAndBundle(certSecret string, caSecret string) {
	// it should use PEM store type
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
		candidate.Spec.Console.SSLSecret = certSecret
		candidate.Spec.Console.KeyStoreType = "PEM"
		candidate.Spec.Console.TrustSecret = &caSecret
		candidate.Spec.Console.TrustStoreType = "PEM"
	})
	pod0Name := createdBrokerCr.Name + "-ss-0"
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
		checkReadPodStatus(pod0Name, createdBrokerCr.Name, g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

	By("Deploying the broker cr exposing acceptor ssl and connector ssl")
	brokerCr, createdBrokerCr = DeployCustomBroker(defaultNamespace, func(candidate *brokerv1beta1.ActiveMQArtemis) {

		candidate.Name = brokerCrName
		candidate.Spec.DeploymentPlan.Size = common.Int32ToPtr(1)
		candidate.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       1,
			TimeoutSeconds:      5,
		}
		candidate.Spec.Acceptors = []brokerv1beta1.AcceptorType{{
			Name:           "new-acceptor",
			Port:           62666,
			Protocols:      "all",
			Expose:         true,
			SSLEnabled:     true,
			SSLSecret:      certSecret,
			TrustSecret:    &caSecret,
			KeyStoreType:   "PEM",
			TrustStoreType: "PEM",
		}}
		candidate.Spec.Connectors = []brokerv1beta1.ConnectorType{{
			Name:           "new-connecor",
			Port:           62666,
			Expose:         true,
			SSLEnabled:     true,
			SSLSecret:      certSecret,
			TrustSecret:    &caSecret,
			KeyStoreType:   "PEM",
			TrustStoreType: "PEM",
		}}
	})

	crdRef := types.NamespacedName{
		Namespace: brokerCr.Namespace,
		Name:      brokerCr.Name,
	}

	By("checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())

		condition := meta.FindStatusCondition(createdBrokerCr.Status.Conditions, brokerv1beta1.DeployedConditionType)
		g.Expect(condition).NotTo(BeNil())
		g.Expect(condition.Status).Should(Equal(metav1.ConditionTrue))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("checking the broker message send and receive")
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())
		checkMessagingInPod(pod0Name, createdBrokerCr.Name, "62666", "/etc/"+caBundleName+"-volume/"+caPemTrustStoreName, g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("checking connector parameters")
	//","trustStoreType":"PEM","keyStorePath":"server-cert-pkcs12-secret.pemcfg"
	Eventually(func(g Gomega) {
		connectorCfg := getConnectorConfig(pod0Name, createdBrokerCr.Name, "new-connector", g)
		g.Expect(connectorCfg).NotTo(BeNil())
		g.Expect(connectorCfg.Factory).To(Equal("org.apache.activemq.artemis.core.remoting.impl.netty.NettyConnectorFactory"))
		g.Expect(connectorCfg.Params["keyStoreType"]).To(Equal("PEMCFG"))
		g.Expect(connectorCfg.Params["port"]).To(Equal("62666"))
		g.Expect(connectorCfg.Params["sslEnabled"]).To(Equal("true"))
		g.Expect(connectorCfg.Params["host"]).To(Equal(pod0Name))
		g.Expect(connectorCfg.Params["trustStorePath"]).To(Equal("/etc/" + caBundleName + "-volume/" + caPemTrustStoreName))
		g.Expect(connectorCfg.Params["trustStoreType"]).To(Equal("PEM"))
		g.Expect(connectorCfg.Params["keyStorePath"]).To(Equal(certSecret + ".pemcfg"))

	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
}

func testCertWithKeystoreConfiguredPem(certLoc string) {
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
		candidate.Spec.Console.SSLSecret = certLoc + "-secret"
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
			Name:           "new-acceptor",
			Port:           62666,
			Protocols:      "all",
			VerifyHost:     true,
			SNIHost:        candidate.Name + "-ss-0",
			Expose:         true,
			SSLEnabled:     true,
			KeyStoreType:   "PEM",
			TrustStoreType: "PEM",
			SSLSecret:      certLoc + "-secret",
		}}
		candidate.Spec.Connectors = []brokerv1beta1.ConnectorType{
			{
				Name:             "new-connector",
				Host:             candidate.Name + "-ss-0",
				Port:             62666,
				EnabledProtocols: "all",
				SSLEnabled:       true,
				Expose:           true,
				SSLSecret:        certLoc + "-secret",
				TrustStoreType:   "PEM",
				KeyStoreType:     "PEM",
			},
		}

		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.UseClientAuth = false
		candidate.Spec.Console.SSLSecret = certLoc + "-secret"
	})

	By("Checking the broker status reflect the truth")
	Eventually(func(g Gomega) {
		crdRef := types.NamespacedName{
			Namespace: brokerCr.Namespace,
			Name:      brokerCr.Name,
		}
		g.Expect(k8sClient.Get(ctx, crdRef, createdBrokerCr)).Should(Succeed())
		podLog := LogsOfPod(pod0Name, brokerCr.Name, defaultNamespace, g)
		g.Expect(len(createdBrokerCr.Status.PodStatus.Ready)).Should(BeEquivalentTo(1), podLog)
	}, existingClusterTimeout, existingClusterInterval*2).Should(Succeed())

	By("Checking acceptor handling request")
	Eventually(func(g Gomega) {
		// here use pem store to check message sending and receiving
		g.Expect(false).To(BeTrue())
		//checkMessagingInPod(pod0Name, createdBrokerCr.Name, "62666", "/etc/"+certLoc+"-secret-volume/truststore.p12", storePassword, g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("Checking connector having correct ssl parameters")
	Eventually(func(g Gomega) {
		command := []string{"sh", "-c", "echo $AMQ_CONNECTORS"}

		result := ExecOnPod(pod0Name, createdBrokerCr.Name, defaultNamespace, command, g)
		g.Expect(result).To(ContainSubstring("new-connector"))
		g.Expect(result).NotTo(ContainSubstring("keyStorePassword="))
		g.Expect(result).NotTo(ContainSubstring("trustStorePassword="))
		g.Expect(result).To(ContainSubstring("sslEnabled=true"))
		g.Expect(result).To(ContainSubstring("keyStorePath=\\/etc\\/" + certLoc + "-secret-volume\\/keystore.pem"))
		g.Expect(result).To(ContainSubstring("trustStorePath=\\/etc\\/" + certLoc + "-secret-volume\\/truststore.pem"))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)
}

func testCertWithCaBundleConfiguredPem(certLoc string) {
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
		candidate.Spec.Console.SSLSecret = certLoc + "-secret"
		candidate.Spec.Console.TrustSecret = &caBundleName
		// console how to use pem?
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
			Name:           "new-acceptor",
			Port:           62666,
			Protocols:      "all",
			VerifyHost:     true,
			SNIHost:        candidate.Name + "-ss-0",
			Expose:         true,
			SSLEnabled:     true,
			SSLSecret:      certLoc + "-secret",
			TrustSecret:    &caBundleName,
			KeyStoreType:   "PEM",
			TrustStoreType: "PEM",
		}}
		candidate.Spec.Connectors = []brokerv1beta1.ConnectorType{
			{
				Name:             "new-connector",
				Host:             candidate.Name + "-ss-0",
				Port:             62666,
				EnabledProtocols: "all",
				SSLEnabled:       true,
				Expose:           true,
				SSLSecret:        certLoc + "-secret",
				TrustSecret:      &caBundleName,
				KeyStoreType:     "PEM",
				TrustStoreType:   "PEM",
			},
		}

		candidate.Spec.Console.Expose = true
		candidate.Spec.Console.SSLEnabled = true
		candidate.Spec.Console.UseClientAuth = false
		candidate.Spec.Console.SSLSecret = certLoc + "-secret"
		candidate.Spec.Console.TrustSecret = &caBundleName
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
		//check message sending/receiveing using pem
		Expect(false).To(BeTrue())
		//checkMessagingInPod(pod0Name, createdBrokerCr.Name, "62666", "/etc/"+caBundleName+"-volume/"+caPemTrustStoreName, "changeit", g)
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	By("Checking connector having correct ssl parameters")
	Eventually(func(g Gomega) {
		command := []string{"sh", "-c", "echo $AMQ_CONNECTORS"}

		result := ExecOnPod(pod0Name, createdBrokerCr.Name, defaultNamespace, command, g)
		g.Expect(result).To(ContainSubstring("new-connector"))
		g.Expect(result).NotTo(ContainSubstring("keyStorePassword="))
		g.Expect(result).NotTo(ContainSubstring("trustStorePassword="))
		g.Expect(result).To(ContainSubstring("sslEnabled=true"))
		g.Expect(result).To(ContainSubstring("keyStorePath=\\/etc\\/" + certLoc + "-secret-volume\\/keystore.pem"))
		g.Expect(result).To(ContainSubstring("trustStorePath=\\/etc\\/" + caBundleName + "-volume\\/" + caPemTrustStoreName))
	}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

	CleanResource(createdBrokerCr, brokerCr.Name, createdBrokerCr.Namespace)

}
