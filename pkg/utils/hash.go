package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"os"
)

func FileMD5Hash(filePath string) ([]byte, error) {
	LogTime(fmt.Sprintf("utils.FileMD5Hash %s start", filePath))
	defer LogTime(fmt.Sprintf("utils.FileMD5Hash %s end", filePath))

	// get checksum
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func FileHash(filePath string) (string, error) {
	LogTime(fmt.Sprintf("utils.FileHash %s start", filePath))
	defer LogTime(fmt.Sprintf("utils.FileHash %s end", filePath))
	hash, err := FileMD5Hash(filePath)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

func StringHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	h := hex.EncodeToString(hash[:])
	return h
}
