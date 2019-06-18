package main

import (
	"context"
	"fmt"
	"time"

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

//TimeBeforeCond TimeBeforeCond
type TimeBeforeCond struct {
	Before int64 `bson:"$lt"`
}

//DeleteCond DeleteCond
type DeleteCond struct {
	beforeCond TimeBeforeCond `bson:"timePoint.startTime"`
}

func main() {
	var (
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
		delCond    *DeleteCond
		delResult  *mongo.DeleteResult
		err        error
	)

	if client, err = mongo.Connect(context.TODO(), "mongodb://47.102.115.229:27017", options.Client().SetConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	database = client.Database("cron")

	collection = database.Collection("log")

	//删除开始时间早于当前时间的所有日志
	//delete({"timePoint.startTime":{"$lt":当前时间}})

	delCond = &DeleteCond{beforeCond: TimeBeforeCond{
		Before: time.Now().Unix(),
	}}

	if delResult, err = collection.DeleteMany(context.TODO(), delCond); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("删除的行数:", delResult.DeletedCount)

}
