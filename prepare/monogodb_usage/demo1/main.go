package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo/options"

	"github.com/mongodb/mongo-go-driver/mongo"
)

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
	)

	//1.建立连接
	if client, err = mongo.Connect(context.TODO(), "mongodb://47.102.115.229:27017", options.Client().SetConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	//2.选择数据库my_db
	database = client.Database("my_db")

	//3.选择表my_collection  //会自动建表
	collection = database.Collection("my_collection")

	fmt.Println(collection)
}
