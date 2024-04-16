package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/emersion/go-maildir"
	"github.com/urfave/cli/v2"
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
		return archiveSingleFolder(path, ctx.String("archive-folder"), ctx.Args().First(), *ctx.Timestamp("cutoff"))
	})
	return err
}

func archiveSingleFolder(srcFolder string, archiveFolder string, mboxName string, cutOff time.Time) error {
	d := maildir.Dir(srcFolder)
	if !isMailDir(d) {
		slog.Info("This is not a mailDir", "folder", srcFolder)
		return nil
	}
	ks, err := d.Keys()
	if err != nil {
		return err
	}
	for _, k := range ks {
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
		msgDate, err := parseDate(msg.Header.Get("Date"))
		if err != nil {
			msgDate, err = getDateFromReceived(msg.Header.Get("Received"))
			if err != nil {
				return err
			}
		}
		if msgDate.After(cutOff) {
			continue
		}
		archivePath := filepath.Join(archiveFolder, msgDate.Format("2006"), msgDate.Format("01"), mboxName)
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
	}
	return nil
}

func isMailDir(dir maildir.Dir) bool {
	files, err := os.ReadDir(string(dir))
	if err != nil {
		return false
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		if file.Name() == "cur" {
			return true
		}
	}
	return false
}

func isDateOK(t time.Time) bool {
	if t.IsZero() {
		return false
	}

	name, offset := t.Zone()
	if offset != 0 {
		return true
	}
	if name == "UTC" || name == "GMT" || name == "" {
		return true
	}

	return false
}

func parseDate(date string) (time.Time, error) {
	// 'UT' timezone is really UTC
	reUT := regexp.MustCompile(`(.*) UT$`)
	date = reUT.ReplaceAllString(date, `$1 UTC`)
	// 'Pacific Standard Time' timezone is really -0800
	rePST := regexp.MustCompile(`(.*) Pacific Standard Time$`)
	date = rePST.ReplaceAllString(date, `$1 -0800`)
	// 'PST' timezone is really -0800
	rePST2 := regexp.MustCompile(`(.*) PST$`)
	date = rePST2.ReplaceAllString(date, `$1 -0800`)
	// 'PDT' timezone is really -0700
	rePDT := regexp.MustCompile(`(.*) PDT$`)
	date = rePDT.ReplaceAllString(date, `$1 -0700`)
	// 'CDT' timezone is really -0500
	reCDT := regexp.MustCompile(`(.*) CDT$`)
	date = reCDT.ReplaceAllString(date, `$1 -0500`)
	// EST and EDT have issues. Replace them with -0500 and -0400 respectively
	reEST := regexp.MustCompile(`(.*) EST$`)
	date = reEST.ReplaceAllString(date, `$1 -0500`)
	reEDT := regexp.MustCompile(`(.*) EDT$`)
	date = reEDT.ReplaceAllString(date, `$1 -0400`)
	// Drop (METDST) if found
	reMETDST := regexp.MustCompile(`(.*) \(METDST\)$`)
	date = reMETDST.ReplaceAllString(date, `$1`)
	// Drop (ora solare Europa occidentale) if found
	reOSEO := regexp.MustCompile(`(.*) \(ora solare Europa occidentale\)$`)
	date = reOSEO.ReplaceAllString(date, `$1`)
	// Drop (ora legale Europa occidentale) if found
	reOLEO := regexp.MustCompile(`(.*) \(ora legale Europa occidentale\)$`)
	date = reOLEO.ReplaceAllString(date, `$1`)

	layouts := []string{
		"02 Jan 06 15:04 MST",                        // RFC822
		"02 Jan 06 15:04 -0700",                      // RFC822Z
		"Mon, 02 Jan 2006 15:04:05 MST",              // RFC1123
		"Mon, 02 Jan 2006 15:04:05 -0700",            // RFC1123Z
		"Mon, 2 Jan 2006 15:04:05 -0700",             // RFC1123Z variation
		"Mon 2 Jan 2006 15:04:05 -0700",              // RFC1123Z variation
		"02 Jan 2006 15:04:05 (MST)",                 // RFC822 variation
		"2 Jan 2006 15:04:05 MST",                    // RFC822 variation
		"2 Jan 2006 15:04:05 -0700",                  // RFC822Z variation
		"Mon, 2 Jan 2006 15:04:05 MST",               // RFC1123 variation
		"Mon, _2 Jan 2006 15:04:05 MST",              // RFC1123 variation
		"Mon, 02 Jan 2006 15:04:05 \"MST\"",          // RFC1123 variation
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",       // RFC1123 variation
		"Mon, _2 Jan 2006 15:04:05 -0700 (MST)",      // RFC1123 variation
		"Mon, 2 Jan 2006 15:04:05 -0700 (GMT-07:00)", // RFC1123 variation
		"Mon, 02 Jan 2006 15:04:05 -0700 (-07:00)",   // RFC1123 variation
		"Mon, 2 Jan 2006 15:04 MST",                  // RFC1123 variation
		"Mon, 2 Jan 2006 15:04 -0700",                // RFC1123Z variation
		"Mon, Jan 02 2006 15:04:05 -0700",            // RFC1123Z variation
		"Mon, 2 Jan 06 15:04:05 GMT-0700",            // RFC1123 variation
		"Mon, 2 Jan 06 15:04:05 -0700",               // RFC1123 variation
		"02 Jan 06   15:04:05",                       // RFC822 variation
		"Mon, 02 Jan 2006 15:04:05",                  // RFC1123 variation
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, date); err == nil {
			if isDateOK(t) {
				return t, nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("date '%s' is not in a known format", date)
}

func getDateFromReceived(received string) (time.Time, error) {
	reSplit := regexp.MustCompile(`(.*);(.*)$`)
	received = reSplit.ReplaceAllString(received, `$2`)
	received = strings.Trim(received, " ")
	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", received)
}
