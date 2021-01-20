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
			addLog(g, err)
		}
		v.Title = panelStreamName
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		err = getStreamNames(g, v)
		if err != nil {
			addLog(g, err)
		}
		if _, err := setCurrentViewOnTop(g, panelStreamName); err != nil {
			return err
		}
	}

	if v, err := g.SetView(panelMessage, maxX/4, 1, 2*maxX/4-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = panelMessage
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}

	if v, err := g.SetView(panelData, 2*maxX/4, 1, maxX-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			addLog(g, err)
		}
		v.Title = panelData
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}

	if helpV, err := g.SetView(panelHelp, 1, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		genHelp(helpV, map[string]string{"q": "quit", "e": "export json", "i": "insert record"})
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
		if err := g.SetKeybinding(n, gocui.MouseLeft, gocui.ModNone, mouseClick); err != nil {
			return err
		}
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
	for _, n := range []string{panelData} {
		if err := g.SetKeybinding(n, 'e', gocui.ModNone, exportJSON); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding(panelPopUp, gocui.KeyEsc, gocui.ModNone, closePopup); err != nil {
		return err
	}
	if err := g.SetKeybinding(panelMessage, 'i', gocui.ModNone, addNewRecord); err != nil {
		return err
	}
	if err := g.SetKeybinding(panelPopUp, gocui.KeyEnter, gocui.ModNone, conformPopup); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func printDebug(g *gocui.Gui, v *gocui.View) error {
	addLog(g, msgDict)
	return nil
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
	v.MoveCursor(0, -1, true)

	return nil
}

func listItemDown(g *gocui.Gui, v *gocui.View) error {
	l, err := getLine(v, 1)
	if err != nil {
		addLog(g, err)
		return err
	}
	if l != "" {
		v.MoveCursor(0, 1, true)
	}
	return nil
}

func listItemBack(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case panelData:
		swapFocus(g, v, panelMessage)
		setCurrentViewOnTop(g, panelMessage)
	case panelMessage:
		setCurrentViewOnTop(g, panelStreamName)
	}
	return nil
}

func listItemSelect(g *gocui.Gui, v *gocui.View) error {
	l, err := getLine(v, 0)
	if err != nil {
		addLog(g, err)
	}

	switch v.Name() {
	case panelStreamName:
		if err := clearView(g, panelMessage); err != nil {
			addLog(g, err)
		}
		if err := clearView(g, panelData); err != nil {
			addLog(g, err)
		}
		if _, err := setCurrentViewOnTop(g, panelMessage); err != nil {
			addLog(g, err)
		}
		go populateList(g, l)
	case panelMessage:
		if err := clearView(g, panelData); err != nil {
			addLog(g, err)
		}
		if _, err := setCurrentViewOnTop(g, panelData); err != nil {
			addLog(g, err)
		}
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

func getLine(v *gocui.View, modifier int) (l string, err error) {
	_, cy := v.Cursor()
	if l, err = v.Line(cy + modifier); err != nil {
		l = ""
	}
	return
}

func genHelp(view *gocui.View, hotkeymap map[string]string) {
	for k, v := range hotkeymap {
		fmt.Fprintf(view, "%v \033[32;7m%v\033[0m ", k, v)
	}
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
		return err
	}
	l, err := getLine(msgV, 0)
	if err != nil {
		addLog(g, err.Error())
		return err
	}
	fileName := l + ".json"
	f, err := os.Create(fileName)
	if err != nil {
		addLog(g, err.Error())
		return err
	}
	defer f.Close()

	p := make([]byte, 5)
	v.Rewind()
	for {
		n, err := v.Read(p)
		if n > 0 {
			if _, err := f.Write(p[:n]); err != nil {
				addLog(g, err.Error())
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			addLog(g, err.Error())
			return err
		}
	}
	addLog(g, "saved to "+fileName)
	return nil
}

// warning: msg and input false
// selection: msg and input true
// input: nil and input true
func popUp(g *gocui.Gui, title string, msg []string, input bool) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(panelPopUp, maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = title
		for _, e := range msg {
			fmt.Fprintln(v, e)
		}
		if msg != nil && input {
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

		}
		if msg == nil && input {
			v.Editable = true
		}
		if _, err := g.SetCurrentView(panelPopUp); err != nil {
			return err
		}
	}
	return nil
}

func closePopup(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(panelPopUp); err != nil {
		return err
	}
	if _, err := g.SetCurrentView(panelMessage); err != nil {
		return err
	}
	return nil
}

func addNewRecord(g *gocui.Gui, v *gocui.View) error {
	if err := popUp(g, "New Record", nil, true); err != nil {
		return err
	}
	return nil
}

func conformPopup(g *gocui.Gui, v *gocui.View) error {
	msgV, err := g.View(panelStreamName)
	if err != nil {
		addLog(g, err.Error())
		return err
	}
	streamName, err := getLine(msgV, 0)
	if err != nil {
		addLog(g, err.Error())
		return err
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
			addLog(g, "successfully put new record")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			addLog(g, err.Error())
			return err
		}
	}

	return closePopup(g, v)
}
