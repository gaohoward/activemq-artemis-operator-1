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

/*
As usual, we start with the necessary imports. We also define some utility variables.
*/
package controllers

import (
	"container/list"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	brokerv2alpha4 "github.com/artemiscloud/activemq-artemis-operator/api/v2alpha4"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/environments"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/secrets"
	ss "github.com/artemiscloud/activemq-artemis-operator/pkg/resources/statefulsets"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"

	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/artemiscloud/activemq-artemis-operator/version"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/cr2jinja2"

	"github.com/Azure/go-amqp"
)

//Uncomment this and the "test" import if you want to debug this set of tests
//func TestArtemisController(t *testing.T) {
//	RegisterFailHandler(Fail)
//	RunSpecs(t, "Artemis Controller Suite")
//}

var boolFalse = false
var boolTrue = true

var _ = Describe("Simple test", Label("simple"), func() {

	Context("security test", func() {
		It("deploy broker with custom readiness probe that relies on security", func() {
			By("Creating broker")
			ctx := context.Background()
			crd := generateArtemisSpec(defaultNamespace)

			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "secret"
			crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
				Handler: corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/home/jboss/amq-broker/bin/artemis",
							"check queue",
							"--name readinessqueue",
							"--produce 1",
							"--consume 1",
							"--silent",
							"--user batch-user",
							"--password mysecret",
						},
					},
				},
				InitialDelaySeconds: 30,
				TimeoutSeconds:      5,
			}
			crd.Spec.DeploymentPlan.Size = 1
			crd.Spec.DeploymentPlan.Clustered = &boolFalse
			crd.Spec.DeploymentPlan.RequireLogin = boolTrue
			crd.Spec.DeploymentPlan.JolokiaAgentEnabled = false
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:      "internal",
					Protocols: "core,openwire",
					Port:      61616,
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, createdSs)
				return err == nil
			}, timeout, interval).Should(Equal(true))

			By("Deploying security")
			secCrd, createdSecCrd := DeploySecurity("", defaultNamespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
			})

			fmt.Printf("ori: %v created: %v", secCrd, createdSecCrd)

			By("Checking that Stateful Set is updated " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: defaultNamespace}
				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false
				}

				initContainer := createdSs.Spec.Template.Spec.InitContainers[0]
				securityFound := false
				for _, argStr := range initContainer.Args {
					fmt.Println("Checking arg str === " + argStr)
					if strings.Contains(argStr, "/opt/amq-broker/script/cfg/config-security.sh") {
						securityFound = true
						break
					}
				}
				return securityFound
			}, timeout, interval).Should(BeTrue())
			fmt.Printf("scrCrd %v created %v", secCrd, createdSecCrd)

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd)).To(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, defaultNamespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdSecCrd)).To(Succeed())
			Eventually(func() bool {
				return checkCrdDeleted(secCrd.Name, defaultNamespace, createdSecCrd)
			}, timeout, interval).Should(BeTrue())

		})
	})
})

