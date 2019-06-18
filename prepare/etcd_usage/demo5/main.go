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
		err     error
		delResp *clientv3.DeleteResponse
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

	delResp, err = kv.Delete(context.TODO(), "cron/jobs/job1")
	if err != nil {
		fmt.Printf("Delete err : %v", err)
		return
	}

	//delResp.PrevKvs 被删除之前的value
	if len(delResp.PrevKvs) != 0 {
		for _, kvpair := range delResp.PrevKvs {
			fmt.Println("删除了:", string(kvpair.Key), string(kvpair.Value))
		}
	}

}
