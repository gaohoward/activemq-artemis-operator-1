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
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	brokerv1beta1 "github.com/artemiscloud/activemq-artemis-operator/api/v1beta1"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/namer"
	coreV1 "k8s.io/api/core/v1"
)

var _ = Describe("artemis controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		namespace = "default"
		timeout   = time.Second * 10
		duration  = time.Second * 10
		interval  = time.Millisecond * 250
	)

	Context("With delopyed controller", func() {
		It("Expect pod desc", func() {
			name := "t1"
			By("By creating a new crd")
			ctx := context.Background()
			crd := &brokerv1beta1.ActiveMQArtemis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ActiveMQArtemis",
					APIVersion: brokerv1beta1.GroupVersion.Identifier(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "t1",
					Namespace: namespace,
				},
			}
			Expect(k8sClient.Create(ctx, crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			Eventually(func() bool {
				key := types.NamespacedName{Name: name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdCrd)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(name))

			// would like more status updates on createdCrd

			By("By checking the status of stateful set")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					fmt.Printf("Error getting ss: %v\n", err)
					return -1, err
				}
				fmt.Printf("CR: %v\n", createdSs)

				// presence is good enough... check on this status just for kicks
				return int(createdSs.Status.Replicas), err
			}, duration, interval).Should(Equal(0))

			By("Checking stopped status of CR because we expect it to fail to deploy")
			Eventually(func() (int, error) {
				key := types.NamespacedName{Name: name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdCrd)

				if err != nil {
					fmt.Printf("Error getting CR: %v\n", err)
					return -1, err
				}
				fmt.Printf("CR: %v\n", createdCrd)

				if len(createdCrd.Status.PodStatus.Stopped) > 0 {
					fmt.Printf("Stopped: %v\n", createdCrd.Status.PodStatus.Stopped[0])
				}
				return len(createdCrd.Status.PodStatus.Stopped), nil
			}, timeout, interval).Should(Equal(1))

		})

		It("Expect vol mount via config map", func() {
			name := "t2"
			By("By creating a new crd with broker props")
			ctx := context.Background()
			crd := &brokerv1beta1.ActiveMQArtemis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ActiveMQArtemis",
					APIVersion: brokerv1beta1.GroupVersion.Identifier(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: brokerv1beta1.ActiveMQArtemisSpec{
					BrokerProperties: map[string]string{
						"globalMaxSize": "512MB",
					},
				},
			}
			Expect(k8sClient.Create(ctx, crd)).Should(Succeed())

			createdCrd := &brokerv1beta1.ActiveMQArtemis{}

			Eventually(func() bool {
				key := types.NamespacedName{Name: name, Namespace: namespace}
				err := k8sClient.Get(ctx, key, createdCrd)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(createdCrd.Name).Should(Equal(name))

			By("By finding a new config map with broker props")
			configMap := &coreV1.ConfigMap{}
			nameOfConfigMap := namer.CrToBpCM(name)
			Eventually(func() bool {

				key := types.NamespacedName{Name: nameOfConfigMap, Namespace: namespace}
				err := k8sClient.Get(ctx, key, configMap)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(configMap.ObjectMeta.Name).Should(Equal(nameOfConfigMap))

			By("By checking the container of stateful set for java opts")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					fmt.Printf("Error getting ss: %v\n", err)
					return false, err
				}
				fmt.Printf("SS: %v\n", createdSs)

				found := false
				for _, container := range createdSs.Spec.Template.Spec.InitContainers {
					fmt.Printf("Container: %v\n", container.Name)
					for _, env := range container.Env {
						fmt.Printf("Env: %v\n", env)
						if env.Name == "JAVA_OPTS" {
							if strings.Contains(env.Value, "broker.properties") {
								found = true
							}
						}
					}
				}

				return found, err
			}, duration, interval).Should(Equal(true))

			By("By checking the container of stateful set for volume mount path")
			Eventually(func() (bool, error) {
				key := types.NamespacedName{Name: namer.CrToSS(createdCrd.Name), Namespace: namespace}
				createdSs := &appsv1.StatefulSet{}

				err := k8sClient.Get(ctx, key, createdSs)
				if err != nil {
					fmt.Printf("Error getting ss: %v\n", err)
					return false, err
				}
				fmt.Printf("SS: %v\n", createdSs)

				found := false
				for _, container := range createdSs.Spec.Template.Spec.Containers {
					fmt.Printf("Container: %v\n", container.Name)
					for _, vm := range container.VolumeMounts {
						fmt.Printf("Volume mount: %v\n", vm)
						// mount path can't have a .
						if strings.Contains(vm.MountPath, "broker-properties") {
							found = true
						}
					}
				}

				return found, err
			}, duration, interval).Should(Equal(true))

		})

	})
})
