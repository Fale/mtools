package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/fale/mtools/pkg/tgz"
)

func compress(ctx *cli.Context) error {
	err := filepath.Walk(ctx.String("archive-folder"), func(path string, info fs.FileInfo, err error) error {
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

		slog.Info("processing a dir", "path", path)

		pathSlice := strings.Split(fmt.Sprintf(".%s", strings.TrimPrefix(path, filepath.Clean(ctx.String("archive-folder")))), "/")
		f, err := os.OpenFile(filepath.Join(ctx.String("compressed-folder"), fmt.Sprintf("%s-%s-%s.tar.gz", pathSlice[1], pathSlice[2], pathSlice[3])), os.O_CREATE|os.O_RDWR, os.FileMode(0o600))
		if err != nil {
			return err
		}
		defer f.Close()
		return tgz.Compress(path, f, true)
	})
	return err
}
