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
	Connection *ssh.ServerConn
	Actual     RemoteForwardChannel
	IsPaused   bool
}

type MappingDisplayModel struct {
	SourceSubdomain string
	TargetPort      uint32
	IsConflict      bool
	Actual          MappingModel
}

func (m *MappingDisplayModel) TogglePause() {
	here.mappingsMu.Lock()
	defer here.mappingsMu.Unlock()

	mapping := here.mappings[m.SourceSubdomain]

	mapping.IsPaused = !m.Actual.IsPaused
	m.Actual.IsPaused = !m.Actual.IsPaused

	here.mappings[m.SourceSubdomain] = mapping
}

func (m *MappingDisplayModel) RenameSubdomain(newDomain string) {
	if m.SourceSubdomain == newDomain {
		return
	}

	here.mappingsMu.Lock()
	defer here.mappingsMu.Unlock()

	oldDomain := m.SourceSubdomain

	m.SourceSubdomain = newDomain
	_, isAlreadyExist := here.mappings[newDomain]

	if !m.IsConflict {
		delete(here.mappings, oldDomain)
	}

	if isAlreadyExist {
		m.IsConflict = true
		m.Actual.IsPaused = true
		return
	}

	m.IsConflict = false
	m.Actual.IsPaused = false
	here.mappings[newDomain] = MappingModel{
		Connection: m.Actual.Connection,
		Actual:     m.Actual.Actual,
		IsPaused:   m.Actual.IsPaused,
	}
}
