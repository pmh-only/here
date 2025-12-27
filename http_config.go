package main

import (
	"os"
)

func (here *HereServer) getHTTPAddr() string {
	addr, ok := os.LookupEnv("HTTP_LISTEN_ADDR")
	if !ok {
		return ":8080"
	}

	return addr
}

func (here *HereServer) getHostSuffix() string {
	addr, ok := os.LookupEnv("HTTP_HOST_SURFFIX")
	if !ok {
		return ".local"
	}

	return addr
}

func (here *HereServer) getHostPerfix() string {
	addr, ok := os.LookupEnv("HTTP_HOST_PREFIX")
	if !ok {
		return "http://"
	}

	return addr
}
