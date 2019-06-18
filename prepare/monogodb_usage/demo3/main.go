package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

//TimePoint 任务的执行时间
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

//LogRecord 一条日志
type LogRecord struct {
	JobName   string    `bson:"jobName"`   //任务名
	Command   string    `bson:"command"`   //shell命令
	Err       string    `bson:"err"`       //脚本错误
	Content   string    `bson:"content"`   //脚本输出
	TimePoint TimePoint `bson:"timePoint"` //执行时间信息
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		record     *LogRecord
		result     *mongo.InsertManyResult
		logArr     []interface{}
		insertid   interface{}
		docid      primitive.ObjectID
		err        error
	)

	if client, err = mongo.Connect(context.TODO(), "mongodb://47.102.115.229:27017", options.Client().SetConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	database = client.Database("cron")

	collection = database.Collection("log")

	//批量插入多条document

	record = &LogRecord{
		JobName: "job10",
		Command: "echo hello",
		Err:     "",
		Content: "hello",
		TimePoint: TimePoint{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}

	logArr = []interface{}{
		record,
		record,
		record,
	}

	if result, err = collection.InsertMany(context.TODO(), logArr); err != nil {
		fmt.Println(err)
		return
	}

	//snowflake:毫秒/微秒的当前时间 + 机器ID + 当前毫秒/微秒内的自增ID(每当毫秒变化了,会重置成0,继续自增)
	for _, insertid = range result.InsertedIDs {
		docid = insertid.(primitive.ObjectID)
		fmt.Println("自增id:", docid.Hex())
	}

}
