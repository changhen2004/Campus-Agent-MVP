package chat

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

const DefaultMaxWindowSize = 10

// SessionMemory holds the sliding-window message history for a session.
type SessionMemory struct {
	mu            sync.Mutex
	Messages      []*schema.Message
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
		Messages:      make([]*schema.Message, 0, DefaultMaxWindowSize),
		MaxWindowSize: DefaultMaxWindowSize,
	}
	sessionStore.Store(id, mem)
	return mem
}

func (m *SessionMemory) append(userMsg, assistantMsg *schema.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Messages = append(m.Messages, userMsg, assistantMsg)
	if len(m.Messages) > m.MaxWindowSize {
		m.Messages = m.Messages[len(m.Messages)-m.MaxWindowSize:]
	}
}

func (m *SessionMemory) snapshot() []*schema.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make([]*schema.Message, len(m.Messages))
	copy(cp, m.Messages)
	return cp
}
