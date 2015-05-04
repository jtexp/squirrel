package main

import (
	"encoding/json"
	"os"
)

type masterConfig struct {
	ListenAddress             string
	Network                   string
	MobilityManager           string
	MobilityManagerParameters map[string]interface{}
	September                 string
	SeptemberParameters       map[string]interface{}
}

func parseMasterConfig(filename string) (config *masterConfig, err error) {
	config = &masterConfig{}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	err = json.NewDecoder(file).Decode(config)
	return
}