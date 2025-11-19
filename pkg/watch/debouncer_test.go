package watch

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncer_Trigger(t *testing.T) {
	var counter atomic.Int32
	debouncer := NewDebouncer(50 * time.Millisecond)

	// Trigger multiple times rapidly
	for i := 0; i < 5; i++ {
		debouncer.Trigger(func() {
			counter.Add(1)
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to complete
	time.Sleep(100 * time.Millisecond)

	// Should only execute once
	if counter.Load() != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter.Load())
	}
}

func TestDebouncer_Stop(t *testing.T) {
	var counter atomic.Int32
	debouncer := NewDebouncer(50 * time.Millisecond)

	debouncer.Trigger(func() {
		counter.Add(1)
	})

	// Stop before execution
	debouncer.Stop()

	// Wait past debounce duration
	time.Sleep(100 * time.Millisecond)

	// Should not execute
	if counter.Load() != 0 {
		t.Errorf("Expected counter to be 0 after stop, got %d", counter.Load())
	}
}

func TestDebouncer_MultipleTriggers(t *testing.T) {
	var counter atomic.Int32
	debouncer := NewDebouncer(30 * time.Millisecond)

	// First batch
	for i := 0; i < 3; i++ {
		debouncer.Trigger(func() {
			counter.Add(1)
		})
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for first batch to execute
	time.Sleep(60 * time.Millisecond)

	// Second batch
	for i := 0; i < 3; i++ {
		debouncer.Trigger(func() {
			counter.Add(1)
		})
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for second batch to execute
	time.Sleep(60 * time.Millisecond)

	// Should execute twice (once per batch)
	if counter.Load() != 2 {
		t.Errorf("Expected counter to be 2, got %d", counter.Load())
	}
}
