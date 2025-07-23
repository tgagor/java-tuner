package tuner

import (
	"os"
	"runtime"
	"strconv"
)

// MemoryLimit checks container memory limit from cgroup files or system RAM.
func MemoryLimit() uint64 {
	// Try /sys/fs/cgroup/memory/memory.limit_in_bytes
	if data, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		if val, err := strconv.ParseUint(string(bytesTrim(data)), 10, 64); err == nil {
			return val
		}
	}
	// Try /sys/fs/cgroup/memory.max (cgroup v2)
	if data, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		s := string(bytesTrim(data))
		if s == "max" {
			// No limit, fallback to system RAM
			return systemRAM()
		}
		if val, err := strconv.ParseUint(s, 10, 64); err == nil {
			return val
		}
	}
	// Fallback to system RAM
	return systemRAM()
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
func bytesTrim(b []byte) []byte {
	return []byte(string(b))
}
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
