package recovery

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
)

type Recovery struct {
	log logger.Logger
}

func New() *Recovery {
	return &Recovery{
		log: logger.Default(),
	}
}

func (r *Recovery) Wrap(fn func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = PanicError(rec)
		}
	}()
	return fn()
}

func (r *Recovery) WrapMain(fn func()) (exitCode int) {
	defer func() {
		if rec := recover(); rec != nil {
			r.log.Error("panic recovered", "panic", rec, "stack", stackTrace())
			exitCode = 1
		}
	}()
	fn()
	return 0
}

func (r *Recovery) Safe(fn func()) {
	defer func() {
		if rec := recover(); rec != nil {
			r.log.Error("panic recovered", "panic", rec, "stack", stackTrace())
		}
	}()
	fn()
}

func WrapMain(fn func()) int {
	return New().WrapMain(fn)
}

// RecoverWithError captures a panic and returns it as an error.
// IMPORTANT: This function MUST be called directly from a deferred function.
// recover() only works when called directly in a deferred function, not in
// functions called from a deferred function.
//
// Usage:
//
//	defer func() {
//		if err := recovery.RecoverWithError(); err != nil {
//			// handle panic as error
//		}
//	}()
func RecoverWithError() error {
	if rec := recover(); rec != nil {
		return PanicError(rec)
	}
	return nil
}

func PanicError(r any) error {
	return &panicError{
		value: r,
		stack: stackTrace(),
	}
}

type panicError struct {
	value any
	stack string
}

func (e *panicError) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.value, e.stack)
}

func stackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&sb, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return sb.String()
}
