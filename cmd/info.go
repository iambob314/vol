package main

import (
	"fmt"
	"github.com/iambob314/vol"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var infoCmd = &cobra.Command{
	Use:  "info volfile [volfile ...]",
	Long: "vol info summarizes the contents of a .vol file",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, fn := range args {
			var v vol.File

			if data, err := ioutil.ReadFile(fn); err != nil {
				return fmt.Errorf("could not read file %s: %w", fn, err)
			} else if err := v.Parse(data); err != nil {
				return fmt.Errorf("could not parse file %s: %w", fn, err)
			}

			fmt.Printf("%s contains %d files:\n", fn, len(v.Items))
			for _, item := range v.Items {
				fmt.Printf("%s:\t%d bytes\t(compression: %s)\n", item.Filename, len(item.Payload), item.Compression)

				if item.Compression == vol.LZH {
					//data := append(make([]byte, 4), item.Payload...)
					//binary.LittleEndian.PutUint32(data, uint32(len(item.Payload)))
					//
					//newLen := C.DecodedLength((*C.uchar)(&data[0]))
					//fmt.Println(newLen)
					//
					//out := make([]byte, newLen)
					//n := C.Decode((*C.uchar)(&data[0]), (*C.uchar)(&out[0]))
					//fmt.Println(n, string(out))
				}
			}
		}
		return nil
	},
}
