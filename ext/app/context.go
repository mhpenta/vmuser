// Package app contains general application management functions
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var ErrContextCancelled = errors.New("context has been cancelled or has expired")

// ContextCancelled is a utility function to check if a context has been cancelled.
func ContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		slog.Info("Context has been cancelled or has expired")
		return true
	default:
		return false
	}
}

type DebugContext struct {
	context.Context
	mu   sync.Mutex
	data map[interface{}]interface{}
}

func (d *DebugContext) WithValue(key, val interface{}) *DebugContext {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.data == nil {
		d.data = make(map[interface{}]interface{})
	}
	d.data[key] = val

	return &DebugContext{
		Context: context.WithValue(d.Context, key, val),
		data:    d.data,
	}
}

func (d *DebugContext) PrintValues() {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Println("Context values - DebugContext")
	for k, v := range d.data {
		fmt.Println("Key:", k, "Value:", v)
	}
}
