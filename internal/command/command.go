package command

import "fmt"

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

// MapItem is an item in a MapSlice.
type MapItem struct {
	Key, Value any
}

// MapSlice encodes and decodes as a YAML map.
// The order of keys is preserved when encoding and decoding.
type MapSlice []MapItem

func (c *Commands) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&c.Cmds)
	if err != nil {
		return err
	}

	var items MapSlice

	var ok bool

	err = unmarshal(&items)
	if err != nil {
		return fmt.Errorf("cannot parse cmds: %w", err)
	}

	c.Names = make([]string, len(items))
	for i, item := range items {
		c.Names[i], ok = item.Key.(string)
		if !ok {
			return fmt.Errorf("assertion to string failed")
		}
	}

	return nil
}

func (c *Commands) Get(name string) (Command, bool) {
	cmd, ok := c.Cmds[name]
	return cmd, ok
}
