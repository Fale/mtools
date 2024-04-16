package tgz_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fale/mtools/pkg/tgz"
)

func TestCompress(t *testing.T) {
	// Setup mock folder
	tmpSrcDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpSrcDir, "empty"), os.ModePerm); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpSrcDir, "withContent"), os.ModePerm); err != nil {
		t.Error(err)
	}
	if f, err := os.Create(filepath.Join(tmpSrcDir, "withContent", "emptyFile")); err != nil {
		t.Error(err)
	} else {
		f.Close()
	}

	tmpDstDir := t.TempDir()
	f, err := os.OpenFile(filepath.Join(tmpDstDir, "compress.tar.gz"), os.O_CREATE|os.O_RDWR, os.FileMode(0o600))
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	if err := tgz.Compress(tmpSrcDir, f, false); err != nil {
		t.Error(err)
	}
	fmt.Println(filepath.Join(tmpDstDir, "compress.tar.gz"))
}
