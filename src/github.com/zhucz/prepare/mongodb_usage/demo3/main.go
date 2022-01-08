package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// TimeBeforeCond {"$lt": timestamp}
type TimeBeforeCond struct {
	Before int64 `bson:"$lt"`
}

// DeleteCond 删除
// {"TimePoint.creat_at":{"$lt": timestamp}}
type DeleteCond struct {
	beforeCond TimeBeforeCond `bson:"TimePoint.creat_at"`
}

func main() {
	var (
		client       *mongo.Client
		database     *mongo.Database
		collection   *mongo.Collection
		deleteCond   *DeleteCond
		deleteResult *mongo.DeleteResult
		err          error
	)
	if client, err = mongo.Connect(context.TODO(), &options.ClientOptions{Hosts: []string{"192.168.5.9:27017"}}); err != nil {
		fmt.Println(err)
		return
	}
	database = client.Database("cron")
	collection = database.Collection("log")

	// 删除开始时间早于当前时间的日志
	deleteCond = &DeleteCond{
		TimeBeforeCond{
			Before: time.Now().Unix(),
		},
	}
	if deleteResult, err = collection.DeleteMany(context.TODO(), deleteCond); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("删除的行数：", deleteResult.DeletedCount)
}
