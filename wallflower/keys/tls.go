package keys

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

func (sk SK) TLS() (cert tls.Certificate, err error) {
	// fuck the secret key, as long as it's d e t e r m i n i s t i c
	skr := bytes.NewReader(append(sk[:], sk[:]...)) // long enough lmaoooo

	esk, err := ecdsa.GenerateKey(elliptic.P384(), skr)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate ECDSA key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour) // 1 day validity

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate serial number: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"wallflower"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		IPAddresses: []net.IP{net.IPv6loopback},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &esk.PublicKey, esk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create X.509 certificate: %v", err)
	}

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	skBytes, err := x509.MarshalECPrivateKey(esk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal ECDSA private key: %v", err)
	}
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: skBytes})

	if cert, err = tls.X509KeyPair(certPem, keyPem); err != nil {
		return tls.Certificate{}, fmt.Errorf("create TLS certificate: %v", err)
	}
	return
}
