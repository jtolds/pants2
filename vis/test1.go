package main

import (
	"fmt"
	"image"
	"image/color"

	wde "github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/xgb"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	go func() {
		w, err := wde.NewWindow(100, 100)
		if err != nil {
			panic(err)
		}
		w.SetTitle("yay")
		w.LockSize(true)
		(&font.Drawer{
			Dst:  w.Screen(),
			Src:  image.NewUniform(color.RGBA{0, 255, 0, 255}),
			Face: basicfont.Face7x13,
			Dot:  fixed.Point26_6{fixed.Int26_6(0 * 64), fixed.Int26_6(13 * 64)},
		}).DrawString("hello world")
		w.FlushImage(image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{100, 100}})
		w.Show()
		for ev := range w.EventChan() {
			switch ev := ev.(type) {
			default:
				fmt.Printf("%#v\n", ev)
			case wde.CloseEvent:
				wde.Stop()
				break
			}
		}
	}()
	wde.Run()
}
