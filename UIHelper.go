package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/jroimartin/gocui"
	"github.com/tidwall/pretty"
)

func addLog(g *gocui.Gui, msg string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(panelLog)
		if err != nil {
			return err
		}
		fmt.Fprintln(v, msg)
		return nil
	})
}

func logError(g *gocui.Gui, err error) error {
	if err != nil {
		addLog(g, err.Error())
	}
	return err
}

func logFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func genHelp(view *gocui.View, hotkeymap map[string]string) {
	for k, v := range hotkeymap {
		fmt.Fprintf(view, "%v \033[32;7m%v\033[0m ", k, v)
	}
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

func populateList(g *gocui.Gui, name string) {
	addLog(g, "getting shards...")
	msgDict = make(map[string][]byte)
	shard, err := listShards(name)
	if err != nil {
		addLog(g, err.Error())
	}
	addLog(g, "getting records...")
	err = getRecords(g, name, shard, nil, recordPageCounter)
	if err != nil {
		addLog(g, err.Error())
	}
	addLog(g, "done loading records")

	if len(msgDict) == 0 {
		addLog(g, "no record found")
	}
}

func showMessage(g *gocui.Gui, sequenceNumber string) error {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View(panelData)
		if err != nil {
			return err
		}
		jByte := pretty.Pretty(msgDict[sequenceNumber])
		fmt.Fprintln(v, string(jByte))
		return nil
	})
	return nil
}

func getStreamNames(g *gocui.Gui, v *gocui.View) (err error) {
	names, err := listStream()
	if err != nil {
		return err
	}

	g.Update(func(g *gocui.Gui) error {
		for _, e := range names {
			fmt.Fprintln(v, e)
		}

		return nil
	})
	return
}

func updateList(g *gocui.Gui, records *kinesis.GetRecordsOutput) {
	g.Update(func(g *gocui.Gui) error {
		if len(records.Records) == 0 {
			return nil
		}

		msgV, err := g.View(panelMessage)
		if err != nil {
			return err
		}

		for _, r := range records.Records {
			key := (*r.ApproximateArrivalTimestamp).Format(time.RFC1123)
			fmt.Fprintln(msgV, key)

			msgDict[key] = r.Data
		}

		return nil
	})
}

func initClient(g *gocui.Gui, v *gocui.View) {
	_, err := getClient()
	if err != nil {
		addLog(g, err.Error())
	}
	getStreamNames(g, v)
	logError(g, err)
}
