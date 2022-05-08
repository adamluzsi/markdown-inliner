package markdowninliner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Inliner struct {
	FS FS
}

type FS interface {
	fs.FS
	Write(name string)
}

const (
	mdTagInline = "markdown:inline"
	mdTagEnd    = "markdown:end"
)

// TODO: possible enhancement is to use os new line for the files
//       or the current file's new line character

func (i Inliner) Update(name string) error {
	file, err := i.FS.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return nil
	}

	bs, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	bs, err = i.removeInlines(bs)
	if err != nil {
		return err
	}

	bs, err = i.addInlines(bs)
	if err != nil {
		return err
	}
	_ = bs

	os.WriteFile()

	return nil
	//io.ReadAll(file)
	//
	//bs, err := i.removeInlines(bs)
	//
	//scanner := bufio.NewScanner(bytes.NewReader(bs))
	//scanner.Split(bufio.ScanLines)
	//
	//for scanner.Scan() {
	//	scanner.Text()
	//}
	//if err := scanner.Err(); err != nil {
	//	return nil, false, err
	//}
	//
	//return output.Bytes(), ok, nil
}

func (i Inliner) forEachLine(bs []byte, fn func(string) error) error {
	scanner := bufio.NewScanner(bytes.NewReader(bs))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fn(scanner.Text())
	}
	return scanner.Err()
}

func (i Inliner) addInlines(bs []byte) ([]byte, error) {
	output := &bytes.Buffer{}

	err := i.forEachLine(bs, func(line string) error {
		_, _ = fmt.Fprintln(output, line)

		if !strings.Contains(line, mdTagInline) {
			return nil // continue
		}

		markdownInlineDeclaration, err := i.parseMarkdownInline(line)
		if err != nil {
			fmt.Print("err of parsing", err.Error())
			return err
		}

		file, err := i.FS.Open(markdownInlineDeclaration.File)
		if err != nil {
			return err
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(output, "\n\n```%s\n%s\n```\n", markdownInlineDeclaration.FencedCodeBlockType, string(content))

		return nil
	})
	return output.Bytes(), err
}

type MarkdownInlineDeclaration struct {
	File                string
	FencedCodeBlockType string
	Section             string
}

var (
	/*
		example comments

		[comment]: <> (This is a comment, it will not be included)
		[comment]: <> (in  the output file unless you use it in)
		[comment]: <> (a reference style link.)
		[//]: <> (This is also a comment.)
		[//]: # (This may be the most platform independent comment)
	*/
	markdownComment = regexp.MustCompile(`\((.+)\)`)
	inlineFilePath  = regexp.MustCompile(`markdown:inline\s+(.+)(?:\s+\w+:)?(?:\s+)?`)
)

func (i Inliner) parseMarkdownInline(line string) (MarkdownInlineDeclaration, error) {
	commentMatch := markdownComment.FindAllStringSubmatch(line, -1)
	if len(commentMatch) == 0 {
		return MarkdownInlineDeclaration{}, fmt.Errorf("invalid markdown error comment format around markdown:inline")
	}

	content := commentMatch[0][1]

	match := inlineFilePath.FindAllStringSubmatch(content, -1)

	if len(match) == 0 {
		return MarkdownInlineDeclaration{}, fmt.Errorf("markdown:inline syntax")
	}

	filePath := match[0][1]

	const localPathPrefix = "./"
	if strings.HasPrefix(filePath, localPathPrefix) {
		filePath = strings.TrimPrefix(filePath, localPathPrefix)
	}

	return MarkdownInlineDeclaration{
		File:                filePath,
		FencedCodeBlockType: filepath.Ext(filePath),
		Section:             "",
	}, nil
}

func (i Inliner) removeInlines(bs []byte) ([]byte, error) {
	output := &bytes.Buffer{}

	var (
		duringInline   bool
		uncertainLines []any
	)
	if err := i.forEachLine(bs, func(line string) error {

		if strings.Contains(line, mdTagInline) {
			duringInline = true
		}

		if strings.Contains(line, mdTagEnd) {
			duringInline = false
			uncertainLines = nil
		}

		if duringInline {
			uncertainLines = append(uncertainLines, fmt.Sprintln(line))
			return nil // continue
		}

		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}

		return nil

	}); err != nil {
		return nil, err
	}

	if uncertainLines != nil {
		if _, err := fmt.Fprintln(output, uncertainLines...); err != nil {
			return nil, err
		}
	}

	return output.Bytes(), nil
}
