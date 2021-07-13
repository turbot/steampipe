package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"os"
)

func FileHash(filePath string) (string, error) {
	LogTime(fmt.Sprintf("utils.FileHash %s start", filePath))
	defer LogTime(fmt.Sprintf("utils.FileHash %s end", filePath))

	// get checksum
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func StringHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
