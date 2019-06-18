package worker

import (
	"context"
	"net"
	"time"
	"wangcron/crontab/common"

	"go.etcd.io/etcd/clientv3"
)

//Register 注册节点到etcd : /cron/worker/IP地址
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string //本机IP
}

var (
	G_register *Register
)

//获取本机网卡IP:第一个非lo的网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIPNet bool
	)
	//获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	//取第一个非lo的网卡IP
	for _, addr = range addrs {
		if ipNet, isIPNet = addr.(*net.IPNet); isIPNet && !ipNet.IP.IsLoopback() {
			//跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}

	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

//注册到/cron/workers/IP,并自动续租
func (register *Register) keepOnline() {
	var (
		regKey            string
		leaseGrantResp    *clientv3.LeaseGrantResponse
		err               error
		keepAliveChan     <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResponse *clientv3.LeaseKeepAliveResponse
		cancleCtx         context.Context
		cancleFunc        context.CancelFunc
	)

	for {

		//注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancleFunc = nil

		//创建租约
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		//自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancleCtx, cancleFunc = context.WithCancel(context.TODO())

		//注册到etcd
		if _, err = register.kv.Put(cancleCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		//处理续租应答
		for {
			select {
			case keepAliveResponse = <-keepAliveChan:
				if keepAliveResponse == nil { //续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancleFunc != nil {
			cancleFunc()
		}
	}

}

//InitRegister 初始化服务注册
func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIP string
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	if localIP, err = getLocalIP(); err != nil {
		return
	}

	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}

	// 服务注册
	go G_register.keepOnline()

	return
}
