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

		// draw the slice elements
		d.wrap("╔", "╗")
		d.push("\n")
		d.middle()
		d.wrap("╚", "╝")

		// dont put a new line after the last slice
		if i+1 < len(slices) {
			d.push("\n")
		}
	}

	render(Writer, buf)
}

// create initializes a new drawing struct.
func create(slice interface{}, buf *strings.Builder) *drawing {
	s := reflect.ValueOf(slice)

	multiple := true
	if s.Kind() != reflect.Slice {
		s = makeSlice(s)

		// don't draw slice details for one item
		multiple = false
	}

	return &drawing{
		slice: s,
		// this contains the backing array's data, after the slice's pointer.
		backer:   s.Slice(0, s.Cap()),
		multiple: multiple,
		buf:      buf,
	}
}

// header draws the header information about the slice with a message
func (d *drawing) header(msg string) {
	var info string
	if d.multiple {
		info = fmt.Sprintf(
			" (len:%-2d cap:%-2d ptr:%-4d)",
			d.slice.Len(), d.slice.Cap(), d.pointer(),
		)
	}

	d.push(ColorHeader.Sprintf("%-35s%26s \n", " "+msg, info))
}

// wrap draws the header and the footer depending on the left and right values
func (d *drawing) wrap(left, right string) {
	for i, v := range over(d.backer) {
		c, l, r, m := ColorSlice, left, right, "═"
		if d.backing(i) {
			c, l, r, m = ColorBacker, "+", "+", "-"
		}

		// draw the horizontal line
		// +2 is for the left and right vertical bars
		w := strings.Repeat(m, slen(v)+2)

		d.push(c.Sprintf("%s%s%s", l, w, r))
	}
}

// middle draws the item's value wrapped between pipes
func (d *drawing) middle() {
	for i, v := range over(d.backer) {
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
func (d *drawing) pointer() int64 {
	var s int64 = 1
	if d.slice.Len() > 0 {
		s = int64(d.slice.Index(0).Type().Size()) // normalize to the size
	}

	return (int64(d.slice.Pointer()) / s) % 10000 // get rid of the leading digits
}

// backing is true if the index belongs to the backing array
func (d *drawing) backing(index int) bool {
	return index+1 > d.slice.Len()
}

// push appends a new string into the drawing's buffer
func (d *drawing) push(s string) {
	d.buf.WriteString(s)
}

// render draws the drawings into the Writer
func render(w io.Writer, buf *strings.Builder) {
	// if the Writer supports WriteString method then use it
	if w, ok := Writer.(stringWriter); ok {
		w.WriteString(buf.String() + "\n")
		return
	}
	// or print it using Fprintln
	fmt.Fprintln(Writer, buf)
}

// just for checking whether the Writer implements the WriteString method
type stringWriter interface {
	WriteString(s string) (n int, err error)
}

// slen gets the length of a utf-8 string.
// this is a func because it doesn't use the struct's data. it's stateless.
func slen(s string) int {
	return utf8.RuneCountInString(s)
}

// over helps you to range over reflect.Values easily
// note: it converts the reflect.Values to []string
//       you can't get byte, and rune as numbers from this
func over(slice reflect.Value) []string {
	values := make([]string, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		var (
			v = slice.Index(i)
			s string
		)

		switch v.Interface().(type) {
		case byte:
			s = string(v.Uint())
		case rune:
			s = string(v.Int())
		default:
			s = fmt.Sprintf("%v", v)
		}
		values[i] = s
	}
	return values
}

func makeSlice(v reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, 1)
	slice = reflect.Append(slice, v)
	return slice
}
