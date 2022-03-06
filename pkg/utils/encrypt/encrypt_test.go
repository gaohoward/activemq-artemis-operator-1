package encrypt

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestConfigUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryption Test Suite")
}

var _ = BeforeSuite(func() {
	fmt.Println("=======Before Encrypt Suite========")
})

var _ = AfterSuite(func() {
	fmt.Println("=======After Encrypt Suite========")
})

var _ = Describe("Config Util Test", func() {
	Context("Test Encryption", func() {
		secretName := types.NamespacedName{
			Name:      "secret",
			Namespace: "operator",
		}
		encrypter, err := NewPasswordEncrypter(secretName, nil)
		Expect(err).To(BeNil())
		It("Encryption and Decryption", func() {
			password := "my49sla-djd"
			ciphertext, err := encrypter.Encrypt(password)
			Expect(err).To(BeNil())
			Expect(ciphertext).NotTo(Equal(password))
			deciphered, err := encrypter.Decrypt(ciphertext)
			Expect(err).To(BeNil())
			Expect(deciphered).To(Equal(password))
		})
	})
})
