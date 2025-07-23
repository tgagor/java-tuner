package tuner

import (
	"fmt"
	"os"
	"strconv"
)

// Options holds calculated JVM options.
type Options struct {
	MemoryOpts []string
	CPUOpts    []string
	OtherOpts  []string
}

// DetectResources reads env vars and returns CPU/mem info.
func DetectResources() (cpuCount int, memLimit uint64, memPercentage float64, err error) {
	// Try env vars first
	cpuStr := os.Getenv("JAVA_TUNER_CPU")
	percentStr := os.Getenv("JAVA_TUNER_MEM_PERCENTAGE")

	if cpuStr != "" {
		cpuCount, err = strconv.Atoi(cpuStr)
		if err != nil {
			return
		}
	} else {
		cpuCount = 2 // Default fallback
	}

	memLimit = MemoryLimit()

	// Default percentage is 80.0
	memPercentage = 80.0
	if percentStr != "" {
		// Try float first
		memPercentage, err = strconv.ParseFloat(percentStr, 64)
		if err != nil {
			// Try int
			var intVal int
			intVal, err = strconv.Atoi(percentStr)
			if err == nil {
				memPercentage = float64(intVal)
			} else {
				memPercentage = 80.0 // fallback
				err = nil
			}
		}
	}
	return
}

// Tune returns JVM options based on detected resources and user flags.
func Tune(cpuCount int, memLimit uint64, memPercentage float64, extra map[string]string) Options {
	opts := Options{}
	// Memory options
	xmx := fmt.Sprintf("-Xmx%dM", memLimit/(1024*1024))
	opts.MemoryOpts = append(opts.MemoryOpts, xmx)
	opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf("-XX:MaxRAMPercentage=%.2f", memPercentage))
	// CPU options
	opts.CPUOpts = append(opts.CPUOpts, fmt.Sprintf("-XX:ActiveProcessorCount=%d", cpuCount))
	// Other options (can be extended)
	for k, v := range extra {
		opts.OtherOpts = append(opts.OtherOpts, fmt.Sprintf("-D%s=%s", k, v))
	}
	return opts
}

// FormatOptions returns a slice of JVM arguments.
func FormatOptions(opts Options) []string {
	args := []string{}
	args = append(args, opts.MemoryOpts...)
	args = append(args, opts.CPUOpts...)
	args = append(args, opts.OtherOpts...)
	return args
}
