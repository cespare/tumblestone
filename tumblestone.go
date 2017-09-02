package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type color byte

const (
	yellow color = 'Y'
	red    color = 'R'
	green  color = 'G'
	purple color = 'P'
	black  color = '.'
)

type board [5][]color

func (c color) String() string {
	return string(c)
}

func parseBoard(f *os.File) (*board, error) {
	scanner := bufio.NewScanner(f)
	var b board
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 9 {
			return nil, fmt.Errorf("bad line: %q", scanner.Text())
		}
		line += strings.Repeat(" ", 9-len(line))
		for i := range b {
			field := line[i*2]
			if field == ' ' {
				continue
			}
			b[i] = append(b[i], color(field))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for i := range b {
		reverse(b[i])
	}
	return &b, nil
}

func reverse(s []color) {
	for i := 0; i < len(s)/2; i++ {
		j := len(s) - i - 1
		s[i], s[j] = s[j], s[i]
	}
}

func (b *board) String() string {
	var buf bytes.Buffer
	for i := len(b) - 1; i >= 0; i-- {
		fmt.Fprintln(&buf, b[i])
	}
	return buf.String()
}

type state [5]int

type move struct {
	col   int
	color color
}

func (mv move) String() string {
	return fmt.Sprintf("%d (%s)", mv.col, mv.color)
}

type triple [3]move

var errStuck = errors.New("stuck")

func (b *board) empty(st state) bool {
	for i := range b {
		if st[i] < len(b[i]) {
			return false
		}
	}
	return true
}

func (b *board) advance(st state, col int) (color, state, bool) {
	if st[col] >= len(b[col]) {
		return 0, st, false
	}
	c := b[col][st[col]]
	st[col]++
	// Check if we can get rid of a black block.
	// We might be able to remove multiple stacked on top of each other.
	for {
		removed := false
		for i := range b {
			canRemove := false
			if st[i] < len(b[i]) && b[i][st[i]] == black {
				canRemove = true
				depth := len(b[i]) - st[i]
				for j := range b {
					if j == i {
						continue
					}
					if len(b[j])-st[j] >= depth {
						canRemove = false
						break
					}
				}
			}
			if canRemove {
				removed = true
				st[i]++
			}
		}
		if !removed {
			return c, st, true
		}
	}
}

func (b *board) solve(st state) ([]triple, bool) {
	if b.empty(st) {
		return nil, true // done
	}
	var c color
	for i := range b {
		c1, st1, ok := b.advance(st, i)
		if !ok {
			continue
		}
		if c1 == black {
			continue
		}
		c = c1
		for j := range b {
			c2, st2, ok := b.advance(st1, j)
			if !ok {
				continue
			}
			if c2 != c {
				continue
			}
			for k := range b {
				c3, st3, ok := b.advance(st2, k)
				if !ok {
					continue
				}
				if c3 != c {
					continue
				}
				trip := triple{
					move{col: i, color: c},
					move{col: j, color: c},
					move{col: k, color: c},
				}
				if soln, ok := b.solve(st3); ok {
					return append([]triple{trip}, soln...), true
				}
			}
		}
	}
	return nil, false
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	board, err := parseBoard(f)
	if err != nil {
		log.Fatal(err)
	}

	soln, ok := board.solve(state{})
	if !ok {
		log.Fatal("no solution")
	}
	for _, mv := range soln {
		fmt.Println(mv)
	}
}
