package slices

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

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
	ColorIndex = color.New(color.FgHiBlack)

	// MaxPerLine is the max allowed slice items on a line
	MaxPerLine = 0

	// PrettyByteRune prints byte and rune elements as chars
	PrettyByteRune = true

	// Writer controls where to draw the slices
	Writer io.Writer = os.Stdout
)

// drawing pretty draws a slice
type drawing struct {
	slice, backer reflect.Value

	buf *strings.Builder

	// draw multiple items or just one?
	multiple bool
}

// Show pretty prints slices
func Show(msg string, slices ...interface{}) {
	buf := new(strings.Builder)

	for i, slice := range slices {
		d := create(slice, buf)

		// only draw the message for the first item (grouping)
		if i > 0 {
			msg = ""
		}
		d.header(msg)

		if s := d.slice; s.IsNil() {
			d.push(" <nil slice>\n")
			continue
		} else if s.Len() == 0 {
			d.push(" <empty slice>\n")
			continue
		}

		// draw the slice elements
		d.wrap("╔", "╗")
		d.middle()
		d.wrap("╚", "╝")
		d.indexes()
	}

	// WriteString already checks for WriteString method
	io.WriteString(Writer, buf.String())
}

// Colors is used to enable/disable the color data from the output
func Colors(enabled bool) {
	colors := []*color.Color{
		ColorHeader, ColorSlice, ColorBacker,
	}

	for _, color := range colors {
		if enabled {
			color.EnableColor()
		} else {
			color.DisableColor()
		}
	}
}

// create initializes a new drawing struct.
func create(slice interface{}, buf *strings.Builder) drawing {
	s := reflect.ValueOf(slice)

	multiple := true
	if s.Kind() != reflect.Slice {
		s = makeSlice(s)

		// don't draw slice details for one item
		multiple = false
	}

	return drawing{
		slice: s,
		// this contains the backing array's data, after the slice's pointer.
		backer:   s.Slice(0, s.Cap()),
		multiple: multiple,
		buf:      buf,
	}
}

// header draws the header information about the slice with a message
func (d drawing) header(msg string) {
	var info string
	if d.multiple {
		info = fmt.Sprintf(
			"(len:%d cap:%d ptr:%d)",
			d.slice.Len(), d.slice.Cap(), d.pointer(),
		)
	}

	d.push(ColorHeader.Sprintf("%s %s", msg, info))
	d.push("\n")
}

// indexes draws the index numbers on top of the slice elements
func (d drawing) indexes() {
	for i, v := range over(d.backer) {
		if enough(i) {
			break
		}

		m := 4 + len(v)
		s := strings.Repeat(" ", m/2)
		if len(v) == 0 {
			s = " "
		}
		d.push(ColorIndex.Sprintf("%s%-*d", s, m-len(s), i))
	}
	d.push("\n")
}

// wrap draws the header and the footer depending on the left and right values
func (d drawing) wrap(left, right string) {
	for i, v := range over(d.backer) {
		if enough(i) {
			break
		}

		c, l, r, m := ColorSlice, left, right, "═"
		if d.backing(i) {
			c, l, r, m = ColorBacker, "+", "+", "-"
		}

		// draw the horizontal line
		// +2 is for the left and right vertical bars
		w := strings.Repeat(m, slen(v)+2)

		d.push(c.Sprintf("%s%s%s", l, w, r))
	}
	d.push("\n")
}

// middle draws the item's value wrapped between pipes
func (d drawing) middle() {
	for i, v := range over(d.backer) {
		if enough(i) {
			d.push(ColorBacker.Sprintf(" ..."))
			break
		}

		p, c := "║", ColorSlice
		if d.backing(i) {
			p, c = "|", ColorBacker
		}

		// Left Vertical : %-2[3]s
		// Item Value    : %-[1]*v
		//   (its width is dynamically adjusted: slen(v))
		// Right Vertical: %2[3]s
		d.push(
			c.Sprintf("%-2[3]s%-[1]*v%2[3]s",
				slen(v), v, p),
		)
	}
	d.push("\n")
}

// pointer simplifies the pointer data for easy viewing
func (d drawing) pointer() int64 {
	var s int64 = 1
	if d.slice.Len() > 0 {
		s = int64(d.slice.Index(0).Type().Size()) // normalize to the size
	}

	return (int64(d.slice.Pointer()) / s) % 10000 // get rid of the leading digits
}

// backing is true if the index belongs to the backing array
func (d drawing) backing(index int) bool {
	return index >= d.slice.Len()
}

// push appends a new string into the drawing's buffer
func (d drawing) push(s string) {
	d.buf.WriteString(s)
}

// slen gets the length of a utf-8 string.
// this is a func because it doesn't use the struct's data. it's stateless.
func slen(s string) int {
	return utf8.RuneCountInString(s)
}

// over range overs a reflect.Value as []string
func over(slice reflect.Value) []string {
	values := make([]string, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		s := fmt.Sprintf("%v", v)

		if PrettyByteRune {
			switch v.Interface().(type) {
			case byte:
				s = string(v.Uint())
			case rune:
				s = string(v.Int())
			}
		}
		values[i] = s
	}
	return values
}

// enough is true if the current is > MaxPerLine
func enough(index int) bool {
	return MaxPerLine > 0 && index >= MaxPerLine
}

func makeSlice(v reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, 1)
	slice = reflect.Append(slice, v)
	return slice
}
