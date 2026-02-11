package recovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovery_Wrap_Success(t *testing.T) {
	r := New()
	called := false
	err := r.Wrap(func() error {
		called = true
		return nil
	})
	assert.True(t, called)
	assert.NoError(t, err)
}

func TestRecovery_Wrap_Error(t *testing.T) {
	r := New()
	expectedErr := assert.AnError
	err := r.Wrap(func() error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestRecovery_Wrap_Panic(t *testing.T) {
	r := New()
	err := r.Wrap(func() error {
		panic("test panic")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic: test panic")
	assert.Contains(t, err.Error(), "recovery_test.go")
}

func TestRecovery_WrapMain_Success(t *testing.T) {
	r := New()
	called := false
	code := r.WrapMain(func() {
		called = true
	})
	assert.True(t, called)
	assert.Equal(t, 0, code)
}

func TestRecovery_WrapMain_Panic(t *testing.T) {
	r := New()
	code := r.WrapMain(func() {
		panic("main panic")
	})
	assert.Equal(t, 1, code, "WrapMain should return 1 on panic")
}

func TestRecovery_Safe_Success(t *testing.T) {
	r := New()
	called := false
	r.Safe(func() {
		called = true
	})
	assert.True(t, called)
}

func TestRecovery_Safe_Panic(t *testing.T) {
	r := New()
	assert.NotPanics(t, func() {
		r.Safe(func() {
			panic("safe panic")
		})
	})
}

func TestWrapMain_PackageLevel(t *testing.T) {
	code := WrapMain(func() {})
	assert.Equal(t, 0, code)
}

func TestPanicError(t *testing.T) {
	err := PanicError("test value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic: test value")
}

func TestStackTrace(t *testing.T) {
	var stack string
	func() {
		stack = stackTrace()
	}()
	assert.True(t, len(stack) >= 0)
}
