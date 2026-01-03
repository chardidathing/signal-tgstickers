// Package converter handles conversion of Telegram sticker formats to Signal-compatible formats.
package converter

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OutputFormat string

const (
	FormatPNG  OutputFormat = "png"
	FormatWebP OutputFormat = "webp"
	FormatAPNG OutputFormat = "apng"
)

type Options struct {
	OutputDir string
	MaxSize   int
	Format    OutputFormat
	Animated  bool
	Verbose   bool
}

type Converter struct {
	opts Options
}

// New returns a converter with the given options.
func New(opts Options) *Converter {
	if opts.MaxSize <= 0 {
		opts.MaxSize = 512
	}
	if opts.Format == "" {
		opts.Format = FormatPNG
	}
	return &Converter{opts: opts}
}

// CheckFFmpeg returns an error if ffmpeg isn't in PATH.
func CheckFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	return err
}

// IsSupportedFormat returns true for .webp, .tgs, and .webm extensions.
func IsSupportedFormat(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".webp" || ext == ".tgs" || ext == ".webm"
}

// Convert processes a sticker file and returns the output path.
func (c *Converter) Convert(inputPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)

	var outputExt string
	switch c.opts.Format {
	case FormatAPNG:
		outputExt = ".png"
	case FormatWebP:
		outputExt = ".webp"
	default:
		outputExt = ".png"
	}

	outputPath := filepath.Join(c.opts.OutputDir, baseName+outputExt)

	switch ext {
	case ".webp":
		return outputPath, c.convertWebP(inputPath, outputPath)
	case ".tgs":
		return outputPath, c.convertTGS(inputPath, outputPath)
	case ".webm":
		return outputPath, c.convertWebM(inputPath, outputPath)
	default:
		return "", fmt.Errorf("unsupported format: %s", ext)
	}
}

func (c *Converter) convertWebP(input, output string) error {
	args := []string{
		"-i", input,
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
		"-y",
		output,
	}
	return c.runFFmpeg(args)
}

// convertTGS handles animated Lottie stickers (gzipped JSON).
func (c *Converter) convertTGS(input, output string) error {
	tempDir, err := os.MkdirTemp("", "tgs-convert-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// decompress the tgs file to get the lottie json
	lottieJSON := filepath.Join(tempDir, "animation.json")
	if err := c.decompressTGS(input, lottieJSON); err != nil {
		return fmt.Errorf("failed to decompress TGS: %w", err)
	}

	// try lottie2gif first, fall back to ffmpeg
	if _, err := exec.LookPath("lottie2gif"); err == nil {
		return c.convertTGSWithLottie2Gif(lottieJSON, output, tempDir)
	}
	return c.convertTGSWithFFmpeg(lottieJSON, output)
}

func (c *Converter) decompressTGS(input, output string) error {
	inFile, err := os.Open(input)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gzReader, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gzReader)
	return err
}

func (c *Converter) convertTGSWithLottie2Gif(lottieJSON, output, tempDir string) error {
	gifPath := filepath.Join(tempDir, "animation.gif")

	cmd := exec.Command("lottie2gif", lottieJSON, "-o", gifPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("lottie2gif failed: %w", err)
	}

	if c.opts.Animated {
		args := []string{
			"-i", gifPath,
			"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
			"-plays", "0",
			"-y",
			output,
		}
		return c.runFFmpeg(args)
	}

	// extract first frame only
	args := []string{
		"-i", gifPath,
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
		"-vframes", "1",
		"-y",
		output,
	}
	return c.runFFmpeg(args)
}

func (c *Converter) convertTGSWithFFmpeg(lottieJSON, output string) error {
	// this only works if ffmpeg was compiled with rlottie support
	if c.opts.Animated {
		args := []string{
			"-i", lottieJSON,
			"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
			"-plays", "0",
			"-y",
			output,
		}
		return c.runFFmpeg(args)
	}

	args := []string{
		"-i", lottieJSON,
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
		"-vframes", "1",
		"-y",
		output,
	}

	err := c.runFFmpeg(args)
	if err != nil {
		return fmt.Errorf("ffmpeg lottie conversion failed (may need rlottie support): %w\n"+
			"try installing lottie2gif: pip install lottie[gif]", err)
	}
	return nil
}

func (c *Converter) convertWebM(input, output string) error {
	if c.opts.Animated {
		args := []string{
			"-i", input,
			"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
			"-plays", "0",
			"-y",
			output,
		}
		return c.runFFmpeg(args)
	}

	// extract first frame only
	args := []string{
		"-i", input,
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", c.opts.MaxSize, c.opts.MaxSize),
		"-vframes", "1",
		"-y",
		output,
	}
	return c.runFFmpeg(args)
}

func (c *Converter) runFFmpeg(args []string) error {
	if c.opts.Verbose {
		fmt.Printf("Running: ffmpeg %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command("ffmpeg", args...)
	if c.opts.Verbose {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
