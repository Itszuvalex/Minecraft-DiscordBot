package mcdiscord

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	MessageType string = "msg"
	StatusType  string = "status"
)

type Header struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Message struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Sender    string `json:"sender"`
}

type JsonMessageHandler func(interface{}) error

type JsonHandler struct {
	handlers map[string][]JsonMessageHandler
}

func NewJsonHandler() *JsonHandler {
	handler := new(JsonHandler)
	handler.handlers = make(map[string][]JsonMessageHandler)
	return handler
}

func (jsonhandler *JsonHandler) RegisterHandler(msg string, handler JsonMessageHandler) {
	jsonhandler.handlers[msg] = append(jsonhandler.handlers[msg], handler)
}

func (jsonhandler *JsonHandler) HandleJson(header Header) error {
	switch header.Type {
	case MessageType:
		var message Message
		if err := UnmarshallMessage(&message, header.Data); err != nil {
			fmt.Println("Error unmarshalling Message, ", err)
			return err
		}
		for _, handler := range jsonhandler.handlers[header.Type] {
			if err := handler(&message); err != nil {
				fmt.Println("Error calling Message handler, ", err)
			}
		}
	case StatusType:
		var status McServerData
		if err := UnmarshallStatus(&status, header.Data); err != nil {
			fmt.Println("Error unmarshalling Status, ", err)
			return err
		}
		for _, handler := range jsonhandler.handlers[header.Type] {
			if err := handler(&status); err != nil {
				fmt.Println("Error calling Status handler, ", err)
			}
		}
	default:
		fmt.Println("Error unmarshalling Header, unknown type ", header.Type)
		return errors.New("Error unmarshalling Header, unknown type " + header.Type)
	}
	return nil
}

func UnmarshallStatus(obj interface{}, data json.RawMessage) error {
	serverdata, ok := obj.(*McServerData)
	if !ok {
		fmt.Println("Unmarshall Status passed non *McServerData obj")
		return errors.New("Unmarshall Status passed non *McServerData obj")
	}

	if err := json.Unmarshal(data, serverdata); err != nil {
		fmt.Println("Error unmarshalling McServerData, ", err)
		return err
	}
	return nil
}

func UnmarshallMessage(obj interface{}, data json.RawMessage) error {
	message, ok := obj.(*Message)
	if !ok {
		fmt.Println("Unmarshall Message passed non *Message obj")
		return errors.New("Unmarshall Message passed non *Message obj")
	}

	if err := json.Unmarshal(data, message); err != nil {
		fmt.Println("Error unmarshalling McServerData, ", err)
		return err
	}
	return nil
}

func MarshalMessage(message *Message) ([]byte, error) {
	msgdata, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling Message, ", err)
		return nil, err
	}
	return msgdata, err
}

func MarshallMessageToHeader(message *Message, header *Header) error {
	header.Type = MessageType
	msgdata, err := MarshalMessage(message)
	if err != nil {
		return err
	}
	header.Data = msgdata
	return nil
}

func MarshalMessageInHeader(message *Message) ([]byte, error) {
	var header Header
	err := MarshallMessageToHeader(message, &header)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(&header)
	if err != nil {
		fmt.Println("Error marshalling Message Header, ", err)
		return nil, err
	}
	return data, nil
}

func MarshallStatus(status *McServerData) ([]byte, error) {
	statusdata, err := json.Marshal(status)
	if err != nil {
		fmt.Println("Error marshalling Status, ", err)
		return nil, err
	}
	return statusdata, err
}

func MarshalStatusToHeader(status *McServerData, header *Header) error {
	header.Type = StatusType
	statusdata, err := MarshallStatus(status)
	if err != nil {
		return nil
	}
	header.Data = statusdata
	return nil
}

func MarshalStatusInHeader(status *McServerData) ([]byte, error) {
	var header Header
	err := MarshalStatusToHeader(status, &header)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(&header)
	if err != nil {
		fmt.Println("Error marshalling Status Header, ", err)
		return nil, err
	}
	return data, nil
}
