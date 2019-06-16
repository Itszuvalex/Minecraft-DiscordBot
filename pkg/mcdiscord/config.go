package mcdiscord // "github.com/itszuvalex/mcdiscord/pkg/mcdiscord"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/itszuvalex/mcdiscord/pkg/api"
)

type configFile struct {
	File          string
	ReadHandlers  map[string]api.ConfigReadHandler
	WriteHandlers map[string]api.ConfigWriteHandler
}

func NewConfig(file string) api.IConfig {
	return &configFile{
		File:          file,
		ReadHandlers:  make(map[string]api.ConfigReadHandler),
		WriteHandlers: make(map[string]api.ConfigWriteHandler),
	}
}

func (cfile *configFile) AddReadHandler(key string, handler api.ConfigReadHandler) {
	cfile.ReadHandlers[key] = handler
}

func (cfile *configFile) AddWriteHandler(key string, handler api.ConfigWriteHandler) {
	cfile.WriteHandlers[key] = handler
}

func (cfile *configFile) Read() error {
	_, err := os.Stat(cfile.File)
	if os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(cfile.File)
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
		if handler, ok := cfile.ReadHandlers[key]; ok {
			err = handler(innerFields[key])
			if err != nil {
				fmt.Println("Error when handling json:"+key+",", err)
			}
		}
	}

	return nil
}

func (cfile *configFile) Write() error {
	file, err := os.Create(cfile.File)
	if err != nil {
		return err
	}
	defer file.Close()

	innerFields := make(map[string]json.RawMessage)
	for key := range cfile.WriteHandlers {
		json, err := cfile.WriteHandlers[key]()
		if err != nil {
			fmt.Println("Error when writing json:"+key+",", err)
		}
		innerFields[key] = json
	}
	data, err := json.MarshalIndent(innerFields, "", "\t")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
