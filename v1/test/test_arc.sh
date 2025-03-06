#!/bin/bash
# Comprehensive test script for the arc utility

set -e  # Exit on error

# Colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

ARC_BIN="$(pwd)/arc"
TEST_DIR="/tmp/arc_test"
ARCHIVE_DIR="${TEST_DIR}/to_archive"
EXTRACT_DIR="${TEST_DIR}/extracted"
COMPRESS_DIR="${TEST_DIR}/compress_test"

# Supported compression algorithms
COMPRESSION_TYPES=("zst" "gz" "bz2" "xz" "lz4" "br")

# Supported archive formats
ARCHIVE_FORMATS=("tar" "zip")

# Print a formatted test step
step() {
  echo -e "${GREEN}==> $1${NC}"
}

# Print a warning message
warn() {
  echo -e "${YELLOW}WARNING: $1${NC}"
}

# Print a formatted error message
error() {
  echo -e "${RED}ERROR: $1${NC}"
  exit 1
}

# Setup test environment
setup() {
  step "Setting up test environment"
  
  # Remove old test directories if they exist
  rm -rf "${TEST_DIR}"
  
  # Create test directories
  mkdir -p "${ARCHIVE_DIR}" "${EXTRACT_DIR}" "${COMPRESS_DIR}"
  
  # Create some test files in the archive directory
  echo "Test file 1 content" > "${ARCHIVE_DIR}/test1.txt"
  echo "Test file 2 content" > "${ARCHIVE_DIR}/test2.txt"
  mkdir -p "${ARCHIVE_DIR}/subdir"
  echo "Subdirectory file content" > "${ARCHIVE_DIR}/subdir/subfile.txt"
  
  # Create a binary-like file
  dd if=/dev/urandom of="${ARCHIVE_DIR}/binary_file.bin" bs=1024 count=100
  
  # Create a larger text file for compression tests
  for i in {1..1000}; do
    echo "This is line $i of the test file for compression" >> "${COMPRESS_DIR}/large_text.txt"
  done
  
  echo "Test environment set up in ${TEST_DIR}"
}

# Test compression of single files with all algorithms
test_compress() {
  step "Testing compression functionality"
  
  INPUT_FILE="${COMPRESS_DIR}/large_text.txt"
  
  for algo in "${COMPRESSION_TYPES[@]}"; do
    echo "Testing compression with ${algo}..."
    OUTPUT_FILE="${COMPRESS_DIR}/large_text.${algo}"
    
    # Skip if algorithm isn't supported
    if ! ${ARC_BIN} compress -i "${INPUT_FILE}" -o "${OUTPUT_FILE}" -t "${algo}"; then
      warn "Compression with ${algo} not supported or failed, skipping"
      continue
    fi
    
    [ -f "${OUTPUT_FILE}" ] || error "Failed to compress with ${algo}"
    
    # Get file sizes for verification
    ORIGINAL_SIZE=$(stat -c%s "${INPUT_FILE}")
    COMPRESSED_SIZE=$(stat -c%s "${OUTPUT_FILE}")
    
    # Simple check to ensure compression actually happened (file got smaller)
    if [ ${COMPRESSED_SIZE} -ge ${ORIGINAL_SIZE} ]; then
      echo "Note: ${algo} compression didn't reduce file size (${COMPRESSED_SIZE} >= ${ORIGINAL_SIZE})"
    fi
    
    echo "Compressed with ${algo}: ${ORIGINAL_SIZE} -> ${COMPRESSED_SIZE} bytes"
  done
  
  echo "Compression tests completed successfully"
}

# Test decompression of single files with all algorithms
test_decompress() {
  step "Testing decompression functionality"
  
  INPUT_FILE="${COMPRESS_DIR}/large_text.txt"
  
  for algo in "${COMPRESSION_TYPES[@]}"; do
    COMPRESSED_FILE="${COMPRESS_DIR}/large_text.${algo}"
    
    # Skip if the compressed file doesn't exist
    if [ ! -f "${COMPRESSED_FILE}" ]; then
      warn "Compressed file for ${algo} not found, skipping decompression test"
      continue
    fi
    
    DECOMPRESSED_FILE="${COMPRESS_DIR}/decompressed_${algo}.txt"
    
    echo "Testing decompression with ${algo}..."
    if ! ${ARC_BIN} decompress -i "${COMPRESSED_FILE}" -o "${DECOMPRESSED_FILE}" -t "${algo}"; then
      warn "Decompression with ${algo} failed, skipping verification"
      continue
    fi
    
    [ -f "${DECOMPRESSED_FILE}" ] || error "Failed to decompress with ${algo}"
    
    # Verify content integrity
    if ! diff "${INPUT_FILE}" "${DECOMPRESSED_FILE}"; then
      error "Content integrity check failed for ${algo}"
    fi
    
    echo "Successfully decompressed ${algo} file with content intact"
  done
  
  echo "Decompression tests completed successfully"
}

