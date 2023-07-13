/*
Copyright 2021.

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

// move the methods out and cancel this package!
package certutil

import (
	"fmt"
	"strings"

	"github.com/artemiscloud/activemq-artemis-operator/pkg/resources/environments"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	cert_annotation_key   = "cert-manager.io/issuer-name"
	bundle_annotation_key = "trust.cert-manager.io/hash"
	console_web_prefix    = "webconfig.bindings.artemis."
)

var defaultKeyStorePassword = "password"

type SslArguments struct {
	KeyStoreType       string
	KeyStorePath       string
	KeyStorePassword   *string
	TrustStoreType     string
	TrustStorePath     string
	TrustStorePassword *string
	PemCfgs            []string
	IsConsole          bool
}

func (s *SslArguments) ToSystemProperties(currentSS *appsv1.StatefulSet) string {
	sslFlags := ""

	if s.KeyStorePath != "" {
		if s.KeyStoreType == "PEM" {

			amqName := environments.Retrieve(currentSS.Spec.Template.Spec.InitContainers, "AMQ_NAME")
			ksPath := "/home/jboss/" + amqName.Value + "/etc/" + s.KeyStorePath
			sslFlags = "-D" + console_web_prefix + "keyStorePath=" + ksPath
		} else {
			sslFlags = "-D" + console_web_prefix + "keyStorePath=" + s.KeyStorePath
		}
	}

	if s.KeyStorePassword != nil {
		sslFlags = sslFlags + " -D" + console_web_prefix + "keyStorePassword=" + *s.KeyStorePassword
	}

	if s.KeyStoreType == "PEM" {
		sslFlags = sslFlags + " -D" + console_web_prefix + "keyStoreType=PEMCFG"
	} else if s.KeyStoreType != "" {
		sslFlags = sslFlags + " -D" + console_web_prefix + "keyStoreType=" + s.KeyStoreType
	}

	sslFlags = sslFlags + " -D" + console_web_prefix + "trustStorePath=" + s.TrustStorePath
	if s.TrustStorePassword != nil {
		sslFlags = sslFlags + " -D" + console_web_prefix + "trustStorePassword=" + *s.TrustStorePassword
	}
	if s.TrustStoreType != "" {
		sslFlags = sslFlags + " -D" + console_web_prefix + "trustStoreType=" + s.TrustStoreType
	}

	sslFlags = sslFlags + " -D" + console_web_prefix + "uri=" + getConsoleUri()
	return sslFlags
}

func getConsoleUri() string {
	return "https://FQ_HOST_NAME:8161"
}

func (s *SslArguments) ToFlags() string {
	sslFlags := ""
	if s.IsConsole {
		sslFlags = sslFlags + " " + "--ssl-key" + " " + s.KeyStorePath
		if s.KeyStorePassword != nil {
			sslFlags = sslFlags + " " + "--ssl-key-password" + " " + *s.KeyStorePassword
		}
		sslFlags = sslFlags + " " + "--ssl-trust" + " " + s.TrustStorePath
		if s.TrustStorePassword != nil {
			sslFlags = sslFlags + " " + "--ssl-trust-password" + " " + *s.TrustStorePassword
		}
		return sslFlags
	}

	sslFlags = "sslEnabled=true"
	sslFlags = sslFlags + ";" + "keyStorePath=" + s.KeyStorePath
	if s.KeyStorePassword != nil {
		sslFlags = sslFlags + ";" + "keyStorePassword=" + *s.KeyStorePassword
	}

	if s.KeyStoreType == "PEM" {
		sslFlags = sslFlags + ";" + "keyStoreType=PEMCFG"
	} else {
		sslFlags = sslFlags + ";" + "keyStoreType=" + s.KeyStoreType
	}

	sslFlags = sslFlags + ";" + "trustStorePath=" + s.TrustStorePath
	if s.TrustStorePassword != nil {
		sslFlags = sslFlags + ";" + "trustStorePassword=" + *s.TrustStorePassword
	}
	sslFlags = sslFlags + ";" + "trustStoreType=" + s.TrustStoreType

	return sslFlags
}

func IsSecretFromCert(secret *corev1.Secret) bool {
	_, exist := secret.Annotations[cert_annotation_key]
	return exist
}

func isSecretFromBundle(secret *corev1.Secret) bool {
	_, exist := secret.Annotations[bundle_annotation_key]
	return exist
}

func getBundleNameFromSecret(secret *corev1.Secret) string {
	//extract the key of the secret's only entry
	bundleName := ""
	for key := range secret.Data {
		bundleName = key
		break
	}
	return bundleName
}

func GetSslArgumentsFromSecret(sslSecret *corev1.Secret, keyStoreType string, trustStoreType string, trustSecret *corev1.Secret, isConsole bool) (*SslArguments, error) {
	sslArgs := SslArguments{
		IsConsole: isConsole,
	}

	if keyStoreType != "" {
		sslArgs.KeyStoreType = keyStoreType
	}

	sep := "/"
	if !isConsole {
		sep = "\\/"
	}

	volumeDir := sep + "etc" + sep + sslSecret.Name + "-volume"

	if sslArgs.KeyStoreType == "PEM" {
		uniqueName := sslSecret.Name + ".pemcfg"
		sslArgs.KeyStorePath = uniqueName
		sslArgs.PemCfgs = []string{
			sslArgs.KeyStorePath,
			"/etc/" + sslSecret.Name + "-volume/tls.key",
			"/etc/" + sslSecret.Name + "-volume/tls.crt",
		}
	} else {
		// if it is the cert-secret, we throw an error
		if IsSecretFromCert(sslSecret) {
			return nil, fmt.Errorf("certificate only supports PEM keystore type")
		}

		// old user secret
		sslArgs.KeyStorePassword = &defaultKeyStorePassword
		sslArgs.KeyStorePath = volumeDir + sep + "broker.ks"
		if passwordString := string(sslSecret.Data["keyStorePassword"]); passwordString != "" {
			if !isConsole {
				passwordString = strings.ReplaceAll(passwordString, "/", sep)
			}
			sslArgs.KeyStorePassword = &passwordString
		}
		if keyPathString := string(sslSecret.Data["keyStorePath"]); keyPathString != "" {
			if !isConsole {
				keyPathString = strings.ReplaceAll(keyPathString, "/", sep)
			}
			sslArgs.KeyStorePath = keyPathString
		}
	}

	sslArgs.TrustStoreType = trustStoreType

	if trustSecret == nil {
		trustSecret = sslSecret
	}

	trustVolumeDir := sep + "etc" + sep + trustSecret.Name + "-volume"

	if isSecretFromBundle(trustSecret) {
		bundleName := getBundleNameFromSecret(trustSecret)
		sslArgs.TrustStorePath = trustVolumeDir + sep + bundleName
	} else {
		//old user Secret
		sslArgs.TrustStorePassword = &defaultKeyStorePassword
		sslArgs.TrustStorePath = trustVolumeDir + sep + "client.ts"
		if trustPassword := string(trustSecret.Data["trustStorePassword"]); trustPassword != "" {
			if !isConsole {
				trustPassword = strings.ReplaceAll(trustPassword, "/", sep)
			}
			sslArgs.TrustStorePassword = &trustPassword
		}
		if trustStorePath := string(trustSecret.Data["trustStorePath"]); trustStorePath != "" {
			if !isConsole {
				trustStorePath = strings.ReplaceAll(trustStorePath, "/", sep)
			}
			sslArgs.TrustStorePath = trustStorePath
		}
	}

	return &sslArgs, nil
}
