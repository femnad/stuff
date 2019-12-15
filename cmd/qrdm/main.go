package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alexflint/go-arg"
)

const bufferSizeBytes = 8192

var args struct {
	URL string `arg:"required"`
}

func hashURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error fetching URL %s: %w", url, err)
	}
	sha256Hash := sha256.New()
	buffer := make([]byte, bufferSizeBytes)
	done := false
	for {
		read, err := resp.Body.Read(buffer)
		if errors.Is(err, io.EOF) {
			done = true
		} else if err != nil {
			return "", fmt.Errorf("Error reading response: %w", err)
		}
		if read < bufferSizeBytes {
			buffer = buffer[:read]
		}
		_, err = sha256Hash.Write(buffer)
		if err != nil {
			return "", fmt.Errorf("Error updating hash: %w", err)
		}
		if done {
			break
		} else {
			buffer = make([]byte, bufferSizeBytes)
		}
	}
	return hex.EncodeToString(sha256Hash.Sum(nil)), nil
}

func main() {
	arg.MustParse(&args)
	sha256Checksum, err := hashURL(args.URL)
	if err != nil {
		panic(err)
	}
	fmt.Println(sha256Checksum)
}
