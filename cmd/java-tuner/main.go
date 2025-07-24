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

const JAVA_TUNER_DEFAULT_PREFIX = "JAVA_TUNER_"

var cmd = &cobra.Command{
	Use:   appName,
	Short: "A simple to use Java tuning wrapper, that simplifies running Java apps in containers.",
	Long: `A CLI tool for tuning JVM options and running Java applications in containers.

Automatically detects system resources and generates optimal JVM flags for containerized environments.

Environment Variables:
  JAVA_TUNER_PREFIX         Change env var prefix (default: JAVA_TUNER_)
  JAVA_TUNER_CPU_COUNT      Override detected CPU count (same as --cpu-count)
  JAVA_TUNER_MEM_PERCENTAGE Override detected memory percentage (same as --mem-percentage)
  JAVA_TUNER_OPTS           Additional JVM flags (same as --opts)
  JAVA_TUNER_NO_COLOR       Disable color output (same as --no-color)
  JAVA_TUNER_VERBOSE        Increase verbosity (same as --verbose)
  JAVA_TUNER_LOG_FORMAT     Log format to use (plain, json, console)
  JAVA_TUNER_JAVA_BIN       Path to the Java binary to use (same as --java-bin)
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config file requirement if --version is provided
		if flags.PrintVersion {
			return nil
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch flags.LogFormat {
		case "plain":
			initLogger(flags.Verbose)
		case "json":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), NoColor: flags.NoColor})
			if flags.Verbose {
				log.Logger = log.With().Caller().Logger()
			} else {
				log.Logger = log.With().Logger()
			}
		}

		// If version flag is provided, show the version and exit.
		if flags.PrintVersion {
			// v0.6.3 (go1.23.4 on darwin/arm64; gc)
			fmt.Printf("%s (%s on %s/%s; %s)\n", BuildVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
			return
		}

		// Main logic goes here
		if flags.Verbose {
			log.Debug().Msg("Verbose mode enabled.")
		}

		// Use tuner package to detect resources and print JVM options
		cpuCount, memBytes, memPercentage, otherFlags, err := tuner.DetectResources(flags)
		if err != nil {
			log.Error().Err(err).Msg("Failed to detect resources")
			os.Exit(1)
		}

		opts := tuner.Tune(cpuCount, memBytes, memPercentage, otherFlags)
		jvmArgs := tuner.FormatOptions(opts)

		log.Info().Strs("jvmArgs", jvmArgs).Msg("Will start Java with options:")
		java := runner.New().FindJava(flags.JavaBin).Arg(jvmArgs...).SetVerbose(flags.Verbose)

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
			output, err := java.Run()
			if err != nil {
				log.Error().Err(err).Msg("Failed to run Java command")
				os.Exit(1)
			}
			log.Info().Str("output", output).Msg("Java command executed successfully")
		} else {
			log.Info().Msg("Dry run enabled, not executing command.")
			log.Debug().Str("cmd", java.String()).Msg("Dry run command")
		}
	},
}

func init() {
	if BuildVersion == "" {
		BuildVersion = "development" // Fallback if not set during build
	}

	viper.SetEnvPrefix(getPrefix(JAVA_TUNER_DEFAULT_PREFIX))
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "d", false, "Print actions but don't execute them")
	if err := viper.BindPFlag("dry-run", cmd.Flags().Lookup("dry-run")); err != nil {
		log.Error().Err(err).Msg("Failed to bind dry-run flag")
		os.Exit(1)
	}
	cmd.Flags().BoolVar(&flags.NoColor, "no-color", false, "Disable color output")
	if err := viper.BindPFlag("no-color", cmd.Flags().Lookup("no-color")); err != nil {
		log.Error().Err(err).Msg("Failed to bind no-color flag")
		os.Exit(1)
	}
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	if err := viper.BindPFlag("verbose", cmd.Flags().Lookup("verbose")); err != nil {
		log.Error().Err(err).Msg("Failed to bind verbose flag")
		os.Exit(1)
	}
	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
	cmd.Flags().StringVarP(&flags.LogFormat, "log-format", "l", "plain", "Log format to use (plain, json or console)")
	if err := viper.BindPFlag("log-format", cmd.Flags().Lookup("log-format")); err != nil {
		log.Error().Err(err).Msg("Failed to bind log-format flag")
		os.Exit(1)
	}
	cmd.Flags().Int("cpu-count", 0, "Override detected CPU count")
	if err := viper.BindPFlag("cpu-count", cmd.Flags().Lookup("cpu-count")); err != nil {
		log.Error().Err(err).Msg("Failed to bind cpu-count flag")
		os.Exit(1)
	}
	cmd.Flags().Float64("mem-percentage", 80.0, "Override detected memory percentage")
	if err := viper.BindPFlag("mem-percentage", cmd.Flags().Lookup("mem-percentage")); err != nil {
		log.Error().Err(err).Msg("Failed to bind mem-percentage flag")
		os.Exit(1)
	}
	cmd.Flags().StringSliceVar(&flags.JvmOpts, "opts", []string{}, "Additional JVM flags to pass")
	if err := viper.BindPFlag("opts", cmd.Flags().Lookup("opts")); err != nil {
		log.Error().Err(err).Msg("Failed to bind opts flag")
		os.Exit(1)
	}
	cmd.Flags().StringVar(&flags.JavaBin, "java-bin", "auto-detect", "Path to the Java binary to use (default: auto-detect)")
	if err := viper.BindPFlag("java-bin", cmd.Flags().Lookup("java-bin")); err != nil {
		log.Error().Err(err).Msg("Failed to bind java-bin flag")
		os.Exit(1)
	}

}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}

func initLogger(verbose bool) {
	// Console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:     colorable.NewColorableStdout(),
		NoColor: flags.NoColor,
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

func getPrefix(fallback string) string {
	prefix := os.Getenv("JAVA_TUNER_PREFIX")
	if len(prefix) == 0 {
		prefix = fallback
	}
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}
	return prefix
}
