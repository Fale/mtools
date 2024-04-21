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

// Compress a folder to a provided io.Writer buffer as a .tar.gz.
// RelPath is used to define if the paths in the .tar.gz are relative to the folder or absolutes.
func Compress(source string, w io.Writer, relPath bool) error {
	gb := gzip.NewWriter(w)
	defer gb.Close()
	tb := tar.NewWriter(gb)
	defer tb.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(path)
		if relPath {
			header.Name = fmt.Sprintf(".%s", strings.TrimPrefix(header.Name, filepath.Clean(source)))
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
}
