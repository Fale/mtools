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

// CompressFiles compresses selected files from a source maildir and includes
// the maildir directory structure needed when extracting the archive.
func CompressFiles(source string, files []string, w io.Writer) error {
	gb := gzip.NewWriter(w)
	defer gb.Close()
	tb := tar.NewWriter(gb)
	defer tb.Close()

	paths := []string{source, filepath.Join(source, "cur"), filepath.Join(source, "new"), filepath.Join(source, "tmp")}
	paths = append(paths, files...)
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}
		if rel == "." {
			header.Name = "."
		} else {
			header.Name = filepath.ToSlash(filepath.Join(".", rel))
		}
		if err := tb.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		data, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tb, data)
		closeErr := data.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
