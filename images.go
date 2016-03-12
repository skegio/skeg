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

func listImages(images []*BaseImage) error {
	for _, im := range images {
		fmt.Printf("%s: %s\n  Tags:\n", im.Name, im.Description)
		for _, tag := range im.Tags {
			var pulled string
			var preferred string
			if tag.Pulled {
				pulled = " (pulled)"
			}
			if tag.Preferred {
				preferred = " (preferred)"
			}
			fmt.Printf("    %s%s%s\n", tag.Name, pulled, preferred)
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
