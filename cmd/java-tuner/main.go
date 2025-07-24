package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/tgagor/java-tuner/pkg/config"
	"github.com/tgagor/java-tuner/pkg/runner"
	"github.com/tgagor/java-tuner/pkg/tuner"
)

var BuildVersion string // Will be set dynamically at build time.
var appName string = "java-tuner"
var flags config.Flags

var cmd = &cobra.Command{
	Use:   appName,
	Short: "A simple to use Java tunning wrapper, that simplifies Java for running in containers.",
	Long: `A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config file requirement if --version is provided
		if flags.PrintVersion {
			return nil
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		initLogger(flags.Verbose)

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
		cpuCount, memBytes, memPercentage, otherFlags, err := tuner.DetectResources()
		if err != nil {
			log.Error().Err(err).Msg("Failed to detect resources")
			os.Exit(1)
		}

		opts := tuner.Tune(cpuCount, memBytes, memPercentage, otherFlags)
		jvmArgs := tuner.FormatOptions(opts)

		log.Info().Strs("jvmArgs", jvmArgs).Msg("Will start Java with options:")
		java := runner.New()
		java.Arg(jvmArgs...).SetVerbose(flags.Verbose)

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

	cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "d", false, "Print actions but don't execute them")
	cmd.Flags().BoolVar(&flags.NoColor, "no-color", false, "Disable color output")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
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
