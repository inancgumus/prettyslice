package prettyslice

import (
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	// ColorHeader sets the color for the header
	ColorHeader = color.New(
		color.BgHiBlack,
		color.FgMagenta,
		color.Bold)

	// ColorSlice sets the color for the slice's items
	ColorSlice = color.New(color.FgCyan)

	// ColorBacker sets the color for the backing array's items
	ColorBacker = color.New(color.FgHiBlack)

	// ColorIndex sets the color for the index numbers of the elements
	ColorIndex = ColorBacker

	// ColorAddr sets the color for the element addresses
	ColorAddr = ColorBacker

	// MaxPerLine is maximum number of slice items on a line
	MaxPerLine = 0

	// MaxElements limits the number of elements printed
	// 0 means print all the elements.
	MaxElements = 0

	// Width is the width of the header
	// It will separate the header message and the slice details with empty spaces
	Width = 0

	// PrettyByteRune prints byte and rune elements as chars
	PrettyByteRune = true

	// PrintBacking prints the backing array if it's true
	PrintBacking = false

	// PrintElementAddr prints the addresses of each element
	PrintElementAddr = false

	// PrintHex prints the pointers in hexadecimals
	//
	// When it's false, only the last 4 digits of the pointers will be printed as decimals.
	//
	// When it's true, all the digits of the pointers will be printed as hexadecimals.
	PrintHex = false

	// NormalizePointers prints pointers as if they're contiguous.
	//
	// Let's say you've []int64{1, 2}
	//
	//              Memory addresses
	// 1st element: 8000
	// 2nd element: 8008
	//
	// This option prints them like this instead:
	//
	//              Memory addresses
	// 1st element: 8000
	// 2nd element: 8001
	//
	// So, it basically normalizes by the element type size.
	NormalizePointers = false

	// Writer controls where to draw the slices
	Writer io.Writer = os.Stdout
)

// Colors is used to enable/disable the color data from the output
func Colors(enabled bool) {
	colors := []*color.Color{
		ColorHeader, ColorSlice, ColorBacker, ColorIndex,
	}

	for _, color := range colors {
		if enabled {
			color.EnableColor()
		} else {
			color.DisableColor()
		}
	}
}
