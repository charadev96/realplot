package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var terminalColors = [16]string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"lightgray",
	"gray",
	"lightred",
	"lightgreen",
	"lightyellow",
	"lightblue",
	"lightmagenta",
	"lightcyan",
	"white",
}

type color struct {
	Value tcell.Color
}

func (c *color) UnmarshalText(b []byte) error {
	s := string(b)
	s = strings.ToLower(s)

	for i, col := range terminalColors {
		if s == col {
			c.Value = tcell.ColorValid + tcell.Color(i)
			return nil
		}
	}

	if s[0] == '#' {
		hex, err := strconv.ParseInt(s[1:], 16, 32)
		if err != nil {
			return fmt.Errorf("color %q is not a valid hex color")
		}
		c.Value = tcell.NewHexColor(int32(hex))
		return nil
	}

	return fmt.Errorf("color %q is not a valid terminal color", s)
}
