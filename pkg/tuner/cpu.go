package tuner

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// CPULimit detects how many CPUs are assigned to the container.
func CPULimit() int {
	// Try cgroup v1: /sys/fs/cgroup/cpu/cpu.cfs_quota_us and cpu.cfs_period_us
	quotaPath := "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	periodPath := "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	quota, err1 := readIntFromFile(quotaPath)
	log.Debug().Str("quotaPath", quotaPath).Int("quota", quota).Msg("Read CPU quota")
	period, err2 := readIntFromFile(periodPath)
	log.Debug().Str("periodPath", periodPath).Int("period", period).Msg("Read CPU period")

	if err1 != nil || err2 != nil {
		log.Debug().Err(err1).Err(err2).Msg("Failed to read CPU limits from cgroup v1")

		// Try cgroup v2: /sys/fs/cgroup/cpu.max
		if data, err := os.ReadFile("/sys/fs/cgroup/cpu.max"); err == nil {
			parts := strings.Fields(string(data))
			log.Debug().Str("cpu.max", string(data)).Err(err).Msg("Read CPU max")
			if len(parts) == 2 {
				if parts[0] == "max" {
					return runtime.NumCPU()
				}
				quota, err1 = strconv.Atoi(parts[0])
				period, err2 = strconv.Atoi(parts[1])
			}
		}
	}

	if err1 == nil && err2 == nil && quota > 0 && period > 0 {
		cpus := float64(quota) / float64(period)
		if cpus < 1 {
			log.Warn().Float64("cpus", cpus).Msg("Detected less than 1 CPU, rounding up to 1")
			return 1
		}
		log.Debug().Float64("cpus", cpus).Msg("Detected CPU limit")
		return int(cpus + 0.5) // round up
	}

	// Fallback: use system CPU count
	cpus := runtime.NumCPU()
	log.Warn().Int("cpus", cpus).Msg("Failed to detect CPU limit, using system CPU count")
	return cpus
}

func readIntFromFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	return strconv.Atoi(s)
}
