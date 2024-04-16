package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/emersion/go-maildir"
	"github.com/urfave/cli/v2"
)

func markRead(ctx *cli.Context) error {
	if len(ctx.Args().First()) == 0 {
		return fmt.Errorf("missing folder")
	}

	mbSkips := []string{"new", "cur", "tmp"}
	err := filepath.Walk(ctx.Args().First(), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error("prevent panic by handling failure accessing a path", "path", path, "error", err)
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if slices.Contains(mbSkips, info.Name()) {
			slog.Info("skipping a dir", "path", path)
			return nil
		}
		slog.Info("processing a dir", "path", path)

		if _, err := os.Stat(filepath.Join(path, "cur")); err != nil {
			slog.Debug("ignoring folder since is not a maildir", "directory", path)
			return nil
		}
		d := maildir.Dir(path)

		if ctx.Bool("include-new") {
			if _, err := d.Unseen(); err != nil {
				return err
			}
		}

		err = d.Walk(func(key string, flags []maildir.Flag) error {
			if slices.Contains(flags, maildir.FlagSeen) {
				slog.Debug("no flags changes needed", "key", key)
				return nil
			}
			flags = append(flags, maildir.FlagSeen)
			slog.Debug("mark as read", "key", key)
			return d.SetFlags(key, flags)
		})
		return err
	})
	return err
}
