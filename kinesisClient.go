package main

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/jroimartin/gocui"
)

var (
	kinesisClient     *kinesis.Client = nil
	recordPageCounter                 = 5
	msgDict                           = make(map[string][]byte)
	partitionKey                      = "partition-1"
)

func listStream() ([]string, error) {
	streamsInput := &kinesis.ListStreamsInput{}
	streamsOutput, err := kinesisClient.ListStreams(ctx, streamsInput)
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
	kinesisClient = kinesis.NewFromConfig(config)
	return kinesisClient, nil
}

func listShards(name string) (*string, error) {
	listShardInput := &kinesis.ListShardsInput{StreamName: &name}
	shards, err := kinesisClient.ListShards(ctx, listShardInput)
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
	records, err := kinesisClient.GetRecords(ctx, recordsInput)
	if err != nil {
		return err
	}
	nextIterator := records.NextShardIterator
	updateList(g, records)
	counter--
	return getRecords(g, name, shardID, nextIterator, counter)
}

func getIterator(shardID *string, streamName string) (*string, error) {
	shardIteratorInput := &kinesis.GetShardIteratorInput{
		ShardId:           shardID,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		StreamName:        &streamName,
	}
	iterator, err := kinesisClient.GetShardIterator(ctx, shardIteratorInput)
	if err != nil {
		return nil, err
	}
	return iterator.ShardIterator, nil
}

func insertRecord(streamName string, data []byte) (string, error) {
	putRecordInput := &kinesis.PutRecordInput{
		Data:         data,
		PartitionKey: &partitionKey,
		StreamName:   &streamName,
	}
	out, err := kinesisClient.PutRecord(ctx, putRecordInput)
	if err != nil {
		return "", err
	}
	return *out.SequenceNumber, nil
}
