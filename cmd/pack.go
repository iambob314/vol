package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var packCmd = &cobra.Command{
	Use:  "pack volfile [file...]",
	Long: "vol pack packs files into a new or existing .vol file",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volFN, fns := args[0], args[1:]

		var v vol.File
		if data, err := os.ReadFile(volFN); os.IsNotExist(err) { // file does not exist
			// nothing to do; leave v empty
		} else if err != nil { // file could not be read
			return fmt.Errorf("could not read vol file %s: %w", volFN, err)
		} else if err := v.Parse(data); err != nil { // file could not be parsed
			return fmt.Errorf("could not parse vol file %s: %w", volFN, err)
		}

		// Keep a table of filename-to-itemidx for the vol file, so we can error or overwrite on duplicate
		filesInVol := make(map[string]int, len(v.Items))
		for i, it := range v.Items {
			filesInVol[filepath.Clean(it.Filename)] = i
		}

		// Expand fileglobs (for Windows, which does not do this in the shell...)
		var expandedFNs []string
		for _, fn := range fns {
			if expanded, err := filepath.Glob(fn); err != nil {
				return fmt.Errorf("invalid filename or glob pattern %s: %w", fn, err)
			} else if expanded != nil {
				expandedFNs = append(expandedFNs, expanded...)
			}
		}
		fns = expandedFNs

		// Load and append all files as vol items (overwriting existing items where needed/allowed)
		for _, fn := range fns {
			fn = filepath.Clean(fn)

			fnInPack := fn
			if packStripPaths {
				fnInPack = filepath.Base(fn)
			}

			if data, err := os.ReadFile(fn); err != nil {
				return fmt.Errorf("could not read input file %s: %w", fn, err)
			} else if idxInVol, overwrite := filesInVol[fnInPack]; overwrite && !allowOverwrite {
				return fmt.Errorf("file %s already exists in vol file %s; use --overwrite to overwrite", fnInPack, volFN)
			} else {
				newItem := vol.Item{
					Compression: vol.None,
					Filename:    fnInPack,
					Payload:     data,
				}

				msg := "packing " + fn
				if fn != fnInPack {
					msg += " (as " + fnInPack + ")"
				}
				if overwrite {
					msg += " (overwrite)"
				}
				fmt.Println(msg)

				if overwrite {
					v.Items[idxInVol] = newItem
				} else {
					filesInVol[fnInPack] = len(v.Items)
					v.Items = append(v.Items, newItem)
				}
			}
		}

		var newVolData vol.ByteBuffer
		v.Store(&newVolData)

		if err := os.WriteFile(volFN, newVolData, 0666); err != nil {
			return fmt.Errorf("could not create or overwrite vol file %s: %w", volFN, err)
		}
		return nil
	},
}

var (
	packStripPaths bool
	allowOverwrite bool
)

func init() {
	packCmd.Flags().BoolVar(&packStripPaths, "strip-paths", false, "remove file paths when packing files into the vol; keep only filenames")
	packCmd.Flags().BoolVar(&allowOverwrite, "overwrite", false, "allow overwriting files when packing into an existing vol; if absent, error on attempted overwrite")
}
