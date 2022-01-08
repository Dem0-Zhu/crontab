package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 建立连接
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		err        error
	)
	if client, err = mongo.Connect(context.TODO(), &options.ClientOptions{Hosts: []string{"192.168.5.16:27017"}}); err != nil {
		fmt.Println(err)
		return
	}

	// 选择数据库
	database = client.Database("cron")

	// 选择表
	collection = database.Collection("log")
	collection = collection
}
