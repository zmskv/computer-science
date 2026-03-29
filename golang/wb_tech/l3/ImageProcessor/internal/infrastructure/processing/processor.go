package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	stdimagedraw "image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
	"go.uber.org/zap"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Processor struct {
	logger *zap.Logger
}

func NewProcessor(logger *zap.Logger) *Processor {
	return &Processor{logger: logger}
}

func (p *Processor) Process(
	ctx context.Context,
	source []byte,
	format string,
	options entity.ProcessingOptions,
) (dto.ProcessedImage, error) {
	if err := ctx.Err(); err != nil {
		return dto.ProcessedImage{}, err
	}

	img, detectedFormat, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		return dto.ProcessedImage{}, fmt.Errorf("decode image: %w", err)
	}

	if format == "" {
		format = normalizeFormat(detectedFormat)
	}

	resized := resizeToFit(img, options.MaxWidth, options.MaxHeight)
	watermarked := applyWatermark(resized, options.WatermarkText)
	thumbnail := resizeToFit(watermarked, options.ThumbnailSize, options.ThumbnailSize)

	processedBytes, err := encodeImage(watermarked, format)
	if err != nil {
		return dto.ProcessedImage{}, fmt.Errorf("encode processed image: %w", err)
	}

	thumbnailBytes, err := encodeImage(thumbnail, format)
	if err != nil {
		return dto.ProcessedImage{}, fmt.Errorf("encode thumbnail image: %w", err)
	}

	p.logger.Debug("image processed", zap.String("format", format))

	return dto.ProcessedImage{
		Processed: processedBytes,
		Thumbnail: thumbnailBytes,
		Format:    format,
	}, nil
}

func resizeToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	if maxWidth <= 0 || maxHeight <= 0 {
		return cloneToNRGBA(src)
	}

	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= maxWidth && srcHeight <= maxHeight {
		return cloneToNRGBA(src)
	}

	scale := math.Min(float64(maxWidth)/float64(srcWidth), float64(maxHeight)/float64(srcHeight))
	dstWidth := max(1, int(math.Round(float64(srcWidth)*scale)))
	dstHeight := max(1, int(math.Round(float64(srcHeight)*scale)))

	dst := image.NewNRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	return dst
}

func applyWatermark(src image.Image, text string) image.Image {
	dst := cloneToNRGBA(src)
	if text == "" {
		return dst
	}

	margin := 12
	face := basicfont.Face7x13
	textWidth := font.MeasureString(face, text).Round()
	textHeight := face.Metrics().Height.Ceil()

	x := max(margin, dst.Bounds().Dx()-textWidth-margin)
	y := max(textHeight+margin/2, dst.Bounds().Dy()-margin)

	bgRect := image.Rect(
		max(0, x-margin/2),
		max(0, y-textHeight-margin/3),
		min(dst.Bounds().Dx(), x+textWidth+margin/2),
		min(dst.Bounds().Dy(), y+margin/3),
	)

	stdimagedraw.Draw(dst, bgRect, &image.Uniform{C: color.NRGBA{R: 15, G: 23, B: 42, A: 120}}, image.Point{}, stdimagedraw.Over)

	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.NRGBA{R: 255, G: 248, B: 220, A: 230}),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(text)

	return dst
}

func cloneToNRGBA(src image.Image) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	stdimagedraw.Draw(dst, dst.Bounds(), src, bounds.Min, stdimagedraw.Src)
	return dst
}

func encodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer

	switch normalizeFormat(format) {
	case "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, err
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "gif":
		if err := gif.Encode(&buf, img, &gif.Options{NumColors: 256}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	return buf.Bytes(), nil
}

func normalizeFormat(format string) string {
	switch format {
	case "jpeg":
		return "jpg"
	default:
		return format
	}
}