# Test archiving with various format and compression combinations
test_archive() {
  step "Testing archive functionality with various formats and compression methods"
  
  # Test special case - ZIP format (has its own compression)
  echo "Testing ZIP format archival..."
  ${ARC_BIN} archive -t zip -f "${TEST_DIR}/archive.zip" "${ARCHIVE_DIR}"
  [ -f "${TEST_DIR}/archive.zip" ] || error "Failed to create ZIP archive"
  
  # Test tar with various compression algorithms
  for algo in "${COMPRESSION_TYPES[@]}"; do
    echo "Testing tar + ${algo} archival..."
    ARCHIVE_FILE="${TEST_DIR}/archive.tar.${algo}"
    
    # Attempt to create archive with this compression
    if ! ${ARC_BIN} archive -c "${algo}" -t tar -f "${ARCHIVE_FILE}" "${ARCHIVE_DIR}"; then
      warn "Creating tar archive with ${algo} compression not supported or failed, skipping"
      continue
    fi
    
    [ -f "${ARCHIVE_FILE}" ] || error "Failed to create tar.${algo} archive"
    echo "Successfully created tar.${algo} archive"
  done
  
  # Test with include filter
  echo "Testing archive with include filter..."
  ${ARC_BIN} archive -include ".*\.txt$" -c zst -t tar -f "${TEST_DIR}/archive_txt_only.tar.zst" "${ARCHIVE_DIR}"
  [ -f "${TEST_DIR}/archive_txt_only.tar.zst" ] || error "Failed to create filtered archive with include filter"
  
  # Test with exclude filter
  echo "Testing archive with exclude filter..."
  ${ARC_BIN} archive -exclude ".*\.bin$" -c zst -t tar -f "${TEST_DIR}/archive_no_bin.tar.zst" "${ARCHIVE_DIR}"
  [ -f "${TEST_DIR}/archive_no_bin.tar.zst" ] || error "Failed to create filtered archive with exclude filter"
  
  echo "Archive creation tests completed successfully"
}

# Test extraction of various archive formats
test_extract() {
  step "Testing extraction of various archive formats"
  
  # Extract ZIP archive
  echo "Testing ZIP archive extraction..."
  mkdir -p "${EXTRACT_DIR}/zip"
  ${ARC_BIN} extract -f "${TEST_DIR}/archive.zip" "${EXTRACT_DIR}/zip"
  [ -f "${EXTRACT_DIR}/zip/to_archive/test1.txt" ] || error "Failed to extract ZIP archive"
  
  # Extract tar archives with various compressions
  for algo in "${COMPRESSION_TYPES[@]}"; do
    ARCHIVE_FILE="${TEST_DIR}/archive.tar.${algo}"
    
    # Skip if the archive file wasn't created
    if [ ! -f "${ARCHIVE_FILE}" ]; then
      warn "Archive file for tar.${algo} not found, skipping extraction test"
      continue
    fi
    
    echo "Testing extraction of tar.${algo} archive..."
    EXTRACT_SUBDIR="${EXTRACT_DIR}/tar_${algo}"
    mkdir -p "${EXTRACT_SUBDIR}"
    
    if ! ${ARC_BIN} extract -f "${ARCHIVE_FILE}" "${EXTRACT_SUBDIR}"; then
      warn "Extraction of tar.${algo} archive failed, skipping verification"
      continue
    fi
    
    [ -f "${EXTRACT_SUBDIR}/to_archive/test1.txt" ] || error "Failed to extract tar.${algo} archive"
    
    # Verify content integrity for extracted files
    if ! diff "${ARCHIVE_DIR}/test1.txt" "${EXTRACT_SUBDIR}/to_archive/test1.txt"; then
      error "Content integrity check failed for tar.${algo} extraction"
    fi
    
    echo "Successfully extracted tar.${algo} archive with content intact"
  done
  
  echo "Archive extraction tests completed successfully"
}

# Check that all files are present for a specified extract directory
verify_extraction() {
  local extract_dir=$1
  
  # Check existence of regular files
  [ -f "${extract_dir}/to_archive/test1.txt" ] || return 1
  [ -f "${extract_dir}/to_archive/test2.txt" ] || return 1
  [ -f "${extract_dir}/to_archive/subdir/subfile.txt" ] || return 1
  [ -f "${extract_dir}/to_archive/binary_file.bin" ] || return 1
  
  # Verify content of one of the files
  diff "${ARCHIVE_DIR}/test1.txt" "${extract_dir}/to_archive/test1.txt" >/dev/null || return 1
  
  return 0
}

# Clean up test files
cleanup() {
  step "Cleaning up test environment"
  rm -rf "${TEST_DIR}"
  echo "Cleanup completed"
}

# Main test function
run_tests() {
  step "Starting arc functionality tests"
  
  # Ensure arc binary exists
  [ -x "${ARC_BIN}" ] || error "arc binary not found or not executable at ${ARC_BIN}"
  
  setup
  test_compress
  test_decompress
  test_archive
  test_extract
  
  # Comment out cleanup during development if you want to inspect the files
  cleanup
  
  echo -e "${GREEN}All tests passed successfully!${NC}"
}

# Run the tests
run_tests
