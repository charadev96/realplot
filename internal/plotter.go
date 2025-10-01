package plotter

import (
	"bufio"
	"math"
	"strconv"

	draw "github.com/cjbassi/gotop/src/termui/drawille-go"
	"github.com/gammazero/deque"
	"github.com/gdamore/tcell/v2"
	"github.com/mazznoer/colorgrad"
)

const (
	boxTopLeft     = '╭'
	botBottomLeft  = '╰'
	boxTopRight    = '╮'
	boxBottomRight = '╯'

	boxV = '│'
	boxH = '─'

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
	BoundMin    int
	BoundMax    int
	NoBorder    bool
	ColorBorder tcell.Color
	ColorsGraph []colorgrad.Color
}

type Plotter struct {
	size     int
	config   PlotConfig
	errs     chan<- error
	screen   tcell.Screen
	scanner  *bufio.Scanner
	deque    *deque.Deque[int]
	gradient colorgrad.Gradient
}

func New(
	screen tcell.Screen, scanner *bufio.Scanner,
	errs chan<- error, config PlotConfig,
) (*Plotter, error) {
	deque := new(deque.Deque[int])
	width, _ := screen.Size()
	size := width * brailleWidth

	grad, err := colorgrad.NewGradient().Colors(config.ColorsGraph...).Build()
	if err != nil {
		return nil, err
	}

	deque.SetBaseCap(size)
	p := Plotter{
		size:     size,
		config:   config,
		errs:     errs,
		screen:   screen,
		scanner:  scanner,
		deque:    deque,
		gradient: grad,
	}
	return &p, nil
}

func (p *Plotter) Run() func() {
	f := func() {
		defer close(p.errs)
		for p.scanner.Scan() {
			text := p.scanner.Text()
			num, err := strconv.Atoi(text)
			if err != nil {
				p.errs <- err
				continue
			}
			p.push(num)
			p.draw()
			p.screen.Show()
		}
		if err := p.scanner.Err(); err != nil {
			p.errs <- err
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

	if p.config.NoBorder {
		p.plot(0, 0, width, height)
	} else {
		p.plot(1, 1, width-2, height-2)
		p.drawBox(0, 0, width-1, height-1)
	}
}

func (p *Plotter) plot(x, y, w, h int) {
	bottom := h*brailleHeight - 1
	right := w*brailleWidth - 1

	styles := make(map[int]tcell.Style)

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

	for oy := range h {
		percent := float64(oy) / float64(h)
		color := tcell.GetColor(p.gradient.At(percent).HexString())
		styles[oy*w] = tcell.StyleDefault.Foreground(color)
	}

	p.drawString(x, y, c.String(), styles)
}

func (p *Plotter) drawString(x, y int, s string, styles map[int]tcell.Style) {
	ox, oy := x, y
	pos := 0
	style := tcell.StyleDefault
	for _, r := range s {
		if st, ok := styles[pos]; ok {
			style = st
		}
		if r == '\n' {
			ox = x
			oy++
			continue
		}
		p.screen.SetContent(ox, oy, r, nil, style)
		ox++
		pos++
	}
}

func (p *Plotter) drawHLine(x, y, l int, style tcell.Style) {
	for ox := x; ox < x+l; ox++ {
		p.screen.SetContent(ox, y, boxH, nil, style)
	}
}

func (p *Plotter) drawVLine(x, y, l int, style tcell.Style) {
	for oy := y; oy < y+l; oy++ {
		p.screen.SetContent(x, oy, boxV, nil, style)
	}
}

func (p *Plotter) drawBox(x, y, w, h int) {
	style := tcell.StyleDefault.Foreground(p.config.ColorBorder)

	p.screen.SetContent(x, y, boxTopLeft, nil, style)
	p.screen.SetContent(x+w, y, boxTopRight, nil, style)
	p.screen.SetContent(x+w, y+h, boxBottomRight, nil, style)
	p.screen.SetContent(x, y+h, botBottomLeft, nil, style)

	p.drawHLine(x+1, y, w-1, style)
	p.drawHLine(x+1, y+h, w-1, style)
	p.drawVLine(x, y+1, h-1, style)
	p.drawVLine(x+w, y+1, h-1, style)
}

func (p *Plotter) push(x int) {
	if p.deque.Len() == p.size {
		p.deque.PopBack()
	}
	p.deque.PushFront(x)
}
