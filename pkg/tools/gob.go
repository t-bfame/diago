package tools

import (
	"bytes"
	"encoding/gob"
)

func GobEncode(obj interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GobDecode(e interface{}, data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(e)
	if err != nil {
		return err
	}
	return nil
}
