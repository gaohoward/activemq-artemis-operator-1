package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// extra kinds
const (
	RouteKind              = "Route"
	OpenShiftAPIServerKind = "OpenShiftAPIServer"
)

var theManager manager.Manager
var theEncrypter Encrypter
var commonLog = ctrl.Log.WithName("common")

type Encrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(text string) (string, error)
}

type ValueInfo struct {
	Name    string
	value   string
	AutoGen bool
	Encrypt bool
}

//debug
func (v *ValueInfo) DumpInfo() {
	fmt.Println("****** begin ValueInfo " + v.Name + "******")
	fmt.Printf("= AutoGen? %v\n", v.AutoGen)
	fmt.Printf("= encrypted enabled: %v\n", v.Encrypt)
	fmt.Printf("= value: %v\n", v.value)
	if v.Encrypt {
		plain, err := v.GetPlainValue()
		fmt.Printf("= decrypted value: %v and any err? %v\n", plain, err)
	}
	fmt.Println("****** end ValueInfo " + v.Name + "******")
}

func (v *ValueInfo) GetValue() string {
	return v.value
}

func (v *ValueInfo) IsValueEmpty() bool {
	if plainText, err := v.GetPlainValue(); err == nil {
		return plainText == ""
	}
	return false
}

func (v *ValueInfo) SetValue(newValue string) error {
	commonLog.Info("Set value for "+v.Name, "newval", newValue, "enc", v.Encrypt)
	if v.Encrypt {
		cipherText, err := theEncrypter.Encrypt(newValue)
		if err != nil {
			return err
		}
		v.value = cipherText
		commonLog.Info("set encrypted value", "vsl", v.value)
	} else {
		v.value = newValue
		commonLog.Info("plaintext value", "vsl", v.value)
	}
	return nil
}

func (v *ValueInfo) GetPlainValue() (string, error) {
	commonLog.Info("Getting plain value for "+v.Name, "enc", v.Encrypt)
	if v.Encrypt {
		plainValue, err := theEncrypter.Decrypt(v.value)
		if err != nil {
			return "", err
		}
		commonLog.Info("returning decripted " + plainValue)
		return plainValue, nil
	}
	commonLog.Info("returning non enc " + v.value)
	return v.value, nil
}

func NewValueInfo(name string, val string, gen bool, isEncrypt bool) (*ValueInfo, error) {
	commonLog.Info("Creating a new value info "+name, "value", val, "gen", gen, "enc", isEncrypt)
	if theEncrypter == nil {
		commonLog.Info("The encrypter is not set up, no encryption")
		isEncrypt = false
	}
	valueInfo := ValueInfo{
		Name:    name,
		value:   val,
		AutoGen: gen,
		Encrypt: isEncrypt,
	}
	if isEncrypt {
		var err error
		commonLog.Info("Encript value", "val", valueInfo.value)
		valueInfo.value, err = theEncrypter.Encrypt(valueInfo.value)
		if err != nil {
			return nil, err
		}
		commonLog.Info("valued encrypted", "now", valueInfo.value)
	}
	return &valueInfo, nil
}

func compareQuantities(resList1 corev1.ResourceList, resList2 corev1.ResourceList, keys []corev1.ResourceName) bool {

	for _, key := range keys {
		if q1, ok1 := resList1[key]; ok1 {
			if q2, ok2 := resList2[key]; ok2 {
				if q1.Cmp(q2) != 0 {
					return false
				}
			} else {
				return false
			}
		} else {
			if _, ok2 := resList2[key]; ok2 {
				return false
			}
		}
	}
	return true
}

func CompareRequiredResources(res1 *corev1.ResourceRequirements, res2 *corev1.ResourceRequirements) bool {

	resNames := []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceStorage, corev1.ResourceEphemeralStorage}
	if !compareQuantities(res1.Limits, res2.Limits, resNames) {
		return false
	}

	if !compareQuantities(res1.Requests, res2.Requests, resNames) {
		return false
	}
	return true
}

func ToJson(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err == nil {
		return string(bytes), nil
	}
	return "", err
}

func FromJson(jsonStr *string, obj interface{}) error {
	return json.Unmarshal([]byte(*jsonStr), obj)
}

func SetManager(mgr manager.Manager) {
	theManager = mgr
}

func GetManager() manager.Manager {
	return theManager
}

func GetEncrypter() Encrypter {
	return theEncrypter
}

func SetEncrypter(enc Encrypter) {
	theEncrypter = enc
}

func Base64Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
