package cli

import "fmt"

type Command struct {
	Description string
	Handle      func([]string) error
}

type Commands map[string]Command

func (c Commands) Parse(args []string) error {
	if len(args) == 0 {
		return c.Usage()
	}

	command, ok := c[args[0]]
	if !ok {
		return c.Usage()
	}

	return command.Handle(args[1:])
}

func (c Commands) Usage() error {
	fmt.Println("Commands:")
	for name, cmd := range c {
		fmt.Printf("\t%s\t\t\t%s\n", name, cmd.Description)
	}
	return nil
}
