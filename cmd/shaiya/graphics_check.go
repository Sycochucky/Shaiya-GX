package main

import (
	"fmt"
	"log"

	img "github.com/Spelunkern/Shaiya-GX/internal/data/image"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

func checkGraphics(options render.InitOptions) error {
	win, err := platform.Open(platform.Config{Title: "Shaiya GX graphics check", Width: 128, Height: 128, Debug: options.Debug, Hidden: true})
	if err != nil {
		return err
	}
	defer win.Close()
	if err := render.InitWithOptions(options); err != nil {
		return err
	}
	scene, err := render.NewScene3D()
	if err != nil {
		return err
	}
	defer scene.Delete()
	skin, err := render.NewSkinner()
	if err != nil {
		return err
	}
	defer skin.Delete()
	sprites, err := render.NewSprites(1)
	if err != nil {
		return err
	}
	defer sprites.Delete()
	texture := render.Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: []byte{255, 0, 0, 255}}})
	defer texture.Delete()

	w, h := win.GetFramebufferSize()
	if w <= 0 || h <= 0 {
		return fmt.Errorf("graphics check: invalid framebuffer size %dx%d", w, h)
	}
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	sprites.Begin(w, h)
	sprites.DrawImage(texture, float32(w/2), float32(h/2))
	sprites.DrawImage(texture, float32(w/2+1), float32(h/2))
	sprites.End()
	var pixels [2][4]byte
	gl.ReadPixels(int32(w/2), int32(h-1-h/2), 2, 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0][0]))
	if code := gl.GetError(); code != gl.NO_ERROR {
		return fmt.Errorf("graphics check: OpenGL error %#x", code)
	}
	red := [4]byte{255, 0, 0, 255}
	if pixels != [2][4]byte{red, red} {
		return fmt.Errorf("graphics check: expected two opaque red pixels, got %v", pixels)
	}
	win.SwapBuffers()
	log.Printf("OpenGL startup check passed (framebuffer=%dx%d)", w, h)
	return nil
}
