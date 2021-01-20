package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/jroimartin/gocui"
	"github.com/tidwall/pretty"
)

var (
	client            *kinesis.Client = nil
	recordPageCounter                 = 5
	msgDict                           = make(map[string][]byte)
)

func getStreamNames(g *gocui.Gui, v *gocui.View) (err error) {
	client, err = getClient()
	if err != nil {
		return err
	}

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

func listStream() ([]string, error) {
	streamsInput := &kinesis.ListStreamsInput{}
	streamsOutput, err := client.ListStreams(ctx, streamsInput)
	if err != nil {
		return nil, err
	}

	return streamsOutput.StreamNames, nil
}

func getClient() (*kinesis.Client, error) {
	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return kinesis.NewFromConfig(config), nil
}

func populateList(g *gocui.Gui, name string) {
	addLog(g, "getting shards...")
	shard, err := listShards(name)
	if err != nil {
		addLog(g, err)
	}
	addLog(g, "getting records...")
	err = getRecords(g, name, shard, nil, recordPageCounter)
	if err != nil {
		addLog(g, err)
	}
}

func listShards(name string) (*string, error) {
	listShardInput := &kinesis.ListShardsInput{StreamName: &name}
	shards, err := client.ListShards(ctx, listShardInput)
	if err != nil {
		return nil, err
	}
	return shards.Shards[0].ShardId, nil
}

func getRecords(g *gocui.Gui, name string, shardID *string, iterator *string, counter int) error {
	if counter < 0 {
		return nil
	}

	var err error
	if iterator == nil {
		iterator, err = getIterator(shardID, name)
		if err != nil {
			return err
		}
	}

	recordsInput := &kinesis.GetRecordsInput{ShardIterator: iterator}
	records, err := client.GetRecords(ctx, recordsInput)
	if err != nil {
		return err
	}
	nextIterator := records.NextShardIterator
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
	counter--
	return getRecords(g, name, shardID, nextIterator, counter)
}

func getIterator(shardID *string, streamName string) (*string, error) {
	shardIteratorInput := &kinesis.GetShardIteratorInput{
		ShardId:           shardID,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		StreamName:        &streamName,
	}
	iterator, err := client.GetShardIterator(ctx, shardIteratorInput)
	if err != nil {
		return nil, err
	}
	return iterator.ShardIterator, nil
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
