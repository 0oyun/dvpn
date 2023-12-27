package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Transactions struct {
	Txs []Transaction
}

type TransactionType string

type Transaction struct {
	TxType   TransactionType
	Data     []byte
	AddrFrom string
}

func (t *Transactions) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(t)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (t *Transactions) Deserialize(d []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(t)
	if err != nil {
		fmt.Println("deserialize transactions err:", err)
		panic(err)
	}
}
