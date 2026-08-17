package objects

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Hash Function
func Hash(objType string, data []byte) (sha string, full []byte) {
	header := fmt.Sprintf("%s %d\x00", objType, len(data))
	full = append([]byte(header), data...)
	sum := sha1.Sum(full)
	return hex.EncodeToString(sum[:]), full
}

// Write the object to a folder
func WriteObject(gitDir, objType string, data []byte) (string, error) {
	sha, full := Hash(objType, data)
	objPath := filepath.Join(gitDir, "objects", sha[:2], sha[2:])

	if _, err := os.Stat(objPath); err == nil {
		return sha, nil
	}

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	if _, err := w.Write(full); err != nil {
		return "", fmt.Errorf("compress object: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("compress object: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(objPath), 0755); err != nil {
		return "", fmt.Errorf("create object dir: %w", err)
	}
	if err := os.WriteFile(objPath, compressed.Bytes(), 0444); err != nil {
		return "", fmt.Errorf("write object: %w", err)
	}

	return sha, nil
}

// Read Object from the folder
func ReadObject(gitDir, sha string) (objType string, data []byte, err error) {
	if len(sha) != 40 {
		return "", nil, fmt.Errorf("invalid sha %q: want 40 hex chars", sha)
	}
	objPath := filepath.Join(gitDir, "objects", sha[:2], sha[2:])

	compressed, err := os.ReadFile(objPath)
	if err != nil {
		return "", nil, fmt.Errorf("read object: %w", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return "", nil, fmt.Errorf("decompress object: %w", err)
	}
	defer r.Close()

	full, err := io.ReadAll(r)
	if err != nil {
		return "", nil, fmt.Errorf("decompress object: %w", err)
	}

	nullIdx := bytes.IndexByte(full, 0)
	if nullIdx == -1 {
		return "", nil, fmt.Errorf("invalid object %s: missing header null byte", sha)
	}

	var size int
	if _, err := fmt.Sscanf(string(full[:nullIdx]), "%s %d", &objType, &size); err != nil {
		return "", nil, fmt.Errorf("invalid object %s header: %w", sha, err)
	}

	data = full[nullIdx+1:]
	if len(data) != size {
		return "", nil, fmt.Errorf("invalid object %s: header size %d != actual %d", sha, size, len(data))
	}

	return objType, data, nil
}
