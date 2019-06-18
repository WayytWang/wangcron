package main

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		getResp *clientv3.GetResponse
		err     error
	)

	config = clientv3.Config{
		DialTimeout: 5 * time.Second,
		Endpoints:   []string{"47.102.115.229:2379"},
	}

	client, err = clientv3.New(config)
	if err != nil {
		fmt.Printf("client connect err : %v", err)
	}

	kv = clientv3.NewKV(client)

	//读取 cron/jobs/为前缀的所有key
	getResp, err = kv.Get(context.TODO(), "cron/jobs/", clientv3.WithPrefix())

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getResp.Kvs)
	}

}
