package shutdown

import (
	"context"
	"errors"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Register(t *testing.T) {
	m := NewManager()

	m.Register("handler1", func() error { return nil })
	m.Register("handler2", func() error { return nil })

	assert.Len(t, m.handlers, 2)
	assert.Equal(t, "handler1", m.handlers[0].name)
	assert.Equal(t, "handler2", m.handlers[1].name)
}

func TestManager_Shutdown_ReverseOrder(t *testing.T) {
	m := NewManager()
	var order []string

	m.Register("first", func() error {
		order = append(order, "first")
		return nil
	})
	m.Register("second", func() error {
		order = append(order, "second")
		return nil
	})
	m.Register("third", func() error {
		order = append(order, "third")
		return nil
	})

	err := m.Shutdown(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"third", "second", "first"}, order)
}

func TestManager_Shutdown_CollectsErrors(t *testing.T) {
	m := NewManager()

	m.Register("success", func() error { return nil })
	m.Register("fail1", func() error { return errors.New("error 1") })
	m.Register("fail2", func() error { return errors.New("error 2") })

	err := m.Shutdown(context.Background())
	require.Error(t, err)

	assert.Contains(t, err.Error(), "fail2")
	assert.Contains(t, err.Error(), "fail1")
}

func TestManager_Shutdown_Timeout(t *testing.T) {
	m := NewManager()
	m.SetTimeout(50 * time.Millisecond)

	m.Register("slow", func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	err := m.Shutdown(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestManager_Wait_ContextCancellation(t *testing.T) {
	m := NewManager()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := m.Wait(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestManager_Wait_Signal(t *testing.T) {
	m := NewManager()

	ctx := context.Background()

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := m.Wait(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "received signal")
}

func TestDefaultManager(t *testing.T) {
	oldDefault := defaultManager
	defaultManager = NewManager()
	defer func() { defaultManager = oldDefault }()

	var called atomic.Bool
	Register("test", func() error {
		called.Store(true)
		return nil
	})

	err := Shutdown(context.Background())
	require.NoError(t, err)
	assert.True(t, called.Load())
}

func TestSetTimeout(t *testing.T) {
	oldDefault := defaultManager
	defaultManager = NewManager()
	defer func() { defaultManager = oldDefault }()

	SetTimeout(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, defaultManager.timeout)
}

func TestManager_Shutdown_EmptyHandlers(t *testing.T) {
	m := NewManager()

	err := m.Shutdown(context.Background())
	require.NoError(t, err)
}
