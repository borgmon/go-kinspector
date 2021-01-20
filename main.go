package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

const (
	panelStreamName = "StreamName"
	panelMessage    = "Messages"
	panelData       = "Data"
	panelHelp       = "Help"
	panelLog        = "Log"
)

var (
	ctx        = context.TODO()
	panelOrder = []string{panelStreamName, panelMessage, panelData}
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
	if v, err := g.SetView(panelLog, -1, maxY-10, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = panelLog
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintln(v, "starting...")
	}
	if v, err := g.SetView(panelStreamName, -1, -1, maxX/4, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = panelStreamName
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.Autoscroll = true
		err = getStreamNames(g, v)
		if err != nil {
			addLog(g, err)
		}
		if _, err := setCurrentViewOnTop(g, panelStreamName); err != nil {
			return err
		}
	}

	if v, err := g.SetView(panelMessage, maxX/4, -1, 2*maxX/4, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = panelMessage
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.Autoscroll = true

	}

	if v, err := g.SetView(panelData, 2*maxX/4, -1, maxX, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = panelData
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.Autoscroll = true

	}

	if helpV, err := g.SetView(panelHelp, -1, maxY-2, maxX, maxY); err != nil {
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

	for _, n := range []string{panelStreamName, panelMessage} {
		if err := g.SetKeybinding(n, gocui.KeyArrowUp, gocui.ModNone, listItemUp); err != nil {
			return err
		}
	}
	for _, n := range []string{panelStreamName, panelMessage} {
		if err := g.SetKeybinding(n, gocui.KeyArrowDown, gocui.ModNone, listItemDown); err != nil {
			return err
		}
	}
	for _, n := range []string{panelMessage, panelData} {
		if err := g.SetKeybinding(n, gocui.KeyArrowLeft, gocui.ModNone, listItemBack); err != nil {
			return err
		}
	}
	for _, n := range []string{panelStreamName, panelMessage} {
		if err := g.SetKeybinding(n, gocui.KeyArrowRight, gocui.ModNone, listItemSelect); err != nil {
			return err
		}
	}
	for _, n := range []string{panelStreamName, panelMessage} {
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
		v, err := g.View(panelLog)
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

func listItemBack(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case panelData:
		swapFocus(g, v, panelMessage)
		setCurrentViewOnTop(g, panelMessage)
	case panelMessage:
		swapFocus(g, v, panelStreamName)
		setCurrentViewOnTop(g, panelStreamName)
	}
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
	case panelStreamName:
		swapFocus(g, v, panelMessage)
		clearView(g, panelMessage)
		setCurrentViewOnTop(g, panelMessage)
		if err := populateList(g, l); err != nil {
			addLog(g, err)
		}
	case panelMessage:
		swapFocus(g, v, panelData)
		clearView(g, panelData)
		setCurrentViewOnTop(g, panelData)
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

func swapFocus(g *gocui.Gui, oldV *gocui.View, new string) error {
	newV, err := g.View(new)
	if err != nil {
		return nil
	}

	newV.Highlight = true
	oldV.Highlight = false

	return nil
}

func clearView(g *gocui.Gui, name string) error {
	v, err := g.View(name)
	if err != nil {
		return nil
	}
	v.Clear()
	return nil
}
