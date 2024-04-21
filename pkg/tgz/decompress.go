package tgz

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func Decompress(dest string, r io.Reader) error {
	gb, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gb.Close()
	tb := tar.NewReader(gb)

	for {
		header, err := tb.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header == nil {
			continue
		}

		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err == nil {
				continue
			}
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tb); err != nil {
				return err
			}
			f.Close()
		}
	}
}
