package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/big"
	"time"
)

type Block struct {
	//上一个区块的hash
	PreHash []byte
	//数据data
	Transactions []Transaction
	//时间戳
	TimeStamp int64
	//区块高度
	Height int
	//本区块hash
	Hash []byte
}

func mineBlock(transaction []Transaction, preHash []byte, height int) (*Block, error) {
	timeStamp := time.Now().Unix()
	block := Block{preHash, transaction, timeStamp, height, nil}
	block.Hash = calculateBlockHash(&block)

	fmt.Println("hash verify : ", block.Verify())
	fmt.Printf("new block generate successfully! Now, block height is: %d\n", block.Height)
	return &block, nil
}

func newGenesisBlock(transaction []Transaction) *Block {
	preHash := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	genesisBlock, err := mineBlock(transaction, preHash, 1)
	if err != nil {
		fmt.Printf("generate genesis block failed:%v\n", err)
		panic(err)
	}
	return genesisBlock
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (v *Block) Deserialize(d []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(v)
	if err != nil {
		panic(err)
	}
}

func (b *Block) Verify() bool {
	var hashInt big.Int
	hashInt.SetBytes(b.Hash)
	var hashCalc big.Int
	hashCalc.SetBytes(calculateBlockHash(b))
	if hashInt.Cmp(&hashCalc) != 0 {
		return false
	}
	return true
}

func isGenesisBlock(block *Block) bool {
	var hashInt big.Int
	hashInt.SetBytes(block.PreHash)
	if big.NewInt(0).Cmp(&hashInt) == 0 {
		return true
	}
	return false
}

func calculateBlockHash(block *Block) []byte {
	blockBytes := block.getBytes()
	hashBytes := sha256.Sum256(blockBytes)
	return hashBytes[:]
}

func (b *Block) getBytes() []byte {
	blockBytes := []byte{}
	blockBytes = append(blockBytes, b.PreHash...)
	blockBytes = append(blockBytes, Int64ToBytes(b.TimeStamp)...)
	blockBytes = append(blockBytes, Int64ToBytes(int64(b.Height))...)
	for _, v := range b.Transactions {
		blockBytes = append(blockBytes, v.getTransBytes()...)
	}
	return blockBytes
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
