// Package shutdown provides graceful shutdown management with cleanup handlers.
//
// This package is useful for long-running services that need to register multiple
// cleanup handlers (e.g., closing database connections, flushing logs) and execute
// them in reverse order on shutdown.
//
// For simple CLI tools that only need context cancellation on signals, prefer using
// signal.NotifyContext directly:
//
//	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//	defer cancel()
//
// Use this package when you need:
//   - Multiple cleanup handlers executed in LIFO order
//   - Named handlers for debugging/logging
//   - Configurable shutdown timeout
//   - A centralized shutdown manager across packages
package shutdown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const DefaultTimeout = 30 * time.Second

type handler struct {
	name string
	fn   func() error
}

type Manager struct {
	mu       sync.Mutex
	handlers []handler
	sigChan  chan os.Signal
	timeout  time.Duration
}

func NewManager() *Manager {
	return &Manager{
		handlers: make([]handler, 0),
		sigChan:  make(chan os.Signal, 1),
		timeout:  DefaultTimeout,
	}
}

func (m *Manager) SetTimeout(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeout = d
}

func (m *Manager) Register(name string, fn func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler{name: name, fn: fn})
}

func (m *Manager) Wait(ctx context.Context) error {
	signal.Notify(m.sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(m.sigChan)

	select {
	case sig := <-m.sigChan:
		return fmt.Errorf("received signal: %v", sig)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	handlers := make([]handler, len(m.handlers))
	copy(handlers, m.handlers)
	timeout := m.timeout
	m.mu.Unlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- m.executeHandlers(handlers)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("shutdown timed out after %v", timeout)
	}
}

func (m *Manager) executeHandlers(handlers []handler) error {
	var errs []error
	for i := len(handlers) - 1; i >= 0; i-- {
		h := handlers[i]
		if err := h.fn(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", h.name, err))
		}
	}
	return errors.Join(errs...)
}

var defaultManager = NewManager()

func Default() *Manager {
	return defaultManager
}

func Register(name string, fn func() error) {
	defaultManager.Register(name, fn)
}

func Wait(ctx context.Context) error {
	return defaultManager.Wait(ctx)
}

func Shutdown(ctx context.Context) error {
	return defaultManager.Shutdown(ctx)
}

func SetTimeout(d time.Duration) {
	defaultManager.SetTimeout(d)
}
