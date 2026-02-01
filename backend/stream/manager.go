package stream

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Manager struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func NewManager() *Manager {
	return &Manager{conns: map[*websocket.Conn]struct{}{}}
}

func (m *Manager) Register(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[conn] = struct{}{}
}

func (m *Manager) Unregister(conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, conn)
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for conn := range m.conns {
		_ = conn.Close()
		delete(m.conns, conn)
	}
}
