package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"os"
	"path"
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

		var fnsToDump map[string]bool
		if len(fns) > 0 {
			fnsToDump = make(map[string]bool, len(fns))
			for _, fn := range fns {
				fnsToDump[path.Clean(fn)] = true
			}
		}

		for _, item := range v.Items {
			fn := path.Clean(item.Filename)
			if fnsToDump == nil || fnsToDump[fn] {
				if len(fnsToDump) != 1 {
					fmt.Printf("*** %s ***\n", fn)
				}
				switch item.Compression {
				case vol.None:
					fmt.Printf("%s", item.Payload)
					if l := len(item.Payload); l == 0 || (item.Payload[l-1] != '\r' && item.Payload[l-1] != '\n') {
						fmt.Println()
						fmt.Println("(no newline at end of file)")
					}
				default:
					fmt.Printf("(content is compressed; decompression unsupported at this time)\n")
				}
			}
		}

		return nil
	},
}

func init() {}
