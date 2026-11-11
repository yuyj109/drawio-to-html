package drawio

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConvertFileAndForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "日本語 infra.drawio.svg")
	original := wrapSVG(model)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := ConvertFile(path, false)
	if err != nil || output != path+".html" {
		t.Fatal(output, err)
	}
	first, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	changed := wrapSVG(strings.Replace(model, "日本語", "変更後", 1))
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ConvertFile(path, false); !errors.Is(err, ErrOutputExists) {
		t.Fatal("existing output must be preserved", err)
	}
	after, err := os.ReadFile(output)
	if err != nil || !bytes.Equal(first, after) {
		t.Fatal("output changed without --force", err)
	}
	if _, err = ConvertFile(path, true); err != nil {
		t.Fatal(err)
	}
	after, err = os.ReadFile(output)
	if err != nil || bytes.Equal(first, after) || !bytes.Contains(after, []byte("変更後")) {
		t.Fatal("forced output not updated", err)
	}
	source, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(source, changed) {
		t.Fatal("source SVG was modified", err)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 2 {
		t.Fatal("temporary file left behind", err)
	}
}

func TestInvalidFileCreatesNoOutput(t *testing.T) {
	for _, name := range []string{"invalid.svg", "wrong.drawio"} {
		dir := t.TempDir()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(`<svg/>`), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ConvertFile(path, false); err == nil {
			t.Fatal("invalid file accepted")
		}
		if _, err := os.Stat(path + ".html"); !os.IsNotExist(err) {
			t.Fatal("invalid conversion created an output", err)
		}
	}
	if _, err := ConvertFile(filepath.Join(t.TempDir(), "missing.svg"), false); err == nil {
		t.Fatal("missing input accepted")
	}
}

func TestConcurrentCreateDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "output.html")
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			if err := writeOutput(path, []byte("complete HTML"), false); err == nil {
				wins.Add(1)
			} else if !errors.Is(err, ErrOutputExists) {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal("expected exactly one output", wins.Load())
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "complete HTML" {
		t.Fatal("partial output", err)
	}
}
