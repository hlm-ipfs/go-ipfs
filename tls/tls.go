package tls

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	cmdhttp "github.com/ipfs/go-ipfs-cmds/http"
)

//go:embed server.key
var ServerKey string

//go:embed server.pem
var ServerPem string

//go:embed client.key
var ClientKey string

//go:embed client.pem
var ClientPem string

func init() {
	if Enable() {
		tlsConf, err := ClientTlsConfig()
		if err != nil {
			panic(err)
		}
		cmdhttp.TLSConfig = tlsConf
	}
}

func Enable() bool {
	/*if str, ok := os.LookupEnv("IPFS_DisableTls"); ok && strings.ToLower(str) == "true" {
		return false
	}
	return true*/
	return false
}

func ServerTlsConfig(clientAuth bool) (*tls.Config, error) {
	if len(ServerKey) == 0 || len(ServerPem) == 0 {
		return nil, errors.New("server cert invalid")
	}

	cert, err := tls.X509KeyPair([]byte(ServerPem), []byte(ServerKey))
	if err != nil {
		return nil, err
	}
	conf := &tls.Config{Certificates: []tls.Certificate{cert},MaxVersion: tls.VersionTLS12}
	if clientAuth {
		if len(ClientKey) == 0 || len(ClientPem) == 0 {
			return nil, errors.New("client cert invalid")
		}

		clientCertPool := x509.NewCertPool()
		ok := clientCertPool.AppendCertsFromPEM([]byte(ClientPem))
		if !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		conf.ClientCAs = clientCertPool
		conf.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return conf, nil
}

func ClientTlsConfig() (*tls.Config, error) {
	if len(ClientPem) == 0 || len(ClientKey) == 0 {
		return nil, errors.New("client cert invalid")
	}
	cert, err := tls.X509KeyPair([]byte(ClientPem), []byte(ClientKey))
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM([]byte(ClientPem))
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		MaxVersion: tls.VersionTLS12,
	}
	return conf, nil
}