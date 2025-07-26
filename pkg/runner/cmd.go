package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

type Cmd struct {
	cmd      string
	args     []string
	verbose  bool
	preText  string
	postText string
	output   string
}

func New(c ...string) *Cmd {
	if len(c) > 0 {
		return &Cmd{
			cmd:      c[0],
			verbose:  false,
			preText:  "",
			postText: "",
		}
	}

	return &Cmd{
		verbose:  false,
		preText:  "",
		postText: "",
	}
}

func (c *Cmd) SetCommand(cmd string) *Cmd {
	c.cmd = cmd
	return c
}

func (c *Cmd) FindJava(defaultPath string) *Cmd {
	if defaultPath == "" || defaultPath == "auto-detect" {
		path, err := exec.LookPath("java")
		if err != nil {
			log.Error().Err(err).Msg("Java executable not found")
			os.Exit(1)
		} else {
			log.Info().Str("path", path).Msg("Java executable found")
		}

		c.cmd = path
	} else {
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			log.Error().Str("path", defaultPath).Msg("Java executable not found at specified path")
			os.Exit(1)
		}
		c.cmd = defaultPath
		log.Info().Str("path", c.cmd).Msg("Using specified Java executable")
	}
	return c
}

func (c *Cmd) Equal(cmd *Cmd) bool {
	return c.String() == cmd.String()
}

func (c *Cmd) Arg(args ...string) *Cmd {
	c.args = append(c.args, args...)
	return c
}

func (c *Cmd) SetVerbose(verbosity bool) *Cmd {
	c.verbose = verbosity
	return c
}

func (c *Cmd) PreInfo(msg string) *Cmd {
	c.preText = msg
	return c
}

func (c *Cmd) PostInfo(msg string) *Cmd {
	c.postText = msg
	return c
}

func (c *Cmd) Exec() (string, error) {
	// pipe the commands output to the applications
	if c.cmd == "" {
		return "", errors.New("command not set")
	}
	if c.preText != "" {
		log.Info().Msg(c.preText)
	}

	log.Debug().Str("cmd", c.cmd).Interface("args", c.args).Msg("Exec replacing process")

	// Prepare argv: first arg is the command itself
	argv := append([]string{c.cmd}, c.args...)
	env := os.Environ()

	// Use syscall.Exec to replace the current process
	// Note: This call does not return if successful
	err := syscall.Exec(c.cmd, argv, env)
	// If Exec fails, log and return error
	log.Error().Err(err).Str("cmd", c.cmd).Interface("args", c.args).Msg("Could not exec command")
	return "", err
}

func (c *Cmd) Run() (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if c.cmd == "" {
		return "", errors.New("command not set")
	}
	if c.preText != "" {
		log.Info().Msg(c.preText)
	}

	cmd := exec.CommandContext(ctx, c.cmd, c.args...)

	// pipe the commands output to the applications
	var b bytes.Buffer
	if c.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &b
		cmd.Stderr = &b
	}

	log.Debug().Str("cmd", c.cmd).Interface("args", c.args).Msg("Running")
	err := cmd.Run()

	// Check for context cancellation or timeout
	if ctx.Err() != nil {
		// If the context was canceled, suppress output and return context error
		if ctx.Err() == context.Canceled {
			log.Warn().Str("cmd", c.cmd).Msg("Command was cancelled")
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Warn().Str("cmd", c.cmd).Msg("Command timed out")
		}
		return "", ctx.Err()
	}

	// Handle other errors
	if err != nil {
		log.Error().Err(err).Str("cmd", c.cmd).Interface("args", c.args).Msg("Could not run command")
		// c.setOutput(&b)
		c.output = b.String()
		log.Error().Msg(c.output)
		return c.output, err
	}
	c.output = b.String()

	if c.postText != "" {
		log.Info().Msg(c.postText)
	}
	return c.output, nil
}

func (c *Cmd) String() string {
	return strings.Trim(fmt.Sprintf("%s %s", c.cmd, strings.Join(c.args, " ")), " ")
}

func (c *Cmd) Output() (string, error) {
	cmd := exec.Command(c.cmd, c.args...)

	// pipe the commands output to the applications
	var b bytes.Buffer
	if c.verbose {
		mw := io.MultiWriter(os.Stdout, &b)
		cmd.Stdout = mw
		cmd.Stderr = mw
	} else {
		cmd.Stdout = &b
		cmd.Stderr = &b
	}

	log.Debug().Str("cmd", c.cmd).Interface("args", c.args).Msg("Running")
	err := cmd.Run()

	// Handle other errors
	if err != nil {
		log.Error().Err(err).Str("cmd", c.cmd).Interface("args", c.args).Msg("Could not run command")
		c.output = b.String()
		log.Error().Msg(c.output)
		return c.output, err
	}
	c.output = b.String()

	return c.output, nil
}
