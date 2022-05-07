package markdowninliner_test

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	markdowninliner "github.com/adamluzsi/markdown-inliner"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

//go:embed internal/assets/example/*
var examples embed.FS

func getExample(tb testing.TB, name string) []byte {
	bytes, err := examples.ReadFile(filepath.Join("internal", "assets", "example", name))
	assert.Must(tb).Nil(err)
	return bytes
}

func TestInliner_Update(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	fileContent := testcase.Var[[]byte]{ID: "fileContent"}

	fs := testcase.Let(s, func(t *testcase.T) fs.FS {
		return fstest.MapFS{
			"file.md": {
				Data:    fileContent.Get(t),
				Mode:    fs.ModePerm,
				ModTime: time.Now(),
			},
			"file.txt": {
				Data:    []byte("hello, world"),
				Mode:    0,
				ModTime: time.Now(),
			},
			"examples": {
				Mode:    os.ModeDir,
				ModTime: time.Now(),
			},
			filepath.Join("examples", "full.go"): {
				Data:    getExample(t, "full.go"),
				Mode:    fs.ModePerm,
				ModTime: time.Now(),
			},
			filepath.Join("examples", "partials.go"): {
				Data:    getExample(t, "partials.go"),
				Mode:    fs.ModePerm,
				ModTime: time.Now(),
			},
		}
	})

	path := testcase.Var[string]{ID: "path"}

	subject := func(t *testcase.T) error {
		return markdowninliner.Inliner{FS: fs.Get(t)}.Update(path.Get(t))
	}
	onSuccess := func(t *testcase.T) {
		t.Must.Nil(subject(t))
	}

	s.When("the path points to a non markdown file", func(s *testcase.Spec) {
		path.LetValue(s, "file.txt")
		fileContent.Let(s, func(t *testcase.T) []byte {
			return nil
		})

		s.Then("it will leave the file as is", func(t *testcase.T) {
			onSuccess(t)

			t.Must.Equal("hello, world", getFileContent(t, fs.Get(t), path.Get(t)))
		})
	})

	s.When("the path points to a markdown file", func(s *testcase.Spec) {
		path.LetValue(s, "file.md")

		s.And("the file content has no inline parts", func(s *testcase.Spec) {
			fileContent.Let(s, func(t *testcase.T) []byte {
				return []byte("no inline here")
			})

			s.Then("it succeed without affecting the file", func(t *testcase.T) {
				onSuccess(t)
				actualContent := getFileContent(t, fs.Get(t), "file.md")
				t.Must.Equal("no inline here", actualContent)
			})
		})

		s.And("the file content has inline part for full file", func(s *testcase.Spec) {
			fileContent.Let(s, func(t *testcase.T) []byte {
				return []byte(strings.Join(
					[]string{
						"some string",
						"",
						"[//]: # (markdown:inline ./examples/full.go)",
					},
					"\n",
				))
			})

			s.Then("it succeed while updating the affecting the file", func(t *testcase.T) {
				onSuccess(t)

				actualContent := getFileContent(t, fs.Get(t), "file.md")

				t.Must.Equal(strings.Join(
					[]string{
						"some string",
						"",
						"[//]: # (markdown:include ./examples/full.go)",
						"```go",
						getFileContent(t, fs.Get(t), "examples/full.go"),
						"```",
						"[//]: # (markdown:include:end)",
					},
					"\n",
				), actualContent)
			})
		})
	})

}

func getFileContent(tb testing.TB, fs fs.FS, name string) string {
	file, err := fs.Open(name)
	assert.Must(tb).Nil(err)
	bytes, err := io.ReadAll(file)
	assert.Must(tb).Nil(err)
	return string(bytes)
}
