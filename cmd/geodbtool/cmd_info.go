package main

import (
	"fmt"
	"strings"

	"github.com/anexia-it/geodbtools"
	"github.com/spf13/cobra"
)

var cmdInfo = &cobra.Command{
	Use:   "info <database>",
	Short: `Print information about a GeoIP database file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbPath := args[0]
		formatName, _ := cmd.Flags().GetString("format")

		var source geodbtools.ReaderSource

		if source, err = geodbtools.NewFileReaderSource(dbPath); err != nil {
			return
		}
		defer source.Close()

		var format geodbtools.Format

		if formatName == "auto" {
			if format, err = geodbtools.DetectFormat(source); err != nil {
				return
			}
		} else if format, err = geodbtools.LookupFormat(formatName); err != nil {
			return
		}

		var meta geodbtools.Metadata
		if _, meta, err = format.NewReaderAt(source); err != nil {
			return
		}

		cmd.Printf("format         : %s\n", format.FormatName())
		cmd.Printf("type           : %s\n", meta.Type)
		cmd.Printf("description    : %s\n", meta.Description)
		cmd.Printf("format version : %d.%d\n", meta.MajorFormatVersion, meta.MinorFormatVersion)
		cmd.Printf("build time     : %s\n", meta.BuildTime)
		cmd.Printf("IP version     : %d\n", meta.IPVersion)

		return
	},
}

func init() {
	cmdInfo.Flags().StringP("format", "f", "auto", fmt.Sprintf("database format (auto|%s)", strings.Join(geodbtools.FormatNames(), "|")))
	cmdRoot.AddCommand(cmdInfo)
}
