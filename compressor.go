package arc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/mholt/archives"
)

// Compress compresses input data using specified compressor.
func Compress(data []byte, compression archives.Compression) ([]byte, error) {
	var compressedBuf bytes.Buffer
	logging("Compressing data using %s", compression.Extension())

	// Wrap the buffer with a compressor
	compressor, err := compression.OpenWriter(&compressedBuf)
	if err != nil {
		return nil, fmt.Errorf("Compress: Failed to create compressor: %w", err)
	}
	defer compressor.Close()

	// Writes to compressor will be compressed
	_, err = compressor.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Compress: Write to compressor failed: %w", err)
	}

	// without this line, the compressed data will be incomplete
	compressor.Close()

	return compressedBuf.Bytes(), nil
}

// Decompress decompresses input compressed data.
func Decompress(data []byte, compression archives.Compression) ([]byte, error) {
	stream := bytes.NewReader(data)

	// Open a reader for decompression using the provided decompressor
	rc, err := compression.OpenReader(stream)
	if err != nil {
		return nil, fmt.Errorf("Decompress: Failed to open decompression reader: %w", err)
	}
	defer rc.Close()

	// Read decompressed data into a buffer
	var decompressedBuf bytes.Buffer
	_, err = io.Copy(&decompressedBuf, rc)
	if err != nil {
		return nil, fmt.Errorf("Decompress: Failed to read from decompressor: %w", err)
	}

	return decompressedBuf.Bytes(), nil
}

// CompressBz2 compresses input data using BZ2 compressor.
func CompressBz2(data []byte) ([]byte, error) {
	return Compress(data, archives.Bz2{})
}

// DecompressBz2 decompresses input compressed data using BZ2 decompressor.
func DecompressBz2(data []byte) ([]byte, error) {
	return Decompress(data, archives.Bz2{})
}

// CompressXz compresses input data using XZ compressor.
func CompressXz(data []byte) ([]byte, error) {
	return Compress(data, archives.Xz{})
}

// DecompressXz decompresses input compressed data using XZ decompressor.
func DecompressXz(data []byte) ([]byte, error) {
	return Decompress(data, archives.Xz{})
}
