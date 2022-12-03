package internal

import (
	"crypto/tls"
	"embed"
	"io/ioutil"
)

//go:embed default.*
var fs embed.FS

func GetDefaultCertificate() (noCert tls.Certificate, err error) {
	cert, err := fs.Open("default.crt")
	if err != nil {
		return noCert, err
	}

	certBytes, err := ioutil.ReadAll(cert)
	if err != nil {
		return noCert, err
	}

	key, err := fs.Open("default.key")
	if err != nil {
		return noCert, err
	}

	keyBytes, err := ioutil.ReadAll(key)
	if err != nil {
		return noCert, err
	}

	return tls.X509KeyPair(certBytes, keyBytes)
}

func GetDefaultTLSConfig() (*tls.Config, error) {
	pair, err := GetDefaultCertificate()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{pair},
	}, nil
}
