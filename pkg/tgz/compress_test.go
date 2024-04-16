package tgz_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/fale/mtools/pkg/tgz"
)

func TestCompress(t *testing.T) {
	// Setup mock folder
	tmpSrcDir := t.TempDir()
	if err := os.MkdirAll(path.Join(tmpSrcDir, "empty"), os.ModePerm); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll(path.Join(tmpSrcDir, "withContent"), os.ModePerm); err != nil {
		t.Error(err)
	}
	if f, err := os.Create(path.Join(tmpSrcDir, "withContent", "emptyFile")); err != nil {
		t.Error(err)
	} else {
		f.Close()
	}

	tmpDstDir := t.TempDir()
	f, err := os.OpenFile(path.Join(tmpDstDir, "compress.tar.gz"), os.O_CREATE|os.O_RDWR, os.FileMode(0o600))
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	if err := tgz.Compress(tmpSrcDir, f, false); err != nil {
		t.Error(err)
	}
	fmt.Println(path.Join(tmpDstDir, "compress.tar.gz"))
}
