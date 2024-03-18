package main

import (
	"fmt"
	"os"
	"script_validation/commands"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "Conversation Evaluation",
		Description: "Know what LLM to use",
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "start the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Usage:       "The port to run on",
						Value:       "3000",
					},
					&cli.BoolFlag{
						Name:        "dev",
						Usage:       "Enable Dev mode",
						Value:       false,
					},
				},
				Action: func(cCtx *cli.Context) error {
					port := cCtx.String("port")
					dev := cCtx.Bool("dev")
					commands.StartServer(port, dev)
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
