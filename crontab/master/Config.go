package master

import (
	"encoding/json"
	"io/ioutil"
)

//Config 配置
type Config struct {
	ApiPort               int      `json:"apiPort"`
	ApiReadTimeOut        int      `json:"apiReadTimeout"`
	ApiWriteTimeOut       int      `json:"apiWriteTimeout"`
	EtcdEndpoints         []string `json:"etcdEndpoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	WebRoot               string   `json:"webroot"`
	MongodbURI            string   `json:"mongodbURI"`
	MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
}

var (
	G_config *Config
)

//InitConfig 初始化配置文件
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	//1.读配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	//2.json反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	G_config = &conf

	return
}
