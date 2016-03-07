package main

import "fmt"

type ImagesCommand struct {
	// nothing yet
}

var imagesCommand ImagesCommand

func (x *ImagesCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	baseImages, err := BaseImages(dc)
	if err != nil {
		return err
	}

	return listImages(baseImages)
}

func listImages(images []BaseImage) error {
	for _, im := range images {
		fmt.Printf("%s: %s\n  Tags:\n", im.Name, im.Description)
		for _, tag := range im.Tags {
			var pulled string
			if tag.Pulled {
				pulled = " (pulled)"
			}
			fmt.Printf("    %s%s\n", tag.Name, pulled)
		}
	}
	return nil
}

func init() {
	_, err := parser.AddCommand("images",
		"List base images.",
		"",
		&imagesCommand)

	if err != nil {
		fmt.Println(err)
	}
}
