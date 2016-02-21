package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"io/ioutil"
	"log"

	"github.com/goburrow/netforward"
)

type tlsConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string

	SkipVerify bool
}

// addCert adds certificate and private key file.
func addCert(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	config.Certificates = append(config.Certificates, cert)
	return nil
}

// addCertAuthorities adds Root CAs from file.
func addCertAuthorities(config *tls.Config, caFile string) error {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}
	var caPool *x509.CertPool
	if config.RootCAs == nil {
		caPool = x509.NewCertPool()
	} else {
		caPool = config.RootCAs
	}
	if !caPool.AppendCertsFromPEM(caCert) {
		return errors.New("could not append certs")
	}
	if config.RootCAs == nil {
		config.RootCAs = caPool
	}
	return nil
}

// addClientCertAuthorities adds client CAs from file.
func addClientCertAuthorities(config *tls.Config, caFile string) error {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}
	var caPool *x509.CertPool
	if config.ClientCAs == nil {
		caPool = x509.NewCertPool()
	} else {
		caPool = config.ClientCAs
	}
	if !caPool.AppendCertsFromPEM(caCert) {
		return errors.New("could not append certs")
	}
	if config.ClientCAs == nil {
		config.ClientCAs = caPool
	}
	config.ClientAuth = tls.RequireAndVerifyClientCert
	return nil
}

var (
	local, remote netforward.Endpoint
)

func parseArgs() error {
	var localTLS, remoteTLS tlsConfig

	flag.StringVar(&local.Network, "network", "tcp", "network protocol")
	flag.StringVar(&local.Address, "address", "localhost:7000", "listen address")
	flag.StringVar(&localTLS.CertFile, "certFile", "", "certificate file")
	flag.StringVar(&localTLS.KeyFile, "keyFile", "", "certificate key file")
	flag.StringVar(&localTLS.CAFile, "caFile", "", "client certificate authorities file")

	flag.StringVar(&remote.Network, "remote.network", "tcp", "network protocol")
	flag.StringVar(&remote.Address, "remote.address", "localhost:8000", "remote address")
	flag.StringVar(&remoteTLS.CertFile, "remote.certFile", "", "certificate file")
	flag.StringVar(&remoteTLS.KeyFile, "remote.keyFile", "", "certificate key file")
	flag.StringVar(&remoteTLS.CAFile, "remote.caFile", "", "server certificate authorities file")
	flag.BoolVar(&remoteTLS.SkipVerify, "remote.skipVerify", false, "Not to verify remote server certificate")

	flag.Parse()

	var err error
	if localTLS.CertFile != "" || localTLS.CAFile != "" {
		local.TLS = &tls.Config{}
		if localTLS.CertFile != "" {
			err = addCert(local.TLS, localTLS.CertFile, localTLS.KeyFile)
			if err != nil {
				return err
			}
		}
		if localTLS.CAFile != "" {
			addClientCertAuthorities(local.TLS, localTLS.CAFile)
			if err != nil {
				return err
			}
		}
	}
	if remoteTLS.CertFile != "" || remoteTLS.CAFile != "" {
		remote.TLS = &tls.Config{}
		if remoteTLS.CertFile != "" {
			err = addCert(remote.TLS, remoteTLS.CertFile, remoteTLS.KeyFile)
			if err != nil {
				return err
			}
		}
		if remoteTLS.CAFile != "" {
			err = addCertAuthorities(remote.TLS, remoteTLS.CertFile)
			if err != nil {
				return err
			}
		}
		if remoteTLS.SkipVerify {
			remote.TLS.InsecureSkipVerify = true
		}
	}
	return nil
}

func main() {
	err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}
	f := netforward.NetForwarder{
		Local: local,
	}
	err = f.Listen()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("forwarding %s://%s->%s://%s", local.Network, local.Address, remote.Network, remote.Address)
	defer f.Close()
	err = f.Forward(&remote)
	if err != nil {
		log.Fatal(err)
	}
}
