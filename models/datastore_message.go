package models

import (
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"
)

// globalMessage Global message key-value store object
var globalMessage *DatastoreMessage

// DatastoreMessage is adapter between actual datastore and datastore client
type DatastoreMessage struct {
	store Datastore
}

// NewDatastoreMessage create DatastoreTopic object
func NewDatastoreMessage(cfg *Config) (*DatastoreMessage, error) {
	d, err := LoadDatastore(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load datastore")
	}
	return &DatastoreMessage{
		store: d,
	}, nil
}

// InitDatastoreMessage initialize global datastore object
func InitDatastoreMessage() error {
	d, err := NewDatastoreMessage(globalConfig)
	if err != nil {
		return err
	}
	globalMessage = d
	return nil
}

// decodeRawMessage return Message from encode raw data
func decodeRawMessage(r interface{}) (*Message, error) {
	switch a := r.(type) {
	case []byte:
		return decodeGobMessage(a)
	default:
		return nil, ErrNotMatchTypeMessage
	}
}

// decodeGobMessage return Message from gob encode data
func decodeGobMessage(e []byte) (*Message, error) {
	var res *Message
	buf := bytes.NewReader(e)
	if err := gob.NewDecoder(buf).Decode(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *DatastoreMessage) Get(key string) (*Message, error) {
	v, err := d.store.Get(key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNotFoundEntry
	}
	return decodeRawMessage(v)
}

func (d *DatastoreMessage) Set(m *Message) error {
	v, err := EncodeGob(m)
	if err != nil {
		return err
	}
	return d.store.Set(m.ID, v)
}

func (d *DatastoreMessage) Delete(key string) error {
	return d.store.Delete(key)
}
