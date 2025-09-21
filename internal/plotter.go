package plotter

import (
	"bufio"
	"math"
	"strconv"

	draw "github.com/cjbassi/gotop/src/termui/drawille-go"
	"github.com/gammazero/deque"
	"github.com/gdamore/tcell/v2"
)

const (
	brailleHeight = 4
	brailleWidth  = 2
)

func clamp(value, low, up int) int {
	if value < low {
		return low
	}
	if value > up {
		return up
	}
	return value
}

func mapRange(value, min1, max1, min2, max2 int) int {
	value = clamp(value, min1, max1)
	fraction := float64(value-min1) / float64(max1-min1)
	final := int(math.Round(fraction*float64(max2-min2))) + min2
	final = clamp(final, min2, max2)
	return final
}

type PlotConfig struct {
	BoundMin int
	BoundMax int
}

type Plotter struct {
	size    int
	config  PlotConfig
	err     chan<- error
	screen  tcell.Screen
	scanner *bufio.Scanner
	deque   *deque.Deque[int]
}

func New(
	screen tcell.Screen, scanner *bufio.Scanner,
	err chan<- error, config PlotConfig,
) *Plotter {
	deque := new(deque.Deque[int])
	width, _ := screen.Size()
	size := width * brailleWidth

	deque.SetBaseCap(size)
	p := Plotter{
		size:    size,
		config:  config,
		err:     err,
		screen:  screen,
		scanner: scanner,
		deque:   deque,
	}
	return &p
}

func (p *Plotter) Run() func() {
	f := func() {
		defer close(p.err)
		for p.scanner.Scan() {
			text := p.scanner.Text()
			num, err := strconv.Atoi(text)
			if err != nil {
				p.err <- err
				continue
			}
			p.push(num)
			p.draw()
			p.screen.Show()
		}
		if err := p.scanner.Err(); err != nil {
			p.err <- err
			return
		}
	}
	return f
}

func (p *Plotter) Resize() {
	width, _ := p.screen.Size()
	width *= brailleWidth
	p.deque.SetBaseCap(width)
	p.size = width
	p.draw()
}

func (p *Plotter) draw() {
	width, height := p.screen.Size()
	p.plot(0, 0, width, height)
}

func (p *Plotter) plot(x, y, w, h int) {
	bottom := h*brailleHeight - 1
	right := w*brailleWidth - 1

	c := draw.NewCanvas()

	ox := right
	for v := range p.deque.Iter() {
		mapped := mapRange(v, p.config.BoundMin, p.config.BoundMax, 0, bottom)
		c.DrawLine(ox, bottom, ox, bottom-mapped)
		ox--
		if ox == -1 {
			break
		}
	}
	p.drawString(x, y, c.String())
}

func (p *Plotter) drawString(x, y int, s string) {
	ox, oy := x, y
	for _, r := range s {
		if r == '\n' {
			ox = x
			oy++
			continue
		}
		p.screen.SetContent(ox, oy, r, nil, tcell.StyleDefault)
		ox++
	}
}

func (p *Plotter) push(x int) {
	if p.deque.Len() == p.size {
		p.deque.PopBack()
	}
	p.deque.PushFront(x)
}
