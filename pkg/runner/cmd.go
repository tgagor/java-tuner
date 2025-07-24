package runner

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

type Cmd struct {
	cmd     string
	args    []string
	verbose bool
	preText string
}

func New(c string) *Cmd {
	return &Cmd{
		cmd:     c,
		verbose: false,
		preText: "",
	}
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

func (c *Cmd) Run() (string, error) {
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

func (c *Cmd) String() string {
	return strings.Trim(fmt.Sprintf("%s %s", c.cmd, strings.Join(c.args, " ")), " ")
}
