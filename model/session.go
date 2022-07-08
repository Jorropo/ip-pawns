package model

import (
	"fmt"
	"net"

	"github.com/hashicorp/yamux"
)

type Session struct {
	yamux *yamux.Session
}

type ForwarderServer struct {
	sessions map[string]*Session
}

func CreateSessionManager() *ForwarderServer {
	server := &ForwarderServer{
		make(map[string]*Session),
	}
	return server
}

func (sm *ForwarderServer) CheckAccess(user, password string) bool {
	if user == "api" && password == "apipass" {
		return true
	} else if sm.sessions[user] != nil && password == "sesionpass" {
		return true
	}
	return false
}

func (sm *ForwarderServer) HandleConnection(conn net.Conn) error {
	addr := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	if _, ok := sm.sessions[addr]; ok {
		return fmt.Errorf("concurrent connection detected")
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		return err
	}

	sm.sessions[addr] = &Session{
		session,
	}
	fmt.Printf("%s connected!\n", addr)
	go func() {
		<-session.CloseChan()
		delete(sm.sessions, addr)
		fmt.Printf("%s disconnected\n", addr)
	}()
	return nil
}

func (sm *ForwarderServer) SessionCount() int {
	return len(sm.sessions)
}

func (sm *ForwarderServer) Sessions() map[string]*Session {
	return sm.sessions
}

func (sm *ForwarderServer) OpenConnection(sourceAddress string, destinationAddress string) (net.Conn, error) {
	session := sm.sessions[sourceAddress]
	if session != nil {
		conn, err := session.yamux.Open()
		fmt.Printf("%s -> %s\n", sourceAddress, destinationAddress)
		if err != nil {
			return nil, err
		}

		err = writeConnectionAddress(conn, destinationAddress)
		if err != nil {
			return nil, err
		}

		err = readError(conn)
		if err != nil {
			return nil, err
		}

		return conn, nil
	} else {
		return nil, fmt.Errorf("no such connection")
	}
}
