package command

import "fmt"

type Type string

const (
	TypeSet    Type = "SET"
	TypeDelete Type = "DELETE"
)

type Command struct {
	Type  Type   `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (c Command) Validate() error {
	if c.Key == "" {
		return fmt.Errorf("command key cannot be empty")
	}

	switch c.Type {
	case TypeSet, TypeDelete:
		return nil
	default:
		return fmt.Errorf("unsupported command type %q", c.Type)
	}
}
