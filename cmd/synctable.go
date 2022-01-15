package cmd

import (
	"github.com/urfave/cli/v2"

	"kdqueue/initial"
	"kdqueue/store"
)

func SyncTableForTest() *cli.Command {
	return &cli.Command{
		Name:  `synctable`,
		Usage: `sync table for local test`,
		Action: func(c *cli.Context) error {
			db := initial.InitGoOrm()
			db.AutoMigrate(&store.Message{})
			return nil
		},
	}
}
