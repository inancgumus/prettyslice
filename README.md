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

	s.Colors(false)
	s.Writer = f
	s.Show("nums", nums)
}
````

Have fun!
