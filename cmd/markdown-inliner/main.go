package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	iterate(currentDirectory)
}

func iterate(dirPath string) error {
	i := Inliner{}
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		const markdownFileExt = "md"
		if strings.ToLower(filepath.Ext(path)) != markdownFileExt {
			return nil // continue
		}

		bs, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		bs, ok, err := i.Update(bs)
		if err != nil {
			return err
		}
		if ok {
			if err := os.WriteFile(path, bs, info.Mode().Perm()); err != nil {
				return err
			}
		}

		return nil
	})
}
