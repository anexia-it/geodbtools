package main

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/anexia-it/geodbtools"
	"github.com/spf13/cobra"
)

var cmdLookup = &cobra.Command{
	Use:   "lookup <ip address>",
	Short: "Look up GeoIP information for an IP address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ip := net.ParseIP(args[0])
		if ip == nil {
			err = errors.New("invalid IP address")
			return
		}

		var formatName string
		var dbPath string
		var verbose bool
		dbPath, _ = cmd.Flags().GetString("db")
		formatName, _ = cmd.Flags().GetString("format")
		verbose, _ = cmd.Flags().GetBool("verbose")

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
			if verbose {
				cmd.Printf("detected format: %s\n", format.FormatName())
			}
		} else if format, err = geodbtools.LookupFormat(formatName); err != nil {
			return
		}

		var reader geodbtools.Reader
		if reader, _, err = format.NewReaderAt(source); err != nil {
			return
		}

		var record geodbtools.Record
		if record, err = reader.LookupIP(ip); err != nil {
			return
		}

		printRecord(cmd, record, verbose)
		return
	},
}

func printRecord(cmd *cobra.Command, rec geodbtools.Record, verbose bool) {

	switch t := rec.(type) {
	case geodbtools.CountryRecord:
		cmd.Printf("country          : %s\n", t.GetCountryCode())
	}

	if verbose {
		if rec.GetNetwork() != nil {
			cmd.Printf("matching network : %s\n", rec.GetNetwork())
		}
	}
}

func init() {
	cmdLookup.Flags().BoolP("verbose", "v", false, "enables verbose output")
	cmdLookup.Flags().StringP("db", "d", "", "database file")
	cmdLookup.Flags().StringP("format", "f", "auto", fmt.Sprintf("database format (auto|%s)", strings.Join(geodbtools.FormatNames(), "|")))

	cmdRoot.AddCommand(cmdLookup)
}
