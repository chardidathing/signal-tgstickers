package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chardidathing/signal-tgstickers/internal/converter"
	"github.com/chardidathing/signal-tgstickers/internal/telegram"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack [sticker-set-name]",
	Short: "Download and convert a Telegram sticker pack in one step",
	Long: `Download a Telegram sticker pack and convert it to Signal format in one step.

This is the recommended way to use signal-tgstickers. It will:
  1. Download all stickers from the Telegram pack
  2. Convert them to Signal-compatible PNG format
  3. Clean up the raw downloaded files (optional)

Requires a Telegram Bot API token via --token or TELEGRAM_BOT_TOKEN env var.

Examples:
  signal-tgstickers pack CoolCats --token YOUR_BOT_TOKEN
  signal-tgstickers pack AnimatedDogs -o ./my-stickers/ --animated
  signal-tgstickers pack VideoMemes --keep-raw`,
	Args: cobra.ExactArgs(1),
	RunE: runPack,
}

var (
	packToken    string
	keepRaw      bool
	packAnimated bool
	packFormat   string
)

func init() {
	rootCmd.AddCommand(packCmd)
	packCmd.Flags().StringVarP(&packToken, "token", "t", "", "Telegram Bot API token (or set TELEGRAM_BOT_TOKEN)")
	packCmd.Flags().BoolVar(&keepRaw, "keep-raw", false, "Keep the raw downloaded files after conversion")
	packCmd.Flags().BoolVarP(&packAnimated, "animated", "a", false, "Preserve animation (outputs APNG for animated stickers)")
	packCmd.Flags().StringVarP(&packFormat, "format", "f", "png", "Output format: png, webp, apng")
}

func runPack(cmd *cobra.Command, args []string) error {
	stickerSetName := args[0]
	outputDir, _ := cmd.Flags().GetString("output")
	size, _ := cmd.Flags().GetInt("size")
	verbose, _ := cmd.Flags().GetBool("verbose")

	token := packToken
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("telegram bot token required: use --token flag or set TELEGRAM_BOT_TOKEN environment variable")
	}

	if err := converter.CheckFFmpeg(); err != nil {
		return fmt.Errorf("ffmpeg not found: %w\nPlease install ffmpeg: https://ffmpeg.org/download.html", err)
	}

	rawDir := filepath.Join(outputDir, ".raw")
	convertedDir := outputDir

	if err := os.MkdirAll(rawDir, 0755); err != nil {
		return fmt.Errorf("failed to create raw directory: %w", err)
	}
	if err := os.MkdirAll(convertedDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// download
	fmt.Println("Downloading sticker pack...")
	api := telegram.NewBotAPI(token)

	if err := api.DownloadStickerSet(stickerSetName, rawDir); err != nil {
		return fmt.Errorf("failed to download sticker set: %w", err)
	}

	// convert
	fmt.Println("\nConverting to Signal format...")
	opts := converter.Options{
		OutputDir: convertedDir,
		MaxSize:   size,
		Format:    converter.OutputFormat(packFormat),
		Animated:  packAnimated,
		Verbose:   verbose,
	}

	entries, err := os.ReadDir(rawDir)
	if err != nil {
		return fmt.Errorf("failed to read raw directory: %w", err)
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

		inputFile := filepath.Join(rawDir, entry.Name())
		c := converter.New(opts)

		outputPath, err := c.Convert(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Failed: %s (%v)\n", entry.Name(), err)
			failed++
			continue
		}

		if verbose {
			fmt.Printf("  %s -> %s\n", entry.Name(), filepath.Base(outputPath))
		}
		converted++
	}

	// cleanup unless --keep-raw
	if !keepRaw {
		if err := os.RemoveAll(rawDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up raw files: %v\n", err)
		}
	} else {
		fmt.Printf("\nRaw files kept in: %s\n", rawDir)
	}

	fmt.Printf("\nDone! Converted %d stickers", converted)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Printf("\nOutput: %s\n", convertedDir)

	return nil
}
