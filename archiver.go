package arc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
	logging("Starting the archival process for directory: %s", dir)

	// remove outfile
	logging("Removing any existing output file: %s", outfile)
	if err := os.RemoveAll(outfile); err != nil {
		errMsg := fmt.Errorf("failed to remove existing output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}

	if !isExist(dir) {
		errMsg := fmt.Errorf("directory '%s' does not exist, cannot proceed with archival", dir)
		logging("%s", errMsg.Error())
		return errMsg
	}

	// map files on disk to their paths in the archive
	logging("Mapping files in directory: %s", dir)
	archiveDirName := filepath.Base(filepath.Clean(dir))
	if dir == "." {
		archiveDirName = ""
	}
	files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
		dir: archiveDirName,
	})
	if err != nil {
		errMsg := fmt.Errorf("error mapping files from directory '%s': %w", dir, err)
		logging("%s", errMsg.Error())
		return errMsg
	}
	logging("Successfully mapped files for directory: %s", dir)

	// create the output file we'll write to
	logging("Creating output file: %s", outfile)
	outf, err := os.Create(outfile)
	if err != nil {
		errMsg := fmt.Errorf("error creating output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}
	defer func() {
		logging("Closing output file: %s", outfile)
		outf.Close()
	}()

	// define the archive format
	logging("Defining the archive format with compression: %T and archival: %T", compression, archival)
	format := archives.CompressedArchive{
		Compression: compression,
		Archival:    archival,
	}

	// create the archive
	logging("Starting archive creation: %s", outfile)
	err = format.Archive(context.Background(), outf, files)
	if err != nil {
		errMsg := fmt.Errorf("error during archive creation for output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}
	logging("Archive created successfully: %s", outfile)
	return nil
}

// ArchiveWithFilter is a function that archives the files in a directory
// while excluding certain files based on a filter
// dir: the directory to Archive
// outfile: the output file
// compression: the compression to use (gzip, bzip2, etc.)
// archival: the archival to use (tar, zip, etc.)
// filter: a function that returns true for files to be excluded
func ArchiveWithFilter(dir, outfile string, compression archives.Compression, archival archives.Archival, filter func(string) bool) error {
	logging("Starting the archival process for directory: %s with filter", dir)

	// remove outfile
	logging("Removing any existing output file: %s", outfile)
	if err := os.RemoveAll(outfile); err != nil {
		errMsg := fmt.Errorf("failed to remove existing output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}

	if !isExist(dir) {
		errMsg := fmt.Errorf("directory '%s' does not exist, cannot proceed with archival", dir)
		logging("%s", errMsg.Error())
		return errMsg
	}

	// map files on disk to their paths in the archive
	logging("Mapping files in directory: %s with filter", dir)
	archiveDirName := filepath.Base(filepath.Clean(dir))
	if dir == "." {
		archiveDirName = ""
	}
	files, err := archives.FilesFromDisk(context.Background(), nil, map[string]string{
		dir: archiveDirName,
	})
	if err != nil {
		errMsg := fmt.Errorf("error mapping files from directory '%s': %w", dir, err)
		logging("%s", errMsg.Error())
		return errMsg
	}

	// apply the filter to exclude certain files
	filteredFiles := make([]archives.FileInfo, 0, len(files))
	for _, fi := range files {
		if !filter(fi.Name()) {
			filteredFiles = append(filteredFiles, fi)
		}
	}
	logging("Successfully mapped and filtered files for directory: %s", dir)

	// create the output file we'll write to
	logging("Creating output file: %s", outfile)
	outf, err := os.Create(outfile)
	if err != nil {
		errMsg := fmt.Errorf("error creating output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}
	defer func() {
		logging("Closing output file: %s", outfile)
		outf.Close()
	}()

	// define the archive format
	logging("Defining the archive format with compression: %T and archival: %T", compression, archival)
	format := archives.CompressedArchive{
		Compression: compression,
		Archival:    archival,
	}

	// create the archive
	logging("Starting archive creation: %s", outfile)
	err = format.Archive(context.Background(), outf, filteredFiles)
	if err != nil {
		errMsg := fmt.Errorf("error during archive creation for output file '%s': %w", outfile, err)
		logging("%s", errMsg.Error())
		return errMsg
	}
	logging("Archive created successfully: %s", outfile)
	return nil
}

// ExcludeFilesFilter returns a filter function that excludes files matching the given regex patterns
func ExcludeFilesFilter(excludePatterns []string) (func(string) bool, error) {
	excludeRegexes := make([]*regexp.Regexp, len(excludePatterns))
	for i, pattern := range excludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		excludeRegexes[i] = re
	}
	return func(name string) bool {
		for _, re := range excludeRegexes {
			if re.MatchString(name) {
				return true
			}
		}
		return false
	}, nil
}

// IncludeFilesFilter returns a filter function that includes only files matching the given regex patterns
func IncludeFilesFilter(includePatterns []string) (func(string) bool, error) {
	includeRegexes := make([]*regexp.Regexp, len(includePatterns))
	for i, pattern := range includePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		includeRegexes[i] = re
	}
	return func(name string) bool {
		for _, re := range includeRegexes {
			if re.MatchString(name) {
				return false
			}
		}
		return true
	}, nil
}
