package main

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		config         clientv3.Config
		client         *clientv3.Client
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseID        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		keepResp       *clientv3.LeaseKeepAliveResponse
		ctx            context.Context
		cancelFunc     context.CancelFunc
		kv             clientv3.KV
		txn            clientv3.Txn
		txnResp        *clientv3.TxnResponse
		err            error
	)

	config = clientv3.Config{
		Endpoints:   []string{"47.102.115.229:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//lease实现锁自动过期
	//op操作
	//txn事务: if else then

	//1.上锁(创建租约,自动续租,拿着租约去抢占一个key)
	//申请一个lease(租约)
	lease = clientv3.NewLease(client)

	//5秒租约
	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println(err)
	}

	leaseID = leaseGrantResp.ID

	//准备一个用于取消自动续租的context
	ctx, cancelFunc = context.WithCancel(context.TODO())

	//确保函数退出后,自动续约会停止
	defer cancelFunc()
	//确保租约在程序结束后立即终止
	defer lease.Revoke(context.TODO(), leaseID)

	if keepRespChan, err = lease.KeepAlive(ctx, leaseID); err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					fmt.Println("租约已经失效")
					goto END
				} else {
					fmt.Println("收到自动续租应答:", keepResp.ID)
				}
			}
		}
	END:
	}()

	//if 不存在key,then 设置它,else 抢锁失败
	kv = clientv3.NewKV(client)

	//创建事务
	txn = kv.Txn(context.TODO())

	//如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job9", "XXX", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet("/cron/lock/job9"))

	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	//判断是否抢到了锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//2.处理业务
	//在锁内

	fmt.Println("处理任务")
	time.Sleep(5 * time.Second)

	//3.释放锁(取消自动续租,释放租约)
	//defer 会把租约释放掉,关联的kv就被删除了
}
