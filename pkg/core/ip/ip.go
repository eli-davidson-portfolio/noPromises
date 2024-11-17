package ip

// Type represents the type of Information Packet
type Type int

const (
    // Normal represents a standard data packet
    Normal Type = iota
    // OpenBracket represents the start of a substream
    OpenBracket
    // CloseBracket represents the end of a substream
    CloseBracket
    // Initial represents an Initial Information Packet
    Initial
)

// IP represents an Information Packet in the FBP network
type IP[T any] struct {
    Type     Type
    Data     T
    Metadata map[string]any
}

// New creates a new normal IP with the given data
func New[T any](data T) *IP[T] {
    return &IP[T]{
        Type:     Normal,
        Data:     data,
        Metadata: make(map[string]any),
    }
}

// NewInitial creates a new Initial Information Packet
func NewInitial[T any](data T) *IP[T] {
    return &IP[T]{
        Type:     Initial,
        Data:     data,
        Metadata: make(map[string]any),
    }
}
