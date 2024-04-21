package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/mail"
	"os"
	"path/filepath"
	"slices"

	"github.com/emersion/go-maildir"
	"github.com/urfave/cli/v2"

	"github.com/fale/mtools/pkg/mailDate"
)

func archive(ctx *cli.Context) error {
	if len(ctx.Args().First()) == 0 {
		return fmt.Errorf("missing maildir-name")
	}

	err := filepath.Walk(filepath.Join(ctx.String("maildir-folder"), ctx.Args().First()), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error("prevent panic by handling failure accessing a path", "path", path, "error", err)
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "cur")); err != nil {
			slog.Debug("ignoring folder since is not a maildir", "directory", path)
			return nil
		}
		if slices.Contains(ctx.StringSlice("skip-folders"), info.Name()) {
			slog.Info("skipping a dir", "path", path)
			return nil
		}
		slog.Info("processing a dir", "path", path)

		d := maildir.Dir(path)

		err = d.Walk(func(k string, flags []maildir.Flag) error {
			slog.Info("Processing email", "name", k)
			rdr, err := d.Open(k)
			if err != nil {
				return err
			}
			msg, err := mail.ReadMessage(rdr)
			if err != nil {
				return err
			}
			slog.Debug("Email read", "name", k, "date", msg.Header.Get("Date"))
			msgDate, err := mailDate.GetDate(*msg)
			if err != nil {
				return err
			}
			if msgDate.After(*ctx.Timestamp("cutoff")) {
				return nil
			}
			archivePath := filepath.Join(ctx.String("archive-folder"), msgDate.Format("2006"), msgDate.Format("01"), ctx.Args().First())
			if err := os.MkdirAll(archivePath, os.ModePerm); err != nil {
				return err
			}
			archiveDir := maildir.Dir(archivePath)
			if err := archiveDir.Init(); err != nil {
				return err
			}
			if err := d.Move(archiveDir, k); err != nil {
				return err
			}
			slog.Debug("Email processed", "name", k, "date", msg.Header.Get("Date"))
			return nil
		})
		return err
	})
	return err
}
