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
	// Define subcommands
	archiveCommand := flag.NewFlagSet("archive", flag.ExitOnError)
	extractCommand := flag.NewFlagSet("extract", flag.ExitOnError)
	compressCommand := flag.NewFlagSet("compress", flag.ExitOnError)
	decompressCommand := flag.NewFlagSet("decompress", flag.ExitOnError)

	// Global flags
	verboseFlag := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	// Enable verbose mode if -v is set
	if *verboseFlag {
		arc.DEBUG = true
	}

	// Check if a subcommand is provided
	if len(flag.Args()) < 1 {
		printUsage()
		return
	}

	// Handle subcommands
	switch flag.Args()[0] {
	case "archive":
		handleArchive(archiveCommand, flag.Args()[1:])
	case "extract":
		handleExtract(extractCommand, flag.Args()[1:])
	case "compress":
		handleCompress(compressCommand, flag.Args()[1:])
	case "decompress":
		handleDecompress(decompressCommand, flag.Args()[1:])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  arc [options] <command> [command options]")
	fmt.Println("\nGlobal Options:")
	fmt.Println("  -v\tVerbose mode")
	fmt.Println("\nCommands:")
	fmt.Println("  archive\tCreate an archive with optional compression")
	fmt.Println("  extract\tExtract an archive")
	fmt.Println("  compress\tCompress a single file")
	fmt.Println("  decompress\tDecompress a single file")
	fmt.Println("\nFor help with a specific command, use:")
	fmt.Println("  arc <command> -h")
}

