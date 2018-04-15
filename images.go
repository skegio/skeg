package main

import "fmt"

type ImagesCommand struct {
	Base    bool   `short:"b" long:"base" description:"List base images."`
	Type    string `short:"t" long:"type" description:"Type of environment."`
	Version string `short:"v" long:"version" description:"Version of environment type."`
	Image   string `short:"i" long:"image" description:"Image to use for creating environment."`
	Prune   bool   `short:"p" long:"prune" description:"Prune unused user images."`
}

var imagesCommand ImagesCommand

func (x *ImagesCommand) Execute(args []string) error {
	dc, err := NewDockerClient(globalOptions.toConnectOpts())
	if err != nil {
		return err
	}

	if len(imagesCommand.Type) > 0 || len(imagesCommand.Image) > 0 {
		sc, err := NewSystemClient()
		if err != nil {
			return err
		}

		userImages, err := UserImages(dc, sc, ImageOpts{
			Type:    imagesCommand.Type,
			Version: imagesCommand.Version,
			Image:   imagesCommand.Image,
		}, -1)
		if err != nil {
			return err
		}

		if imagesCommand.Prune {
			for _, im := range userImages {
				if im.EnvCount == 0 {
					fmt.Printf("Removing %s...\n", im.Name)
					err = RemoveUserImage(dc, im)
					if err != nil {
						return err
					}
				}
			}
			return nil
		}
		return listUserImages(userImages)
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

func listUserImages(images []UserImage) error {
	for _, im := range images {
		fmt.Printf("%s (ver: %d) (%d envs)\n", im.Name, im.Version, im.EnvCount)
		fmt.Printf("  build time: %s\n", im.Labels["skeg.io/image/buildtime"])
		fmt.Printf("  time zone: %s\n", im.Labels["skeg.io/image/timezone"])
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
