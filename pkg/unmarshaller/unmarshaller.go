package unmarshaller

func Unmarshal(unmarshal func(any) error) (map[string]any, error) {
	var items map[string]any

	err := unmarshal(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}
