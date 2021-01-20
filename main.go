package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
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

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if mainV, err := g.SetView("main_view", -1, -1, maxX, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		mainV.Autoscroll = true
		err = listStream(g, mainV)
		if err != nil {
			addLog(g, err)
		}
	}
	if logV, err := g.SetView("log", -1, maxY-10, maxX, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(logV, "starting...")
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

func listStream(g *gocui.Gui, v *gocui.View) error {

	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := kinesis.NewFromConfig(config)

	streamsInput := &kinesis.ListStreamsInput{}
	streamsOutput, err := client.ListStreams(ctx, streamsInput)
	if err != nil {
		return err
	}

	g.Update(func(g *gocui.Gui) error {
		_, err := drawSelectPopup(g, streamsOutput.StreamNames)
		if err != nil {
			return err
		}

		return nil
	})
	return nil

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
func drawSelectPopup(g *gocui.Gui, ls []string) (*gocui.View, error) {
	maxX, maxY := g.Size()
	if v, err := g.SetView("select_popup", maxX/2-20, (maxY-len(ls))/2, maxX/2+20, ((maxY-len(ls))/2)+len(ls)); err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		for _, e := range ls {
			fmt.Fprintln(v, e)
		}
		return v, nil
	}
	return nil, errors.New("fail create popup")
}
