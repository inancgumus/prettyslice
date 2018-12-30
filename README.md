[![Go Report Card](https://goreportcard.com/badge/github.com/inancgumus/prettyslice)](https://goreportcard.com/report/github.com/inancgumus/prettyslice) [![Go Doc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](https://godoc.org/github.com/inancgumus/prettyslice)

# Pretty Slice Printer
It pretty prints **any type of** slices to any [io.Writer](https://golang.org/pkg/io/#Writer) with adjustable **coloring** features.

## Example

```go
package main

import s "github.com/inancgumus/prettyslice"

func main() {
	nums := []int{1, 3, 5, 2, 4, 8}
	odds := nums[:3]
	evens := nums[3:]

	nums[1], nums[3] = 9, 6
	s.Show("nums", nums)
	s.Show("odds : nums[:3]", odds)
	s.Show("evens: nums[3:]", evens)
}
```

### Output:
![](https://github.com/inancgumus/prettyslice/raw/master/slices.png)

## Example #2 — Render Colorless

```go
package main

import s "github.com/inancgumus/prettyslice"

func main() {
	// Render colorless output to a file
	f, _ := os.Create("out.txt")
	defer f.Close()

	nums := []int{1, 3, 5, 2, 4, 8}

	s.Writer = f
	s.Colors(false)
	s.Show("nums", nums)
}
```

---

## Printing Options

* **Writer:** Control where to draw the output. _Default: os.Stdout._
* **PrintBacking:** Whether to print the backing array. _Default: false._
* **PrettyByteRune:** Prints the bytes and runes as characters instead of numbers. _Default: true._
* **MaxPerLine:** Maximum number of slice items on a line. _Default: 5._
* **MaxElements:** Limits the number of elements printed. 0 means printing all elements. _Default: 0._
* **Width:** Number of space characters (_padding_) between the header message and the slice details like len, cap and ptr. _Default: 45._
* **NormalizePointers:** Prints the addresses of the slice elements as if they're contiguous. It basically normalizes by the element type size. See the source code for more information. _Default: false._
* **PrintHex:** Prints the pointers as hexadecimals. _Default: false._
* **PrintElementAddr:** Prints the element addresses. _Default: false._

## Coloring Options

* **ColorHeader:** Sets the color for the header. _Default: color.New(color.BgHiBlack, color.FgMagenta, color.Bold)._
* **ColorSlice:** Sets the color for the slice elements. _Default: color.New(color.FgCyan)._
* **ColorBacker:** Sets the color for the backing array elements. _Default: color.New(color.FgHiBlack)._
* **ColorIndex:** Sets the color for the index numbers. _Default: ColorBacker._
* **ColorAddr:** Sets the color for the element addresses. _Default: ColorBacker._

Have fun!
