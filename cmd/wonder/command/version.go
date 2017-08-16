package command

import (
	"bytes"
	"fmt"

	"github.com/mitchellh/cli"
)

type VersionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	Ui                cli.Ui
}

func (c *VersionCommand) Help() string {
	return "Usage: wonder version"
}

func (c *VersionCommand) Run(args []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Wonder version v%s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", c.VersionPrerelease)
		if c.Revision != "" {
			fmt.Fprintf(&versionString, "(revision %s)", c.Revision)
		}
	}

	c.Ui.Output(versionString.String())

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "show current version string"
}
