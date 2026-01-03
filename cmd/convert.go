package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chardidathing/signal-tgstickers/internal/converter"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert [input-path]",
	Short: "Convert stickers from a directory or single file",
	Long: `Convert Telegram stickers to Signal-compatible PNG format.

Supports:
  - .webp (static stickers)
  - .tgs  (animated Lottie stickers - extracts first frame)
  - .webm (video stickers - extracts first frame or converts to APNG)

Examples:
  signal-tgstickers convert ./telegram-pack/
  signal-tgstickers convert sticker.tgs -o ./converted/
  signal-tgstickers convert ./pack/ --animated`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

var animatedOutput bool
var format string

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().BoolVarP(&animatedOutput, "animated", "a", false, "Preserve animation (outputs APNG for animated stickers)")
	convertCmd.Flags().StringVarP(&format, "format", "f", "png", "Output format: png, webp, apng")
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputDir, _ := cmd.Flags().GetString("output")
	size, _ := cmd.Flags().GetInt("size")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if err := converter.CheckFFmpeg(); err != nil {
		return fmt.Errorf("ffmpeg not found: %w\nPlease install ffmpeg: https://ffmpeg.org/download.html", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	opts := converter.Options{
		OutputDir: outputDir,
		MaxSize:   size,
		Format:    converter.OutputFormat(format),
		Animated:  animatedOutput,
		Verbose:   verbose,
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to access input path: %w", err)
	}

	if info.IsDir() {
		return convertDirectory(inputPath, opts)
	}
	return convertFile(inputPath, opts)
}

func convertDirectory(dirPath string, opts converter.Options) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var converted, failed int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if !converter.IsSupportedFormat(ext) {
			continue
		}

		inputFile := filepath.Join(dirPath, entry.Name())
		if err := convertFile(inputFile, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to convert %s: %v\n", entry.Name(), err)
			failed++
			continue
		}
		converted++
	}

	fmt.Printf("\nConversion complete: %d successful, %d failed\n", converted, failed)
	return nil
}

func convertFile(inputPath string, opts converter.Options) error {
	c := converter.New(opts)

	outputPath, err := c.Convert(inputPath)
	if err != nil {
		return err
	}

	if opts.Verbose {
		fmt.Printf("Converted: %s -> %s\n", inputPath, outputPath)
	}
	return nil
}
