package main

import (
	"errors"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalBannerGIFComposesFrames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gx-banner.gif")
	if _, err := readBannerGIF(t.Context(), path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing override: %v", err)
	}
	palette := color.Palette{color.Black, color.White}
	full := image.Rect(0, 0, 12, 8)
	first := image.NewPaletted(full, palette)
	second := image.NewPaletted(image.Rect(0, 0, 6, 8), palette)
	for i := range second.Pix {
		second.Pix[i] = 1
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	err = gif.EncodeAll(file, &gif.GIF{
		Image: []*image.Paletted{first, second}, Delay: []int{0, 8},
		Disposal: []byte{gif.DisposalNone, gif.DisposalNone},
		Config:   image.Config{Width: full.Dx(), Height: full.Dy(), ColorModel: palette},
	})
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	frames, err := readBannerGIF(t.Context(), path)
	if err != nil || len(frames) != 2 {
		t.Fatalf("decoded %d frames: %v", len(frames), err)
	}
	if frames[0].delay != 20*time.Millisecond || frames[1].delay != 80*time.Millisecond {
		t.Fatalf("unexpected frame delays: %s, %s", frames[0].delay, frames[1].delay)
	}
	if first := frames[0].image.RGBAAt(40, 35); first.R != 0 {
		t.Fatalf("first frame was overwritten: %v", first)
	}
	if left, right := frames[1].image.RGBAAt(40, 35), frames[1].image.RGBAAt(350, 35); left.R != 255 || right.R != 0 {
		t.Fatalf("partial frame was not composed: left %v, right %v", left, right)
	}
}
