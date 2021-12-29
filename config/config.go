package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Location  string `json:"location"`
	IPAddress string `json:"ip_address"`
}

func GetConfig() Config {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fileBytes, err := os.ReadFile(fmt.Sprintf("%s/config.json", wd))
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := json.Unmarshal(fileBytes, &cfg); err != nil {
		panic(err)
	}

	return cfg
}
