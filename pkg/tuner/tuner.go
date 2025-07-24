package tuner

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Options holds calculated JVM options.
type Options struct {
	MemoryOpts []string
	CPUOpts    []string
	OtherOpts  []string
}

// DetectResources reads env vars and returns CPU/mem info.
func DetectResources() (cpuCount int, memLimit uint64, memPercentage float64, otherFlags []string, err error) {
	// Try env vars first
	cpuStr := os.Getenv("JAVA_TUNER_CPU")
	percentStr := os.Getenv("JAVA_TUNER_MEM_PERCENTAGE")
	// Fallback to JAVA_OPTS if JAVA_TUNER_OPTS is empty
	flagsStr := os.Getenv("JAVA_TUNER_OPTS")
	if flagsStr == "" {
		flagsStr = os.Getenv("JAVA_OPTS")
	}
	log.Debug().Str("JAVA_TUNER_CPU", cpuStr).Str("JAVA_TUNER_MEM_PERCENTAGE", percentStr).Str("JAVA_TUNER_OPTS", flagsStr).Msg("Read environment variables")

	if cpuStr != "" {
		cpuCount, err = strconv.Atoi(cpuStr)
		log.Debug().Str("JAVA_TUNER_CPU", cpuStr).Int("cpuCount", cpuCount).Msg("Read CPU count from env")
		if err != nil {
			return
		}
	} else {
		log.Debug().Msg("JAVA_TUNER_CPU not set, detecting CPU count")
		cpuCount = CPULimit()
	}
	log.Debug().Int("cpuCount", cpuCount).Msg("Detected CPU count")

	memLimit = MemoryLimit()
	log.Debug().Uint64("memLimit", memLimit).Msg("Detected memory limit")

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
				log.Warn().Err(err).Msg("Failed to parse JAVA_TUNER_MEM_PERCENTAGE, using default 80.0")
			}
		}
		log.Debug().Str("JAVA_TUNER_MEM_PERCENTAGE", percentStr).Float64("memPercentage", memPercentage).Msg("Read memory percentage from env")
	}

	// Split flagsStr into a slice of strings
	otherFlagsList := []string{}
	if flagsStr != "" {
		otherFlagsList = strings.Split(flagsStr, "")
	}
	otherFlags = otherFlagsList
	log.Debug().Strs("otherFlags", otherFlags).Msg("Parsed other flags")

	return
}

// Tune returns JVM options based on detected resources and user flags.
func Tune(cpuCount int, memLimit uint64, memPercentage float64, extra []string) Options {
	log.Debug().Msg("Tuning JVM options")
	opts := Options{}
	// Collecting default JVM options
	defaults := []string{
		"-XX:+AlwaysActAsServerClassMachine",     // Always use server JVM
		"-Dnetworkaddress.cache.ttl=10",          // DNS cache
		"-Dnetworkaddress.cache.negative.ttl=10", // Negative DNS cache
		"-XX:+UseStringDeduplication",            // Enable string deduplication
		"-Xshare:off",                            // Disable class data sharing as it's just single JVM
	}

	// Memory options
	memDefaults := []string{
		"-XX:InitialRAMPercentage=25.0",
		"-XX:MinRAMPercentage=25.0",
	}

	opts.MemoryOpts = append(opts.MemoryOpts, memDefaults...)
	log.Debug().Strs("memDefaults", memDefaults).Msg("Using default memory options")
	opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf("-XX:MaxRAMPercentage=%.1f", memPercentage))
	log.Debug().Float64("memPercentage", memPercentage).Msg("Using memory percentage for MaxRAMPercentage")

	if memLimit < 128*1024*1024 { // Less than 128MB
		log.Warn().Uint64("memLimit", memLimit).Msg("Memory limit is less than 128MB, setting -XX:MaxRAM would not allow to start JVM, skipping it")
	} else {
		maxRAM := memLimit/1024/1024 - 100 // Leave 100MB for OS
		opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf("-XX:MaxRAM=%dm", maxRAM))
		log.Debug().Uint64("memLimit", memLimit).Msg("Using memory limit for MaxRAM")
	}

	// CPU options
	opts.CPUOpts = append(opts.CPUOpts, fmt.Sprintf("-XX:ActiveProcessorCount=%d", cpuCount))
	log.Debug().Int("cpuCount", cpuCount).Msg("Using CPU count for ActiveProcessorCount")

	// Other options
	opts.OtherOpts = append(opts.OtherOpts, defaults...)
	log.Debug().Strs("defaults", defaults).Msg("Using default JVM options")
	opts.OtherOpts = append(opts.OtherOpts, extra...)
	log.Debug().Strs("extra", extra).Msg("Using extra JVM options")
	return opts
}

// FormatOptions returns a slice of JVM arguments.
func FormatOptions(opts Options) []string {
	args := []string{}
	args = append(args, opts.OtherOpts...)
	args = append(args, opts.CPUOpts...)
	args = append(args, opts.MemoryOpts...)
	return args
}
