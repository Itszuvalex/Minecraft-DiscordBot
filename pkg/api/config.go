package api // "github.com/itszuvalex/mcdiscord/pkg/api"

import "encoding/json"

type ConfigReadHandler func(data json.RawMessage) error
type ConfigWriteHandler func() (json.RawMessage, error)

type IConfig interface {
	AddReadHandler(key string, handler ConfigReadHandler)
	AddWriteHandler(key string, handler ConfigWriteHandler)
	Read() error
	Write() error
}
