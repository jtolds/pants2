package vis2d

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/jtolds/pants2/interp"
	wde "github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/xgb"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	windowHeight = 768
	windowWidth  = 1024
)

var (
	colors = map[string]color.RGBA{
		"red":   color.RGBA{255, 0, 0, 255},
		"green": color.RGBA{0, 255, 0, 255},
		"blue":  color.RGBA{0, 0, 255, 255},
		"white": color.RGBA{255, 255, 255, 255},
		"black": color.RGBA{0, 0, 0, 255},
	}
)

type mod struct {
	w     wde.Window
	color color.RGBA
	row   int
}

func Mod() (map[string]interp.Value, error) {
	w, err := wde.NewWindow(windowWidth, windowHeight)
	if err != nil {
		return nil, err
	}
	m := &mod{
		w:     w,
		color: colors["white"],
	}

	w.SetTitle("Pants")
	w.LockSize(true)
	w.Show()

	go func() {
		for ev := range w.EventChan() {
			switch ev.(type) {
			case wde.CloseEvent:
				wde.Stop()
				break
			}
		}
	}()

	return map[string]interp.Value{
		"color": interp.ProcCB(m.Color),
		"print": interp.ProcCB(m.Print),
		"pixel": interp.ProcCB(m.Pixel),
		"line":  interp.ProcCB(m.Line),
		"clear": interp.ProcCB(m.Clear),
	}, nil
}

func Run() {
	wde.Run()
}

func Stop() {
	wde.Stop()
}

func (m *mod) Color(args []interp.Value) error {
	if len(args) != 1 {
		return fmt.Errorf("expected only one argument")
	}
	arg, ok := args[0].(interp.ValString)
	if !ok {
		return fmt.Errorf("unexpected value: %#v", args[0])
	}
	color, ok := colors[strings.ToLower(arg.Val)]
	if !ok {
		return fmt.Errorf("unknown color: %#v", arg.Val)
	}
	m.color = color
	return nil
}

func (m *mod) Pixel(args []interp.Value) error {
	if len(args) != 2 {
		return fmt.Errorf("expected two arguments")
	}
	for _, arg := range args {
		if _, ok := arg.(interp.ValNumber); !ok {
			return fmt.Errorf("unexpected value: %#v", arg)
		}
	}
	x, _ := args[0].(interp.ValNumber).Val.Float64()
	y, _ := args[1].(interp.ValNumber).Val.Float64()
	if x < 0 {
		x = 0
	}
	if x > windowWidth {
		x = windowWidth - 1
	}
	if y < 0 {
		y = 0
	}
	if y > windowHeight {
		y = windowHeight - 1
	}
	m.w.Screen().Set(int(x), int(y), m.color)

	m.w.FlushImage(image.Rectangle{
		Min: image.Point{int(x), int(y)},
		Max: image.Point{int(x) + 1, int(y) + 1}})
	return nil
}

func (m *mod) Line(args []interp.Value) error {
	if len(args) != 4 {
		return fmt.Errorf("expected four arguments")
	}
	for _, arg := range args {
		if _, ok := arg.(interp.ValNumber); !ok {
			return fmt.Errorf("unexpected value: %#v", arg)
		}
	}
	x1f, _ := args[0].(interp.ValNumber).Val.Float64()
	y1f, _ := args[1].(interp.ValNumber).Val.Float64()
	x2f, _ := args[2].(interp.ValNumber).Val.Float64()
	y2f, _ := args[3].(interp.ValNumber).Val.Float64()
	x1, x2, y1, y2 := int(x1f), int(x2f), int(y1f), int(y2f)
	if x2 < x1 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}
	y := y1
	ydir := 1
	if y2 < y1 {
		ydir = -1
	}
	for x := x1; x <= x2; x++ {
		denom := x2 - x1
		num := x - x1
		ynext := y1
		if num == denom {
			ynext = y2
		} else {
			ynext += (y2 - y1) * num / denom
		}
		for ; y != ynext; y += ydir {
			if 0 <= x && x < windowWidth &&
				0 <= y && y < windowHeight {
				m.w.Screen().Set(x, y, m.color)
			}
		}
		if y == ynext {
			if 0 <= x && x < windowWidth &&
				0 <= y && y < windowHeight {
				m.w.Screen().Set(x, y, m.color)
			}
		}
	}

	ymin, ymax := y1, y2
	if y2 < y1 {
		ymin, ymax = y2, y1
	}

	m.w.FlushImage(image.Rectangle{
		Min: image.Point{clamp(x1, 0, windowWidth-1),
			clamp(ymin, 0, windowHeight-1)},
		Max: image.Point{clamp(x2, 0, windowWidth-1) + 1,
			clamp(ymax, 0, windowHeight-1) + 1}})
	return nil
}

func clamp(val, min, max int) int {
	if val < min {
		val = min
	}
	if val > max {
		val = max
	}
	return val
}

func (m *mod) Print(args []interp.Value) error {
	if len(args) != 1 {
		return fmt.Errorf("expected only one argument")
	}
	arg, ok := args[0].(interp.ValString)
	if !ok {
		return fmt.Errorf("unexpected value: %#v", args[0])
	}
	(&font.Drawer{
		Dst:  m.w.Screen(),
		Src:  image.NewUniform(m.color),
		Face: basicfont.Face7x13,
		Dot: fixed.Point26_6{
			fixed.Int26_6(0 * 64),
			fixed.Int26_6((m.row + 1) * 13 * 64)},
	}).DrawString(arg.Val)
	m.w.FlushImage(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{windowWidth, windowHeight}})
	m.row += 1
	return nil
}

func (m *mod) Clear(args []interp.Value) error {
	if len(args) != 0 {
		return fmt.Errorf("expected no arguments")
	}
	rect := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{windowWidth, windowHeight}}
	m.w.Screen().CopyRGBA(image.NewRGBA(rect), rect)
	m.w.FlushImage(rect)
	return nil
}
