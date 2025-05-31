package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileUtils struct{}

func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

func (fu *FileUtils) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func (fu *FileUtils) FindSVGFiles(dir string) ([]string, error) {
	var svgFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".svg" {
			svgFiles = append(svgFiles, path)
		}

		return nil
	})

	return svgFiles, err
}
