package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

func MakeSelfSignedKeyStore(password string) (*x509.Certificate, []byte, error) {
	fmt.Println("======setting up cert")
	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	if err := caPrivKey.Validate(); err != nil {
		return nil, nil, err
	}

	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"www.artemiscloud.io"},
			Country:       []string{"US"},
			Province:      []string{"MA"},
			Locality:      []string{"Westford"},
			StreetAddress: []string{"Littleton Rd"},
			PostalCode:    []string{"01886"},
			CommonName:    "artemiscloud",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
		IsCA:      true,
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		panic(err)
	}

	cert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		panic(err)
	}

	pfxBytes, err := pkcs12.Encode(rand.Reader, caPrivKey, cert, []*x509.Certificate{}, password)
	if err != nil {
		panic(err)
	}

	return cert, pfxBytes, nil
}

func MakeSelfSignedTrustStore(cert *x509.Certificate, password string) ([]byte, error) {

	pfxBytes, err := pkcs12.EncodeTrustStore(rand.Reader, []*x509.Certificate{cert}, string(password))

	if err != nil {
		panic(err)
	}

	if err != nil {
		return nil, err
	}

	return pfxBytes, nil

}
