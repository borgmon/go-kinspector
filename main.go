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

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if logV, err := g.SetView("log", -1, maxY-10, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(logV, "starting...")
	}
	if mainV, err := g.SetView("select_name", -1, -1, maxX/5, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		mainV.Autoscroll = true
		err = listStream(g, mainV)
		if err != nil {
			addLog(g, err)
		}
	}

	if _, err := g.SetView("select_msg", maxX/5, -1, 2*maxX/5, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
	}

	if _, err := g.SetView("select_detail", 2*maxX/5, -1, maxX, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
	}

	if helpV, err := g.SetView("help", -1, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(helpV, "q to quit; e to export line as json file")
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
