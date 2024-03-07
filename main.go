package main

import (
	"fmt"
	"os"
	"script_validation/commands"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "Conversation Validator",
		Description: "Know what LLM to use",
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "start the server",
				Action: func(cCtx *cli.Context) error {
					commands.StartServer()
					return nil
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "Generate a new template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name to insert into the template",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					commands.TraverseFolder("./models")
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
