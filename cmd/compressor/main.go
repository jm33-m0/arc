package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jm33-m0/arc"
)

func main() {
	file := flag.String("f", "", "Compressed file to decompress")
	output := flag.String("o", "", "Output file")
	compressionType := flag.String("t", "xz", "Compression type (e.g., bz2, gz, xz, zst, lz4, br)")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	if *file == "" || *output == "" {
		log.Fatal("Both --file and --output must be specified")
	}

	compression, ok := arc.CompressionMap[strings.ToLower(*compressionType)]
	if !ok {
		log.Fatalf("Unsupported compression type: %s", *compressionType)
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", *file, err)
	}

	out, err := arc.Decompress(data, compression)
	if err != nil {
		log.Fatalf("Error decompressing file %s: %v", *file, err)
	}

	if err := os.WriteFile(*output, out, 0o644); err != nil {
		log.Fatalf("Error writing to file %s: %v", *output, err)
	}

	log.Printf("File decompressed successfully to %s", *output)
}