var _ = Describe("artemis controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		namespace = defaultNamespace
	)

	// see what has changed from the controllers perspective
	var wi watch.Interface
	BeforeEach(func() {

		var err error
		wc, err := client.NewWithWatch(restConfig, client.Options{})
		if err != nil {
			fmt.Printf("Err on watch client:  %v\n", err)
			return
		}

		// see what changed
		wi, err = wc.Watch(ctx, &brokerv1beta1.ActiveMQArtemisList{}, &client.ListOptions{})
		if err != nil {
			fmt.Printf("Err on watch:  %v\n", err)
			return
		}
		go func() {
			for event := range wi.ResultChan() {
				switch co := event.Object.(type) {
				case client.Object:
					if verobse {
						fmt.Printf("%v ActiveMQArtemisList CRD: ResourceVersion: %v Generation: %v, OR: %v\n", event.Type, co.GetResourceVersion(), co.GetGeneration(), co.GetOwnerReferences())
						fmt.Printf("Object: %v\n", event.Object)
					}
				}
			}
		}()
	})

	AfterEach(func() {
		if wi != nil {
			wi.Stop()
		}
	})

	Context("New address settings options", func() {
		It("Deploy broker with new address settings", func() {

			By("By creating a crd with address settings in spec")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			// add address settings, to an existing crd
			ma := "merge_all"
			dlqabc := "dlqabc"
			maxSize := "10m"
			maxMessages := int64(5000)
			redistributionDelay := int32(5000)
			configDeleteDiverts := "OFF"

			crd.Spec.AddressSettings = brokerv1beta1.AddressSettingsType{
				ApplyRule: &ma,
				AddressSetting: []brokerv1beta1.AddressSettingType{
					{
						Match:               "abc#",
						DeadLetterAddress:   &dlqabc,
						MaxSizeBytes:        &maxSize,
						MaxSizeMessages:     &maxMessages,
						ConfigDeleteDiverts: &configDeleteDiverts,
						RedistributionDelay: &redistributionDelay,
					},
					{
						Match:               "#",
						DeadLetterAddress:   &dlqabc,
						MaxSizeBytes:        &maxSize,
						MaxSizeMessages:     &maxMessages,
						ConfigDeleteDiverts: &configDeleteDiverts,
						RedistributionDelay: &redistributionDelay,
					},
				},
			}
			crd.Spec.DeploymentPlan.Size = 1

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

			By("tracking the yaconfig init command with user_address_settings and verifying new options are in")
			key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
			var initArgsString string
			Eventually(func() bool {
				createdSs := &appsv1.StatefulSet{}

				if k8sClient.Get(ctx, key, createdSs) != nil {
					return false
				}

				initArgsString = strings.Join(createdSs.Spec.Template.Spec.InitContainers[0].Args, ",")
				if !strings.Contains(initArgsString, "max_size_messages: 5000") {
					return false
				}

				value := cr2jinja2.GetUniqueShellSafeSubstution(configDeleteDiverts)
				fullString := "config_delete_diverts: " + value
				return strings.Contains(initArgsString, fullString)

			}, timeout, interval).Should(BeTrue())

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("verying started")
				brokerKey := types.NamespacedName{Name: createdCrd.Name, Namespace: createdCrd.Namespace}
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}

			// cleanup
			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("Versions Test", func() {
		It("default image to use latest", func() {
			crd := generateArtemisSpec(namespace)
			imageToUse := determineImageToUse(&crd, "Kubernetes")
			Expect(imageToUse).To(Equal(version.LatestKubeImage), "actual", imageToUse)

			imageToUse = determineImageToUse(&crd, "Init")
			Expect(imageToUse).To(Equal(version.LatestInitImage), "actual", imageToUse)
			brokerCr := generateArtemisSpec(namespace)
			compactVersionToUse := determineCompactVersionToUse(&brokerCr)
			yacfgProfileVersion = version.YacfgProfileVersionFromFullVersion[version.FullVersionFromCompactVersion[compactVersionToUse]]
			Expect(yacfgProfileVersion).To(Equal("2.21.0"))
		})
	})

	Context("ClientId autoshard Test", func() {

		produceMessage := func(url string, clientId string, linkAddress string, messageId int, g Gomega) {

			client, err := amqp.Dial(url, amqp.ConnContainerID(clientId), amqp.ConnSASLPlain("dummy-user", "dummy-pass"))

			g.Expect(err).Should(BeNil())
			g.Expect(client).ShouldNot(BeNil())
			defer client.Close()

			session, err := client.NewSession()

			g.Expect(err).Should(BeNil())
			g.Expect(session).ShouldNot(BeNil())

			sender, err := session.NewSender(
				amqp.LinkTargetAddress(linkAddress),
			)
			g.Expect(err).Should(BeNil())
			ctx, cancel := context.WithTimeout(ctx, 4*time.Second)

			defer sender.Close(ctx)
			defer cancel()

			msg := amqp.NewMessage([]byte("Hello! from:" + clientId))
			msg.ApplicationProperties = make(map[string]interface{})
			msg.ApplicationProperties["MessageID"] = messageId
			msg.ApplicationProperties["ClientID"] = clientId

			err = sender.Send(ctx, msg)
			g.Expect(err).Should(BeNil())
		}

		consumeMatchingMessage := func(url string, linkAddress string, receivedTracker map[string]*list.List, g Gomega) {

			client, err := amqp.Dial(url, amqp.ConnSASLPlain("dummy-user", "dummy-pass"))
			g.Expect(err).Should(BeNil())
			g.Expect(client).ShouldNot(BeNil())

			defer client.Close()

			session, err := client.NewSession()
			g.Expect(err).Should(BeNil())
			g.Expect(session).ShouldNot(BeNil())

			receiver, err := session.NewReceiver(
				amqp.LinkSourceAddress(linkAddress),
				amqp.LinkSourceCapabilities("queue"),
				amqp.LinkCredit(50),
			)
			g.Expect(err).Should(BeNil())
			g.Expect(receiver).ShouldNot(BeNil())

			ctx, cancel := context.WithTimeout(ctx, 600*time.Millisecond)
			defer receiver.Close(ctx)
			defer cancel()

			// Receive messages till error or nil
			for {
				msg, err := receiver.Receive(ctx)

				if err != nil || msg == nil {
					break
				}

				g.Expect(err).Should(BeNil())
				g.Expect(msg).ShouldNot(BeNil())

				var senderClientId = msg.ApplicationProperties["ClientID"].(string)
				receivedTracker[senderClientId].PushBack(msg.ApplicationProperties["MessageID"])

				err = receiver.AcceptMessage(ctx, msg)
				g.Expect(err).Should(BeNil())

			}
		}

		It("deploy 2 with clientID auto sharding", func() {

			isClusteredBoolean := false
			NOT := false
			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       5,
				TimeoutSeconds:      5,
			}
			crd.Spec.DeploymentPlan.LivenessProbe = &corev1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       5,
				TimeoutSeconds:      5,
			}

			crd.Spec.DeploymentPlan.Size = 2
			crd.Spec.DeploymentPlan.Clustered = &isClusteredBoolean
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{{
				Name:                "tcp",
				Port:                62616,
				Expose:              false,
				BindToAllInterfaces: &NOT,
			}}

			linkAddress := "LB.TESTQ"

			crd.Spec.BrokerProperties = []string{
				"connectionRouters.autoShard.keyType=CLIENT_ID",
				"connectionRouters.autoShard.localTargetFilter=NULL|${STATEFUL_SET_ORDINAL}|-${STATEFUL_SET_ORDINAL}",
				"connectionRouters.autoShard.policyConfiguration=CONSISTENT_HASH_MODULO",
				"connectionRouters.autoShard.policyConfiguration.properties.MODULO=2",
				"acceptorConfigurations.tcp.params.router=autoShard",           // matching spec.acceptor
				"addressesSettings.\"LB.#\".defaultAddressRoutingType=ANYCAST", // b/c cannot set linkTargetCapabilities from amqp client
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			crdNsName := types.NamespacedName{
				Name:      crd.Name,
				Namespace: namespace,
			}
			deployedCrd := &brokerv1beta1.ActiveMQArtemis{}

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("Finding cluster host")
				baseUrl, err := url.Parse(testEnv.Config.Host)
				Expect(err).Should(BeNil())

				ipAddressNet, err := net.LookupIP(baseUrl.Hostname())
				Expect(err).Should(BeNil())
				ipAddress := ipAddressNet[0].String()

				By("verying started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, crdNsName, deployedCrd)).Should(Succeed())
					g.Expect(len(deployedCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(2))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("exposing ss-0 via NodePort")
				ss0Service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: crd.Name + "-nodeport-0", Namespace: namespace},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"ActiveMQArtemis":                    crd.Name,
							"statefulset.kubernetes.io/pod-name": namer.CrToSS(crd.Name) + "-0",
						},
						Type: "NodePort",
						Ports: []corev1.ServicePort{
							{
								Port:       61617,
								TargetPort: intstr.FromInt(62616),
							},
						},
						ExternalIPs: []string{ipAddress},
					},
				}
				Expect(k8sClient.Create(ctx, ss0Service)).Should(Succeed())

				By("exposing ss-1 via NodePort")
				ss1Service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: crd.Name + "-nodeport-1", Namespace: namespace},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"ActiveMQArtemis":                    crd.Name,
							"statefulset.kubernetes.io/pod-name": namer.CrToSS(crd.Name) + "-1",
						},
						Type: "NodePort",
						Ports: []corev1.ServicePort{
							{
								Port:       61618,
								TargetPort: intstr.FromInt(62616),
							},
						},
						ExternalIPs: []string{string(ipAddress)},
					},
				}
				Expect(k8sClient.Create(ctx, ss1Service)).Should(Succeed())

				urls := []string{"amqp://" + baseUrl.Hostname() + ":61617", "amqp://" + baseUrl.Hostname() + ":61618"}

				By("verify partition by sending eventually... expect auto-shard error if we get the wrong broker")
				numberOfMessagesToSendPerClientId := 5
				sentMessageSequenceId := 0
				urlBalancerCounter := 1 // round robin over urls if len(urls) > 0
				clientIds := []string{"W", "ONE", "TWO", "THREE", "FOUR", "0299s-99", "D-0301-c-e-SomeHostBla"}
				for _, id := range clientIds {
					// send messages on new connection with each clientId
					for i := 0; i < numberOfMessagesToSendPerClientId; i++ {
						Eventually(func(g Gomega) {
							urlBalancerCounter++
							produceMessage(urls[urlBalancerCounter%len(urls)], id, linkAddress, sentMessageSequenceId+1, g)
							sentMessageSequenceId++ // on success
						}, timeout*4, interval).Should(Succeed())
					}
				}

				Expect(len(clientIds) * numberOfMessagesToSendPerClientId).Should(BeEquivalentTo(sentMessageSequenceId))

				By("verify partition by consuming messages from each broker")
				receivedIdTracker := map[string]*list.List{}
				for _, id := range clientIds {
					receivedIdTracker[id] = list.New()
				}
				for _, url := range urls {
					Eventually(func(g Gomega) {
						consumeMatchingMessage(url, linkAddress, receivedIdTracker, g)
					}).Should(Succeed())
				}

				for _, list := range receivedIdTracker {
					Expect(list.Len()).Should(BeEquivalentTo(5))

					By("verifying received in order per clientId, ie: producer was sharded nicely")
					receivedIdTracker := list.Front().Value.(int64)
					for e := list.Front(); e != nil; e = e.Next() {
						if e.Value == receivedIdTracker {
							receivedIdTracker++
						}
					}

					By("verifying all in order")
					Expect(receivedIdTracker - list.Front().Value.(int64)).Should(BeEquivalentTo(numberOfMessagesToSendPerClientId))
				}

				Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, ss0Service)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, ss1Service)).Should(Succeed())

			}
		})
	})

	Context("SS delete recreate Test", func() {
		It("deploy, delete ss, verify", func() {

			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Size = 1

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			key := types.NamespacedName{
				Name:      namer.CrToSS(crd.Name),
				Namespace: namespace,
			}

			currentSS := &appsv1.StatefulSet{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, currentSS)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			brokerKey := types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: namespace}

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("manually deleting SS quickly")
			ssVersion := currentSS.ResourceVersion
			Expect(k8sClient.Delete(ctx, currentSS)).Should(Succeed())

			By("checking new version created")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, key, currentSS)).Should(Succeed())
				g.Expect(currentSS.ResourceVersion).ShouldNot(Equal(ssVersion))
				ssVersion = currentSS.ResourceVersion
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("verying started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("manually deleting SS again once ready")
				Expect(k8sClient.Delete(ctx, currentSS)).Should(Succeed())

				By("checking new version created again")
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, key, currentSS)).Should(Succeed())
					g.Expect(currentSS.ResourceVersion).ShouldNot(Equal(ssVersion))
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}

			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
		})
	})

	Context("PVC gc test", func() {
		It("deploy, verify, undeploy, verify", func() {

			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Size = 1
			crd.Spec.DeploymentPlan.PersistenceEnabled = true

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("Deploying the CRD " + crd.ObjectMeta.Name)
				Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

				createdCrd := &brokerv1beta1.ActiveMQArtemis{}
				brokerKey := types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: namespace}

				By("verifing started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("finding PVC")
				pvcKey := types.NamespacedName{Namespace: namespace, Name: crd.Name + "-" + namer.CrToSS(crd.Name) + "-0"}
				pvc := &corev1.PersistentVolumeClaim{}
				Expect(k8sClient.Get(ctx, pvcKey, pvc)).Should(Succeed())
				Expect(len(pvc.OwnerReferences)).Should(BeEquivalentTo(1))

				By("undeploying CR")
				Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())

				By("now not finding PVC b/c  it has been gc'ed")
				Eventually(func(g Gomega) {
					g.Expect(k8sClient.Get(ctx, pvcKey, &corev1.PersistentVolumeClaim{})).ShouldNot(Succeed())
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())
			}
		})
	})

	Context("Tolerations Existing Cluster", func() {
		It("Toleration of artemis", func() {

			By("Creating a crd with tolerations")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Size = 1
			crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       5,
			}
			crd.Spec.DeploymentPlan.Tolerations = []corev1.Toleration{
				{
					Key:    "artemis",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			brokerKey := types.NamespacedName{Name: crd.Name, Namespace: crd.Namespace}
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			// some required services on crc get evicted which invalidates this test of taints
			isOpenshift, err := environments.DetectOpenshift()
			Expect(err).Should(BeNil())
			if !isOpenshift && os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("veryify pod started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, timeout*5, interval).Should(Succeed())

				By("veryify no taints on node")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(0))
				}, timeout*2, interval).Should(Succeed())

				By("veryify pod status still started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, timeout*2, interval).Should(Succeed())

				By("applying taints to node")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(0))
					node.Spec.Taints = []corev1.Taint{{Key: "artemis", Effect: corev1.TaintEffectPreferNoSchedule}}

					g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
				}, timeout*2, interval).Should(Succeed())

				By("veryify pod status still started")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, timeout*2, interval).Should(Succeed())

				By("reverting taints on node")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(1))
					node.Spec.Taints = []corev1.Taint{}
					g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
				}, timeout*2, interval).Should(Succeed())

			}

			Expect(k8sClient.Delete(ctx, createdCrd))

		})

		It("Toleration of artemis required add/remove verify status", func() {

			By("Creating a crd with tolerations")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Size = 1
			crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
				InitialDelaySeconds: 1,
				PeriodSeconds:       5,
			}

			crd.Spec.DeploymentPlan.Tolerations = []corev1.Toleration{
				{
					Key:      "artemis",
					Value:    "ok",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoExecute,
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			brokerKey := types.NamespacedName{Name: crd.Name, Namespace: crd.Namespace}
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			// some required services on crc get evicted which invalidates this test of taints
			isOpenshift, err := environments.DetectOpenshift()
			Expect(err).Should(BeNil())
			if !isOpenshift && os.Getenv("USE_EXISTING_CLUSTER") == "true" {

				By("veryify pod started as no taints in play")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, timeout*5, interval).Should(Succeed())

				By("apply matching taint with wrong value, force eviction")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(0))

					node.Spec.Taints = []corev1.Taint{{Key: "artemis", Value: "no", Effect: corev1.TaintEffectNoExecute}}
					g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("veryify pod status now starting, evicted!")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Starting)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("updating taint to match key and value")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(1))
					node.Spec.Taints = []corev1.Taint{{Key: "artemis", Value: "ok", Effect: corev1.TaintEffectNoExecute}}
					g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("veryify ready status on CR, started again")
				Eventually(func(g Gomega) {

					g.Expect(k8sClient.Get(ctx, brokerKey, createdCrd)).Should(Succeed())
					g.Expect(len(createdCrd.Status.PodStatus.Ready)).Should(BeEquivalentTo(1))

				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

				By("reverting taints on node")
				Eventually(func(g Gomega) {

					// find our node, take the first one...
					nodes := &corev1.NodeList{}
					g.Expect(k8sClient.List(ctx, nodes, &client.ListOptions{})).Should(Succeed())
					g.Expect(len(nodes.Items) > 0).Should(BeTrue())

					node := nodes.Items[0]
					g.Expect(len(node.Spec.Taints)).Should(BeEquivalentTo(1))
					node.Spec.Taints = []corev1.Taint{}
					g.Expect(k8sClient.Update(ctx, &node)).Should(Succeed())
				}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			}

			Expect(k8sClient.Delete(ctx, createdCrd))

		})

	})

	Context("Console secret Test", func() {

		crd := generateArtemisSpec(namespace)
		crd.Spec.Console.Expose = true
		crd.Spec.Console.SSLEnabled = true

		namespacedName := types.NamespacedName{
			Name:      crd.Name,
			Namespace: namespace,
		}
		reconcilerImpl := &ActiveMQArtemisReconcilerImpl{}
		theFsm := MakeActiveMQArtemisFSM(&crd, namespacedName, brokerReconciler)

		It("deploy broker with ssl enabled console", func() {
			os.Setenv("OPERATOR_OPENSHIFT", "true")
			defer os.Unsetenv("OPERATOR_OPENSHIFT")

			defaultConsoleSecretName := crd.Name + "-console-secret"
			currentSS := &appsv1.StatefulSet{}
			currentSS.Name = namer.CrToSS(crd.Name)
			currentSS.Namespace = namespace
			currentSS.Spec.Template.Spec.InitContainers = []corev1.Container{{
				Name: "main-container",
			}}
			currentSS.Spec.Template.Spec.Containers = []corev1.Container{{
				Name: "init-container",
			}}

			reconcilerImpl.ProcessConsole(theFsm, brokerReconciler.Client, brokerReconciler.Scheme, currentSS)
			secretName := theFsm.GetConsoleSecretName()
			internalSecretName := secretName + "-internal"
			consoleArgs := "AMQ_CONSOLE_ARGS"

			foundSecret := false
			foundInternalSecret := false
			defaultSecretPath := "/etc/" + defaultConsoleSecretName + "-volume/"
			defaultSslArgs := " --ssl-key " + defaultSecretPath + "broker.ks --ssl-key-password password --ssl-trust " + defaultSecretPath + "client.ts --ssl-trust-password password"
			for _, reqres := range theFsm.requestedResources {
				if reqres.GetObjectKind().GroupVersionKind().Kind == "Secret" {
					secret := reqres.(*corev1.Secret)
					if secret.Name == secretName {
						foundSecret = true
					}
					if secret.Name == internalSecretName {
						foundInternalSecret = true
						consoleSslValue := secret.StringData[consoleArgs]
						Expect(consoleSslValue).To(Equal(defaultSslArgs))
					}
				}
			}
			Expect(foundSecret).To(BeTrue())
			Expect(foundInternalSecret).To(BeTrue())

			foundSecretRef := false
			foundSecretKey := false
			for _, evar := range currentSS.Spec.Template.Spec.InitContainers[0].Env {
				if evar.Name == consoleArgs {
					if evar.ValueFrom.SecretKeyRef.Name == internalSecretName {
						foundSecretRef = true
					}
					if evar.ValueFrom.SecretKeyRef.Key == consoleArgs {
						foundSecretKey = true
					}
				}
			}
			Expect(foundSecretRef).To(BeTrue())
			Expect(foundSecretKey).To(BeTrue())
		})
	})

	Context("PodSecurityContext Test", func() {
		It("Setting the pods PodSecurityContext", func() {
			By("Creating a CR instance with PodSecurityContext configured")

			podSecurityContext := corev1.PodSecurityContext{}
			retrievedCR := &brokerv1beta1.ActiveMQArtemis{}
			createdSS := &appsv1.StatefulSet{}

			context := context.Background()
			defaultCR := generateArtemisSpec(namespace)
			defaultCR.Spec.DeploymentPlan.PodSecurityContext = &podSecurityContext

			By("Deploying the CR " + defaultCR.ObjectMeta.Name)
			Expect(k8sClient.Create(context, &defaultCR)).Should(Succeed())

			By("Making sure that the CR gets deployed " + defaultCR.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(defaultCR.ObjectMeta.Name, namespace, retrievedCR)
			}, timeout, interval).Should(BeTrue())
			Expect(retrievedCR.Name).Should(Equal(defaultCR.ObjectMeta.Name))

			By("Checking that the StatefulSet has been created with a PodSecurityContext field " + namer.CrToSS(retrievedCR.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(retrievedCR.Name), Namespace: namespace}
				if err := k8sClient.Get(context, key, createdSS); nil != err {
					return false
				}
				return nil != createdSS.Spec.Template.Spec.SecurityContext
			})

			By("Check for default CR instance deletion")
			Expect(k8sClient.Delete(context, retrievedCR))
			Eventually(func() bool {
				return checkCrdDeleted(defaultCR.ObjectMeta.Name, namespace, retrievedCR)
			}, timeout, interval).Should(BeTrue())

			By("Creating a non-default SELinuxOptions CR instance")
			nonDefaultCR := generateArtemisSpec(namespace)

			credentialSpecName := "CredentialSpecName0"
			credentialSpec := "CredentialSpec0"
			runAsUserName := "RunAsUserName0"
			hostProcess := false
			var runAsUser int64 = 1000
			var runAsGroup int64 = 1001
			runAsNonRoot := true
			var supplementalGroupA int64 = 2000
			var supplementalGroupB int64 = 2001
			supplementalGroups := []int64{supplementalGroupA, supplementalGroupB}
			var fsGroup int64 = 3000
			sysctlA := corev1.Sysctl{
				Name:  "NameA",
				Value: "ValueA",
			}
			sysctlB := corev1.Sysctl{
				Name:  "NameB",
				Value: "ValueB",
			}
			sysctls := []corev1.Sysctl{sysctlA, sysctlB}

			fsGCPString := "GroupChangePolicy0"
			fsGCP := corev1.PodFSGroupChangePolicy(fsGCPString)
			localhostProfile := "LocalhostProfile0"
			seccompProfile := corev1.SeccompProfile{
				Type:             corev1.SeccompProfileTypeUnconfined,
				LocalhostProfile: &localhostProfile,
			}

			nonDefaultCR.Spec.DeploymentPlan.PodSecurityContext = &corev1.PodSecurityContext{
				SELinuxOptions: &corev1.SELinuxOptions{
					User:  "TestUser0",
					Role:  "TestRole0",
					Type:  "TestType0",
					Level: "TestLevel0",
				},
				WindowsOptions: &corev1.WindowsSecurityContextOptions{
					GMSACredentialSpecName: &credentialSpecName,
					GMSACredentialSpec:     &credentialSpec,
					RunAsUserName:          &runAsUserName,
					HostProcess:            &hostProcess,
				},
				RunAsUser:           &runAsUser,
				RunAsGroup:          &runAsGroup,
				RunAsNonRoot:        &runAsNonRoot,
				SupplementalGroups:  supplementalGroups,
				FSGroup:             &fsGroup,
				Sysctls:             sysctls,
				FSGroupChangePolicy: &fsGCP,
				SeccompProfile:      &seccompProfile,
			}

			By("Deploying the non-default CR instance named " + nonDefaultCR.Name)
			Expect(k8sClient.Create(context, &nonDefaultCR)).Should(Succeed())

			retrievednonDefaultCR := brokerv1beta1.ActiveMQArtemis{}
			By("Checking to ensure that the non-default CR was created with the right values " + nonDefaultCR.Name)
			Eventually(func() bool {
				key := types.NamespacedName{Name: nonDefaultCR.Name, Namespace: namespace}
				retVal := false
				for {
					if err := k8sClient.Get(context, key, &retrievednonDefaultCR); nil != err {
						break
					}
					if !reflect.DeepEqual(nonDefaultCR.Spec.DeploymentPlan.PodSecurityContext, retrievednonDefaultCR.Spec.DeploymentPlan.PodSecurityContext) {
						break
					}

					retVal = true
					break
				}

				return retVal
			})

			By("Checking that the StatefulSet has been created with the non-default PodSecurityContext " + namer.CrToSS(nonDefaultCR.Name))
			Eventually(func() bool {
				nonDefaultSS := &appsv1.StatefulSet{}
				key := types.NamespacedName{Name: namer.CrToSS(nonDefaultCR.Name), Namespace: namespace}
				retVal := false
				for {
					if err := k8sClient.Get(context, key, nonDefaultSS); nil != err {
						break
					}

					if nil == nonDefaultSS.Spec.Template.Spec.SecurityContext {
						break
					}

					if !reflect.DeepEqual(nonDefaultCR, nonDefaultSS.Spec.Template.Spec.SecurityContext) {
						break
					}

					retVal = true
					break
				}

				return retVal
			})

			By("Check for non-default CR instance deletion")
			Expect(k8sClient.Delete(context, &retrievednonDefaultCR))
			Eventually(func() bool {
				return checkCrdDeleted(defaultCR.ObjectMeta.Name, namespace, &retrievednonDefaultCR)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Affinity Test", func() {
		It("setting Pod Affinity", func() {
			By("Creating a crd with pod affinity")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			labelSelector := metav1.LabelSelector{}
			labelSelector.MatchLabels = make(map[string]string)
			labelSelector.MatchLabels["key"] = "value"

			podAffinityTerm := corev1.PodAffinityTerm{}
			podAffinityTerm.LabelSelector = &labelSelector

			podAffinity := corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					podAffinityTerm,
				},
			}
			crd.Spec.DeploymentPlan.Affinity.PodAffinity = &podAffinity

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))
			By("Checking that Stateful Set is Created with the node selectors " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Affinity.PodAffinity != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the pd affinity are correct")
			Expect(createdSs.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"] == "value").Should(BeTrue())

			By("Updating the CR")
			Eventually(func() bool {
				// we need to update the latest version and deal with update failures
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

			original := createdCrd

			labelSelector = metav1.LabelSelector{}
			labelSelector.MatchLabels = make(map[string]string)
			labelSelector.MatchLabels["key"] = "differentvalue"

			podAffinityTerm = corev1.PodAffinityTerm{}
			podAffinityTerm.LabelSelector = &labelSelector
			podAffinity = corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					podAffinityTerm,
				},
			}
			original.Spec.DeploymentPlan.Affinity.PodAffinity = &podAffinity
			By("Redeploying the CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}

				if createdSs.Spec.Template.Spec.Affinity.PodAffinity == nil {
					return false
				}

				if len(createdSs.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 {
					return false
				}

				fmt.Printf("checking value of key: %v\n", createdSs.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"])
				By("Making sure the pd affinity are correct")
				return createdSs.Spec.Template.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"] == "differentvalue"

			}, timeout, interval).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
		It("setting Pod AntiAffinity", func() {
			By("Creating a crd with pod anti affinity")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			labelSelector := metav1.LabelSelector{}
			labelSelector.MatchLabels = make(map[string]string)
			labelSelector.MatchLabels["key"] = "value"

			podAffinityTerm := corev1.PodAffinityTerm{}
			podAffinityTerm.LabelSelector = &labelSelector
			podAntiAffinity := corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					podAffinityTerm,
				},
			}
			crd.Spec.DeploymentPlan.Affinity.PodAntiAffinity = &podAntiAffinity

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))
			By("Checking that Stateful Set is Created with the node selectors " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the pd affinity are correct")
			Expect(createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"] == "value").Should(BeTrue())

			By("Updating the CR")
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			original := createdCrd

			labelSelector = metav1.LabelSelector{}
			labelSelector.MatchLabels = make(map[string]string)
			labelSelector.MatchLabels["key"] = "differentvalue"

			podAffinityTerm = corev1.PodAffinityTerm{}
			podAffinityTerm.LabelSelector = &labelSelector
			podAntiAffinity = corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					podAffinityTerm,
				},
			}
			original.Spec.DeploymentPlan.Affinity.PodAntiAffinity = &podAntiAffinity
			By("Redeploying the CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}

				if createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
					return false
				}

				if len(createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 {
					return false
				}

				fmt.Printf("checking value of key: %v\n", createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"])
				return createdSs.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels["key"] == "differentvalue"

			}, timeout, interval).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
		It("setting Node AntiAffinity", func() {
			By("Creating a crd with node affinity")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			nodeSelectorRequirement := corev1.NodeSelectorRequirement{
				Key:    "foo",
				Values: make([]string, 1),
			}
			nodeSelectorRequirements := [1]corev1.NodeSelectorRequirement{nodeSelectorRequirement}
			nodeSelectorRequirements[0] = nodeSelectorRequirement
			nodeSelectorTerm := corev1.NodeSelectorTerm{MatchExpressions: nodeSelectorRequirements[:]}
			nodeSelectorTerms := [1]corev1.NodeSelectorTerm{nodeSelectorTerm}
			nodeSelector := corev1.NodeSelector{
				NodeSelectorTerms: nodeSelectorTerms[:],
			}
			nodeAffinity := corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
			}
			crd.Spec.DeploymentPlan.Affinity.NodeAffinity = &nodeAffinity

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))
			By("Checking that Stateful Set is Created with the node selectors " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Affinity.NodeAffinity != nil
			}, timeout, interval).Should(BeTrue())

			By("Making sure the node affinity are correct")
			Expect(createdSs.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key).Should(Equal("foo"))

			By("Updating the CR")
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			original := createdCrd

			nodeSelectorRequirement = corev1.NodeSelectorRequirement{
				Key:    "bar",
				Values: make([]string, 2),
			}
			nodeSelectorRequirements = [1]corev1.NodeSelectorRequirement{nodeSelectorRequirement}
			nodeSelectorRequirements[0] = nodeSelectorRequirement
			nodeSelectorTerm = corev1.NodeSelectorTerm{MatchExpressions: nodeSelectorRequirements[:]}
			nodeSelectorTerms = [1]corev1.NodeSelectorTerm{nodeSelectorTerm}
			nodeSelector = corev1.NodeSelector{
				NodeSelectorTerms: nodeSelectorTerms[:],
			}
			nodeAffinity = corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
			}
			original.Spec.DeploymentPlan.Affinity.NodeAffinity = &nodeAffinity
			By("Redeploying the CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			Eventually(func(g Gomega) bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				if k8sClient.Get(ctx, key, createdSs) != nil {
					return false
				}

				if len(createdSs.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
					return false
				}

				By("Making sure the pod affinity is correct")
				return createdSs.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key == "bar"

			}, timeout, interval).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Node Selector Test", func() {
		It("passing in 2 labels", func() {
			By("Creating a crd with 2 selectors")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			nodeSelector := map[string]string{
				"location": "production",
				"type":     "foo",
			}
			crd.Spec.DeploymentPlan.NodeSelector = nodeSelector

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))
			By("Checking that Stateful Set is Created with the node selectors " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return len(createdSs.Spec.Template.Spec.NodeSelector) == 2
			}, timeout, interval).Should(Equal(true))

			By("Making sure the node selectors are correct")
			Expect(createdSs.Spec.Template.Spec.NodeSelector["location"] == "production").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.NodeSelector["type"] == "foo").Should(BeTrue())

			By("Updating the CR")
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			original := createdCrd

			nodeSelector = map[string]string{
				"type": "foo",
			}
			original.Spec.DeploymentPlan.NodeSelector = nodeSelector
			By("Redeploying the CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			Eventually(func() int {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return -1
				}
				return len(createdSs.Spec.Template.Spec.NodeSelector)
			}, timeout, interval).Should(Equal(1))

			By("Making sure the node selectors are correct")
			Expect(createdSs.Spec.Template.Spec.NodeSelector["location"] == "production").Should(BeFalse())
			Expect(createdSs.Spec.Template.Spec.NodeSelector["type"] == "foo").Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Labels Test", func() {
		It("passing in 2 labels", func() {
			By("Creating a crd with 2 labels")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Labels = make(map[string]string)
			crd.Spec.DeploymentPlan.Labels["key1"] = "val1"
			crd.Spec.DeploymentPlan.Labels["key2"] = "val2"

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))
			By("Checking that Stateful Set is Created with the labels " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return len(createdSs.ObjectMeta.Labels) == 4
			}, timeout, interval).Should(Equal(true))

			By("Making sure the labels are correct")
			Expect(createdSs.ObjectMeta.Labels["key1"] == "val1").Should(BeTrue())
			Expect(createdSs.ObjectMeta.Labels["key2"] == "val2").Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Tolerations Test", func() {
		It("passing in 2 tolerations", func() {

			By("Creating a crd with 2 tolerations")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.DeploymentPlan.Tolerations = []corev1.Toleration{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: "NoSchedule",
				},
				{
					Key:    "yes",
					Value:  "No",
					Effect: "NoSchedule",
				},
			}

			By("Deploying the CRD " + crd.ObjectMeta.Name)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Making sure that the CRD gets deployed " + crd.ObjectMeta.Name)
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the tolerations " + namer.CrToSS(createdCrd.Name))
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return len(createdSs.Spec.Template.Spec.Tolerations) == 2
			}, timeout, interval).Should(Equal(true))
			Expect(len(createdSs.Spec.Template.Spec.Tolerations) == 2).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Key == "foo").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Value == "bar").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Effect == "NoSchedule").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[1].Key == "yes").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[1].Value == "No").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[1].Effect == "NoSchedule").Should(BeTrue())

			By("Redeploying the CRD with different Tolerations")

			Eventually(func() bool {

				// fetch, modify and update (we compete with the status updates)

				ok := getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
				if !ok {
					return false
				}

				createdCrd.Spec.DeploymentPlan.Tolerations = []corev1.Toleration{
					{
						Key:    "yes",
						Value:  "No",
						Effect: "NoSchedule",
					},
				}

				err := k8sClient.Update(ctx, createdCrd)
				return err == nil
			}, timeout, interval).Should(Equal(true))

			By("and checking there is just a single Toleration")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return 0, err
				}
				return len(createdSs.Spec.Template.Spec.Tolerations), err
			}, timeout, interval).Should(Equal(1))
			Expect(len(createdSs.Spec.Template.Spec.Tolerations) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Key == "yes").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Value == "No").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Tolerations[0].Effect == "NoSchedule").Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("Liveness Probe Tests", func() {
		It("Override Liveness Probe No Exec", func() {
			By("By creating a crd with Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			livenessProbe := corev1.Probe{}
			livenessProbe.PeriodSeconds = 5
			livenessProbe.InitialDelaySeconds = 6
			livenessProbe.TimeoutSeconds = 7
			livenessProbe.SuccessThreshold = 8
			livenessProbe.FailureThreshold = 9
			crd.Spec.DeploymentPlan.LivenessProbe = &livenessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}

			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Liveness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Liveness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.TCPSocket.Port.String() == "8161").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 5).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 6).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds == 7).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold == 8).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold == 9).Should(BeTrue())

			By("Updating the CR")
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			original := createdCrd

			original.Spec.DeploymentPlan.LivenessProbe.PeriodSeconds = 15
			original.Spec.DeploymentPlan.LivenessProbe.InitialDelaySeconds = 16
			original.Spec.DeploymentPlan.LivenessProbe.TimeoutSeconds = 17
			original.Spec.DeploymentPlan.LivenessProbe.SuccessThreshold = 18
			original.Spec.DeploymentPlan.LivenessProbe.FailureThreshold = 19
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			original.Spec.DeploymentPlan.LivenessProbe.Exec = &exec
			By("Redeploying the CRD")
			By("Redeploying the modified CRD")
			Expect(k8sClient.Update(ctx, original)).Should(Succeed())

			By("Retrieving the new SS to find the modification")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 15
			}, timeout, interval).Should(Equal(true))

			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec != nil).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 15).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 16).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds == 17).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold == 18).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold == 19).Should(BeTrue())

			By("check it has gone")
			Expect(k8sClient.Delete(ctx, createdCrd))
			Eventually(func() bool { return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
		})

		It("Override Liveness Probe Exec", func() {
			By("By creating a crd with Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			livenessProbe := corev1.Probe{}
			livenessProbe.PeriodSeconds = 5
			livenessProbe.InitialDelaySeconds = 6
			livenessProbe.TimeoutSeconds = 7
			livenessProbe.SuccessThreshold = 8
			livenessProbe.FailureThreshold = 9
			livenessProbe.Exec = &exec
			crd.Spec.DeploymentPlan.LivenessProbe = &livenessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Liveness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Liveness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds == 5).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 6).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds == 7).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold == 8).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold == 9).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})

		It("Default Liveness Probe", func() {
			By("By creating a crd without Liveness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that the Liveness Probe is created")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil
			}, timeout, interval).Should(Equal(true))

			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.TCPSocket.Port.String() == "8161").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds).Should(BeEquivalentTo(5))

			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Readiness Probe Tests", func() {
		It("Override Readiness Probe No Exec", func() {
			By("By creating a crd with Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			readinessProbe := corev1.Probe{}
			readinessProbe.PeriodSeconds = 5
			readinessProbe.InitialDelaySeconds = 6
			readinessProbe.TimeoutSeconds = 7
			readinessProbe.SuccessThreshold = 8
			readinessProbe.FailureThreshold = 9
			crd.Spec.DeploymentPlan.ReadinessProbe = &readinessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Readiness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/bin/bash").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[1] == "-c").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[2] == "/opt/amq/bin/readinessProbe.sh").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})

		It("Override Readiness Probe Exec", func() {
			By("By creating a crd with Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			exec := corev1.ExecAction{
				Command: []string{"/broker/bin/artemis check node"},
			}
			readinessProbe := corev1.Probe{}
			readinessProbe.Exec = &exec
			crd.Spec.DeploymentPlan.ReadinessProbe = &readinessProbe
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			By("Deploying the CRD")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Making sure that the CRD gets deployed")
			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that Stateful Set is Created with the Readiness Probe")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/broker/bin/artemis check node").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})

		It("Default Readiness Probe", func() {
			By("By creating a crd without Readiness Probe")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			crd.Spec.AdminUser = "admin"
			crd.Spec.AdminPassword = "password"
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			createdSs := &appsv1.StatefulSet{}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("Checking that the Readiness Probe is created")
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}

				err := k8sClient.Get(ctx, key, createdSs)

				if err != nil {
					return false
				}
				return createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Exec != nil
			}, timeout, interval).Should(Equal(true))

			By("Making sure the Readiness probe is correct")
			Expect(len(createdSs.Spec.Template.Spec.Containers) == 1).Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[0] == "/bin/bash").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[1] == "-c").Should(BeTrue())
			Expect(createdSs.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.Exec.Command[2] == "/opt/amq/bin/readinessProbe.sh").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdCrd))

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Status", func() {
		It("Expect pod desc", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			Eventually(func() bool {
				return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			// would like more status updates on createdCrd

			By("By checking the status of stateful set")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return -1, err
				}

				// presence is good enough... check on this status just for kicks
				return int(createdSs.Status.Replicas), err
			}, duration, interval).Should(Equal(0))

			By("Checking stopped status of CR because we expect it to fail to deploy")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdCrd)

				if err != nil {
					return -1, err
				}

				return len(createdCrd.Status.PodStatus.Stopped), nil
			}, timeout, interval).Should(Equal(1))

			By("Checking presence of secrets")
			secretList := &corev1.SecretList{}
			opts := &client.ListOptions{
				Namespace: namespace,
			}
			Eventually(func() int {
				err := k8sClient.List(ctx, secretList, opts)
				if err != nil {
					fmt.Printf("error getting secretList! %v", err)
				}
				count := 0
				for _, s := range secretList.Items {
					if strings.Contains(s.ObjectMeta.Name, createdCrd.Name) {
						count++
					}
				}
				return count
			}, timeout, interval).Should(Equal(3))

			By("deleting crd")
			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Env var updates TRIGGERED_ROLL_COUNT checksum", func() {
		It("Expect TRIGGERED_ROLL_COUNT count non 0", func() {
			By("By creating a new crd")
			var checkSum string
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.AdminUser = "Joe"
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			By("By checking the container stateful set for TRIGGERED_ROLL_COUNT non zero")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false, err
				}

				found := false
				for _, container := range createdSs.Spec.Template.Spec.Containers {
					for _, env := range container.Env {
						if env.Name == "TRIGGERED_ROLL_COUNT" {
							if env.Value > "0" {
								checkSum = env.Value
								found = true
							}
						}
					}
				}
				return found, err
			}, duration, interval).Should(Equal(true))

			By("update env var")
			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: crd.ObjectMeta.Namespace}, createdCrd)
				if err == nil {

					createdCrd.Spec.AdminUser = "Joseph"

					err = k8sClient.Update(ctx, createdCrd)
					if err != nil {
						fmt.Printf("error on update! %v\n", err)
					}
				}
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("verify different checksum")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false, err
				}

				found := false
				for _, container := range createdSs.Spec.Template.Spec.Containers {
					for _, env := range container.Env {
						if env.Name == "TRIGGERED_ROLL_COUNT" {
							if env.Value != checkSum {
								found = true
							}
						}
					}
				}
				return found, err
			}, duration, interval).Should(Equal(true))

			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("BrokerProperties", func() {
		It("Expect vol mount via config map", func() {
			By("By creating a new crd with BrokerProperties in the spec")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.BrokerProperties = []string{"globalMaxSize=512m"}

			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}
			Eventually(func() bool { return getPersistedVersionedCrd(crd.ObjectMeta.Name, namespace, createdCrd) }, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(crd.ObjectMeta.Name))

			hexShaOriginal := HexShaHashOfMap(crd.Spec.BrokerProperties)

			By("By finding a new config map with broker props")
			configMap := &corev1.ConfigMap{}
			key := types.NamespacedName{Name: crd.ObjectMeta.Name + "-props-" + hexShaOriginal, Namespace: crd.ObjectMeta.Namespace}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, configMap)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("By checking the container stateful set for java opts")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false, err
				}

				found := false
				for _, container := range createdSs.Spec.Template.Spec.InitContainers {
					for _, env := range container.Env {
						if env.Name == "JAVA_OPTS" {
							if strings.Contains(env.Value, "broker.properties") {
								found = true
							}
						}
					}
				}

				return found, err
			}, duration, interval).Should(Equal(true))

			By("By checking the stateful set for volume mount path")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false, err
				}

				found := false
				for _, container := range createdSs.Spec.Template.Spec.Containers {
					for _, vm := range container.VolumeMounts {
						// mount path can't have a .
						if strings.Contains(vm.MountPath, "-props-") {
							found = true
						}
					}
				}

				return found, err
			}, duration, interval).Should(Equal(true))

			By("By checking the container stateful launch set for STATEFUL_SET_ORDINAL")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					return false, err
				}

				found := false
				for _, container := range createdSs.Spec.Template.Spec.Containers {
					for _, command := range container.Command {
						if strings.Contains(command, "STATEFUL_SET_ORDINAL") {
							found = true
						}
					}
				}

				return found, err
			}, duration, interval).Should(Equal(true))

			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())

			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})

		It("Expect new config map on update to BrokerProperties", func() {
			By("By creating a crd with BrokerProperties in the spec")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.BrokerProperties = []string{"globalMaxSize=64g"}
			hexShaOriginal := HexShaHashOfMap(crd.Spec.BrokerProperties)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("By eventualy finding a matching config map with broker props")
			configMapList := &corev1.ConfigMapList{}
			opts := &client.ListOptions{
				Namespace: namespace,
			}
			Eventually(func() bool {
				err := k8sClient.List(ctx, configMapList, opts)
				if err != nil {
					fmt.Printf("error getting list of configopts map! %v", err)
				}
				for _, cm := range configMapList.Items {
					if strings.Contains(cm.ObjectMeta.Name, hexShaOriginal) {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("updating the crd, expect new ConfigMap name")
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			By("pushing the update on the current version...")
			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: crd.ObjectMeta.Namespace}, createdCrd)
				if err == nil {

					// add a new property
					createdCrd.Spec.BrokerProperties = append(createdCrd.Spec.BrokerProperties, "gen="+strconv.FormatInt(createdCrd.ObjectMeta.Generation, 10))

					err = k8sClient.Update(ctx, createdCrd)
					if err != nil {
						fmt.Printf("error on update! %v\n", err)
					}
				}
				return err == nil
			}, timeout, interval).Should(BeTrue())

			hexShaModified := HexShaHashOfMap(createdCrd.Spec.BrokerProperties)

			By("finding the updated config map using the sha")
			Eventually(func() bool {
				err := k8sClient.List(ctx, configMapList, opts)

				if err == nil && len(configMapList.Items) > 0 {

					for _, cm := range configMapList.Items {
						if strings.Contains(cm.ObjectMeta.Name, hexShaModified) {
							return true
						}
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			// cleanup
			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

			// cannot verify no leaks b/c gc is not enabled on envTest
			// and on delete we don't have any state to determine the owner reference
		})

		It("Expect two crs to coexist", func() {
			By("By creating two crds with BrokerProperties in the spec")
			ctx := context.Background()
			crd1 := generateArtemisSpec(namespace)
			crd2 := generateArtemisSpec(namespace)

			Expect(k8sClient.Create(ctx, &crd1)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &crd2)).Should(Succeed())

			By("By eventualy finding two config maps with broker props")
			configMapList := &corev1.ConfigMapList{}
			opts := &client.ListOptions{
				Namespace: namespace,
			}
			Eventually(func() int {
				err := k8sClient.List(ctx, configMapList, opts)
				if err != nil {
					fmt.Printf("error getting list of config opts map! %v", err)
				}

				ret := 0
				for _, cm := range configMapList.Items {
					if strings.Contains(cm.ObjectMeta.Name, crd1.Name) || strings.Contains(cm.ObjectMeta.Name, crd2.Name) {
						ret++
					}
				}
				return ret
			}, timeout, interval).Should(BeEquivalentTo(2))

			// cleanup
			Expect(k8sClient.Delete(ctx, &crd1)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &crd2)).Should(Succeed())
		})

	})

	Context("With address settings via updated cr", func() {
		It("Expect ok deploy", func() {
			By("By creating a crd without address spec")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.BrokerProperties = []string{"globalMaxSize=65g"}

			hexShaOriginal := HexShaHashOfMap(crd.Spec.BrokerProperties)
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("By eventualy finding a matching config map with broker props")
			configMapList := &corev1.ConfigMapList{}
			opts := &client.ListOptions{
				Namespace: namespace,
			}
			Eventually(func() bool {
				err := k8sClient.List(ctx, configMapList, opts)
				if err != nil {
					fmt.Printf("error getting list of configopts map! %v", err)
				}
				for _, cm := range configMapList.Items {
					if strings.Contains(cm.ObjectMeta.Name, hexShaOriginal) {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("updating the crd with address settings")
			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			By("pushing the update on the current version...")
			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: crd.ObjectMeta.Namespace}, createdCrd)
				if err == nil {

					// add a new property
					createdCrd.Spec.BrokerProperties = append(createdCrd.Spec.BrokerProperties, "gen="+strconv.FormatInt(createdCrd.ObjectMeta.Generation, 10))

					// add address settings, to an existing crd
					ma := "merge_all"
					dlq := "dlq"
					dlqabc := "dlqabc"
					maxSize := "10m"

					createdCrd.Spec.AddressSettings = brokerv1beta1.AddressSettingsType{
						ApplyRule: &ma,
						AddressSetting: []brokerv1beta1.AddressSettingType{
							{
								Match:             "#",
								DeadLetterAddress: &dlq,
							},
							{
								Match:             "abc#",
								DeadLetterAddress: &dlqabc,
								MaxSizeBytes:      &maxSize,
							},
						},
					}

					err = k8sClient.Update(ctx, createdCrd)
					if err != nil {
						fmt.Printf("error on update! %v\n", err)
					}
				}
				return err == nil
			}, timeout, interval).Should(BeTrue())

			hexShaModified := HexShaHashOfMap(createdCrd.Spec.BrokerProperties)

			By("finding the updated config map using the sha")

			Eventually(func() bool {
				err := k8sClient.List(ctx, configMapList, opts)

				if err == nil && len(configMapList.Items) > 0 {

					for _, cm := range configMapList.Items {
						if strings.Contains(cm.ObjectMeta.Name, hexShaModified) {
							return true
						}
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("tracking the yaconfig init command with user_address_settings and verifying no change on further update")
			key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
			createdSs := &appsv1.StatefulSet{}
			var initArgsString string
			Eventually(func(g Gomega) {

				createdSs := &appsv1.StatefulSet{}

				g.Expect(k8sClient.Get(ctx, key, createdSs))

				initArgsString = strings.Join(createdSs.Spec.Template.Spec.InitContainers[0].Args, ",")
				g.Expect(initArgsString).Should(ContainSubstring("user_address_settings"))

			}, timeout, interval).Should(Succeed())

			By("pushing another update on the current version...")
			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{Name: crd.ObjectMeta.Name, Namespace: crd.ObjectMeta.Namespace}, createdCrd)
				if err == nil {

					createdCrd.Spec.BrokerProperties = append(createdCrd.Spec.BrokerProperties, "gen2="+strconv.FormatInt(createdCrd.ObjectMeta.Generation, 10))

					// add address settings, to an existing crd
					ma := "merge_all"
					dlq := "dlq"
					dlqabc := "dlqabc"
					maxSize := "10m"

					createdCrd.Spec.AddressSettings = brokerv1beta1.AddressSettingsType{
						ApplyRule: &ma,
						AddressSetting: []brokerv1beta1.AddressSettingType{
							{
								Match:             "#",
								DeadLetterAddress: &dlq,
							},
							{
								Match:             "abc#",
								DeadLetterAddress: &dlqabc,
								MaxSizeBytes:      &maxSize,
							},
						},
					}

					err = k8sClient.Update(ctx, createdCrd)
					if err != nil {
						fmt.Printf("error on update! %v\n", err)
					}
				}
				return err == nil
			}, timeout, interval).Should(BeTrue())

			hexShaModified = HexShaHashOfMap(createdCrd.Spec.BrokerProperties)

			By("again finding the updated config map using the sha")

			Eventually(func() bool {
				err := k8sClient.List(ctx, configMapList, opts)

				if err == nil && len(configMapList.Items) > 0 {

					for _, cm := range configMapList.Items {
						if strings.Contains(cm.ObjectMeta.Name, hexShaModified) {
							return true
						}
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("verifying init command args did not change")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, createdSs); err != nil {
					return false
				}

				updatedInitArgs := strings.Join(createdSs.Spec.Template.Spec.InitContainers[0].Args, ",")
				return initArgsString == updatedInitArgs
			}, timeout, interval).Should(BeTrue())

			// cleanup
			Expect(k8sClient.Delete(ctx, createdCrd)).Should(Succeed())
			By("check it has gone")
			Eventually(func() bool {
				return checkCrdDeleted(crd.ObjectMeta.Name, namespace, createdCrd)
			}, timeout, interval).Should(BeTrue())

		})
	})
	Context("With deployed controller", func() {
		It("Checking acceptor service while expose is false", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:   "new-acceptor",
					Port:   61616,
					Expose: false,
				},
			}
			crd.Spec.Connectors = []brokerv1beta1.ConnectorType{
				{
					Name:   "new-connector",
					Port:   61616,
					Expose: false,
				},
			}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name + "-" + "new-acceptor-0-svc", Namespace: namespace}
				acceptorService := &corev1.Service{}
				err := k8sClient.Get(context.Background(), key, acceptorService)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name + "-" + "new-connector-0-svc", Namespace: namespace}
				connectorService := &corev1.Service{}
				err := k8sClient.Get(context.Background(), key, connectorService)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
	})

	Context("With deployed controller", func() {
		It("Testing acceptor bindToAllInterfaces default", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name: "new-acceptor",
					Port: 61666,
				},
			}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			ssNamespacedName := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}
			Eventually(func(g Gomega) {
				currentStatefulSet, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, nil, k8sClient)
				g.Expect(err).To(BeNil())
				g.Expect(len(currentStatefulSet.Spec.Template.Spec.InitContainers)).Should(BeEquivalentTo(1))
				initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]

				//check AMQ_ACCEPTORS value
				for _, envVar := range initContainer.Env {
					if envVar.Name == "AMQ_ACCEPTORS" {
						secretName := envVar.ValueFrom.SecretKeyRef.Name
						namespaceName := types.NamespacedName{
							Name:      secretName,
							Namespace: namespace,
						}
						secret, err := secrets.RetriveSecret(namespaceName, secretName, make(map[string]string), k8sClient)
						g.Expect(err).To(BeNil())
						data := secret.Data[envVar.ValueFrom.SecretKeyRef.Key]
						//the value is a string of acceptors in xml format:
						//<acceptor name="new-acceptor">...</acceptor><another one>...
						//we need to locate our target acceptor and do the check
						//we use the port as a clue
						g.Expect(strings.Contains(string(data), "ACCEPTOR_IP:61666")).To(BeTrue())
					}
				}
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
		It("Testing acceptor bindToAllInterfaces being false", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			bindTo := false
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:                "new-acceptor",
					Port:                61666,
					BindToAllInterfaces: &bindTo,
				},
			}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			ssNamespacedName := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}

			Eventually(func(g Gomega) {
				currentStatefulSet, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, nil, k8sClient)
				g.Expect(err).To(BeNil())
				g.Expect(len(currentStatefulSet.Spec.Template.Spec.InitContainers)).Should(BeEquivalentTo(1))
				initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]

				//check AMQ_ACCEPTORS value
				for _, envVar := range initContainer.Env {
					if envVar.Name == "AMQ_ACCEPTORS" {
						secretName := envVar.ValueFrom.SecretKeyRef.Name
						namespaceName := types.NamespacedName{
							Name:      secretName,
							Namespace: namespace,
						}
						secret, err := secrets.RetriveSecret(namespaceName, secretName, make(map[string]string), k8sClient)
						g.Expect(err).To(BeNil())
						data := secret.Data[envVar.ValueFrom.SecretKeyRef.Key]
						//the value is a string of acceptors in xml format:
						//<acceptor name="new-acceptor">...</acceptor><another one>...
						//we need to locate our target acceptor and do the check
						//we use the port as a clue
						g.Expect(strings.Contains(string(data), "ACCEPTOR_IP:61666")).To(BeTrue())
					}
				}
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
		It("Testing acceptor bindToAllInterfaces being true", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			bindToAll := true
			notbindToAll := false
			crd.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:                "new-acceptor",
					Port:                61666,
					BindToAllInterfaces: &bindToAll,
				},
				{
					Name: "new-acceptor-1",
					Port: 61777,
				},
				{
					Name:                "new-acceptor-2",
					Port:                61888,
					BindToAllInterfaces: &notbindToAll,
				},
			}
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			ssNamespacedName := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}

			Eventually(func(g Gomega) {
				currentStatefulSet, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, nil, k8sClient)
				g.Expect(err).To(BeNil())
				g.Expect(len(currentStatefulSet.Spec.Template.Spec.InitContainers)).Should(BeEquivalentTo(1))
				initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]
				//check AMQ_ACCEPTORS value
				for _, envVar := range initContainer.Env {
					if envVar.Name == "AMQ_ACCEPTORS" {
						secretName := envVar.ValueFrom.SecretKeyRef.Name
						namespaceName := types.NamespacedName{
							Name:      secretName,
							Namespace: namespace,
						}
						secret, err := secrets.RetriveSecret(namespaceName, secretName, make(map[string]string), k8sClient)
						g.Expect(err).To(BeNil())
						data := secret.Data[envVar.ValueFrom.SecretKeyRef.Key]
						//the value is a string of acceptors in xml format:
						//<acceptor name="new-acceptor">...</acceptor><another one>...
						//we need to locate our target acceptor and do the check
						//we use the port as a clue
						g.Expect(strings.Contains(string(data), "0.0.0.0:61666")).To(BeTrue())
						//the other one not affected
						g.Expect(strings.Contains(string(data), "ACCEPTOR_IP:61777")).To(BeTrue())
						g.Expect(strings.Contains(string(data), "ACCEPTOR_IP:61888")).To(BeTrue())
					}
				}

			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
	})

	Context("With a deployed controller", func() {
		//TODO: Remove the 4x duplication and add all acceptor settings

		It("Testing acceptor keyStoreProvider being set", func() {
			By("By creating a new custom resource instance")
			ctx := context.Background()
			cr := generateArtemisSpec(namespace)

			cr.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			keyStoreProvider := "SunJCE"
			cr.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:             "new-acceptor",
					Port:             61666,
					SSLEnabled:       true,
					KeyStoreProvider: keyStoreProvider,
				},
			}
			Expect(k8sClient.Create(ctx, &cr)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &cr)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
			Eventually(func() bool {
				_, ok := namespacedNameToFSM[key]
				return ok
			}, timeout, interval).Should(BeTrue())

			fsm := namespacedNameToFSM[key]
			ssNamespacedName := fsm.GetStatefulSetNamespacedName()
			Eventually(func() bool {
				_, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				return err == nil
			}).Should(BeTrue())

			Eventually(func() bool {
				currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				len := len(currentStatefulSet.Spec.Template.Spec.InitContainers)
				return len == 1
			}).Should(BeTrue())

			currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
			initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]
			//check AMQ_ACCEPTORS value
			for _, envVar := range initContainer.Env {
				if envVar.Name == "AMQ_ACCEPTORS" {
					secretName := envVar.ValueFrom.SecretKeyRef.Name
					namespaceName := types.NamespacedName{
						Name:      secretName,
						Namespace: namespace,
					}
					//the value is a string of acceptors in xml format:
					//<acceptor name="new-acceptor">...</acceptor><another one>...
					//we need to locate our target acceptor and do the check
					//we use the port as a clue
					checkSecretHasCorrectKeyValue(secretName, namespaceName, envVar.ValueFrom.SecretKeyRef.Key, "keyStoreProvider=SunJCE")
				}
			}
			Expect(k8sClient.Delete(ctx, &cr)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(cr.Name, namespace, &cr), timeout, interval).Should(BeTrue())
		})
		It("Testing acceptor trustStoreType being set", func() {
			By("By creating a new custom resource instance")
			ctx := context.Background()
			cr := generateArtemisSpec(namespace)

			cr.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			trustStoreType := "JCEKS"
			cr.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:           "new-acceptor",
					Port:           61666,
					SSLEnabled:     true,
					TrustStoreType: trustStoreType,
				},
			}
			Expect(k8sClient.Create(ctx, &cr)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &cr)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
			Eventually(func() bool {
				_, ok := namespacedNameToFSM[key]
				return ok
			}, timeout, interval).Should(BeTrue())

			fsm := namespacedNameToFSM[key]
			ssNamespacedName := fsm.GetStatefulSetNamespacedName()
			Eventually(func() bool {
				_, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				return err == nil
			}).Should(BeTrue())

			Eventually(func() bool {
				currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				len := len(currentStatefulSet.Spec.Template.Spec.InitContainers)
				return len == 1
			}).Should(BeTrue())

			currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
			initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]
			//check AMQ_ACCEPTORS value
			for _, envVar := range initContainer.Env {
				if envVar.Name == "AMQ_ACCEPTORS" {
					secretName := envVar.ValueFrom.SecretKeyRef.Name
					namespaceName := types.NamespacedName{
						Name:      secretName,
						Namespace: namespace,
					}
					//the value is a string of acceptors in xml format:
					//<acceptor name="new-acceptor">...</acceptor><another one>...
					//we need to locate our target acceptor and do the check
					//we use the port as a clue
					checkSecretHasCorrectKeyValue(secretName, namespaceName, envVar.ValueFrom.SecretKeyRef.Key, "trustStoreType=JCEKS")
				}
			}
			Expect(k8sClient.Delete(ctx, &cr)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(cr.Name, namespace, &cr), timeout, interval).Should(BeTrue())
		})
		It("Testing acceptor trustStoreProvider being set", func() {
			By("By creating a new custom resource instance")
			ctx := context.Background()
			cr := generateArtemisSpec(namespace)

			cr.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size: 1,
			}
			trustStoreProvider := "SUN"
			cr.Spec.Acceptors = []brokerv1beta1.AcceptorType{
				{
					Name:               "new-acceptor",
					Port:               61666,
					SSLEnabled:         true,
					TrustStoreProvider: trustStoreProvider,
				},
			}
			Expect(k8sClient.Create(ctx, &cr)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &cr)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			key := types.NamespacedName{Name: cr.Name, Namespace: namespace}
			Eventually(func() bool {
				_, ok := namespacedNameToFSM[key]
				return ok
			}, timeout, interval).Should(BeTrue())

			fsm := namespacedNameToFSM[key]
			ssNamespacedName := fsm.GetStatefulSetNamespacedName()
			Eventually(func() bool {
				_, err := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				return err == nil
			}).Should(BeTrue())

			Eventually(func() bool {
				currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
				len := len(currentStatefulSet.Spec.Template.Spec.InitContainers)
				return len == 1
			}).Should(BeTrue())

			currentStatefulSet, _ := ss.RetrieveStatefulSet(ssNamespacedName.Name, ssNamespacedName, fsm.namers.LabelBuilder.Labels(), k8sClient)
			initContainer := currentStatefulSet.Spec.Template.Spec.InitContainers[0]
			//check AMQ_ACCEPTORS value
			for _, envVar := range initContainer.Env {
				if envVar.Name == "AMQ_ACCEPTORS" {
					secretName := envVar.ValueFrom.SecretKeyRef.Name
					namespaceName := types.NamespacedName{
						Name:      secretName,
						Namespace: namespace,
					}
					//the value is a string of acceptors in xml format:
					//<acceptor name="new-acceptor">...</acceptor><another one>...
					//we need to locate our target acceptor and do the check
					//we use the port as a clue
					checkSecretHasCorrectKeyValue(secretName, namespaceName, envVar.ValueFrom.SecretKeyRef.Key, "trustStoreProvider=SUN")
				}
			}
			Expect(k8sClient.Delete(ctx, &cr)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(cr.Name, namespace, &cr), timeout, interval).Should(BeTrue())
		})
	})

	Context("With deployed controller", func() {
		It("verify old ver support", func() {
			By("By creating an old crd")
			ctx := context.Background()

			spec := brokerv2alpha4.ActiveMQArtemisSpec{}
			crd := brokerv2alpha4.ActiveMQArtemis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ActiveMQArtemis",
					APIVersion: brokerv2alpha4.GroupVersion.Identifier(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      randString(),
					Namespace: namespace,
				},
				Spec: spec,
			}

			crd.Spec.DeploymentPlan = brokerv2alpha4.DeploymentPlanType{
				Size:               1,
				PersistenceEnabled: true,
			}

			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			createdSs := &appsv1.StatefulSet{}
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdSs)
				return err == nil
			}, timeout, interval).Should(Equal(true))

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
	})

	Context("With deployed controller", func() {
		It("Checking storageClassName is configured", func() {

			if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
				By("By not requesting PVC that cannot be bound, seems to prevent minkube auto pv allocation")
				return
			}
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size:               1,
				PersistenceEnabled: true,
				Storage: brokerv1beta1.StorageType{
					StorageClassName: "some-storage-class",
				},
			}

			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			createdSs := &appsv1.StatefulSet{}
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdSs)
				return err == nil
			}, timeout, interval).Should(Equal(true))

			volumeTemplates := createdSs.Spec.VolumeClaimTemplates
			Expect(len(volumeTemplates)).To(Equal(1))

			storageClassName := volumeTemplates[0].Spec.StorageClassName
			Expect(*storageClassName).To(Equal("some-storage-class"))

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
	})

	It("populateValidatedUser", func() {

		ctx := context.Background()
		crd := generateArtemisSpec(namespace)
		crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       5,
		}
		crd.Spec.DeploymentPlan.Size = 1
		crd.Spec.DeploymentPlan.RequireLogin = true
		crd.Spec.BrokerProperties = []string{
			"securityEnabled=true",
			"rejectEmptyValidatedUser=true",
			"populateValidatedUser=true",
		}

		propLoginModules := make([]brokerv1beta1.PropertiesLoginModuleType, 1)
		pwd := "activemq"
		user := "Jay"
		moduleName := "prop-module"
		flag := "sufficient"
		propLoginModules[0] = brokerv1beta1.PropertiesLoginModuleType{
			Name: moduleName,
			Users: []brokerv1beta1.UserType{
				{Name: user,
					Password: &pwd,
					Roles:    []string{"jay-role"}},
			},
		}

		brokerDomainName := "activemq"
		loginModules := make([]brokerv1beta1.LoginModuleReferenceType, 1)
		loginModules[0] = brokerv1beta1.LoginModuleReferenceType{
			Name: &moduleName,
			Flag: &flag,
		}
		brokerDomain := brokerv1beta1.BrokerDomainType{
			Name:         &brokerDomainName,
			LoginModules: loginModules,
		}

		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

			By("Deploying security spec")
			_, deployedSecCrd := DeploySecurity("for-jay", namespace, func(secCrdToDeploy *brokerv1beta1.ActiveMQArtemisSecurity) {
				secCrdToDeploy.Spec.LoginModules.PropertiesLoginModules = propLoginModules
				secCrdToDeploy.Spec.SecurityDomains.BrokerDomain = brokerDomain
				secCrdToDeploy.Spec.SecuritySettings.Broker =
					[]brokerv1beta1.BrokerSecuritySettingType{
						{Match: "#",
							Permissions: []brokerv1beta1.PermissionType{
								{OperationType: "send", Roles: []string{"jay-role"}},
								{OperationType: "browse", Roles: []string{"jay-role"}},
							},
						},
					}
			})

			By("Deploying a broker")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Checking ready on SS")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: defaultNamespace}
				sfsFound := &appsv1.StatefulSet{}

				g.Expect(k8sClient.Get(ctx, key, sfsFound)).Should(Succeed())
				g.Expect(sfsFound.Status.ReadyReplicas).Should(BeEquivalentTo(1))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Sending 1")

			podWithOrdinal := namer.CrToSS(crd.Name) + "-0"
			command := []string{"amq-broker/bin/artemis", "producer", "--user", user, "--password", pwd, "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}

			Eventually(func(g Gomega) {
				stdOutContent := execOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)
				g.Expect(stdOutContent).Should(ContainSubstring("Produced: 1 messages"))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Consuming 1")

			command = []string{"amq-broker/bin/artemis", "browser", "--user", user, "--password", pwd, "--url", "tcp://" + podWithOrdinal + ":61616", "--destination", "queue://DLQ", "--verbose"}

			Eventually(func(g Gomega) {
				stdOutContent := execOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)

				Expect(stdOutContent).Should(ContainSubstring("messageID="))
				Expect(stdOutContent).Should(ContainSubstring("_AMQ_VALIDATED_USER=" + user))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			// cleanup
			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, deployedSecCrd)).Should(Succeed())

		}

	})

	It("populateValidatedUser as auto generated guest", func() {

		ctx := context.Background()
		crd := generateArtemisSpec(defaultNamespace)
		crd.Spec.DeploymentPlan.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 1,
			PeriodSeconds:       5,
		}
		crd.Spec.DeploymentPlan.Size = 1
		crd.Spec.BrokerProperties = []string{
			"securityEnabled=true",
			"rejectEmptyValidatedUser=true",
			"populateValidatedUser=true",
		}

		if os.Getenv("USE_EXISTING_CLUSTER") == "true" {

			By("Deploying a broker")
			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			By("Checking ready on SS")
			Eventually(func(g Gomega) {
				key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: defaultNamespace}
				sfsFound := &appsv1.StatefulSet{}

				g.Expect(k8sClient.Get(ctx, key, sfsFound)).Should(Succeed())
				g.Expect(sfsFound.Status.ReadyReplicas).Should(BeEquivalentTo(1))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Sending 1")

			podWithOrdinal := namer.CrToSS(crd.Name) + "-0"
			command := []string{"amq-broker/bin/artemis", "producer", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--message-count", "1", "--destination", "queue://DLQ", "--verbose"}

			Eventually(func(g Gomega) {
				stdOutContent := execOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)
				g.Expect(stdOutContent).Should(ContainSubstring("Produced: 1 messages"))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			By("Consuming 1")

			command = []string{"amq-broker/bin/artemis", "browser", "--user", "Jay", "--password", "activemq", "--url", "tcp://" + podWithOrdinal + ":61616", "--destination", "queue://DLQ", "--verbose"}

			Eventually(func(g Gomega) {
				stdOutContent := execOnPod(podWithOrdinal, crd.Name, defaultNamespace, command, g)

				// ...ActiveMQMessage[ID:3a356c2a-e7e1-11ec-ae1c-5ec4168c91b2]:PERSISTENT/ClientMessageImpl[messageID=19, durable=true, address=DLQ,userID=3a356c2a-e7e1-11ec-ae1c-5ec4168c91b2,properties=TypedProperties[__AMQ_CID=3a2f9fc7-e7e1-11ec-ae1c-5ec4168c91b2,_AMQ_ROUTING_TYPE=1,_AMQ_VALIDATED_USER=73ykuMrb,count=0,ThreadSent=Producer ActiveMQQueue[DLQ], thread=0]]
				Expect(stdOutContent).Should(ContainSubstring("messageID="))
				Expect(stdOutContent).Should(ContainSubstring("_AMQ_VALIDATED_USER="))
			}, existingClusterTimeout, existingClusterInterval).Should(Succeed())

			// cleanup
			k8sClient.Delete(ctx, &crd)

		}

	})

	Context("With deployed controller", func() {
		It("Checking jolokiaAgentOptions is configured", func() {
			By("By creating a new crd")
			ctx := context.Background()
			crd := generateArtemisSpec(namespace)

			jolokiaAgentOptions := "realm:test"

			crd.Spec.DeploymentPlan = brokerv1beta1.DeploymentPlanType{
				Size:                1,
				PersistenceEnabled:  true,
				JolokiaAgentEnabled: true,
				JolokiaAgentOptions: jolokiaAgentOptions,
			}

			Expect(k8sClient.Create(ctx, &crd)).Should(Succeed())

			Eventually(func() bool {
				key := types.NamespacedName{Name: crd.Name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, &crd)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			createdSs := &appsv1.StatefulSet{}
			Eventually(func() bool {
				key := types.NamespacedName{Name: namer.CrToSS(crd.Name), Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdSs)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("By checking the container stateful set for jolokia agent opts")
			Eventually(func() bool {
				found := false
				for _, container := range createdSs.Spec.Template.Spec.InitContainers {
					for _, env := range container.Env {
						if env.Name == "AMQ_JOLOKIA_AGENT_OPTS" {
							if strings.Contains(env.Value, jolokiaAgentOptions) {
								found = true
							}
						}
					}
				}
				return found
			}, timeout, interval).Should(BeTrue())

			Expect(k8sClient.Delete(ctx, &crd)).Should(Succeed())

			By("check it has gone")
			Eventually(checkCrdDeleted(crd.Name, namespace, &crd), timeout, interval).Should(BeTrue())
		})
	})
})

func generateArtemisSpec(namespace string) brokerv1beta1.ActiveMQArtemis {

	spec := brokerv1beta1.ActiveMQArtemisSpec{}

	toCreate := brokerv1beta1.ActiveMQArtemis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ActiveMQArtemis",
			APIVersion: brokerv1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      randString(),
			Namespace: namespace,
		},
		Spec: spec,
	}

	return toCreate
}

func generateOriginalArtemisSpec(namespace string, name string) *brokerv1beta1.ActiveMQArtemis {

	spec := brokerv1beta1.ActiveMQArtemisSpec{}

	toCreate := brokerv1beta1.ActiveMQArtemis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ActiveMQArtemis",
			APIVersion: brokerv1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}

	return &toCreate
}

func DeployBroker(brokerName string, targetNamespace string) (*brokerv1beta1.ActiveMQArtemis, *brokerv1beta1.ActiveMQArtemis) {
	ctx := context.Background()
	brokerCrd := generateOriginalArtemisSpec(targetNamespace, brokerName)

	Expect(k8sClient.Create(ctx, brokerCrd)).Should(Succeed())

	createdBrokerCrd := &brokerv1beta1.ActiveMQArtemis{}

	Eventually(func() bool {
		return getPersistedVersionedCrd(brokerCrd.ObjectMeta.Name, targetNamespace, createdBrokerCrd)
	}, timeout, interval).Should(BeTrue())
	Expect(createdBrokerCrd.Name).Should(Equal(createdBrokerCrd.ObjectMeta.Name))

	return brokerCrd, createdBrokerCrd

}

func randString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz")
	length := 6
	var b strings.Builder
	b.WriteString("broker-")
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func getPersistedVersionedCrd(name string, nameSpace string, object client.Object) bool {
	key := types.NamespacedName{Name: name, Namespace: nameSpace}
	err := k8sClient.Get(ctx, key, object)
	return err == nil
}

func checkCrdDeleted(name string, namespace string, crd client.Object) bool {
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, crd)
	return errors.IsNotFound(err)
}

func checkSecretHasCorrectKeyValue(secName string, ns types.NamespacedName, key string, expectedValue string) {
	Eventually(func() bool {
		secret, err := secrets.RetriveSecret(ns, secName, make(map[string]string), k8sClient)
		if err != nil {
			return false
		}
		data := secret.Data[key]
		return strings.Contains(string(data), expectedValue)
	}, timeout, interval).Should(BeTrue())
}
