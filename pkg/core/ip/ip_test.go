package ip_test

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/elleshadow/noPromises/pkg/core/ip"
)

func TestIP(t *testing.T) {
    t.Run("creation", func(t *testing.T) {
        t.Run("normal packet", func(t *testing.T) {
            packet := ip.New("test")
            assert.Equal(t, ip.TypeNormal, packet.Type())
            assert.Equal(t, "test", packet.Data())
        })

        t.Run("open bracket", func(t *testing.T) {
            packet := ip.NewOpenBracket[string]()
            assert.Equal(t, ip.TypeBracketOpen, packet.Type())
        })

        t.Run("close bracket", func(t *testing.T) {
            packet := ip.NewCloseBracket[string]()
            assert.Equal(t, ip.TypeBracketClose, packet.Type())
        })
    })

    t.Run("ownership", func(t *testing.T) {
        packet := ip.New("test")

        t.Run("initial state", func(t *testing.T) {
            assert.Empty(t, packet.Owner())
        })

        t.Run("set owner", func(t *testing.T) {
            err := packet.SetOwner("proc1")
            require.NoError(t, err)
            assert.Equal(t, "proc1", packet.Owner())
        })

        t.Run("change owner", func(t *testing.T) {
            err := packet.SetOwner("proc2")
            require.NoError(t, err)
            assert.Equal(t, "proc2", packet.Owner())
        })
    })

    t.Run("metadata", func(t *testing.T) {
        packet := ip.New("test")

        t.Run("creation timestamp", func(t *testing.T) {
            val, ok := packet.GetMetadata("created_at")
            assert.True(t, ok)
            _, isTime := val.(time.Time)
            assert.True(t, isTime, "created_at should be a time.Time value")
        })

        t.Run("set and get", func(t *testing.T) {
            packet.SetMetadata("key", "value")
            val, ok := packet.GetMetadata("key")
            assert.True(t, ok)
            assert.Equal(t, "value", val)
        })
    })
}