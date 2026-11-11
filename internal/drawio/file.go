package drawio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrOutputExists indicates that conversion would replace an existing file.
var ErrOutputExists = errors.New("output already exists")

// ConvertFile writes <input>.html and returns its path. Existing output is
// preserved unless force is true. The input file is never modified.
func ConvertFile(input string, force bool) (string, error) {
	if !strings.EqualFold(filepath.Ext(input), ".svg") {
		return "", errors.New("input must be a .svg file containing draw.io diagram data")
	}
	if err := convertFile(input, force); err != nil {
		return "", err
	}
	return input + ".html", nil
}

func convertFile(input string, force bool) error {
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("open SVG: %w", err)
	}
	defer func() { _ = file.Close() }()
	svg, err := io.ReadAll(io.LimitReader(file, maxSVGSize+1))
	if err != nil {
		return fmt.Errorf("read SVG: %w", err)
	}
	if len(svg) > maxSVGSize {
		return fmt.Errorf("SVG exceeds the %d MiB input limit", maxSVGSize>>20)
	}
	output, err := Render(svg, filepath.Base(input))
	if err != nil {
		return err
	}
	return writeOutput(input+".html", output, force)
}

// Write completely before publishing the result. Link prevents accidental
// replacement, including a concurrent conversion; --force explicitly uses rename.
func writeOutput(path string, content []byte, force bool) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".drawio-to-html-*")
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	_, writeErr := io.Copy(file, bytes.NewReader(content))
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if err = errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write HTML: %w", err)
	}
	if force {
		err = os.Rename(temp, path)
	} else {
		err = os.Link(temp, path)
	}
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %s (use --force to overwrite)", ErrOutputExists, path)
	}
	if err != nil {
		return fmt.Errorf("save HTML: %w", err)
	}
	return nil
}
