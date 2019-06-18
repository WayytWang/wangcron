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
		putResp *clientv3.PutResponse
		err     error
	)

	config = clientv3.Config{
		Endpoints:   []string{"47.102.115.229:2379"},
		DialTimeout: 5 * time.Second,
	}

	client, err = clientv3.New(config)
	if err != nil {
		fmt.Printf("connect err : %v \n", err)
		return
	}

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	//clientv3.WithPrevKV() 带前一个kv
	putResp, err = kv.Put(context.TODO(), "/cron/jobs/job1", "hello", clientv3.WithPrevKV())

	if err != nil {
		fmt.Printf("put err : %v \n", err)
	} else {
		fmt.Println("Revision:", putResp.Header.Revision)
		if putResp.PrevKv != nil {
			fmt.Println("PrevValue:", string(putResp.PrevKv.Value))
		}
	}

}
