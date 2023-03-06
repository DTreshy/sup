package supfile

import (
	"fmt"
	"os"

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

type ErrMustUpdate struct {
	Msg string
}

type ErrUnsupportedSupfileVersion struct {
	Msg string
}

func (e ErrMustUpdate) Error() string {
	return fmt.Sprintf("%v\n\nPlease update sup by `go get -u github.com/DTreshy/sup/cmd/sup`", e.Msg)
}

func (e ErrUnsupportedSupfileVersion) Error() string {
	return fmt.Sprintf("%v\n\nCheck your Supfile version (available latest version: v0.5)", e.Msg)
}

// NewSupfile parses configuration file and returns Supfile or error.
func NewSupfile(data []byte) (*Supfile, error) {
	var conf Supfile

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	// API backward compatibility. Will be deprecated in v1.0.
	switch conf.Version {
	case "":
		conf.Version = "0.1"
		fallthrough

	case "0.1":
		for _, cmd := range conf.Commands.Cmds {
			if cmd.RunOnce {
				return nil, ErrMustUpdate{"command.run_once is not supported in Supfile v" + conf.Version}
			}
		}

		fallthrough

	case "0.2":
		for _, cmd := range conf.Commands.Cmds {
			if cmd.Once {
				return nil, ErrMustUpdate{"command.once is not supported in Supfile v" + conf.Version}
			}

			if cmd.Local != "" {
				return nil, ErrMustUpdate{"command.local is not supported in Supfile v" + conf.Version}
			}

			if cmd.Serial != 0 {
				return nil, ErrMustUpdate{"command.serial is not supported in Supfile v" + conf.Version}
			}
		}

		for _, network := range conf.Networks.Nets {
			if network.Inventory != "" {
				return nil, ErrMustUpdate{"network.inventory is not supported in Supfile v" + conf.Version}
			}
		}

		fallthrough

	case "0.3":
		var warning string

		for key, cmd := range conf.Commands.Cmds {
			if cmd.RunOnce {
				warning = "Warning: command.run_once was deprecated by command.once in Supfile v" + conf.Version + "\n"
				cmd.Once = true
				conf.Commands.Cmds[key] = cmd
			}
		}

		if warning != "" {
			fmt.Fprint(os.Stderr, warning)
		}

		fallthrough

	case "0.4", "0.5":

	default:
		return nil, ErrUnsupportedSupfileVersion{"unsupported Supfile version " + conf.Version}
	}

	return &conf, nil
}
