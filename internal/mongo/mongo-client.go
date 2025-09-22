package mongoclient

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoClient struct {
	uri    string
	client *mongo.Client
}

func New(uri string) *MongoClient {
	return &MongoClient{
		uri: uri,
	}
}

func (mc *MongoClient) Connect(ctx context.Context) error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(mc.uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(nil, opts)
	if err != nil {
		return fmt.Errorf("connecting to mongo: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("checking mongo connection: %w", err)
	}

	mc.client = client

	return nil
}

func (mc *MongoClient) Disconnect(ctx context.Context) error {
	if err := mc.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnecting from mongo: %w", err)
	}

	return nil
}

func (mc *MongoClient) Client() *mongo.Client {
	return mc.client
}
