package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type TimePoint struct {
	CreatAt int64 `bson:"creat_at"`
	EndAt   int64 `bson:"end_at"`
}

type LogRecord struct {
	JobName string `bson:"job_name"`
	// shell命令（任务）
	Command string `bson:"command"`
	Err     string `bson:"err"`
	// 返回内容
	Context   string    `bson:"context"`
	TimePoint TimePoint `bson:"time_point"`
}

// jobName过滤条件
type FindByJobName struct {
	JobName string `bson:"job_name"`
}

// 	增查
func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		record     *LogRecord
		result     *mongo.InsertOneResult
		docId      primitive.ObjectID
		err        error
	)
	if client, err = mongo.Connect(context.TODO(), &options.ClientOptions{Hosts: []string{"192.168.5.16:27017"}}); err != nil {
		fmt.Println(err)
		return
	}
	database = client.Database("cron")
	collection = database.Collection("log")

	record = &LogRecord{
		JobName: "job10",
		Command: "echo hello",
		Err:     "",
		Context: "hello",
		TimePoint: TimePoint{
			CreatAt: time.Now().Unix(),
			EndAt:   time.Now().Unix() + 10,
		},
	}
	if result, err = collection.InsertOne(context.TODO(), record); err != nil {
		fmt.Println(err)
		return
	}
	docId = result.InsertedID.(primitive.ObjectID)
	fmt.Println("自增id为：", docId.Hex())

	// -------------------读-----------------------
	var (
		cond      *FindByJobName
		cursor    *mongo.Cursor
		recordRes *LogRecord
	)
	cond = &FindByJobName{
		JobName: "job10",
	}
	skip, limit := int64(0), int64(2)
	if cursor, err = collection.Find(context.TODO(), cond, &options.FindOptions{Skip: &skip, Limit: &limit}); err != nil {
		fmt.Println(err)
		return
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		recordRes = &LogRecord{}
		// 反序列化
		if err = cursor.Decode(recordRes); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(*recordRes)
	}

}