func handleArchive(cmd *flag.FlagSet, args []string) {
	// Flags for archive creation
	compressionType := cmd.String("c", "zst", "Compression type: gzip/gz, bzip2/bz2, xz, zst, lz4, br, etc.")
	archivalType := cmd.String("t", "tar", "Archival type: tar, zip, etc.")
	archiveFile := cmd.String("f", "", "Archive file to create (required)")
	includeFilter := cmd.String("include", "", "Include filter (regex pattern)")
	excludeFilter := cmd.String("exclude", "", "Exclude filter (regex pattern)")
	// New flags for ZIP compression
	compressionLevel := cmd.Int("level", 6, "ZIP compression level (0-9, 0=none, 9=best)")
	compressionMethod := cmd.Int("method", 8, "ZIP compression method (0=store, 8=deflate)")

	cmd.Usage = func() {
		fmt.Println("Usage: arc archive [options] <source_directory>")
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Validate required flags
	if *archiveFile == "" {
		fmt.Println("Error: Archive file (-f) is required")
		cmd.Usage()
		return
	}

	// Get source directory
	if cmd.NArg() < 1 {
		fmt.Println("Error: Source directory is required")
		cmd.Usage()
		return
	}
	source := cmd.Arg(0)

	// Handle ZIP format specifically due to its constraints
	if strings.ToLower(*archivalType) == "zip" {
		// Handle filters for ZIP format
		var filter func(string) bool
		var err error
		if *includeFilter != "" {
			filter, err = arc.IncludeFilesFilter(strings.Split(*includeFilter, ","))
		} else if *excludeFilter != "" {
			filter, err = arc.ExcludeFilesFilter(strings.Split(*excludeFilter, ","))
		}

		if err != nil {
			log.Fatal(err)
		}

		// Use the new Zip function with custom compression options
		if filter != nil {
			err = arc.ZipWithFilter(source, *archiveFile, *compressionLevel, *compressionMethod, filter)
		} else {
			err = arc.Zip(source, *archiveFile, *compressionMethod)
		}

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("ZIP archive created: %s\n", *archiveFile)
		return
	}

	// For other archive types, proceed with normal compression
	compression, ok := arc.CompressionMap[strings.ToLower(*compressionType)]
	if !ok {
		log.Fatalf("Unsupported compression type: %s", *compressionType)
	}

	archival, ok := arc.ArchivalMap[strings.ToLower(*archivalType)]
	if !ok {
		log.Fatalf("Unsupported archival type: %s", *archivalType)
	}

	// Handle filters
	var filter func(string) bool
	var err error
	if *includeFilter != "" {
		filter, err = arc.IncludeFilesFilter(strings.Split(*includeFilter, ","))
	} else if *excludeFilter != "" {
		filter, err = arc.ExcludeFilesFilter(strings.Split(*excludeFilter, ","))
	}

	if err != nil {
		log.Fatal(err)
	}

	// Create archive
	if filter != nil {
		err = arc.ArchiveWithFilter(source, *archiveFile, compression, archival, filter)
	} else {
		err = arc.Archive(source, *archiveFile, compression, archival)
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Archive created: %s\n", *archiveFile)
}

func handleExtract(cmd *flag.FlagSet, args []string) {
	// Flags for archive extraction
	archiveFile := cmd.String("f", "", "Archive file to extract (required)")

	cmd.Usage = func() {
		fmt.Println("Usage: arc extract [options] <destination_directory>")
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Validate required flags
	if *archiveFile == "" {
		fmt.Println("Error: Archive file (-f) is required")
		cmd.Usage()
		return
	}

	// Get destination directory
	destination := "."
	if cmd.NArg() > 0 {
		destination = cmd.Arg(0)
	}

	// Extract archive
	err := arc.Unarchive(*archiveFile, destination)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Archive extracted to: %s\n", destination)
}

func handleCompress(cmd *flag.FlagSet, args []string) {
	// Flags for file compression
	inputFile := cmd.String("i", "", "Input file to compress (required)")
	outputFile := cmd.String("o", "", "Output file (required)")
	compressionType := cmd.String("t", "zst", "Compression type: gzip/gz, bzip2/bz2, xz, zst, lz4, br, etc.")

	cmd.Usage = func() {
		fmt.Println("Usage: arc compress [options]")
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Validate required flags
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Error: Input (-i) and output (-o) files are required")
		cmd.Usage()
		return
	}

	// Get compression type
	compression, ok := arc.CompressionMap[strings.ToLower(*compressionType)]
	if !ok {
		log.Fatalf("Unsupported compression type: %s", *compressionType)
	}

	// Read input file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", *inputFile, err)
	}

	// Compress data
	out, err := arc.Compress(data, compression)
	if err != nil {
		log.Fatalf("Error compressing file %s: %v", *inputFile, err)
	}

	// Write output file
	if err := os.WriteFile(*outputFile, out, 0o644); err != nil {
		log.Fatalf("Error writing to file %s: %v", *outputFile, err)
	}

	log.Printf("File compressed: %s -> %s\n", *inputFile, *outputFile)
}

func handleDecompress(cmd *flag.FlagSet, args []string) {
	// Flags for file decompression
	inputFile := cmd.String("i", "", "Input compressed file (required)")
	outputFile := cmd.String("o", "", "Output file (required)")
	compressionType := cmd.String("t", "", "Compression type: gzip/gz, bzip2/bz2, xz, zst, lz4, br, etc. (required)")

	cmd.Usage = func() {
		fmt.Println("Usage: arc decompress [options]")
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(args); err != nil {
		log.Fatal(err)
	}

	// Validate required flags
	if *inputFile == "" || *outputFile == "" || *compressionType == "" {
		fmt.Println("Error: Input (-i), output (-o), and compression type (-t) are required")
		cmd.Usage()
		return
	}

	// Get compression type
	compression, ok := arc.CompressionMap[strings.ToLower(*compressionType)]
	if !ok {
		log.Fatalf("Unsupported compression type: %s", *compressionType)
	}

	// Read input file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", *inputFile, err)
	}

	// Decompress data
	out, err := arc.Decompress(data, compression)
	if err != nil {
		log.Fatalf("Error decompressing file %s: %v", *inputFile, err)
	}

	// Write output file
	if err := os.WriteFile(*outputFile, out, 0o644); err != nil {
		log.Fatalf("Error writing to file %s: %v", *outputFile, err)
	}

	log.Printf("File decompressed: %s -> %s\n", *inputFile, *outputFile)
}
