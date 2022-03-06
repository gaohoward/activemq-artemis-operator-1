package encrypt

import (
	"context"
	"crypto/aes"
	"crypto/cipher"

	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/secrets"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/common"
	"github.com/artemiscloud/activemq-artemis-operator/pkg/utils/random"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const KEY_BYTES = "encBytes"
const KEY_SECRET = "encSecret"

var log = ctrl.Log.WithName("encrypt")

type PasswordEncrypter struct {
	SecretName types.NamespacedName
	TheSecret  *corev1.Secret
}

func (pe *PasswordEncrypter) Encrypt(plaintext string) (string, error) {
	secret, bytes, err := pe.GetSecretAndBytes()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}
	plainTextAsBytes := []byte(plaintext)

	cfb := cipher.NewCTR(block, bytes)
	cipherText := make([]byte, len(plainTextAsBytes))
	cfb.XORKeyStream(cipherText, plainTextAsBytes)
	return common.Base64Encode(cipherText), nil
}

func (pe *PasswordEncrypter) Decrypt(text string) (string, error) {
	secret, bytes, err := pe.GetSecretAndBytes()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}
	cipherText, err := common.Base64Decode(text)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCTR(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func (pe *PasswordEncrypter) GetSecretAndBytes() (string, []byte, error) {
	log.Info("get encrypt key and iv", "key", pe.TheSecret.Data[KEY_SECRET], "iv", pe.TheSecret.Data[KEY_BYTES])
	return string(pe.TheSecret.Data[KEY_SECRET]), pe.TheSecret.Data[KEY_BYTES], nil
}

func (pe *PasswordEncrypter) Init(clnt client.Client) error {
	log.Info("********************init encrypter")
	bytes := random.GenerateRandomBytes(16)
	keystring := random.GenerateRandomString(16)
	log.Info("8888 creating a random", "key", keystring, "bytes", bytes)
	rawData := make(map[string][]byte)
	rawData[KEY_BYTES] = bytes
	rawData[KEY_SECRET] = []byte(keystring)

	pe.TheSecret = secrets.NewBinarySecret(pe.SecretName, pe.SecretName.Name, rawData, nil)

	if clnt != nil {
		theSecret, err := secrets.RetriveSecret(pe.SecretName, pe.SecretName.Name, nil, clnt)
		if err != nil {
			if errors.IsNotFound(err) {
				log.V(1).Info("Creating the encrypter secret")
				err = clnt.Create(context.TODO(), pe.TheSecret, &client.CreateOptions{})
				return err
			}
		} else {
			log.Info("Using existing secret")
			pe.TheSecret = theSecret
			log.Info("existing key and bytes", "key", pe.TheSecret.Data[KEY_SECRET], "bytes", pe.TheSecret.Data[KEY_BYTES])
		}
	}
	return nil
}

func NewPasswordEncrypter(secretName types.NamespacedName, clnt client.Client) (*PasswordEncrypter, error) {
	encrypter := &PasswordEncrypter{
		SecretName: secretName,
	}
	log.Info("******** Initing encrypter")
	err := encrypter.Init(clnt)
	return encrypter, err
}
