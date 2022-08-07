package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var unpackCmd = &cobra.Command{
	Use:  "unpack volfile [outdir] [filenames...]",
	Long: "vol unpack unpacks the contents of a .vol file",
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
			fnInVol := filepath.Clean(item.Filename)
			if !fnmatch.Match(fnInVol) {
				continue
			}

			if item.Compression != vol.None {
				fmt.Printf("cannot unpack %s; compression %s unsupported\n", item.Filename, item.Compression)
				continue
			}

			fn := fnInVol
			if unpackFlags.StripPaths {
				fn = filepath.Base(fn)
				if fn == "." || fn == "/" {
					fmt.Println("skipping empty filename")
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

			if fnFull == fnInVol {
				fmt.Printf("unpacked %s\n", fnFull)
			} else {
				fmt.Printf("unpacked %s to %s\n", fnInVol, fnFull)
			}
		}

		return nil
	},
}

var unpackFlags struct {
	StripPaths bool
}

func init() {
	unpackCmd.Flags().BoolVar(&unpackFlags.StripPaths, "strip-paths", false, "ignore file paths in the vol; unpack with no subdirectories")
}
