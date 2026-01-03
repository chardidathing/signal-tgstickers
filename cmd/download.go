package cmd

import (
	"fmt"
	"os"

	"github.com/chardidathing/signal-tgstickers/internal/telegram"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [sticker-set-name]",
	Short: "Download a Telegram sticker pack",
	Long: `Download a Telegram sticker pack by its name.

Requires a Telegram Bot API token, which can be provided via:
  - The --token flag
  - The TELEGRAM_BOT_TOKEN environment variable

To get a bot token, message @BotFather on Telegram.

Example:
  signal-tgstickers download AnimatedEmojis --token YOUR_BOT_TOKEN
  TELEGRAM_BOT_TOKEN=xxx signal-tgstickers download AnimatedEmojis`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

var botToken string

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringVarP(&botToken, "token", "t", "", "Telegram Bot API token (or set TELEGRAM_BOT_TOKEN)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	stickerSetName := args[0]
	outputDir, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")

	token := botToken
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("telegram bot token required: use --token flag or set TELEGRAM_BOT_TOKEN environment variable")
	}

	if verbose {
		fmt.Printf("Downloading sticker set: %s\n", stickerSetName)
		fmt.Printf("Output directory: %s\n", outputDir)
	}

	api := telegram.NewBotAPI(token)

	if err := api.DownloadStickerSet(stickerSetName, outputDir); err != nil {
		return fmt.Errorf("failed to download sticker set: %w", err)
	}

	fmt.Printf("\nStickers downloaded to: %s\n", outputDir)
	fmt.Println("Run 'signal-tgstickers convert' to convert them to Signal format.")

	return nil
}
