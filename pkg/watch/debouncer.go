/*
# Module: pkg/watch/debouncer.go
Debouncer for batching rapid file changes.

Provides time-based debouncing to prevent excessive re-processing during
rapid file changes (e.g., auto-save, batch edits).

## Linked Modules
None (standalone utility)

## Tags
watch, debounce, rate-limiting

## Exports
Debouncer, NewDebouncer

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#debouncer.go> a code:Module ;
    code:name "pkg/watch/debouncer.go" ;
    code:description "Debouncer for batching rapid file changes" ;
    code:language "go" ;
    code:layer "watch" ;
    code:exports <#Debouncer>, <#NewDebouncer> ;
    code:tags "watch", "debounce", "rate-limiting" .
<!-- End LinkedDoc RDF -->
*/

package watch

import (
	"sync"
	"time"
)

// Debouncer delays function execution until after a period of inactivity
type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
}

// NewDebouncer creates a new debouncer with the specified duration
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{
		duration: duration,
	}
}

// Trigger schedules fn to run after the debounce duration
// If called again before the duration expires, the timer resets
func (d *Debouncer) Trigger(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.duration, fn)
}

// Stop cancels any pending execution
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
