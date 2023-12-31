package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

const (
	BitSize        = 4096
	KeyPath        = "internal/util/keys"
	PublicKeyFile  = KeyPath + "/public.pub"
	PrivateKeyFile = KeyPath + "/private.pub"
)

func generatePrivateKey(size int) (*rsa.PrivateKey, error) {
	prk, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, err
	}
	if err = prk.Validate(); err != nil {
		return nil, err
	}
	return prk, nil
}
func generatePublicKey(prk *rsa.PublicKey) ([]byte, error) {
	pbk, err := ssh.NewPublicKey(prk)
	if err != nil {
		return nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(pbk)
	return pubKeyBytes, nil
}
func encodePrivateKeyToPEM(prk *rsa.PrivateKey) []byte {
	privDER := x509.MarshalPKCS1PrivateKey(prk)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}
func SeedRS512Keys() error {
	privateKey, err := generatePrivateKey(BitSize)
	if err != nil {
		return err
	}
	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	if _, err := os.Stat(KeyPath); os.IsNotExist(err) {
		if err = os.Mkdir(KeyPath, os.ModePerm); err != nil {
			return err
		}
	}
	if err := os.WriteFile(PublicKeyFile, publicKeyBytes, 0666); err != nil {
		return err
	}
	if err := os.WriteFile(PrivateKeyFile, privateKeyBytes, 0666); err != nil {
		return err
	}
	return nil
}
func GetKeyPair() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	pubFile, err := os.OpenFile(PublicKeyFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, nil, err
	}
	pubBytes, err := io.ReadAll(pubFile)
	if err != nil {
		return nil, nil, err
	}

	privFile, err := os.OpenFile(PrivateKeyFile, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, nil, err
	}
	privBytes, err := io.ReadAll(privFile)
	if err != nil {
		return nil, nil, err
	}

	res, _, _, _, err := ssh.ParseAuthorizedKey(pubBytes)
	if err != nil {
		return nil, nil, err
	}
	parsedCryptoKey := res.(ssh.CryptoPublicKey)
	pubCrypto := parsedCryptoKey.CryptoPublicKey()
	pub := pubCrypto.(*rsa.PublicKey)

	block, _ := pem.Decode(privBytes)
	priv, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	return pub, priv, nil
}
