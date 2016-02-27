package main

import "fmt"

type ListCommand struct {
	// nothing yet
}

var listCommand ListCommand

func (x *ListCommand) Execute(args []string) error {
	dc, err := NewConn()
	if err != nil {
		return err
	}

	return listEnvironments(dc)
}

func init() {
	_, err := parser.AddCommand("list",
		"List environments.",
		"",
		&listCommand)

	if err != nil {
		fmt.Println(err)
	}
}
