package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuyj109/drawio-to-html/internal/drawio"
)

// Execute runs the standalone command.
func Execute() error { return NewCommand().Execute() }

// NewCommand builds the CLI independently of process arguments and streams.
// Conversion itself lives in internal/drawio for later notion-sync integration.
func NewCommand() *cobra.Command {
	var force bool
	root := &cobra.Command{
		Use:           "drawio-to-html <file.drawio.svg>",
		Short:         "Convert a draw.io SVG into a Notion HTML block",
		Long:          "Convert a draw.io SVG into a Notion HTML block.\nCreates <file.drawio.svg>.html beside the source SVG.",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := drawio.ConvertFile(args[0], force)
			if !force && errors.Is(err, drawio.ErrOutputExists) {
				confirmed, confirmErr := confirmOverwrite(cmd, args[0]+".html")
				if confirmErr != nil {
					return confirmErr
				}
				if !confirmed {
					_, err = fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
					return err
				}
				path, err = drawio.ConvertFile(args[0], true)
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
	root.Flags().BoolVar(&force, "force", false, "Overwrite an existing output HTML without confirmation")
	root.CompletionOptions.DisableDefaultCmd = true
	return root
}

func confirmOverwrite(cmd *cobra.Command, path string) (bool, error) {
	if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Output already exists: %s. Overwrite? [y/N] ", path); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read overwrite confirmation: %w", err)
	}
	answer := strings.TrimSpace(line)
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes"), nil
}
