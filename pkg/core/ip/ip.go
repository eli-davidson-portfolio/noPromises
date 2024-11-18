// Package ip implements Information Packets for Flow-Based Programming
package ip

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Type represents the type of Information Packet
type Type int

const (
	TypeNormal  Type = iota
	TypeInitial      // IIP
	TypeBracketOpen
	TypeBracketClose
)

// IP represents an Information Packet with type safety and metadata
type IP[T any] struct {
	id        string
	ipType    Type
	data      T
	metadata  map[string]any
	owner     string
	immutable bool
	mu        sync.RWMutex
}

// New creates a new Information Packet with the given data
func New[T any](data T) *IP[T] {
	return &IP[T]{
		id:       uuid.New().String(),
		ipType:   TypeNormal,
		data:     data,
		metadata: makeInitialMetadata(),
	}
}

// NewIIP creates a new Initial Information Packet
func NewIIP[T any](data T) *IP[T] {
	ip := New(data)
	ip.ipType = TypeInitial
	ip.immutable = true
	return ip
}

// NewOpenBracket creates an opening bracket IP
func NewOpenBracket[T any]() *IP[T] {
	ip := new(IP[T])
	ip.id = uuid.New().String()
	ip.ipType = TypeBracketOpen
	ip.metadata = makeInitialMetadata()
	return ip
}

// NewCloseBracket creates a closing bracket IP
func NewCloseBracket[T any]() *IP[T] {
	ip := new(IP[T])
	ip.id = uuid.New().String()
	ip.ipType = TypeBracketClose
	ip.metadata = makeInitialMetadata()
	return ip
}

// Type returns the IP type
func (ip *IP[T]) Type() Type {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.ipType
}

// Data returns the IP data
func (ip *IP[T]) Data() T {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.data
}

// ID returns the unique identifier of the IP
func (ip *IP[T]) ID() string {
	return ip.id
}

// Owner returns the current owner of the IP
func (ip *IP[T]) Owner() string {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.owner
}

// SetOwner sets the owner of the IP
func (ip *IP[T]) SetOwner(owner string) error {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.immutable {
		return fmt.Errorf("cannot modify owner of immutable IP")
	}

	ip.owner = owner
	return nil
}

// IsImmutable returns whether the IP is immutable
func (ip *IP[T]) IsImmutable() bool {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.immutable
}

// Metadata returns a copy of the IP's metadata
func (ip *IP[T]) Metadata() map[string]any {
	ip.mu.RLock()
	defer ip.mu.RUnlock()

	// Return a copy to prevent external modification
	metadataCopy := make(map[string]any, len(ip.metadata))
	for k, v := range ip.metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}

// SetMetadata sets a metadata value
func (ip *IP[T]) SetMetadata(key string, value any) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.metadata == nil {
		ip.metadata = make(map[string]any)
	}
	ip.metadata[key] = value
}

// GetMetadata gets a metadata value
func (ip *IP[T]) GetMetadata(key string) (any, bool) {
	ip.mu.RLock()
	defer ip.mu.RUnlock()

	if ip.metadata == nil {
		return nil, false
	}
	val, ok := ip.metadata[key]
	return val, ok
}

// makeInitialMetadata creates the initial metadata map
func makeInitialMetadata() map[string]any {
	return map[string]any{
		"created_at": time.Now(),
	}
}

// Clone creates a deep copy of the IP
func (ip *IP[T]) Clone() *IP[T] {
	ip.mu.RLock()
	defer ip.mu.RUnlock()

	newIP := &IP[T]{
		id:        uuid.New().String(), // New ID for the clone
		ipType:    ip.ipType,
		data:      ip.data, // Note: This is a shallow copy of data
		metadata:  make(map[string]any, len(ip.metadata)),
		immutable: ip.immutable,
	}

	// Deep copy metadata
	for k, v := range ip.metadata {
		newIP.metadata[k] = v
	}

	return newIP
}
