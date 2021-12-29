package certs

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"log"
)

//go:embed server.key
var MTLSKey []byte

//go:embed server.pem
var MTLSPem []byte

//go:embed cacert.pem
var CACertPem []byte

func GetTLSConfig() *tls.Config {
	serverCert, err := tls.X509KeyPair(MTLSPem, MTLSKey)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(CACertPem) {
		log.Fatalf("Failed to append trusted certificate to certificate pool. %s.", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}
}
