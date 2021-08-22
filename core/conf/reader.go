// pakcage conf provide config helpers for ~/.zcn/config.yaml, ～/.zcn/network.yaml and ~/.zcn/wallet.json

package conf

import (
	"encoding/json"
)

// Reader a config reader
type Reader interface {
	GetString(key string) string
	GetInt(key string) int
	GetStringSlice(key string) []string
}

// JSONReader read config from json
type JSONReader struct {
	items map[string]json.RawMessage
}

func (r *JSONReader) getRawMessage(key string) (json.RawMessage, bool) {
	if r == nil || r.items == nil {
		return nil, false
	}
	v, ok := r.items[key]
	if !ok {
		return nil, false
	}

	return v, true

}

// GetString read string from key
func (r *JSONReader) GetString(key string) string {
	v, ok := r.getRawMessage(key)
	if !ok {
		return ""
	}

	var s string
	err := json.Unmarshal(v, &s)
	if err != nil {
		return ""
	}

	return s
}

// GetInt read int from key
func (r *JSONReader) GetInt(key string) int {
	v, ok := r.getRawMessage(key)
	if !ok {
		return 0
	}

	var i int

	err := json.Unmarshal(v, &i)
	if err != nil {
		return 0
	}

	return i
}

// GetStringSlice get string slice from key
func (r *JSONReader) GetStringSlice(key string) []string {

	v, ok := r.getRawMessage(key)
	if !ok {
		return nil
	}

	var list []string

	err := json.Unmarshal(v, &list)
	if err != nil {
		return nil
	}

	return list
}

// NewReaderFromJSON create a JSONReader from json string
func NewReaderFromJSON(data string) (Reader, error) {

	var items map[string]json.RawMessage

	err := json.Unmarshal([]byte(data), &items)

	if err != nil {
		return nil, err
	}

	return &JSONReader{
		items: items,
	}, nil
}
