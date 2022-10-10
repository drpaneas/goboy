package io

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"math"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// PixelScale is the multiplier on the pixels on display
var PixelScale float64 = 3

// PixelsIOBinding binds screen output and input using the pixels library.
type PixelsIOBinding struct {
	window  *pixelgl.Window
	picture *pixel.PictureData
}

// NewPixelsIOBinding returns a new Pixelsgl IOBinding
func NewPixelsIOBinding(enableVSync bool, gameboy *gb.Gameboy) *PixelsIOBinding {
	windowConfig := pixelgl.WindowConfig{
		//Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			160, 144,
		),
		VSync:       enableVSync,
		Resizable:   false,
		Undecorated: true,
	}

	window, err := pixelgl.NewWindow(windowConfig)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	// Hack so that pixelgl renders on Darwin
	window.SetPos(window.GetPos().Add(pixel.V(0, 1)))

	picture := &pixel.PictureData{
		Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
		Stride: gb.ScreenWidth,
		Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
	}

	monitor := PixelsIOBinding{
		window:  window,
		picture: picture,
	}

	monitor.updateCamera()

	return &monitor
}

var SC = false

func Screenshot(win *pixelgl.Window) {
	fmt.Println("taking screenshot...")

	f, err := os.Create("screenshot.png")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	img := pixel.PictureDataFromPicture(win)
	err = png.Encode(f, img.Image())
	if err != nil {
		return
	}

	fmt.Println("done")
}

// updateCamera updates the window camera to center the output.
func (mon *PixelsIOBinding) updateCamera() {
	xScale := mon.window.Bounds().W() / 160
	yScale := mon.window.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := mon.window.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	mon.window.SetMatrix(cam)
}

// IsRunning returns if the game should still be running. When
// the window is closed this will be false so the game stops.
func (mon *PixelsIOBinding) IsRunning() bool {
	return !mon.window.Closed()
}

// Render renders the pixels on the screen.
func (mon *PixelsIOBinding) Render(screen *[160][144][3]uint8) {
	for y := 0; y < gb.ScreenHeight; y++ {
		for x := 0; x < gb.ScreenWidth; x++ {
			col := screen[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(gb.ScreenHeight-1-y)*gb.ScreenWidth+x] = rgb
		}
	}

	if mon.window.JustPressed(pixelgl.KeyT) {
		// ----
		// Screenshot(window)
		//if SC {
		eikona := mon.picture.Image()
		if eikona.ColorModel() == color.RGBAModel {
			//32-bit RGBA color, each R,G,B, A component requires 8-bits
			fmt.Println("RGBA")
		} else if eikona.ColorModel() == color.GrayModel {
			//8-bit grayscale
			fmt.Println("Gray")
		} else if eikona.ColorModel() == color.NRGBAModel {
			//32-bit non-alpha-premultiplied RGB color, each R,G,B component requires 8-bits
			fmt.Println("NRGBA")
		} else if eikona.ColorModel() == color.NYCbCrAModel {
			//32-bit non-alpha-premultiplied YCbCr color, each Y,Cb,Cr component requires 8-bits
			fmt.Println("NYCbCrA")
		} else if eikona.ColorModel() == color.YCbCrModel {
			//24-bit YCbCr color, each Y,Cb,Cr component requires 8-bits
			fmt.Println("YCbCr")
		} else {
			fmt.Println("Unknown")
		}
		f, err := os.Create("screenshot.png")
		if err != nil {
			panic(err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
		// encode as .png to the file
		err = png.Encode(f, eikona)
	}
	// ----

	r, g, b := gb.GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.window.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight))
	spr.Draw(mon.window, pixel.IM)

	mon.updateCamera()
	// fmt.Println(mon.window.Bounds().Size())
	mon.window.Update()
}

// SetTitle sets the title of the game window.
func (mon *PixelsIOBinding) SetTitle(title string) {
	mon.window.SetTitle(title)
}

// Toggle the fullscreen window on the main monitor.
func (mon *PixelsIOBinding) toggleFullscreen() {
	mon.window.SetBounds(pixel.R(0, 0, 160, 144))
}

var keyMap = map[pixelgl.Button]gb.Button{
	pixelgl.KeyZ:         gb.ButtonA,
	pixelgl.KeyX:         gb.ButtonB,
	pixelgl.KeyBackspace: gb.ButtonSelect,
	pixelgl.KeyEnter:     gb.ButtonStart,
	pixelgl.KeyRight:     gb.ButtonRight,
	pixelgl.KeyLeft:      gb.ButtonLeft,
	pixelgl.KeyUp:        gb.ButtonUp,
	pixelgl.KeyDown:      gb.ButtonDown,

	pixelgl.KeyEscape: gb.ButtonPause,
	pixelgl.KeyEqual:  gb.ButtonChangePallete,
	pixelgl.KeyQ:      gb.ButtonToggleBackground,
	pixelgl.KeyW:      gb.ButtonToggleSprites,
	pixelgl.KeyE:      gb.ButttonToggleOutputOpCode,
	pixelgl.KeyD:      gb.ButtonPrintBGMap,
	pixelgl.Key7:      gb.ButtonToggleSoundChannel1,
	pixelgl.Key8:      gb.ButtonToggleSoundChannel2,
	pixelgl.Key9:      gb.ButtonToggleSoundChannel3,
	pixelgl.Key0:      gb.ButtonToggleSoundChannel4,
}

// ProcessInput checks the input and process it.
func (mon *PixelsIOBinding) ButtonInput() gb.ButtonInput {

	if mon.window.JustPressed(pixelgl.KeyF) {
		mon.toggleFullscreen()
	}

	var buttonInput gb.ButtonInput

	for handledKey, button := range keyMap {
		if mon.window.JustPressed(handledKey) {
			buttonInput.Pressed = append(buttonInput.Pressed, button)
		}
		if mon.window.JustReleased(handledKey) {
			buttonInput.Released = append(buttonInput.Released, button)
		}
	}

	return buttonInput
}
