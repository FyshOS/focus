package main

import (
	"math"

	"fyne.io/fyne/v2"
)

const (
	pixCount = 25
	pixMin   = 5
)

type highlightLayout struct{}

func (h highlightLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSquareSize(pixCount * pixMin)
}

func (h highlightLayout) Layout(o []fyne.CanvasObject, s fyne.Size) {
	edge := float32(math.Min(float64(s.Width), float64(s.Height)) / pixCount)
	o[0].Resize(fyne.NewSquareSize(edge + 1))
	o[0].Move(fyne.NewPos((s.Width-edge-1)/2, (s.Height-edge-1)/2))
}
