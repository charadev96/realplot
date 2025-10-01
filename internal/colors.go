package plotter

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mazznoer/colorgrad"
	csscolor "github.com/mazznoer/csscolorparser"
)

type ColorType int

const (
	ColorTerminal ColorType = iota
	ColorHex
)

type Color struct {
	value string
	ctype ColorType
}

func (c *Color) UnmarshalText(b []byte) error {
	s := string(b)
	s = strings.ToLower(s)
	c.value = s

	for col := range tcell.ColorNames {
		if s == col {
			c.ctype = ColorTerminal
			return nil
		}
	}

	if s[0] == '#' {
		if runes := utf8.RuneCountInString(s); runes != 7 {
			return fmt.Errorf("color %q is not a valid hex color", s)
		}

		_, err := strconv.ParseInt(s[1:], 16, 32)
		if err != nil {
			return fmt.Errorf("color %q is not a valid hex color", s)
		}
		c.ctype = ColorHex
		return nil
	}

	return fmt.Errorf("color %q is not a valid terminal color", s)
}

func (c *Color) Tcell() tcell.Color {
	switch c.ctype {
	case ColorTerminal:
		return tcell.ColorNames[c.value]
	case ColorHex:
		return tcell.GetColor(c.value)
	default:
		return tcell.ColorDefault
	}
}

func (c *Color) Colorgrad() (colorgrad.Color, error) {
	if c.ctype != ColorHex {
		err := fmt.Errorf("cannot convert non-hex color %q to colorgrad", c.value)
		return colorgrad.Color{}, err
	}
	color, err := csscolor.Parse(c.value)
	if err != nil {
		return colorgrad.Color{}, err
	}
	return color, nil
}

func (c *Color) String() string {
	return c.value
}

func (c *Color) Type() ColorType {
	return c.ctype
}
