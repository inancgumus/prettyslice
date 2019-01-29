package prettyslice

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
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
		d.pushNewline()

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
			if enough(f) {
				d.push(ColorBacker.Sprintf("...%d more...", l-f))
				d.pushNewline()
				break
			}

			t := f + step

			d.wrap("╔", "╗", f, t)
			d.pushNewline()
			d.middle(f, t)
			d.pushNewline()
			d.wrap("╚", "╝", f, t)
			d.pushNewline()
			d.indexes(f, t)
			d.pushNewline()

			if PrintElementAddr {
				d.addresses(f, t)
				d.pushNewline()
			}
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
		f := " (len:%-2d cap:%-2d ptr:%-4d)"
		if PrintHex {
			f = " (len:%-2d cap:%-2d ptr:%-10x)"
		}

		info = fmt.Sprintf(
			f,
			d.slice.Len(), d.slice.Cap(), d.pointer(0),
		)
	}

	msg = " " + msg

	w, l := Width, len(msg)+len(info)
	w -= l
	if l > Width {
		w = 1
	}

	d.push(ColorHeader.Sprintf("%s%*s%s", msg, w, "", info))
}

// indexes draws the index numbers on top of the slice elements
func (d drawing) indexes(from, to int) {
	for i, v := range over(d.backer, from, to) {
		if !PrintBacking && d.backing(from+i) {
			break
		}

		// current index
		ci := i + from

		lp, rp := paddings(len(strconv.Itoa(ci)), slen(v))
		lps := strings.Repeat(" ", lp)

		d.push(ColorIndex.Sprintf("%s%-*d", lps, rp, ci))
	}
}

// addresses draw element addresses
func (d drawing) addresses(from, to int) {
	for i, v := range over(d.backer, from, to) {
		if !PrintBacking && d.backing(from+i) {
			break
		}

		// current index
		ci := i + from

		p := d.pointer(ci)

		lp, rp := paddings(len(strconv.FormatInt(p, 10)), slen(v))
		lps := strings.Repeat(" ", lp)

		d.push(ColorAddr.Sprintf("%s%-*d", lps, rp, p))
	}
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
}

// pointer simplifies the pointer data for easy viewing
func (d drawing) pointer(index int) int64 {
	var s int64 = 1

	if NormalizePointers && d.slice.Len() > 0 {
		s = int64(d.backer.Index(index).Type().Size())
	}

	p := int64(d.slice.Pointer())
	if index != 0 && d.slice.Len() > 0 {
		p = int64(d.backer.Index(index).Addr().Pointer())
	}

	trim := int64(10000) // get rid of the leading digits
	if PrintHex {
		// do not trim the digits: p % p + 1 = p
		trim = p + 1
	}

	return (p / s) % trim
}

// backing is true if the index belongs to the backing array
func (d drawing) backing(index int) bool {
	return index >= d.slice.Len()
}

// push appends a new string into the drawing's buffer
func (d drawing) push(s string) {
	d.buf.WriteString(s)
}

// pushNewline appends a newline into the drawing's buffer
func (d drawing) pushNewline() {
	d.push("\n")
}

// paddings finds out the left and right paddings from two values' lengths
func paddings(a, b int) (lp int, rp int) {
	// middle length
	mli := a / 2

	// total width
	w := b + 4

	// left and right paddings
	lp = w/2 - mli
	rp = w - lp

	if lp < 0 {
		lp = 0
	}
	if rp < 0 {
		rp = 0
	}

	return
}

// slen gets the length of a utf-8 string.
// this is a func because it doesn't use the struct's data. it's stateless.
func slen(s string) int {
	return utf8.RuneCountInString(s)
}

// enough is true if the current is > MaxElements
func enough(index int) bool {
	return MaxElements != 0 && index > MaxElements
}

// over range overs a reflect.Value as []string
// TODO (@inanc): Fix the unnecessary allocation
func over(slice reflect.Value, from, to int) []string {
	size := to - from
	if MaxElements != 0 && size > MaxElements {
		size = MaxElements
	}

	values := make([]string, 0, size)

	if l := slice.Len(); to > l {
		to = l
	}

	for i := from; i < to; i++ {
		if enough(i) {
			break
		}

		v := slice.Index(i)
		s := fmt.Sprintf("%v", v)

		// this will be overwritten if PrettyByteRune
		if PrintBytesHex {
			if b, ok := v.Interface().(byte); ok {
				s = fmt.Sprintf("%02x", b)
			}
		}

		if PrettyByteRune {
			var (
				r      rune
				isRune bool
			)

			switch v.Interface().(type) {
			case byte:
				r = rune(v.Uint())
				isRune = !PrintBytesHex && true
			case rune:
				r = rune(v.Int())
				isRune = true
			}

			if isRune {
				s = string(r)

				switch {
				case unicode.IsSpace(r), unicode.IsControl(r):
					s = ` `
				}
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
