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

	getResp, err = kv.Get(context.TODO(), "cron/jobs/job1")
	if err != nil {
		fmt.Printf("Get err : %v", err)
	} else {
		fmt.Println(getResp.Kvs, getResp.Count)
	}

}
