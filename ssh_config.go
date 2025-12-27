package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

func (here *HereServer) getSSHAddr() string {
	addr, ok := os.LookupEnv("SSH_LISTEN_ADDR")
	if !ok {
		return ":2222"
	}

	return addr
}

func (here *HereServer) getSSHHostKey() (ssh.Signer, error) {
	hostKeyPath := path.Join(here.getDataPath(), "hostkey")
	hostKeyInfo, err := os.Stat(hostKeyPath)
	if err != nil || hostKeyInfo.IsDir() {
		log.Println("no hostkey found, try to create...")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		privateKeyBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyDER,
		}
		privateKeyPEM := pem.EncodeToMemory(privateKeyBlock)

		err = os.WriteFile(hostKeyPath, privateKeyPEM, 0o600)
		if err != nil {
			return nil, err
		}
	}

	hostKey, err := os.ReadFile(hostKeyPath)
	if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(hostKey)
}
