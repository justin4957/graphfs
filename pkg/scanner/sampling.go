/*
# Module: pkg/scanner/sampling.go
Sampling strategies for large codebases.

Provides intelligent sampling to enable quick exploration of massive codebases
without processing everything.

## Linked Modules
- [scanner](./scanner.go) - Main scanner

## Tags
scanner, sampling, performance, large-codebase

## Exports
Sampler, SamplingStrategy, NewSampler

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#sampling.go> a code:Module ;
    code:name "pkg/scanner/sampling.go" ;
    code:description "Sampling strategies for large codebases" ;
    code:language "go" ;
    code:layer "scanner" ;
    code:linksTo <./scanner.go> ;
    code:exports <#Sampler>, <#SamplingStrategy>, <#NewSampler> ;
    code:tags "scanner", "sampling", "performance", "large-codebase" .
<!-- End LinkedDoc RDF -->
*/

package scanner

import (
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// SamplingStrategy defines how files should be sampled
type SamplingStrategy int

const (
	// SampleRandom randomly selects files
	SampleRandom SamplingStrategy = iota
	// SampleStratified samples proportionally from each directory
	SampleStratified
	// SampleRecent samples the most recently modified files
	SampleRecent
)

// ParseSamplingStrategy parses a strategy name string
func ParseSamplingStrategy(name string) SamplingStrategy {
	switch name {
	case "stratified":
		return SampleStratified
	case "recent":
		return SampleRecent
	default:
		return SampleRandom
	}
}

// String returns the string representation of the strategy
func (s SamplingStrategy) String() string {
	switch s {
	case SampleStratified:
		return "stratified"
	case SampleRecent:
		return "recent"
	default:
		return "random"
	}
}

// Sampler provides file sampling capabilities
type Sampler struct {
	strategy SamplingStrategy
	size     int
	seed     int64
	rng      *rand.Rand
}

// NewSampler creates a new sampler with the given strategy and size
func NewSampler(strategy SamplingStrategy, size int, seed int64) *Sampler {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Sampler{
		strategy: strategy,
		size:     size,
		seed:     seed,
		rng:      rand.New(rand.NewSource(seed)),
	}
}

// Sample samples files according to the configured strategy
func (s *Sampler) Sample(files []string) []string {
	if len(files) <= s.size {
		return files
	}

	switch s.strategy {
	case SampleRandom:
		return s.sampleRandom(files)
	case SampleStratified:
		return s.sampleStratified(files)
	case SampleRecent:
		return s.sampleRecent(files)
	default:
		return s.sampleRandom(files)
	}
}

// sampleRandom randomly selects files
func (s *Sampler) sampleRandom(files []string) []string {
	if len(files) <= s.size {
		return files
	}

	// Create a copy of the indices and shuffle
	indices := make([]int, len(files))
	for i := range indices {
		indices[i] = i
	}

	// Fisher-Yates shuffle
	for i := len(indices) - 1; i > 0; i-- {
		j := s.rng.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Take first `size` indices
	sampled := make([]string, s.size)
	for i := 0; i < s.size; i++ {
		sampled[i] = files[indices[i]]
	}

	return sampled
}

// sampleStratified samples proportionally from each directory
func (s *Sampler) sampleStratified(files []string) []string {
	if len(files) <= s.size {
		return files
	}

	// Group files by directory
	byDir := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		byDir[dir] = append(byDir[dir], file)
	}

	// Calculate samples per directory (proportional to directory size)
	totalFiles := len(files)
	sampled := make([]string, 0, s.size)
	remaining := s.size

	// Sort directories for deterministic ordering
	dirs := make([]string, 0, len(byDir))
	for dir := range byDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	for i, dir := range dirs {
		dirFiles := byDir[dir]

		// Calculate proportional sample size for this directory
		var samplesForDir int
		if i == len(dirs)-1 {
			// Last directory gets all remaining slots
			samplesForDir = remaining
		} else {
			// Proportional allocation
			proportion := float64(len(dirFiles)) / float64(totalFiles)
			samplesForDir = int(float64(s.size) * proportion)
			if samplesForDir < 1 && len(dirFiles) > 0 {
				samplesForDir = 1
			}
		}

		// Don't exceed available files in directory
		if samplesForDir > len(dirFiles) {
			samplesForDir = len(dirFiles)
		}

		// Don't exceed remaining slots
		if samplesForDir > remaining {
			samplesForDir = remaining
		}

		if samplesForDir > 0 {
			// Create a sub-sampler for this directory
			subSampler := &Sampler{
				strategy: SampleRandom,
				size:     samplesForDir,
				seed:     s.seed,
				rng:      s.rng,
			}
			dirSampled := subSampler.sampleRandom(dirFiles)
			sampled = append(sampled, dirSampled...)
			remaining -= len(dirSampled)
		}

		if remaining <= 0 {
			break
		}
	}

	return sampled
}

// fileWithMtime holds a file path and its modification time
type fileWithMtime struct {
	path  string
	mtime time.Time
}

// sampleRecent samples the most recently modified files
func (s *Sampler) sampleRecent(files []string) []string {
	if len(files) <= s.size {
		return files
	}

	// Get modification times for all files
	filesWithMtime := make([]fileWithMtime, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			// If we can't stat, use zero time (will be sorted to end)
			filesWithMtime = append(filesWithMtime, fileWithMtime{path: file, mtime: time.Time{}})
		} else {
			filesWithMtime = append(filesWithMtime, fileWithMtime{path: file, mtime: info.ModTime()})
		}
	}

	// Sort by modification time (most recent first)
	sort.Slice(filesWithMtime, func(i, j int) bool {
		return filesWithMtime[i].mtime.After(filesWithMtime[j].mtime)
	})

	// Take first `size` files
	sampled := make([]string, s.size)
	for i := 0; i < s.size; i++ {
		sampled[i] = filesWithMtime[i].path
	}

	return sampled
}

// SampleStats provides statistics about a sampling operation
type SampleStats struct {
	TotalFiles      int
	SampledFiles    int
	Strategy        SamplingStrategy
	DirectoryCounts map[string]int // Files sampled per directory (for stratified)
}

// SampleWithStats samples files and returns statistics
func (s *Sampler) SampleWithStats(files []string) ([]string, *SampleStats) {
	sampled := s.Sample(files)

	stats := &SampleStats{
		TotalFiles:      len(files),
		SampledFiles:    len(sampled),
		Strategy:        s.strategy,
		DirectoryCounts: make(map[string]int),
	}

	// Count files per directory in the sample
	for _, file := range sampled {
		dir := filepath.Dir(file)
		stats.DirectoryCounts[dir]++
	}

	return sampled, stats
}
