package database

import (
	"os"

	"github.com/boltdb/bolt"
)

var ListenPort string

type BucketType string

// TODO: 其他的(网络相关的)类型
const (
	BlockBucket BucketType = "blocks"
)

type BlockchainDB struct {
	ListenPort string
}

func New() *BlockchainDB {
	return &BlockchainDB{ListenPort: ListenPort}
}

func IsBlotExist(nodeID string) bool {
	var DBFileName = "blockchain_" + nodeID + ".db"
	_, err := os.Stat(DBFileName)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func IsBucketExist(bd *BlockchainDB, bt BucketType) bool {
	var isBucketExist bool

	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bt))
		if bucket == nil {
			isBucketExist = false
		} else {
			isBucketExist = true
		}
		return nil
	})
	if err != nil {
		panic("datebase IsBucketExist err:" + err.Error())
	}

	err = db.Close()
	if err != nil {
		panic("db close err :" + err.Error())
	}
	return isBucketExist
}
