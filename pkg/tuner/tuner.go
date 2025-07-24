package tuner

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/java-tuner/pkg/config"
)

// Options holds calculated JVM options.
type Options struct {
	MemoryOpts []string
	CPUOpts    []string
	OtherOpts  []string
}

// DetectResources reads env vars and returns CPU/mem info.
func DetectResources(flags config.Flags) (cpuCount int, memLimit uint64, memPercentage float64, otherFlags []string, err error) {
	cpuCount = flags.CPUCount
	if cpuCount <= 0 {
		log.Debug().Msg("CPU count not set, detecting")
		cpuCount = CPULimit()
	}
	log.Debug().Int("cpuCount", cpuCount).Msg("Detected CPU count")

	memPercentage = flags.MemPercentage
	if memPercentage <= 0 {
		log.Debug().Msg("Memory percentage not set, using default 80.0")
		memPercentage = 80.0
	}
	log.Debug().Float64("memPercentage", memPercentage).Msg("Using memory percentage")

	memLimit = MemoryLimit()
	log.Debug().Uint64("memLimit", memLimit).Msg("Detected memory limit")
	if memLimit <= 0 {
		log.Warn().Msg("Memory limit is 0, using 25% of system RAM")
		memLimit = systemRAM() / 4 // Fallback to 25% of system RAM
		log.Debug().Uint64("memLimit", memLimit).Msg("Using 25% of system RAM as memory limit")
	}

	otherFlags = flags.JvmOpts
	if len(otherFlags) == 0 {
		log.Debug().Msg("No extra JVM options provided")
	} else {
		log.Debug().Strs("otherFlags", otherFlags).Msg("Using extra JVM options")
	}

	return
}

// Tune returns JVM options based on detected resources and user flags.
func Tune(cpuCount int, memLimit uint64, memPercentage float64, otherFlags []string) Options {
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
	opts.OtherOpts = append(opts.OtherOpts, otherFlags...)
	log.Debug().Strs("otherFlags", otherFlags).Msg("Using additional JVM options")
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
