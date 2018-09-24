package main

import (
	"os"

	s "github.com/inancgumus/prettyslice"
)

func main() {
	nums := []int{1, 3, 5, 2, 4, 8}
	odds := nums[:3]
	evens := nums[3:]

	nums[1], nums[3] = 9, 6

	// Render to stdout by default
	s.Show("nums", nums)
	s.Show("odds : nums[:3]", odds)
	s.Show("evens: nums[3:]", evens)

	// Render colorless output to a file
	f, _ := os.Create("out.txt")
	defer f.Close()

	s.Colors(false)
	s.Writer = f
	s.Show("nums", nums)
}
