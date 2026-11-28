package main

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	xdraw "golang.org/x/image/draw"
)

const maxBannerBytes = 16 << 20

var bannerPixels = image.Rect(0, 0, 432, 243)

type bannerFrame struct {
	image *image.RGBA
	delay time.Duration
}

func newBanner(ctx context.Context, root string) fyne.CanvasObject {
	builtIn := drawBanner()
	view := canvas.NewImageFromImage(image.Transparent)
	view.FillMode = canvas.ImageFillContain
	frame := canvas.NewRectangle(color.Transparent)
	frame.StrokeColor = border
	frame.StrokeWidth = 1
	animation := container.NewStack(view, frame)
	animation.Hide()
	go func() {
		path := filepath.Join(root, ".local", "gx-banner.gif")
		frames, err := readBannerGIF(ctx, path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) && ctx.Err() == nil {
				log.Printf("GX banner: %v", err)
			}
			return
		}
		if len(frames) == 0 || ctx.Err() != nil {
			return
		}
		fyne.Do(func() {
			if ctx.Err() == nil {
				view.Image = frames[0].image
				view.Refresh()
				builtIn.Hide()
				animation.Show()
			}
		})
		if len(frames) > 1 {
			playBanner(ctx, view, frames)
		}
	}()
	return container.NewGridWrap(bannerSize, container.NewStack(builtIn, animation))
}

func readBannerGIF(ctx context.Context, path string) ([]bannerFrame, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxBannerBytes {
		return nil, errors.New("banner must be a regular GIF of at most 16 MiB")
	}
	config, err := gif.DecodeConfig(file)
	if err != nil {
		return nil, err
	}
	if config.Width > 1920 || config.Height > 1080 {
		return nil, errors.New("banner dimensions must fit within 1920 by 1080")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	source, err := gif.DecodeAll(io.LimitReader(file, maxBannerBytes))
	if err != nil {
		return nil, err
	}
	layer := image.NewRGBA(image.Rect(0, 0, source.Config.Width, source.Config.Height))
	var background image.Image = image.Transparent
	if palette, ok := source.Config.ColorModel.(color.Palette); ok && int(source.BackgroundIndex) < len(palette) {
		background = image.NewUniform(palette[source.BackgroundIndex])
	}
	frames := make([]bannerFrame, 0, len(source.Image))
	for i, frame := range source.Image {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		area := frame.Bounds()
		var previous *image.RGBA
		if source.Disposal[i] == gif.DisposalPrevious {
			previous = image.NewRGBA(area)
			draw.Draw(previous, area, layer, area.Min, draw.Src)
		}
		draw.Draw(layer, area, frame, area.Min, draw.Over)
		display := image.NewRGBA(bannerPixels)
		xdraw.ApproxBiLinear.Scale(display, display.Bounds(), layer, layer.Bounds(), xdraw.Src, nil)
		frames = append(frames, bannerFrame{display, time.Duration(max(source.Delay[i], 2)) * 10 * time.Millisecond})
		switch source.Disposal[i] {
		case gif.DisposalBackground:
			draw.Draw(layer, area, background, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			draw.Draw(layer, area, previous, area.Min, draw.Src)
		}
	}
	return frames, nil
}

// Prepare drawing-sized frames once, preserving partial-frame disposal.
// Only the installation that owns this canvas keeps its animation running.
func playBanner(ctx context.Context, view *canvas.Image, frames []bannerFrame) {
	timer := time.NewTimer(frames[0].delay)
	defer timer.Stop()
	for index := 1; ; index = (index + 1) % len(frames) {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		frame := frames[index]
		fyne.Do(func() {
			if ctx.Err() == nil {
				view.Image = frame.image
				view.Refresh()
			}
		})
		timer.Reset(frame.delay)
	}
}
