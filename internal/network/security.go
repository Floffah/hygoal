package network

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

func NewSelfSignedServerCert(hosts ...string) (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			//CommonName:   "localhost",
			Organization: []string{"Hygoal Self-Signed"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour), // helps clock skew
		NotAfter:              time.Now().Add(30 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		//DNSNames:    []string{},
		//IPAddresses: []net.IP{},
	}

	//tmpl.DNSNames = append(tmpl.DNSNames, "localhost")

	//for _, h := range hosts {
	//	if ip := net.ParseIP(h); ip != nil {
	//		tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
	//	} else if h != "" {
	//		tmpl.DNSNames = append(tmpl.DNSNames, h)
	//	}
	//}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func NewQUICServerTLSConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},

		NextProtos: []string{"hytale/1"},
	}
}
