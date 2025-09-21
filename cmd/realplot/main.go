package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode"

	"github.com/charadev96/realplot/internal"

	"github.com/alexflint/go-arg"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
)

type args struct {
	Min         int   `arg:"required" help:"lower bound"`
	Max         int   `arg:"required" help:"upper bound"`
	NoBorder    bool  `arg:"--no-border" help:"disables the border"`
	BorderColor color `arg:"--border-color" placeholder:"COLOR" default:"black" help:"color of the border"`
}

func (args) Epilogue() string {
	return fmt.Sprintf(
		"Option COLOR:\n  - Any HEX color (e.g. #ffffff)\n  - Named color: %v",
		terminalColors,
	)
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
	for {
		select {
		case err := <-errs:
			if err == nil {
				return
			}
			log.Printf("plotter error: %v", err)
		default:
			return
		}
	}
}

func main() {
	var args args
	arg.MustParse(&args)
	log.SetFlags(0)

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
	errPlot := make(chan error)
	plot := plotter.New(screen, scan, errPlot, plotter.PlotConfig{
		BoundMin:    args.Min,
		BoundMax:    args.Max,
		NoBorder:    args.NoBorder,
		StyleBorder: tcell.StyleDefault.Foreground(args.BorderColor.Value),
	})

	quit := func() {
		screen.Fini()
		printErrs(errPlot)
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
