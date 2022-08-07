package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

// const unsigned int DecodedLength(unsigned char *in);
// unsigned int Decode(unsigned char *in, unsigned char *out);
//import "C"

var rootCmd = &cobra.Command{
	Version: "v0.1",
	Use:     "vol",
	Long:    "vol manipulates DarkStar .vol files",

	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(unpackCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(dumpCmd)
}

func main() {
	rootCmd.ExecuteContext(context.Background())
}
