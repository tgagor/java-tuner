package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/java-tuner/pkg/tuner"
)

func TestTune_Defaults(t *testing.T) {
	cases := []struct {
		name        string
		javaVersion string
		wantFlags   []string
	}{
		{
			name:        "Java8Defaults",
			javaVersion: "v1.8.0",
			wantFlags:   []string{"-XX:ActiveProcessorCount=2", "-Xmx=819m", "-Xms=819m", "-XX:MaxRAM=924m"},
		},
		{
			name:        "Java11Defaults",
			javaVersion: "v11.0",
			wantFlags:   []string{"-XX:ActiveProcessorCount=2", "-XX:MaxRAMPercentage=80.0", "-XX:MinRAMPercentage=25.0", "-XX:InitialRAMPercentage=25.0", "-XX:MaxRAM=924m"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tuner.Tune(tc.javaVersion, 2, 1024*1024*1024, 80.0, []string{})
			args := tuner.FormatOptions(opts)
			for _, flag := range tc.wantFlags {
				assert.Contains(t, args, flag)
			}
		})
	}
}

func TestTune_LowMemory(t *testing.T) {
	cases := []struct {
		name        string
		javaVersion string
		cpu         int
		mem         uint64
		maxRAMPct   float64
		wantFlags   []string
		notFlags    []string
	}{
		{
			name:        "Java8LowMem",
			javaVersion: "v1.8.0",
			cpu:         1,
			mem:         64 * 1024 * 1024,
			maxRAMPct:   75.0,
			wantFlags:   []string{"-XX:ActiveProcessorCount=1", "-Xmx=48m", "-Xms=48m"},
			notFlags:    []string{"-XX:MaxRAM="},
		},
		{
			name:        "Java11LowMem",
			javaVersion: "v11.0",
			cpu:         1,
			mem:         64 * 1024 * 1024,
			maxRAMPct:   75.0,
			wantFlags:   []string{"-XX:ActiveProcessorCount=1", "-XX:MaxRAMPercentage=75.0"},
			notFlags:    []string{"-XX:MaxRAM="},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tuner.Tune(tc.javaVersion, tc.cpu, tc.mem, tc.maxRAMPct, []string{})
			args := tuner.FormatOptions(opts)
			for _, flag := range tc.wantFlags {
				assert.Contains(t, args, flag)
			}
			for _, notFlag := range tc.notFlags {
				for _, a := range args {
					assert.NotContains(t, a, notFlag)
				}
			}
		})
	}
}

func TestTune_ExtraFlags(t *testing.T) {
	cases := []struct {
		name        string
		javaVersion string
		cpu         int
		mem         uint64
		maxRAMPct   float64
		extra       []string
		wantFlags   []string
	}{
		{
			name:        "Java8ExtraFlags",
			javaVersion: "v1.8.0",
			cpu:         4,
			mem:         2048 * 1024 * 1024,
			maxRAMPct:   90.0,
			extra:       []string{"-Dfoo=bar", "-XX:+UseG1GC"},
			wantFlags:   []string{"-XX:ActiveProcessorCount=4", "-Xmx=1843m", "-Xms=1843m", "-XX:MaxRAM=1948m", "-Dfoo=bar", "-XX:+UseG1GC"},
		},
		{
			name:        "Java11ExtraFlags",
			javaVersion: "v11.0",
			cpu:         4,
			mem:         2048 * 1024 * 1024,
			maxRAMPct:   90.0,
			extra:       []string{"-Dfoo=bar", "-XX:+UseG1GC"},
			wantFlags:   []string{"-XX:ActiveProcessorCount=4", "-XX:MaxRAMPercentage=90.0", "-XX:MaxRAM=1948m", "-Dfoo=bar", "-XX:+UseG1GC"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tuner.Tune(tc.javaVersion, tc.cpu, tc.mem, tc.maxRAMPct, tc.extra)
			args := tuner.FormatOptions(opts)
			for _, flag := range tc.wantFlags {
				assert.Contains(t, args, flag)
			}
		})
	}
}

func TestTune_ZeroCPU(t *testing.T) {
	cases := []struct {
		name        string
		javaVersion string
		cpu         int
		mem         uint64
		maxRAMPct   float64
		wantFlags   []string
	}{
		{
			name:        "Java8ZeroCPU",
			javaVersion: "v1.8.0",
			cpu:         0,
			mem:         512 * 1024 * 1024,
			maxRAMPct:   50.0,
			// detection happen eslewhere, but we still want to see the flags
			wantFlags: []string{"-XX:ActiveProcessorCount=0", "-Xmx=256m", "-Xms=256m", "-XX:MaxRAM=412m"},
		},
		{
			name:        "Java11ZeroCPU",
			javaVersion: "v11.0",
			cpu:         0,
			mem:         512 * 1024 * 1024,
			maxRAMPct:   50.0,
			wantFlags:   []string{"-XX:ActiveProcessorCount=0", "-XX:MaxRAMPercentage=50.0", "-XX:MaxRAM=412m"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tuner.Tune(tc.javaVersion, tc.cpu, tc.mem, tc.maxRAMPct, nil)
			args := tuner.FormatOptions(opts)
			for _, flag := range tc.wantFlags {
				assert.Contains(t, args, flag)
			}
		})
	}
}
