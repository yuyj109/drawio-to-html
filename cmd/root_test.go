package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	svg, err := os.ReadFile("../examples/notion-sample.drawio.svg")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "infra.drawio.svg")
	if err = os.WriteFile(path, svg, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	command := NewCommand()
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{path})
	if err = command.Execute(); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != path+".html\n" || stderr.Len() != 0 {
		t.Fatal("unexpected CLI output", stdout.String(), stderr.String())
	}
	if _, err = os.Stat(path + ".html"); err != nil {
		t.Fatal(err)
	}
	command = NewCommand()
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{path, "--force"})
	if err = command.Execute(); err != nil {
		t.Fatal("Cobra must support flags after the positional argument", err)
	}
}

func TestHelpAndArguments(t *testing.T) {
	for _, args := range [][]string{{}, {"one.svg", "two.svg"}, {"--unknown"}, {"missing.svg"}} {
		var stdout, stderr bytes.Buffer
		command := NewCommand()
		command.SetOut(&stdout)
		command.SetErr(&stderr)
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Fatalf("expected error for %v", args)
		}
		if stdout.Len() != 0 {
			t.Fatal("failure must not print an output path")
		}
	}
	var help bytes.Buffer
	command := NewCommand()
	command.SetOut(&help)
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil || !strings.Contains(help.String(), "drawio-to-html") {
		t.Fatal("help failed", err)
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("broken output") }

func TestCommandPropagatesOutputError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagram.svg")
	data, err := os.ReadFile("../examples/notion-sample.drawio.svg")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := NewCommand()
	command.SetArgs([]string{path})
	command.SetOut(brokenWriter{})
	if err = command.Execute(); err == nil || !strings.Contains(err.Error(), "broken output") {
		t.Fatal("stdout error ignored", err)
	}
}

func TestOverwriteConfirmation(t *testing.T) {
	for _, test := range []struct {
		name, input string
		overwrite   bool
	}{
		{"yes", "y\n", true},
		{"uppercase", "Y\n", true},
		{"word", " yes \n", true},
		{"yes at EOF", "y", true},
		{"no", "n\n", false},
		{"default no", "\n", false},
		{"EOF", "", false},
		{"unrecognized", "maybe\n", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := existingOutput(t)
			var stdout, stderr bytes.Buffer
			command := NewCommand()
			command.SetArgs([]string{path})
			command.SetIn(strings.NewReader(test.input))
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stderr.String(), path+".html. Overwrite? [y/N] ") {
				t.Fatal("overwrite prompt missing", stderr.String())
			}
			output, err := os.ReadFile(path + ".html")
			if err != nil {
				t.Fatal(err)
			}
			if test.overwrite {
				if stdout.String() != path+".html\n" || !bytes.Contains(output, []byte("<!doctype html>")) {
					t.Fatal("confirmed conversion failed", stdout.String())
				}
			} else if stdout.Len() != 0 || string(output) != "existing output" || !strings.Contains(stderr.String(), "Cancelled.") {
				t.Fatal("cancelled conversion changed output or reported success")
			}
		})
	}
}

func TestForceSkipsConfirmation(t *testing.T) {
	path := existingOutput(t)
	var stdout, stderr bytes.Buffer
	command := NewCommand()
	command.SetArgs([]string{"--force", path})
	command.SetIn(brokenReader{})
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 || stdout.String() != path+".html\n" {
		t.Fatal("--force prompted or did not report the output path")
	}
}

func TestConfirmationIOErrorsPreserveOutput(t *testing.T) {
	for _, promptFails := range []bool{false, true} {
		path := existingOutput(t)
		var stdout, stderr bytes.Buffer
		command := NewCommand()
		command.SetArgs([]string{path})
		command.SetIn(brokenReader{})
		command.SetOut(&stdout)
		command.SetErr(&stderr)
		if promptFails {
			command.SetErr(brokenWriter{})
		}
		if err := command.Execute(); err == nil {
			t.Fatal("confirmation I/O failure ignored")
		}
		output, err := os.ReadFile(path + ".html")
		if err != nil || string(output) != "existing output" || stdout.Len() != 0 {
			t.Fatal("confirmation failure changed the output", err)
		}
	}
}

func existingOutput(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "infra.drawio.svg")
	data, err := os.ReadFile("../examples/notion-sample.drawio.svg")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path+".html", []byte("existing output"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("broken input") }
