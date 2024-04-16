package main

import (
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	var vcount int

	programLevel := slog.LevelError
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(logger)
	app := &cli.App{
		Name:                   "mt",
		Usage:                  "mail tools",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbosity (-v, -vv, -vvv)",
				Count:   &vcount,
				Action: func(ctx *cli.Context, v bool) error {
					switch vcount {
					case 1:
						programLevel = slog.LevelWarn
					case 2:
						programLevel = slog.LevelInfo
					case 3:
						programLevel = slog.LevelDebug
					}
					logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))
					slog.SetDefault(logger)
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "archive",
				Usage:  "archive MAILDIR",
				Action: archive,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "maildir-folder",
						Aliases: []string{"m"},
						Usage:   "folder where the maildir to be archived resides",
						Value:   path.Join(Must(os.UserHomeDir()), ".mail"),
					},
					&cli.StringFlag{
						Name:    "archive-folder",
						Aliases: []string{"a"},
						Usage:   "folder where the archive need to resides",
						Value:   path.Join(Must(os.UserHomeDir()), ".mail", "archive"),
					},
					&cli.TimestampFlag{
						Name:     "cutoff",
						Aliases:  []string{"d"},
						Usage:    "date before which the emails should be archived",
						Layout:   "2006-01-02T15:04:05",
						Value:    cli.NewTimestamp(time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.UTC)),
						Timezone: time.Local,
					},
					&cli.StringSliceFlag{
						Name:    "skip-folders",
						Aliases: []string{"s"},
						Usage:   "folders to be skipped",
						Value:   cli.NewStringSlice("Inbox", "Spam"),
					},
				},
			},
			{
				Name:   "mark-read",
				Usage:  "mark-read FOLDER",
				Action: markRead,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "include-new",
						Aliases: []string{"n"},
						Usage:   "apply to the emails not yet processed",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
