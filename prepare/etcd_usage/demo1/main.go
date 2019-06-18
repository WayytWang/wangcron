package main

import (
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
	)

	//客户端配置
	config = clientv3.Config{
		Endpoints:   []string{"47.102.115.229:2379"},
		DialTimeout: 5 * time.Second,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("client:%+v", client)
}
