package arc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archives"
)

// Maps to handle compression and archival types
var CompressionMap = map[string]archives.Compression{
	"gz":  archives.Gz{},
	"bz2": archives.Bz2{},
	"xz":  archives.Xz{},
	"zst": archives.Zstd{},
	"lz4": archives.Lz4{},
	"br":  archives.Brotli{},
}

var ArchivalMap = map[string]archives.Archival{
	"tar": archives.Tar{},
	"zip": archives.Zip{},
}

// check if a path exists
func isExist(path string) bool {
	_, statErr := os.Stat(path)
	return !os.IsNotExist(statErr)
}

// Archive is a function that archives the files in a directory
// dir: the directory to Archive
// outfile: the output file
// compression: the compression to use (gzip, bzip2, etc.)
// archival: the archival to use (tar, zip, etc.)
func Archive(dir, outfile string, compression archives.Compression, archival archives.Archival) error {
	// remove outfile
	os.RemoveAll(outfile)

	if !isExist(dir) {
		return fmt.Errorf("%s does not exist", dir)
	}

	// map files on disk to their paths in the archive
	archive_dir_name := filepath.Base(filepath.Clean(dir))
	if dir == "." {
		archive_dir_name = ""
	}
	files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
		dir: archive_dir_name,
	})
	if err != nil {
		return err
	}

	// create the output file we'll write to
	outf, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer outf.Close()

	// we can use the Archive type to gzip a tarball
	// (compression is not required; you could use Tar directly)
	format := archives.CompressedArchive{
		Compression: compression,
		Archival:    archival,
	}

	// create the archive
	err = format.Archive(context.Background(), outf, files)
	if err != nil {
		return err
	}
	return nil
}
