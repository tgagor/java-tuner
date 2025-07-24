package tuner

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// MemoryLimit checks container memory limit from cgroup files or system RAM.
func MemoryLimit() uint64 {
	// Try /sys/fs/cgroup/memory/memory.limit_in_bytes (cgroup v1)
	if data, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
			log.Debug().Str("memory.limit_in_bytes", string(data)).Msg("Read memory limit")
			return val
		}
	}
	// Try /sys/fs/cgroup/memory.max (cgroup v2)
	if data, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		s := strings.TrimSpace(string(data))
		log.Debug().Str("memory.max", s).Msg("Read memory max")
		if s == "max" {
			// No limit, fallback to system RAM
			log.Warn().Str("memory.max", s).Msg("No memory limit set, using 25% of system RAM")
			return systemRAM() / 4 // Assume 1/4 of system RAM
		}
		if val, err := strconv.ParseUint(s, 10, 64); err == nil {
			return val
		}
	}
	// Fallback to 25% of system RAM
	return systemRAM() / 4
}

// systemRAM returns the total system RAM in bytes.
func systemRAM() uint64 {
	// Use runtime.MemStats as a simple fallback (not always accurate for total RAM)
	var sysinfo runtime.MemStats
	runtime.ReadMemStats(&sysinfo)
	// This is not total system RAM, but Go doesn't provide a portable way.
	// For Linux, we can parse /proc/meminfo
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := bytesSplit(data, '\n')
		for _, line := range lines {
			if bytesHasPrefix(line, []byte("MemTotal:")) {
				fields := bytesFields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseUint(string(fields[1]), 10, 64); err == nil {
						// Value is in kB
						return val * 1024
					}
				}
			}
		}
	}
	// Fallback: return Go heap sys (not accurate)
	return sysinfo.Sys
}

// Helper functions for trimming and splitting bytes
func bytesSplit(b []byte, sep byte) [][]byte {
	var out [][]byte
	start := 0
	for i, c := range b {
		if c == sep {
			out = append(out, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, b[start:])
	}
	return out
}
func bytesHasPrefix(b, prefix []byte) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := range prefix {
		if b[i] != prefix[i] {
			return false
		}
	}
	return true
}
func bytesFields(b []byte) [][]byte {
	var out [][]byte
	start := -1
	for i, c := range b {
		if c == ' ' || c == '\t' {
			if start != -1 {
				out = append(out, b[start:i])
				start = -1
			}
		} else {
			if start == -1 {
				start = i
			}
		}
	}
	if start != -1 {
		out = append(out, b[start:])
	}
	return out
}
