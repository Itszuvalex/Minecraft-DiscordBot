package mcdiscord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ConfigReadHandler func(data json.RawMessage) error
type ConfigWriteHandler func() (json.RawMessage, error)

type Config struct {
	File          string
	ReadHandlers  map[string]ConfigReadHandler
	WriteHandlers map[string]ConfigWriteHandler
}

func NewConfig(file string) *Config {
	return &Config{
		File:          file,
		ReadHandlers:  make(map[string]ConfigReadHandler),
		WriteHandlers: make(map[string]ConfigWriteHandler),
	}
}

func (config *Config) AddReadHandler(key string, handler ConfigReadHandler) {
	config.ReadHandlers[key] = handler
}

func (config *Config) AddWriteHandler(key string, handler ConfigWriteHandler) {
	config.WriteHandlers[key] = handler
}

func (config *Config) Read() error {
	_, err := os.Stat(config.File)
	if os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(config.File)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var innerFields map[string]json.RawMessage
	err = json.Unmarshal(data, &innerFields)
	if err != nil {
		return err
	}

	for key := range innerFields {
		if handler, ok := config.ReadHandlers[key]; ok {
			err = handler(innerFields[key])
			if err != nil {
				fmt.Println("Error when handling json:"+key+",", err)
			}
		}
	}

	return nil
}

func (config *Config) Write() error {
	file, err := os.Create(config.File)
	if err != nil {
		return err
	}
	defer file.Close()

	innerFields := make(map[string]json.RawMessage)
	for key := range config.WriteHandlers {
		json, err := config.WriteHandlers[key]()
		if err != nil {
			fmt.Println("Error when writing json:"+key+",", err)
		}
		innerFields[key] = json
	}
	data, err := json.Marshal(innerFields)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
