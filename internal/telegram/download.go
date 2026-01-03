// Package telegram provides a client for downloading sticker packs via the Bot API.
package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type BotAPI struct {
	Token   string
	BaseURL string
}

type StickerSet struct {
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	StickerType string    `json:"sticker_type"`
	IsAnimated  bool      `json:"is_animated"`
	IsVideo     bool      `json:"is_video"`
	Stickers    []Sticker `json:"stickers"`
}

type Sticker struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Type         string `json:"type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	IsAnimated   bool   `json:"is_animated"`
	IsVideo      bool   `json:"is_video"`
	Emoji        string `json:"emoji"`
	SetName      string `json:"set_name"`
}

type File struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
	FileSize int    `json:"file_size"`
}

type APIResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"description,omitempty"`
}

// NewBotAPI creates a client for the given bot token.
func NewBotAPI(token string) *BotAPI {
	return &BotAPI{
		Token:   token,
		BaseURL: "https://api.telegram.org",
	}
}

// GetStickerSet fetches metadata for a sticker set by name.
func (b *BotAPI) GetStickerSet(name string) (*StickerSet, error) {
	url := fmt.Sprintf("%s/bot%s/getStickerSet?name=%s", b.BaseURL, b.Token, name)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get sticker set: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	var stickerSet StickerSet
	if err := json.Unmarshal(apiResp.Result, &stickerSet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sticker set: %w", err)
	}

	return &stickerSet, nil
}

// GetFile gets the file path needed to download a file by its ID.
func (b *BotAPI) GetFile(fileID string) (*File, error) {
	url := fmt.Sprintf("%s/bot%s/getFile?file_id=%s", b.BaseURL, b.Token, fileID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	var file File
	if err := json.Unmarshal(apiResp.Result, &file); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &file, nil
}

// DownloadFile saves a file to the given path.
func (b *BotAPI) DownloadFile(file *File, destPath string) error {
	url := fmt.Sprintf("%s/file/bot%s/%s", b.BaseURL, b.Token, file.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadStickerSet downloads all stickers from a set to the given directory.
func (b *BotAPI) DownloadStickerSet(setName, outputDir string) error {
	stickerSet, err := b.GetStickerSet(setName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	ext := ".webp"
	if stickerSet.IsAnimated {
		ext = ".tgs"
	} else if stickerSet.IsVideo {
		ext = ".webm"
	}

	fmt.Printf("Downloading sticker set: %s (%s)\n", stickerSet.Title, setName)
	fmt.Printf("Stickers: %d, Type: %s\n", len(stickerSet.Stickers), getTypeString(stickerSet))

	for i, sticker := range stickerSet.Stickers {
		file, err := b.GetFile(sticker.FileID)
		if err != nil {
			fmt.Printf("Warning: failed to get file for sticker %d: %v\n", i+1, err)
			continue
		}

		filename := fmt.Sprintf("%s_%02d%s", setName, i+1, ext)
		destPath := filepath.Join(outputDir, filename)

		if err := b.DownloadFile(file, destPath); err != nil {
			fmt.Printf("Warning: failed to download sticker %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("Downloaded: %s (emoji: %s)\n", filename, sticker.Emoji)
	}

	return nil
}

func getTypeString(set *StickerSet) string {
	if set.IsAnimated {
		return "animated (TGS)"
	}
	if set.IsVideo {
		return "video (WebM)"
	}
	return "static (WebP)"
}
