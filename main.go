package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

const (
	panelStreamName = "StreamName"
	panelMessage    = "Messages"
	panelData       = "Data"
	panelHelp       = "Help"
	panelLog        = "Log"
	panelPopUp      = "PopUp"
)

var (
	ctx = context.TODO()
)

func main() {
	var err error
	g, err := gocui.NewGui(gocui.OutputNormal)
	logFatal(err)
	defer g.Close()

	g.SetManagerFunc(layout)
	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}
	g.Mouse = true
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView(panelLog, 1, maxY-9, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = panelLog
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintln(v, "starting...")
	}

	if v, err := g.SetView(panelStreamName, 1, 1, maxX/4-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = panelStreamName
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		_, err := getClient()
		if err != nil {
			return err
		}
		err = getStreamNames(g, v)
		logError(g, err)
		if _, err := setCurrentViewOnTop(g, panelStreamName); err != nil {
			return err
		}
	}

	if v, err := g.SetView(panelMessage, maxX/4, 1, 2*maxX/4-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = panelMessage
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}

	if v, err := g.SetView(panelData, 2*maxX/4, 1, maxX-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = panelData
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}

	if v, err := g.SetView(panelHelp, 1, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		genHelp(v, map[string]string{"q": "quit", "e": "export json", "i": "insert record"})
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	// global
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	// mouse
	for _, n := range []string{panelStreamName, panelMessage} {
		if err := g.SetKeybinding(n, gocui.MouseLeft, gocui.ModNone, mouseClick); err != nil {
			return err
		}
	}
	// arrow keys
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
	// exportJSON
	for _, n := range []string{panelData} {
		if err := g.SetKeybinding(n, 'e', gocui.ModNone, exportJSON); err != nil {
			return err
		}
	}
	// popups
	if err := g.SetKeybinding(panelPopUp, gocui.KeyEsc, gocui.ModNone, closePopup); err != nil {
		return err
	}
	for _, n := range []string{panelData, panelStreamName, panelMessage} {
		if err := g.SetKeybinding(n, 'i', gocui.ModNone, addNewRecord); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding(panelPopUp, gocui.KeyEnter, gocui.ModNone, conformPopup); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func listItemUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -1, true)

	return nil
}

func listItemDown(g *gocui.Gui, v *gocui.View) error {
	l, err := getLine(v, 1)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}
	if l != "" {
		v.MoveCursor(0, 1, true)
	}
	return nil
}

func listItemBack(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case panelData:
		if err := swapFocus(g, v, panelMessage); err != nil {
			addLog(g, err.Error())
			return nil
		}
		if _, err := setCurrentViewOnTop(g, panelMessage); err != nil {
			addLog(g, err.Error())
			return nil
		}

	case panelMessage:
		if _, err := setCurrentViewOnTop(g, panelStreamName); err != nil {
			addLog(g, err.Error())
			return nil
		}

	}
	return nil
}

func listItemSelect(g *gocui.Gui, v *gocui.View) error {
	l, err := getLine(v, 0)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}

	switch v.Name() {
	case panelStreamName:
		if err := clearView(g, panelMessage); err != nil {
			addLog(g, err.Error())
			return nil
		}
		if err := clearView(g, panelData); err != nil {
			addLog(g, err.Error())
			return nil
		}
		if _, err := setCurrentViewOnTop(g, panelMessage); err != nil {
			addLog(g, err.Error())
			return nil
		}
		go populateList(g, l)
	case panelMessage:
		if err := clearView(g, panelData); err != nil {
			addLog(g, err.Error())
			return nil
		}
		if _, err := setCurrentViewOnTop(g, panelData); err != nil {
			addLog(g, err.Error())
			return nil
		}
		if err := showMessage(g, l); err != nil {
			addLog(g, err.Error())
			return nil
		}
	}
	return nil
}

func mouseClick(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case panelStreamName:
		return listItemSelect(g, v)
	case panelMessage:
		return listItemSelect(g, v)
	}
	return nil
}

func exportJSON(g *gocui.Gui, v *gocui.View) error {
	msgV, err := g.View(panelMessage)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}
	l, err := getLine(msgV, 0)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}
	if l == "" {
		return nil
	}
	fileName := l + ".json"
	f, err := os.Create(fileName)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}
	defer f.Close()

	p := make([]byte, 5)
	v.Rewind()
	for {
		n, err := v.Read(p)
		if n > 0 {
			if _, err := f.Write(p[:n]); err != nil {
				addLog(g, err.Error())
				return nil
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			addLog(g, err.Error())
			return nil
		}
	}
	addLog(g, "saved to "+fileName)
	return nil
}

func closePopup(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(panelPopUp); err != nil {
		addLog(g, err.Error())
		return nil
	}
	if _, err := g.SetCurrentView(panelMessage); err != nil {
		addLog(g, err.Error())
		return nil
	}
	return nil
}

func addNewRecord(g *gocui.Gui, v *gocui.View) error {
	if err := popUp(g, "New Record", nil, true); err != nil {
		addLog(g, err.Error())
		return nil
	}
	return nil
}

func conformPopup(g *gocui.Gui, v *gocui.View) error {
	msgV, err := g.View(panelStreamName)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}
	streamName, err := getLine(msgV, 0)
	if err != nil {
		addLog(g, err.Error())
		return nil
	}

	p := make([]byte, 5)
	v.Rewind()
	for {
		n, err := v.Read(p)
		if n > 0 {
			if _, err := insertRecord(streamName, p[:n]); err != nil {
				addLog(g, err.Error())
				return err
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			addLog(g, err.Error())
			return nil
		}
	}
	addLog(g, "successfully put new record")

	return closePopup(g, v)
}
