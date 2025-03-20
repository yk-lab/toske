package utils

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/yk-lab/toske/static"
)

func PrintAAFromTxt(fp string) error {
	text, err := readFile(filepath.Join("aa", fp))
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

func AAFromText(fp string) (string, error) {
	return readFile(filepath.Join("aa", fp))
}

func readFile(fp string) (string, error) {
	file, err := static.Aa.Open(fp)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}
	return buf.String(), nil
}
