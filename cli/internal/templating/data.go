package templating

import (
	"io"

	"gopkg.in/yaml.v3"
)

func DataFromSource[T any](r io.Reader) (*T, error) {
	var data T
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
