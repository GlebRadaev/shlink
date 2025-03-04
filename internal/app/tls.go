package app

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// GenerateCertificate generates a self-signed TLS certificate and a private key.
func GenerateCertificate(certPath, keyPath string) error {
	certDir := filepath.Dir(certPath)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории для сертификатов: %v", err)
	}

	if _, err := os.Stat(certPath); err == nil {
		log.Println("Сертификат уже существует:", certPath)
		return nil
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"MyApp"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		SubjectKeyId: []byte{1, 2, 3, 4, 5},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("ошибка генерации ключа: %v", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("ошибка создания сертификата: %v", err)
	}

	var certPEM, keyPEM bytes.Buffer
	if err := pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("ошибка кодирования сертификата в PEM: %v", err)
	}

	if err := pem.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return fmt.Errorf("ошибка кодирования ключа в PEM: %v", err)
	}

	if err := os.WriteFile(certPath, certPEM.Bytes(), 0644); err != nil {
		return fmt.Errorf("ошибка записи сертификата: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM.Bytes(), 0600); err != nil {
		return fmt.Errorf("ошибка записи ключа: %v", err)
	}

	log.Println("Сертификаты успешно сгенерированы.")
	return nil
}
