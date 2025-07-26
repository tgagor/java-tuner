package tuner

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/java-tuner/pkg/runner"
	"golang.org/x/mod/semver"
)

// Options holds calculated JVM options.
type Options struct {
	MemoryOpts []string
	CPUOpts    []string
	OtherOpts  []string
}

// DetectResources reads env vars and returns CPU/mem info.
func DetectResources(cpuCount int, memPercentage float64, opts string) (javaVersion string, finalCpuCount int, memLimit uint64, finalMemPercentage float64, finalOpts []string, err error) {
	cmd := runner.New("java").Arg("-version")
	versionOutput, err := cmd.Output()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Java version")
		// TODO: handle error properly
	}
	log.Debug().Str("output", versionOutput).Err(err).Msg("Java version output")

	javaVersion, err = JavaVersion(versionOutput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse Java version")
	}
	log.Debug().Str("version", javaVersion).Err(err).Msg("Detected Java version")

	defaults := GetDefaults(javaVersion)

	finalCpuCount = cpuCount
	if finalCpuCount <= 0 {
		log.Debug().Msg("CPU count not set, detecting")
		finalCpuCount = CPULimit()
	}
	log.Debug().Int("cpuCount", finalCpuCount).Msg("Detected CPU count")

	finalMemPercentage = memPercentage
	if finalMemPercentage <= 0 {
		log.Debug().Msg("Memory percentage not set, using default 80.0")
		finalMemPercentage = defaults.maxRamPercentage
	}
	log.Debug().Float64("memPercentage", finalMemPercentage).Msg("Using memory percentage")

	memLimit = MemoryLimit()
	log.Debug().Uint64("memLimit", memLimit).Msg("Detected memory limit")
	if memLimit <= 0 {
		log.Warn().Msg("Memory limit is 0, using 25% of system RAM")
		memLimit = systemRAM() / 4 // Fallback to 25% of system RAM
		log.Debug().Uint64("memLimit", memLimit).Msg("Using 25% of system RAM as memory limit")
	}

	if len(opts) != 0 {
		// add defaults to opts
		finalOpts = append(finalOpts, defaults.opts...)
		// Parse opts from raw string if provided
		finalOpts = append(finalOpts, strings.Fields(opts)...)
		log.Debug().Strs("otherFlags", finalOpts).Msg("Using extra JVM options")
	} else {
		log.Debug().Msg("No extra JVM options provided")
	}

	return
}

// Tune returns JVM options based on detected resources and user flags.
func Tune(javaVersion string, cpuCount int, memLimit uint64, memPercentage float64, otherFlags []string) Options {
	log.Debug().Msg("Tuning JVM options")
	opts := Options{}

	defaults := GetDefaults(javaVersion)
	opts.OtherOpts = append(opts.OtherOpts, defaults.opts...)

	if semver.Compare(defaults.maxVersion, "v10.0") < 0 { // older Java, calculate limits in MB
		for _, flag := range defaults.maxRamFlags {
			// we take the percentage of max memory limit and convert it to MB
			opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf(flag, float64(memLimit)*memPercentage/100/1024/1024))
			log.Info().Str("flag", flag).Msg("Using max RAM flag")
		}
		for _, flag := range defaults.initialRamFlags {
			opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf(flag, float64(memLimit)*memPercentage/100/1024/1024))
			log.Info().Str("flag", flag).Msg("Using initial RAM flag")
		}
	} else { // Java 10+, use percentage
		for _, flag := range defaults.maxRamFlags {
			opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf(flag, memPercentage))
			log.Info().Str("flag", flag).Msg("Using max RAM percentage flag")
		}
		for _, flag := range defaults.initialRamFlags {
			opts.MemoryOpts = append(opts.MemoryOpts, fmt.Sprintf(flag, defaults.initialRamPercentage))
			log.Info().Str("flag", flag).Msg("Using initial RAM percentage flag")
		}
	}

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
