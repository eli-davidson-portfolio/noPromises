package ip_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/noPromises/pkg/core/ip"
)

func TestIP(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        tests := []struct {
            name string
            data interface{}
        }{
            {"string data", "test"},
            {"integer data", 42},
            {"struct data", struct{ Value string }{"test"}},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                packet := ip.New(tt.data)
                assert.Equal(t, ip.TypeNormal, packet.Type())
                assert.Equal(t, tt.data, packet.Data())
                assert.NotNil(t, packet.Metadata())
            })
        }
    })

    t.Run("brackets", func(t *testing.T) {
        t.Run("open bracket", func(t *testing.T) {
            bracket := ip.NewOpenBracket[string]()
            assert.Equal(t, ip.TypeBracketOpen, bracket.Type())
            assert.Empty(t, bracket.Data())
        })

        t.Run("close bracket", func(t *testing.T) {
            bracket := ip.NewCloseBracket[string]()
            assert.Equal(t, ip.TypeBracketClose, bracket.Type())
            assert.Empty(t, bracket.Data())
        })

        t.Run("bracket sequence", func(t *testing.T) {
            open := ip.NewOpenBracket[string]()
            data := ip.New("test")
            close := ip.NewCloseBracket[string]()

            assert.Equal(t, ip.TypeBracketOpen, open.Type())
            assert.Equal(t, ip.TypeNormal, data.Type())
            assert.Equal(t, ip.TypeBracketClose, close.Type())
        })
    })

    t.Run("type safety", func(t *testing.T) {
        t.Run("string type", func(t *testing.T) {
            packet := ip.New[string]("test")
            data := packet.Data()
            assert.IsType(t, "", data)
        })

        t.Run("int type", func(t *testing.T) {
            packet := ip.New[int](42)
            data := packet.Data()
            assert.IsType(t, int(0), data)
        })
    })

    t.Run("metadata", func(t *testing.T) {
        packet := ip.New("test")

        t.Run("set and get", func(t *testing.T) {
            packet.SetMetadata("key", "value")
            val, ok := packet.GetMetadata("key")
            assert.True(t, ok)
            assert.Equal(t, "value", val)
        })

        t.Run("creation timestamp", func(t *testing.T) {
            ts, ok := packet.GetMetadata("created_at")
            assert.True(t, ok)
            assert.NotNil(t, ts)
        })
    })

    t.Run("ownership", func(t *testing.T) {
        packet := ip.New("test")
        
        t.Run("initial state", func(t *testing.T) {
            assert.Empty(t, packet.Owner())
            assert.NotEmpty(t, packet.ID())
        })

        t.Run("ownership transfer", func(t *testing.T) {
            err := packet.SetOwner("proc1")
            require.NoError(t, err)
            assert.Equal(t, "proc1", packet.Owner())

            err = packet.SetOwner("proc2")
            require.NoError(t, err)
            assert.Equal(t, "proc2", packet.Owner())
        })
    })

    t.Run("IIP", func(t *testing.T) {
        t.Run("creation", func(t *testing.T) {
            iip := ip.NewIIP("test")
            assert.Equal(t, ip.TypeInitial, iip.Type())
            assert.Equal(t, "test", iip.Data())
        })

        t.Run("immutable", func(t *testing.T) {
            iip := ip.NewIIP("test")
            assert.True(t, iip.IsImmutable())
            err := iip.SetOwner("proc1")
            assert.Error(t, err)
        })
    })
}

// Helper function for type assertion checks
func assertType[T any](t *testing.T, value interface{}) {
    t.Helper()
    _, ok := value.(T)
    assert.True(t, ok)
}