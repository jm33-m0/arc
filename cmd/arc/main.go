package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jm33-m0/arc"
)

func main() {
	// Flags for compression and archival types (with default values set to zst and tar)
	compressionType := flag.String("c", "zst", "Compression type: gzip, bzip2, xz, zst, etc.")
	archivalType := flag.String("t", "tar", "Archival type: tar, zip, etc.")

	// Flags for creating or extracting archives
	createFlag := flag.Bool("a", false, "Create an archive") // Changed to -a for creating archives
	extractFlag := flag.Bool("x", false, "Extract an archive")

	// Flag for specifying the archive file (for both create and extract)
	archiveFile := flag.String("f", "", "Archive file (for creating or extracting)")

	// Parse command line flags
	flag.Parse()

	// Remaining arguments (source and destination)
	args := flag.Args()

	// Validate that source and destination are provided
	if len(args) < 1 || *archiveFile == "" {
		flag.Usage()
		fmt.Printf("\n%s [options] -f <archive_file> <source/destination>\n", os.Args[0])
		return
	}

	source := args[0]     // In create mode, this is the source directory; in extract mode, this is the destination directory
	destination := source // In extract mode, this is the destination directory

	if len(args) > 1 {
		destination = args[1] // Only used in extract mode if -C is not set
	}

	// Maps to handle compression and archival types
	compressionMap := arc.CompressionMap
	archivalMap := arc.ArchivalMap

	// Check if we're creating or extracting an archive
	if *createFlag && *extractFlag {
		log.Fatal("Please specify only one mode: either -a for create or -x for extract.")
	}

	// Archiving mode
	if *createFlag {
		if source == "" {
			log.Fatal("Source directory must be specified for creating an archive")
		}

		// Select the correct compression and archival type based on user input
		compression, ok := compressionMap[strings.ToLower(*compressionType)]
		if !ok {
			log.Fatalf("Unsupported compression type: %s", *compressionType)
		}

		archival, ok := archivalMap[strings.ToLower(*archivalType)]
		if !ok {
			log.Fatalf("Unsupported archival type: %s", *archivalType)
		}

		err := arc.Archive(source, *archiveFile, compression, archival)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Archive created: %s\n", *archiveFile)
		return
	}

	// Unarchiving mode
	if *extractFlag {
		if destination == "" {
			log.Fatal("Destination directory must be specified for extracting an archive")
		}

		// Automatically identify archive format during extraction
		err := arc.Unarchive(*archiveFile, destination)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Archive extracted to: %s\n", destination)
		return
	}

	// If neither create nor extract is specified
	log.Fatal("Please specify either -a to create an archive or -x to extract an archive.")
}
