package server

import (
	"crypto/tls"

	"k8s.io/klog/v2"
)

type Config struct {
	CertFile string
	KeyFile  string
}

/*
 * This function was originally copied from the Kubernetes project (https://github.com/kubernetes/kubernetes)
 * and is licensed under Apache License 2.0.
 */
func configTLS(config Config) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
