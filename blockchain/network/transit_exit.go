package network

import (
	"bytes"
	"encoding/gob"
)

type Exit struct {
	Addrfrom string
}

func (v Exit) serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (v *Exit) deserialize(d []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(v)
	if err != nil {
		panic(err)
	}
}
