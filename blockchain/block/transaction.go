package block

import (
	"bytes"
	"encoding/gob"
)

type TransactionType string

type Transaction struct {
	TxType TransactionType
	Data   []byte
}

func (t *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(t)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (t *Transaction) getTransBytes() []byte {
	transBytes := []byte{}
	transBytes = append(transBytes, []byte(t.TxType)...)
	transBytes = append(transBytes, t.Data...)
	return transBytes
}

func isGenesisTransaction(tss []Transaction) bool {
	if tss != nil {
		if tss[0].TxType == "genesis" {
			return true
		}
	}
	return false
}
