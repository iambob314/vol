package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var dumpCmd = &cobra.Command{
	Use:  "dump volfile [files...]",
	Long: "vol dump prints the contents of a .vol file to stdout (only specific files, or whole file if none specified)",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volFN, fns := args[0], args[1:]

		var v vol.File
		if data, err := os.ReadFile(volFN); err != nil {
			return fmt.Errorf("could not read file %s: %w", volFN, err)
		} else if err := v.Parse(data); err != nil {
			return fmt.Errorf("could not parse file %s: %w", volFN, err)
		}

		var fnmatch FilenameSet
		if len(fns) > 0 {
			fnmatch = fns
		}

		for _, item := range v.Items {
			fn := filepath.Clean(item.Filename)
			if !fnmatch.Match(fn) {
				continue
			}

			content := item.Payload
			switch item.Compression {
			case vol.None:
				// content is already correct
			default:
				content = []byte(fmt.Sprintf("(content is %s compressed; decompression unsupported at this time)", item.Compression))
			}

			if dumpFlags.Raw {
				fmt.Printf("*** %s ***\n", fn)
			}
			_, _ = os.Stdout.Write(content)
			if dumpFlags.Raw {
				if l := len(item.Payload); l == 0 || (item.Payload[l-1] != '\r' && item.Payload[l-1] != '\n') {
					fmt.Println()
					fmt.Println("(no newline at end of file)")
				}
			}
		}

		return nil
	},
}

var dumpFlags struct {
	Raw bool
}

func init() {
	dumpCmd.Flags().BoolVar(&dumpFlags.Raw, "raw", false, "dump file content only and exactly (after decompression);\ndo not print filenames, 'no newline at end of file',\nor separators between content of multiple files")
}
