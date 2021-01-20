package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

var (
	ctx = context.TODO()
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("log", -1, maxY-10, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Logs"
		fmt.Fprintln(v, "starting...")
	}
	if v, err := g.SetView("name", -1, -1, maxX/4, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = "Stream Names"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.Autoscroll = true
		err = getStreamNames(g, v)
		if err != nil {
			addLog(g, err)
		}
	}

	if v, err := g.SetView("message", maxX/4, -1, 2*maxX/4, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = "SequenceNumber"
	}

	if v, err := g.SetView("data", 2*maxX/4, -1, maxX, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = "Body"
	}

	if helpV, err := g.SetView("help", -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(helpV, "q to quit; e to export line as json file")
	}

	if _, err := g.SetCurrentView("name"); err != nil {
		return err
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}

	for _, n := range []string{"name", "message"} {
		if err := g.SetKeybinding(n, gocui.KeyArrowUp, gocui.ModNone, listItemUp); err != nil {
			return err
		}
	}
	for _, n := range []string{"name", "message"} {
		if err := g.SetKeybinding(n, gocui.KeyArrowDown, gocui.ModNone, listItemDown); err != nil {
			return err
		}
	}
	for _, n := range []string{"name", "message"} {
		if err := g.SetKeybinding(n, gocui.KeyEnter, gocui.ModNone, listItemSelect); err != nil {
			return err
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func addLog(g *gocui.Gui, msg interface{}) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("log")
		if err != nil {
			return err
		}
		fmt.Fprintln(v, msg)
		return nil
	})
}

func listItemUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -1, false)
	return nil
}

func listItemDown(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, 1, false)
	return nil
}

func listItemSelect(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	switch v.Name() {
	case "name":
		if err := populateList(g, l); err != nil {
			addLog(g, err)
		}
	case "message":
		if err := showMessage(g, l); err != nil {
			addLog(g, err)
		}
	}
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}
