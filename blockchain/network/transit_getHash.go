package network

import (
	"bytes"
	"encoding/gob"
)

type getHash struct {
	Height   int
	AddrFrom string
}

func (v getHash) serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (v *getHash) deserialize(d []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(v)
	if err != nil {
		panic(err)
	}
}
