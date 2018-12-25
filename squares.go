package prettyslice

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
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
			d.push("<nil slice>\n")
			continue
		} else if s.Len() == 0 {
			d.push("<empty slice>\n")
			// keep processing: slice can have elements in the backing array
		}

		// draw the slice elements
		l := d.backer.Len()
		if !PrintBacking {
			l = d.slice.Len()
		}

		step := MaxPerLine
		if step <= 0 {
			step = l
		}

		for f := 0; f < l; f += step {
			t := f + step

			d.wrap("╔", "╗", f, t)
			d.middle(f, t)
			d.wrap("╚", "╝", f, t)
			d.indexes(f, t)
		}
	}

	// WriteString already checks for WriteString method
	io.WriteString(Writer, buf.String())
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
			" (len:%-2d cap:%-2d ptr:%-4d)",
			d.slice.Len(), d.slice.Cap(), d.pointer(),
		)
	}

	d.push(ColorHeader.Sprintf("%*s%s", -Width, " "+msg, info))
	d.push("\n")
}

// indexes draws the index numbers on top of the slice elements
func (d drawing) indexes(from, to int) {
	for i, v := range over(d.backer, from, to) {
		if !PrintBacking && d.backing(from+i) {
			break
		}

		// current index
		ci := i + from

		// middle length of the index's width
		mli := len(strconv.Itoa(ci)) / 2

		// total width
		w := slen(v) + 4

		// left and right paddings
		lpw := w/2 - mli
		lp := strings.Repeat(" ", lpw)
		rp := w - len(lp)

		d.push(ColorIndex.Sprintf("%s%-*d", lp, rp, ci))
	}
	d.push("\n")
}

// wrap draws the header and the footer depending on the left and right values
func (d drawing) wrap(left, right string, from, to int) {
	for i, v := range over(d.backer, from, to) {
		c, l, r, m := ColorSlice, left, right, "═"

		if d.backing(from + i) {
			if !PrintBacking {
				break
			}
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
func (d drawing) middle(from, to int) {
	for i, v := range over(d.backer, from, to) {
		p, c := "║", ColorSlice
		if d.backing(from + i) {
			if !PrintBacking {
				break
			}
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

	if NormalizePointers && d.slice.Len() > 0 {
		s = int64(d.slice.Index(0).Type().Size())
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

// enough is true if the current is > MaxPerLine
func enough(index int) bool {
	return MaxPerLine > 0 && index%MaxPerLine == 0
}

// over range overs a reflect.Value as []string
func over(slice reflect.Value, from, to int) []string {
	values := make([]string, 0, to-from)

	if l := slice.Len(); to > l {
		to = l
	}

	for i := from; i < to; i++ {
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
		values = append(values, s)
	}
	return values
}

func makeSlice(v reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, 1)
	slice = reflect.Append(slice, v)
	return slice
}
