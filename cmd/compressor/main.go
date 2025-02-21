package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jm33-m0/arc"
)

func main() {
	to_compress := flag.String("c", "", "File to compress")
	to_decompress := flag.String("f", "", "Compressed file to decompress")
	output := flag.String("o", "", "Output file")
	compressionType := flag.String("t", "", "Compression type (e.g., bz2, gz, xz, zst, lz4, br)")
	flag.Parse()

	if *output == "" || *compressionType == "" {
		flag.Usage()
		return
	}

	file := *to_decompress
	if file == "" {
		file = *to_compress
	}
	if file == "" {
		log.Fatalf("No file to compress or decompress")
	}

	compression, ok := arc.CompressionMap[strings.ToLower(*compressionType)]
	if !ok {
		log.Fatalf("Unsupported compression type: %s", *compressionType)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", file, err)
	}

	var out []byte
	if *to_decompress != "" {
		out, err = arc.Decompress(data, compression)
		if err != nil {
			log.Fatalf("Error decompressing file %s: %v", file, err)
		}
	} else if *to_compress != "" {
		// out, err = arc.Compress(data, compression)
		out, err = arc.Compress(data, compression)
		if err != nil {
			log.Fatalf("Error compressing file %s: %v", file, err)
		}
	}

	if err := os.WriteFile(*output, out, 0o644); err != nil {
		log.Fatalf("Error writing to file %s: %v", *output, err)
	}

	log.Printf("Success: %s", *output)
}
