package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/emersion/go-maildir"
	"github.com/urfave/cli/v2"

	"github.com/fale/mtools/pkg/mailDate"
	"github.com/fale/mtools/pkg/tgz"
)

func compress(ctx *cli.Context) error {
	tmpl, err := template.New("destination").Parse(ctx.String("destination"))
	if err != nil {
		return fmt.Errorf("impossible to parse the destination string: %w", err)
	}
	archiveFolder := filepath.Clean(ctx.String("archive-folder"))

	err = filepath.Walk(archiveFolder, func(path string, info fs.FileInfo, err error) error {
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
		return compressMaildir(path, info.Name(), tmpl)
	})
	return err
}

func compressMaildir(path, mailDir string, tmpl *template.Template) error {
	type mailGroup struct {
		date  time.Time
		files []string
	}
	groups := make(map[string]*mailGroup)
	d := maildir.Dir(path)
	if err := d.Walk(func(key string, _ []maildir.Flag) error {
		rdr, err := d.Open(key)
		if err != nil {
			return err
		}
		msg, err := mail.ReadMessage(rdr)
		rdr.Close()
		if err != nil {
			return err
		}
		msgDate, err := mailDate.GetDate(*msg)
		if err != nil {
			return err
		}
		groupKey := msgDate.Format("2006-01")
		group := groups[groupKey]
		if group == nil {
			group = &mailGroup{date: msgDate}
			groups[groupKey] = group
		}
		filename, err := d.Filename(key)
		if err != nil {
			return err
		}
		group.files = append(group.files, filename)
		return nil
	}); err != nil {
		return err
	}

	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		group := groups[key]
		var destination bytes.Buffer
		if err := tmpl.Execute(&destination, struct {
			Year    string
			Month   string
			MailDir string
		}{group.date.Format("2006"), group.date.Format("01"), mailDir}); err != nil {
			return fmt.Errorf("impossible to execute the destination template: %w", err)
		}
		fmt.Println(destination.String())
		if err := os.MkdirAll(filepath.Dir(destination.String()), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(destination.String(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(0o600))
		if err != nil {
			return err
		}
		err = tgz.CompressFiles(path, group.files, f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}

	return nil
}
