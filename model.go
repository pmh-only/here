package main

import "golang.org/x/crypto/ssh"

type RemoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

type RemoteForwardSuccess struct {
	BindPort uint32
}

type RemoteForwardChannel struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

type MappingModel struct {
	connection *ssh.ServerConn
	channel    RemoteForwardChannel
}
