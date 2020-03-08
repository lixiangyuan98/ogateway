package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"time"
)

var (
	GlobalConf *Config
)

const (
	baseDir = "./conf"
)

type Config struct {
	IP        string      `json:"ip"`
	Port      uint16      `json:"port"`
	VideoConf VideoConfig `json:"video"`
}

type VideoConfig struct {
	VideoServerTimeout time.Duration `json:"video_server_timeout"`
	GetInfoInterval    time.Duration `json:"get_info_interval"`
}

func readConfFile(filename string, conf interface{}) error {
	b, err := ioutil.ReadFile(path.Join(baseDir, filename))
	if err != nil {
		log.Printf("read configuration file %v error: %v\n", filename, err)
		return err
	}
	if err := json.Unmarshal(b, conf); err != nil {
		log.Printf("resolve configuration file %v error: %v\n", filename, err)
		return err
	}
	log.Printf("resolve configuration file %v successfully\n", filename)
	return nil
}

func Init() {
	log.Println("init configuration")
	GlobalConf = &Config{}
	readConfFile("conf.json", GlobalConf)
	GlobalConf.VideoConf.VideoServerTimeout = GlobalConf.VideoConf.VideoServerTimeout * time.Second
	GlobalConf.VideoConf.GetInfoInterval = GlobalConf.VideoConf.GetInfoInterval * time.Second
	initUser("user_conf.json")
	initVideo("video_conf.json")
}
