package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var extractCmd = &cobra.Command{
	Use:  "extract volfile [outdir] [filenames...]",
	Long: "vol extract unpacks the contents of a .vol file",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fn, outdir := args[0], "."
		if len(args) >= 2 && args[1] != "" {
			outdir = args[1]
		}

		var fnmatch FilenameSet
		if len(args) > 2 {
			fnmatch = args[2:]
		}

		if stat, err := os.Stat(outdir); err != nil {
			return fmt.Errorf("could not access directory %s: %w", outdir, err)
		} else if !stat.IsDir() {
			return fmt.Errorf("%s is not a directory: %w", outdir, err)
		}

		var v vol.File
		if data, err := os.ReadFile(fn); err != nil {
			return fmt.Errorf("could not read file %s: %w", fn, err)
		} else if err := v.Parse(data); err != nil {
			return fmt.Errorf("could not parse file %s: %w", fn, err)
		}

		for _, item := range v.Items {
			fn := filepath.Clean(item.Filename)
			if !fnmatch.Match(fn) {
				continue
			}

			if item.Compression != vol.None {
				log.Printf("cannot extract %s; compression %s unsupported\n", item.Filename, item.Compression)
				continue
			}

			if extractStripPaths {
				fn = filepath.Base(fn)
				if fn == "." || fn == "/" {
					log.Println("skipping empty filename")
					continue
				}
			}

			fnDir, fnBase := filepath.Dir(fn), filepath.Base(fn)
			if fnDir != "." {
				fnFullDir := filepath.Join(outdir, fnDir)
				if err := os.MkdirAll(fnFullDir, 0666|os.ModeDir); err != nil {
					return fmt.Errorf("could not create directory path %s: %w", fnFullDir, err)
				}
			}

			fnFull := filepath.Join(outdir, fnDir, fnBase)
			if err := os.WriteFile(fnFull, item.Payload, 0666); err != nil {
				return fmt.Errorf("could not create file %s: %w", fnFull, err)
			}
		}

		return nil
	},
}

var extractStripPaths bool

func init() {
	extractCmd.Flags().BoolVar(&extractStripPaths, "strip-paths", false, "ignore file paths in the vol; extract with no subdirectories")
}
