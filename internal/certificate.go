package internal

import (
    "embed"
    "crypto/tls"
    "io/ioutil"
)

//go:embed default.*
var fs embed.FS

func GetDefaultTLSConfig() (*tls.Config, error) {
    cert, err := fs.Open("default.crt")
    if err != nil {
        return nil, err
    }

    certBytes, err := ioutil.ReadAll(cert)
    if err != nil {
        return nil, err
    }

    key, err := fs.Open("default.key")
    if err != nil {
        return nil, err
    }

    keyBytes, err := ioutil.ReadAll(key)
    if err != nil {
        return nil, err
    }

    pair, err := tls.X509KeyPair(certBytes, keyBytes)
    if err != nil {
        return nil, err
    }

    return &tls.Config{
        Certificates: []tls.Certificate{pair},
    }, nil
}
