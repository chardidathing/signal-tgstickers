package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "signal-tgstickers",
	Short: "Convert Telegram sticker packs to Signal-compatible formats",
	Long: `A CLI tool to convert Telegram sticker packs (.webp, .tgs, .webm) to formats supported by Signal messenger.

Quick start:
  signal-tgstickers pack <sticker-set-name> --token YOUR_BOT_TOKEN

Commands:
  pack      Download and convert in one step (recommended)
  download  Download a sticker pack from Telegram
  convert   Convert local sticker files

Requires ffmpeg to be installed on your system for conversion.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "./output", "Output directory for converted stickers")
	rootCmd.PersistentFlags().IntP("size", "s", 512, "Maximum dimension for output stickers (Signal max: 512)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}

