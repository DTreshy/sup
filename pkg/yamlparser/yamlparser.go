package yamlparser

// MapItem is an item in a MapSlice.
type MapItem struct {
	Key, Value any
}

// MapSlice encodes and decodes as a YAML map.
// The order of keys is preserved when encoding and decoding.
type MapSlice []MapItem

func Unmarshal(unmarshal func(any) error) (MapSlice, error) {
	var items MapSlice

	err := unmarshal(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}
