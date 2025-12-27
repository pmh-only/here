package main

import (
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshNetConn struct {
	ch ssh.Channel
}

func (c *sshNetConn) Read(p []byte) (int, error)  { return c.ch.Read(p) }
func (c *sshNetConn) Write(p []byte) (int, error) { return c.ch.Write(p) }
func (c *sshNetConn) Close() error                { return c.ch.Close() }
func (c *sshNetConn) LocalAddr() net.Addr         { return dummyAddr("ssh-local") }
func (c *sshNetConn) RemoteAddr() net.Addr        { return dummyAddr("ssh-remote") }

func (c *sshNetConn) SetDeadline(t time.Time) error      { return nil }
func (c *sshNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *sshNetConn) SetWriteDeadline(t time.Time) error { return nil }

type dummyAddr string

func (a dummyAddr) Network() string { return "ssh" }
func (a dummyAddr) String() string  { return string(a) }

func oneShotDial(c net.Conn) func(ctx context.Context, network, addr string) (net.Conn, error) {
	var used bool
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if used {
			return nil, errors.New("upstream connection already used")
		}
		used = true
		return c, nil
	}
}
