package command

import (
	"fmt"

	"github.com/DTreshy/sup/pkg/unmarshaller"
)

// Command represents command(s) to be run remotely.
type Command struct {
	Name   string   `yaml:"-"`      // Command name.
	Desc   string   `yaml:"desc"`   // Command description.
	Local  string   `yaml:"local"`  // Command(s) to be run locally.
	Run    string   `yaml:"run"`    // Command(s) to be run remotely.
	Script string   `yaml:"script"` // Load command(s) from script and run it remotely.
	Upload []Upload `yaml:"upload"` // See Upload struct.
	Stdin  bool     `yaml:"stdin"`  // Attach localhost STDOUT to remote commands' STDIN?
	Once   bool     `yaml:"once"`   // The command should be run "once" (on one host only).
	Serial int      `yaml:"serial"` // Max number of clients processing a task in parallel.

	// API backward compatibility. Will be deprecated in v1.0.
	RunOnce bool `yaml:"run_once"` // The command should be run once only.
}

// Upload represents file copy operation from localhost Src path to Dst
// path of every host in a given Network.
type Upload struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
	Exc string `yaml:"exclude"`
}

// Commands is a list of user-defined commands
type Commands struct {
	Names []string
	Cmds  map[string]Command
}

func (c *Commands) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&c.Cmds)
	if err != nil {
		return err
	}

	items, err := unmarshaller.Unmarshal(unmarshal)
	if err != nil {
		return fmt.Errorf("cannot parse networks: %w", err)
	}

	c.Names = make([]string, len(items))
	i := 0

	for key := range items {
		c.Names[i] = key
		i += 1
	}

	return nil
}

func (c *Commands) Get(name string) (Command, bool) {
	cmd, ok := c.Cmds[name]
	return cmd, ok
}
