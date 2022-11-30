package main

import (
	"context"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/option"
)

var dbName = "projects/PROJECT_ID/instances/INSTANCE_ID/databases/DATABASE_ID"
var client *spanner.Client

func init() {
	ctx := context.Background()
	options := []option.ClientOption{
		option.WithCredentialsFile("PATH_TO_CREDENTIALS_FILE"),
	}
	client, err := spanner.NewClient(ctx, dbName, options...)
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	client.Single().ReadRow(ctx, "Users", spanner.Key{"alice"}, []string{"name"})
}
