package target

import "fmt"

// MapItem is an item in a MapSlice.
type MapItem struct {
	Key, Value any
}

// MapSlice encodes and decodes as a YAML map.
// The order of keys is preserved when encoding and decoding.
type MapSlice []MapItem

// Targets is a list of user-defined targets
type Targets struct {
	Names   []string
	targets map[string][]string
}

func (t *Targets) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&t.targets)
	if err != nil {
		return err
	}

	var items MapSlice

	var ok bool

	err = unmarshal(&items)
	if err != nil {
		return fmt.Errorf("cannot parse targets: %w", err)
	}

	t.Names = make([]string, len(items))
	for i, item := range items {
		t.Names[i], ok = item.Key.(string)
		if !ok {
			return fmt.Errorf("assertion to string failed")
		}
	}

	return nil
}

func (t *Targets) Get(name string) ([]string, bool) {
	cmds, ok := t.targets[name]
	return cmds, ok
}
