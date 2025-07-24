package config

type Flags struct {
	DryRun        bool
	NoColor       bool
	PrintVersion  bool
	Verbose       bool
	Debug         bool
	LogFormat     string
	CPUCount      int
	MemPercentage float64
	MemLimit      uint64
	JvmOpts       []string
	OptsRaw       string
	JavaBin       string
}
