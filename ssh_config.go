package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"
	"time"

	"github.com/charmbracelet/log"
)

func (here *HereServer) getSSHAddr() string {
	addr, ok := os.LookupEnv("SSH_LISTEN_ADDR")
	if !ok {
		return ":2222"
	}

	return addr
}

func (here *HereServer) getSSHPassword() string {
	password, ok := os.LookupEnv("SSH_PASSWORD")
	if !ok {
		return ""
	}

	return password
}

func (here *HereServer) isPasswordRequired() bool {
	return here.getSSHPassword() != ""
}

func (here *HereServer) getUnauthenticatedTimeout() time.Duration {
	timeoutStr, ok := os.LookupEnv("UNAUTHENTICATED_TIMEOUT")
	if !ok {
		return 30 * time.Minute
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		log.Warn("Invalid UNAUTHENTICATED_TIMEOUT value '%s', using default 30m: %v", timeoutStr, err)
		return 30 * time.Minute
	}

	return timeout
}

func (here *HereServer) getSSHHostKey() []byte {
	hostKeyPath := path.Join(here.getDataPath(), "hostkey")
	hostKeyInfo, err := os.Stat(hostKeyPath)
	if err != nil || hostKeyInfo.IsDir() {
		log.Warn("no hostkey found, try to create...")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatal(err)
		}

		privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			log.Fatal(err)
		}

		privateKeyBlock := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyDER,
		}
		privateKeyPEM := pem.EncodeToMemory(privateKeyBlock)

		err = os.WriteFile(hostKeyPath, privateKeyPEM, 0o600)
		if err != nil {
			log.Fatal(err)
		}
	}

	hostKey, err := os.ReadFile(hostKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	return hostKey
}
