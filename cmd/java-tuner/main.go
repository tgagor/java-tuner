package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tgagor/java-tuner/pkg/config"
	"github.com/tgagor/java-tuner/pkg/runner"
	"github.com/tgagor/java-tuner/pkg/tuner"
)

var BuildVersion string // Will be set dynamically at build time.
var appName string = "java-tuner"
var flags config.Flags
var v = viper.New()

const JAVA_TUNER_DEFAULT_PREFIX = "JAVA_TUNER"

var cmd = &cobra.Command{
	Use:   appName,
	Short: "A simple to use Java tuning wrapper, that simplifies running Java apps in containers.",
	Long: `A CLI tool for tuning JVM options and running Java applications in containers.

Automatically detects system resources and generates optimal JVM flags for containerized environments.

Environment Variables:
  JAVA_TUNER_PREFIX         Change env var prefix (default: JAVA_TUNER)
  JAVA_TUNER_CPU_COUNT      Override detected CPU count (same as --cpu-count)
  JAVA_TUNER_MEM_PERCENTAGE Override detected memory percentage (same as --mem-percentage)
  JAVA_TUNER_OPTS           Additional JVM flags (same as --opts)
  JAVA_TUNER_NO_COLOR       Disable color output (same as --no-color)
  JAVA_TUNER_VERBOSE        Increase verbosity (same as --verbose)
  JAVA_TUNER_LOG_FORMAT     Log format to use (plain, json, console)
  JAVA_TUNER_JAVA_BIN       Path to the Java binary to use (same as --java-bin)
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch v.GetString("log-format") {
		case "console":
			initLogger(v.GetBool("verbose"), false)
		case "plain":
			initLogger(v.GetBool("verbose"), true)
		default: // json works out of the box
		}

		log.Debug().Any("settings", v.AllSettings()).Msg("Loaded configuration from environment variables")

		// If version flag is provided, show the version and exit.
		if v.GetBool("version") {
			// v0.6.3 (go1.23.4 on darwin/arm64; gc)
			fmt.Printf("%s (%s on %s/%s; %s)\n", BuildVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
			return
		}

		// Main logic goes here
		if v.GetBool("verbose") {
			log.Debug().Msg("Verbose mode enabled.")
		}

		// Use tuner package to detect resources and print JVM options
		javaVersion, cpuCount, memBytes, memPercentage, otherFlags, err := tuner.DetectResources(
			v.GetInt("cpu-count"),
			v.GetFloat64("mem-percentage"),
			v.GetString("opts"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to detect resources")
			os.Exit(1)
		}

		opts := tuner.Tune(javaVersion, cpuCount, memBytes, memPercentage, otherFlags)
		jvmArgs := tuner.FormatOptions(opts)
		jvmArgs = tuner.FilterBlacklisted(jvmArgs)
		java := runner.New().Arg(jvmArgs...).SetVerbose(flags.Verbose)

		// Find arguments after -- and append them to jvmArgs
		extraArgs := []string{}
		for i, arg := range os.Args {
			if arg == "--" {
				extraArgs = os.Args[i+1:]
				break
			}
		}
		if len(extraArgs) > 0 {
			java.Arg(extraArgs...)
			log.Debug().Strs("extraArgs", extraArgs).Msg("Appended extra arguments after --")
		}

		if !flags.DryRun {
			_, err := java.FindJava(v.GetString("java-bin")).Exec()
			if err != nil {
				log.Error().Err(err).Msg("Failed to run Java command")
				os.Exit(1)
			}
		} else {
			log.Debug().Str("cmd", "java "+java.String()).Msg("Would run")
			log.Info().Msg("Dry run enabled, not executing command.")
		}
	},
}

func init() {
	if BuildVersion == "" {
		BuildVersion = "development" // Fallback if not set during build
	}

	v.SetEnvPrefix(getPrefix())
	replacer := strings.NewReplacer("-", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()

	// Use v to provide default values for flags if env vars are set
	cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "d", false, "Print actions but don't execute them")
	_ = v.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run"))

	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	_ = v.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))

	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
	_ = v.BindPFlag("version", cmd.Flags().Lookup("version"))

	cmd.Flags().StringVarP(&flags.LogFormat, "log-format", "l", "console", "Log format to use (plain, json or console)")
	_ = v.BindPFlag("log-format", cmd.Flags().Lookup("log-format"))

	cmd.Flags().IntVar(&flags.CPUCount, "cpu-count", 0, "Override detected CPU count")
	_ = v.BindPFlag("cpu-count", cmd.Flags().Lookup("cpu-count"))

	cmd.Flags().Float64Var(&flags.MemPercentage, "mem-percentage", 0.0, "Override detected memory percentage")
	_ = v.BindPFlag("mem-percentage", cmd.Flags().Lookup("mem-percentage"))

	cmd.Flags().StringVar(&flags.OptsRaw, "opts", "", "Additional JVM flags to pass (space-separated)")
	_ = v.BindPFlag("opts", cmd.Flags().Lookup("opts"))

	cmd.Flags().StringVar(&flags.JavaBin, "java-bin", "auto-detect", "Path to the Java binary to use (default: auto-detect)")
	_ = v.BindPFlag("java-bin", cmd.Flags().Lookup("java-bin"))

	v.AutomaticEnv()
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}

func initLogger(verbose bool, noColor bool) {
	// Console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:     colorable.NewColorableStdout(),
		NoColor: noColor,
	}
	// Disable timestamps
	zerolog.TimeFieldFormat = ""
	consoleWriter.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	// Base logger
	baseLogger := zerolog.New(consoleWriter).With().Logger()

	// Add caller only for debug level using a hook
	if verbose {
		log.Logger = baseLogger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			if level == zerolog.DebugLevel {
				e.Caller()
			}
		})).Level(zerolog.DebugLevel)
	} else {
		log.Logger = baseLogger.Level(zerolog.InfoLevel)
	}
}

func getPrefix() string {
	fallback := JAVA_TUNER_DEFAULT_PREFIX
	prefix := os.Getenv("JAVA_TUNER_PREFIX")
	if len(prefix) == 0 {
		prefix = fallback
	}
	return prefix
}
