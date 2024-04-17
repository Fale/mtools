package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
		pars := struct {
			Year    string
			Month   string
			MailDir string
		}{
			pathSlice[1],
			pathSlice[2],
			pathSlice[3],
		}
		tmpl, err := template.New("test").Parse(ctx.String("destination"))
		if err != nil {
			slog.Error("impossible to parse the destination string", "error", err)
		}
		var destination bytes.Buffer
		err = tmpl.Execute(&destination, pars)
		if err != nil {
			panic(err)
		}
		fmt.Println(destination.String())
		if err := os.MkdirAll(filepath.Dir(destination.String()), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(destination.String(), os.O_CREATE|os.O_RDWR, os.FileMode(0o600))
		if err != nil {
			return err
		}
		defer f.Close()
		return tgz.Compress(path, f, true)
	})
	return err
}
