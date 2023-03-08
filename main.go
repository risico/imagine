package main

import (
	"github.com/alecthomas/kong"

	"github.com/risico/imagine/cmd"
)

func main() {
	ctx := kong.Parse(&cmd.CLI,
		kong.Name("imagine"),
		kong.Description("Imagine is an image processing server"),
		kong.Configuration(kong.JSON, ".imagine.config.json"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	err := ctx.Run(&cmd.CLI)
	ctx.FatalIfErrorf(err)
}
