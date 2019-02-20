package main

import "github.com/spf13/cobra"

var cmdRoot = &cobra.Command{
	Use:   "geodbtool",
	Short: `GeoIP database swiss army knife`,
}
