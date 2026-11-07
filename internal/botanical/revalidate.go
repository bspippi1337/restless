package botanical

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxCachedManifestLines = 50000

// readCachedRecords loads archived metadata locally; it makes no network requests.
// Repeated URLs keep their newest entry, matching BuildCatalog semantics.
func readCachedRecords(directory string) (map[string]Record, error) {
	records := make(map[string]Record)
	f, err := os.Open(filepath.Join(directory, "index.jsonl"))
	if errors.Is(err, os.ErrNotExist) { return records, nil }
	if err != nil { return nil, err }
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	lines := 0
	for scanner.Scan() {
		lines++
		if lines > maxCachedManifestLines {
			return nil, fmt.Errorf("archive manifest exceeds maximum %d lines", maxCachedManifestLines)
		}
		var record Record
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("invalid cached archive manifest line %d: %w", lines, err)
		}
		if record.URL == "" || record.Status != http.StatusOK {
			continue
		}
		records[record.URL] = record
	}
	if err := scanner.Err(); err != nil { return nil, err }
	return records, nil
}

// verifiedCachedBody reads a local archived representation and checks its hash
// before it can be reused after HTTP 304. It rejects arbitrary paths/symlinks
// into other locations by opening only the expected digest filename.
func verifiedCachedBody(directory string, record Record) ([]byte, error) {
	if len(record.SHA256) != 64 || record.ArchivePath != filepath.Join("sha256", record.SHA256) {
		return nil, fmt.Errorf("invalid cached archive metadata for %s", record.URL)
	}
	if _, err := hex.DecodeString(record.SHA256); err != nil {
		return nil, fmt.Errorf("invalid cached SHA256 for %s", record.URL)
	}
	archive := filepath.Join(directory, "sha256", record.SHA256)
	fi, err := os.Lstat(archive)
	if err != nil { return nil, err }
	if !fi.Mode().IsRegular() || fi.Size() > maxDocumentBytes {
		return nil, fmt.Errorf("cached archive is not a bounded regular file: %s", record.URL)
	}
	b, err := os.ReadFile(archive)
	if err != nil { return nil, err }
	if int64(len(b)) > maxDocumentBytes {
		return nil, fmt.Errorf("cached document exceeds size limit: %s", record.URL)
	}
	hash := sha256.Sum256(b)
	if hex.EncodeToString(hash[:]) != record.SHA256 {
		return nil, fmt.Errorf("cached archive checksum mismatch: %s", record.URL)
	}
	if !supportedType(record.ContentType) {
		return nil, fmt.Errorf("cached document is not an approved content type: %s", record.URL)
	}
	return b, nil
}

// verified conditional request headers are used only with local checked content.
func boundedETag(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 256 || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	return value
}
func boundedLastModified(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 256 || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	if _, err := http.ParseTime(value); err != nil { return "" }
	return value
}
