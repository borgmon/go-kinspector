package main

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/jroimartin/gocui"
)

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
