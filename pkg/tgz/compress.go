package tgz

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(folder string, buf io.Writer, relPath bool) error {
	gb := gzip.NewWriter(buf)
	tb := tar.NewWriter(gb)

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(path)
		if relPath {
			header.Name = fmt.Sprintf(".%s", strings.TrimPrefix(header.Name, filepath.Clean(folder)))
		}

		if err := tb.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(tb, data)
		return err
	})

	if err := tb.Close(); err != nil {
		return err
	}
	return gb.Close()
}
