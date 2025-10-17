package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("com.fyshos.focus")
	w := a.NewWindow("Focus")

	preview := &canvas.Image{}
	preview.ScaleMode = canvas.ImageScalePixels
	preview.FillMode = canvas.ImageFillContain
	highlight := canvas.NewRectangle(color.Transparent)
	highlight.StrokeColor = theme.ErrorColor()
	highlight.StrokeWidth = 1

	output := widget.NewLabel("#000000")
	choose := widget.NewSelect([]string{"Hex", "rgb"}, nil)
	choose.PlaceHolder = "Hex"
	choose.Selected = "Hex"
	copyAction := func() {
		c := w.Clipboard()
		if c == nil {
			return // can happen in some cases...
		}

		c.SetContent(output.Text)
	}
	copyToClip := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), copyAction)
	bar := container.NewBorder(nil, nil, choose, copyToClip, output)
	w.SetContent(container.NewBorder(nil, bar, nil, nil,
		container.NewStack(preview, container.New(highlightLayout{}, highlight))))

	hold := false
	w.Canvas().AddShortcut(&fyne.ShortcutCopy{}, func(_ fyne.Shortcut) { copyAction() })
	w.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyH, Modifier: fyne.KeyModifierShortcutDefault},
		func(_ fyne.Shortcut) {
			hold = !hold
		})

	go func() {
		halfSize := (pixCount - 1) / 2

		for pix := range pollPixels(w) {
			if hold {
				continue
			}
			preview.Image = pix
			fyne.Do(preview.Refresh)

			center := (pixCount*halfSize + halfSize) * 4
			r, g, b := pix.Pix[center], pix.Pix[center+1], pix.Pix[center+2]
			fyne.Do(func() {
				if choose.Selected == "rgb" {
					output.SetText(fmt.Sprintf("rgb(%d, %d, %d)", r, g, b))
				} else {
					output.SetText(fmt.Sprintf("#%02x%02x%02x", r, g, b))
				}
			})
		}
	}()
	w.ShowAndRun()
}

func pollPixels(w fyne.Window) <-chan *image.NRGBA {
	ch := make(chan *image.NRGBA)
	x11, err := xgbutil.NewConn()
	if err != nil {
		dialog.ShowError(err, w)
		return nil
	}
	halfSize := (pixCount - 1) / 2

	go func() {
		t := time.NewTicker(time.Second / 10)
		for range t.C {
			r, _ := xproto.QueryPointer(x11.Conn(), x11.RootWin()).Reply()
			g, _ := xproto.GetGeometry(x11.Conn(), xproto.Drawable(x11.RootWin())).Reply()

			x, y := r.RootX-int16(halfSize), r.RootY-int16(halfSize)
			if x < 0 {
				x = 0
			}
			if y < 0 {
				y = 0
			}
			if x >= int16(g.Width)-pixCount {
				x = int16(g.Width) - pixCount
			}
			if y >= int16(g.Height)-pixCount {
				y = int16(g.Height) - pixCount
			}

			pix, err := xproto.GetImage(x11.Conn(), xproto.ImageFormatZPixmap, xproto.Drawable(x11.RootWin()),
				x, y, pixCount, pixCount, math.MaxUint32).Reply()
			if err != nil {
				fyne.LogError("failed to read pixel data", err)
				continue
			}

			img := image.NewNRGBA(image.Rect(0, 0, pixCount, pixCount))
			// b, g, r, a
			for i := 0; i < 625; i++ {
				img.Pix[i*4] = pix.Data[i*4+2]
				img.Pix[i*4+1] = pix.Data[i*4+1]
				img.Pix[i*4+2] = pix.Data[i*4]
				img.Pix[i*4+3] = 0xff // no transparency on the overall screen
			}

			ch <- img
		}
	}()

	return ch
}
