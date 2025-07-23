package tuner

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// CPULimit detects how many CPUs are assigned to the container.
func CPULimit() int {
	// Try cgroup v1: /sys/fs/cgroup/cpu/cpu.cfs_quota_us and cpu.cfs_period_us
	quotaPath := "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	periodPath := "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	quota, err1 := readIntFromFile(quotaPath)
	period, err2 := readIntFromFile(periodPath)
	if err1 == nil && err2 == nil && quota > 0 && period > 0 {
		cpus := float64(quota) / float64(period)
		if cpus < 1 {
			return 1
		}
		return int(cpus + 0.5) // round up
	}

	// Try cgroup v2: /sys/fs/cgroup/cpu.max
	if data, err := os.ReadFile("/sys/fs/cgroup/cpu.max"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) == 2 {
			if parts[0] == "max" {
				return runtime.NumCPU()
			}
			quota, err1 := strconv.Atoi(parts[0])
			period, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && quota > 0 && period > 0 {
				cpus := float64(quota) / float64(period)
				if cpus < 1 {
					return 1
				}
				return int(cpus + 0.5)
			}
		}
	}

	// Fallback: use system CPU count
	return runtime.NumCPU()
}

func readIntFromFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	return strconv.Atoi(s)
}
