package chat

import (
	"sync"

	"campus-agent/internal/ai/client"
)

const DefaultMaxWindowSize = 10

type SessionMemory struct {
	mu           sync.Mutex
	Messages     []client.ChatMessage
	MaxWindowSize int
}

// sessionStore holds all active sessions.
var sessionStore sync.Map

func loadOrCreateSession(id string) *SessionMemory {
	val, ok := sessionStore.Load(id)
	if ok {
		return val.(*SessionMemory)
	}

	mem := &SessionMemory{
		Messages:      make([]client.ChatMessage, 0, DefaultMaxWindowSize),
		MaxWindowSize: DefaultMaxWindowSize,
	}
	sessionStore.Store(id, mem)
	return mem
}

func (m *SessionMemory) append(userMsg, assistantMsg client.ChatMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Messages = append(m.Messages, userMsg, assistantMsg)
	if len(m.Messages) > m.MaxWindowSize {
		m.Messages = m.Messages[len(m.Messages)-m.MaxWindowSize:]
	}
}

func (m *SessionMemory) snapshot() []client.ChatMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make([]client.ChatMessage, len(m.Messages))
	copy(cp, m.Messages)
	return cp
}
