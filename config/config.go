package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Config struct {
	Daemon struct {
		Listen string `json:"listen"`
	}
	Database struct {
		Pg struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			DBName   string `json:"db_name"`
		}
	}
	JWTKey    []byte `json:"jwt_key"`
	AppServer string `json:"app_server"`
	Logger    struct {
		Debug string `json:"debug"`
		Error string `json:"error"`
		Info  string `json:"info"`
	}
	Queue struct {
		Size               int      `json:"size"`
		WorkerTimeInterval int      `json:"worker_time_interval"`
		Workers            []string `json:"workers"`
	}
}

var config *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		config = loadConfig()
	})
	return config
}

func loadConfig() *Config {
	cfg := &Config{}
	configFile, err := os.Open("/etc/jackiechat/jackiechat.conf")
	if err != nil {
		log.Fatal("Can't load config: ", err.Error())
	}

	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&cfg)

	return cfg
}
