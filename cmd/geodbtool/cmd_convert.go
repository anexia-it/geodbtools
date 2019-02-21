package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/anexia-it/geodbtools"
	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

var cmdConvert = &cobra.Command{
	Use:   "convert <database> <target>",
	Short: `Convert a GeoIP database from one format to another`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var inputFormatName, outputFormatName string
		var ipVersionInt8 int8
		var verify, force bool

		inputFormatName, _ = cmd.Flags().GetString("in-format")
		outputFormatName, _ = cmd.Flags().GetString("out-format")
		ipVersionInt8, _ = cmd.Flags().GetInt8("ip-version")
		verify, _ = cmd.Flags().GetBool("verify")
		force, _ = cmd.Flags().GetBool("force")

		inputPath := args[0]
		outputPath := args[1]

		if outputFileInfo, statErr := os.Stat(outputPath); statErr == nil && !force {
			err = errors.New("target file exists")
			return
		} else if statErr == nil && outputFileInfo.IsDir() {
			err = errors.New("target is a directory")
			return
		}

		ipVersion := geodbtools.IPVersion(ipVersionInt8)

		if ipVersion != geodbtools.IPVersion4 && ipVersion != geodbtools.IPVersion6 {
			err = geodbtools.ErrUnsupportedIPVersion
			return
		}

		var outputFormat geodbtools.Format
		if outputFormat, err = geodbtools.LookupFormat(outputFormatName); err != nil {
			return
		}

		var inputReaderSource geodbtools.ReaderSource
		if inputReaderSource, err = geodbtools.NewFileReaderSource(inputPath); err != nil {
			return
		}
		defer inputReaderSource.Close()

		var inputFormat geodbtools.Format
		if inputFormatName == "auto" {
			if inputFormat, err = geodbtools.DetectFormat(inputReaderSource); err != nil {
				return
			}
			cmd.Printf("detected input format: %s\n", inputFormat.FormatName())
		} else if inputFormat, err = geodbtools.LookupFormat(inputFormatName); err != nil {
			return
		}

		var inputReader geodbtools.Reader
		var meta geodbtools.Metadata
		if inputReader, meta, err = inputFormat.NewReaderAt(inputReaderSource); err != nil {
			return
		}

		var recordTree *geodbtools.RecordTree
		cmd.Println("starting generation of record tree...")
		treeStartAt := time.Now()
		if recordTree, err = inputReader.RecordTree(ipVersion); err != nil {
			return
		}
		cmd.Printf("tree generated after %s\n", time.Since(treeStartAt))

		outputBuffer := bytes.NewBufferString("")
		var outputWriter geodbtools.Writer

		if outputWriter, err = outputFormat.NewWriter(outputBuffer, meta.Type, ipVersion); err != nil {
			return
		}

		cmd.Printf("starting conversion from %s format to %s format...\n", inputFormat.FormatName(), outputFormat.FormatName())
		convertStartAt := time.Now()
		if err = outputWriter.WriteDatabase(meta, recordTree); err != nil {
			return
		}
		cmd.Printf("conversion finished after %s\n", time.Since(convertStartAt))

		if verify {

			var verifyReader geodbtools.Reader

			if verifyReader, _, err = outputFormat.NewReaderAt(geodbtools.NewReaderSourceWrapper(bytes.NewReader(outputBuffer.Bytes()), int64(outputBuffer.Len()))); err != nil {
				return
			}

			cmd.Println("starting verification...")

			var progress *pb.ProgressBar

			progressReports := make(chan *geodbtools.VerificationProgress, 8)
			defer func() {
				if progressReports != nil {
					close(progressReports)
				}

				if progress != nil {
					progress.Finish()
				}
			}()

			progressDoneCtx, progressDone := context.WithCancel(context.Background())
			go func() {
				defer progressDone()

				for report := range progressReports {
					if progress == nil {
						progress = pb.StartNew(report.TotalRecords)
						progress.SetWriter(cmd.OutOrStderr())
					}

					progress.SetCurrent(int64(report.CheckedRecords))
				}
				progress.Finish()
				progress = nil
			}()

			verifyStartAt := time.Now()
			err = geodbtools.Verify(verifyReader, recordTree, progressReports)
			close(progressReports)
			<-progressDoneCtx.Done()
			progressReports = nil
			if err != nil {
				verificationErrors := multierr.Errors(err)
				for i, verificationErr := range verificationErrors {
					cmd.Printf("error #%d: %s\n", i+1, verificationErr.Error())
				}
				err = fmt.Errorf("verification failed with %d errors", len(verificationErrors))
				return
			}
			cmd.Printf("verification finished after %s\n", time.Since(verifyStartAt))
		}

		var outputFile *os.File
		if outputFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC|os.O_EXCL|os.O_CREATE, 0600); err != nil {
			return
		}

		cmd.Println("starting write of output file...")
		writeStartAt := time.Now()
		if _, err = io.Copy(outputFile, outputBuffer); err != nil {
			return
		}
		cmd.Printf("write finished after %s\n", time.Since(writeStartAt))

		return
	},
}

func init() {
	cmdConvert.Flags().StringP("in-format", "I", "", fmt.Sprintf("input format (auto|%s)", strings.Join(geodbtools.FormatNames(), "|")))
	cmdConvert.Flags().StringP("out-format", "O", "", fmt.Sprintf("output format (%s)", strings.Join(geodbtools.FormatNames(), "|")))
	cmdConvert.Flags().Int8P("ip-version", "i", 4, "IP version (4|6)")
	cmdConvert.Flags().BoolP("verify", "V", false, "enables verification of the conversion by checking all records")
	cmdConvert.Flags().BoolP("force", "f", false, "overwrites existing output files")
	cmdRoot.AddCommand(cmdConvert)
}
