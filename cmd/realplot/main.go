package main

import (
	"bufio"
	"log"
	"os"
	"unicode"

	"github.com/charadev96/realplot/internal"

	"github.com/alexflint/go-arg"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/mazznoer/colorgrad"
)

type args struct {
	Min         int             `arg:"required" help:"lower bound"`
	Max         int             `arg:"required" help:"upper bound"`
	NoBorder    bool            `arg:"--no-border" help:"disables the border"`
	ColorBorder plotter.Color   `arg:"--color-border" placeholder:"COLOR" default:"black" help:"color of the border"`
	ColorsGraph []plotter.Color `arg:"--colors-graph" placeholder:"COLOR ..." help:"list of colors for the graph (HEX only)"`
}

func (args) Epilogue() string {
	return `Option COLOR:
  - Any HEX color (e.g. #ffffff)
  - Named terminal color, refer to https://github.com/gdamore/tcell/blob/v2.9.0/color.go#L851`
}

func (args) Description() string {
	return "A simple terminal-based bar graph plotter\n"
}

func eventIsQuit(ev *tcell.EventKey) bool {
	key := ev.Key()
	if key == tcell.KeyRune {
		return unicode.ToLower(ev.Rune()) == 'q'
	}
	if key == tcell.KeyCtrlC {
		return true
	}
	return false
}

func printErrs(errs <-chan error) {
	c := 1
	for {
		select {
		case err := <-errs:
			if err == nil {
				continue
			}
			log.Printf("plotter (%d): %v", c, err)
			c++
		default:
			return
		}
	}
}

func main() {
	var args args
	parser := arg.MustParse(&args)
	log.SetFlags(0)

	var colorsGraph []colorgrad.Color
	for _, c := range args.ColorsGraph {
		if c.Type() == plotter.ColorTerminal {
			parser.Fail("option --graph-colors cannot contain named colors")
		}
		color, err := c.Colorgrad()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		colorsGraph = append(colorsGraph, color)
	}
	if len(args.ColorsGraph) == 0 {
		colorsGraph = append(colorsGraph, colorgrad.Color{})
	}

	if isatty.IsTerminal(os.Stdin.Fd()) {
		log.Fatal("error: requires stdin from pipe to function")
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal("error: %v", err)
	}
	screen.Clear()

	scan := bufio.NewScanner(os.Stdin)
	errs := make(chan error, 64)
	plot, err := plotter.New(screen, scan, errs, plotter.PlotConfig{
		BoundMin:    args.Min,
		BoundMax:    args.Max,
		NoBorder:    args.NoBorder,
		ColorBorder: args.ColorBorder.Tcell(),
		ColorsGraph: colorsGraph,
	})
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	quit := func() {
		screen.Fini()
		printErrs(errs)
		if r := recover(); r != nil {
			log.Fatal(r)
		}
		os.Exit(0)
	}
	defer quit()

	go plot.Run()()

	for {
		ev := screen.PollEvent()
		screen.Show()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			plot.Resize()
			screen.Sync()
		case *tcell.EventKey:
			if eventIsQuit(ev) {
				return
			}
		}
	}
}
