package supfile

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/DTreshy/sup/internal/command"
	"github.com/DTreshy/sup/internal/envs"
	"github.com/DTreshy/sup/internal/network"
	"github.com/DTreshy/sup/internal/target"

	"gopkg.in/yaml.v3"
)

// Supfile represents the Stack Up configuration YAML file.
type Supfile struct {
	Networks network.Networks `yaml:"networks"`
	Commands command.Commands `yaml:"commands"`
	Targets  target.Targets   `yaml:"targets"`
	Env      envs.EnvList     `yaml:"env"`
	Version  string           `yaml:"version"`
}

var (
	ErrMustUpdate                error = errors.New("Please update sup by `go get -u github.com/DTreshy/sup/cmd/sup`")
	ErrUnsupportedSupfileVersion error = errors.New("Check your Supfile version (available latest version: v1.0)")
)

// NewSupfile parses configuration file and returns Supfile or error.
func NewSupfile(data []byte) (*Supfile, error) {
	var conf Supfile

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	if conf.Version != "1.0" {
		return nil, ErrUnsupportedSupfileVersion
	}

	return &conf, nil
}

func (s *Supfile) CmdUsage() {
	w := &tabwriter.Writer{}

	w.Init(os.Stderr, 4, 4, 2, ' ', 0)

	defer w.Flush()

	// Print available targets/commands.
	fmt.Fprintln(w, "Targets:\t")

	for _, name := range s.Targets.Names {
		cmds, _ := s.Targets.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, strings.Join(cmds, " "))
	}

	fmt.Fprintln(w, "\t")
	fmt.Fprintln(w, "Commands:\t")

	for _, name := range s.Commands.Names {
		cmd, _ := s.Commands.Get(name)
		fmt.Fprintf(w, "- %v\t%v\n", name, cmd.Desc)
	}

	fmt.Fprintln(w)
}
