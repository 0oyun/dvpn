package block

import (
	"fmt"
	"math/big"
	"time"

	"github.com/toy-playground/miniDVPN/blockchain/database"
)

type blockchain struct {
	BD *database.BlockchainDB // blot database
}

func NewBlockchain() *blockchain {
	blockchain := blockchain{}
	blockchain.BD = database.New()
	return &blockchain
}

func (bc *blockchain) CreateGenesisBlock(send Sender) {
	txs := []Transaction{{TxType: "genesis", Data: []byte("hello,mini DVPN!")}}
	if len(bc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket)) != 0 {
		fmt.Println("genesis block already exists")
		return
	}
	genesisBlock := newGenesisBlock(txs)
	NewestBlockHeight = 1
	bc.AddBlock(genesisBlock)
	send.SendVersionToPeers(1)
	fmt.Println("genesis block create success!")
}

func (bc *blockchain) CreateTransaction(txtype, data string, send Sender) {
	if len(bc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket)) == 0 {
		fmt.Println("genesis block not exists!")
		return
	}
	fmt.Printf("txtype: %s, data: %s\n", txtype, data)

	send.SendTransToPeers([]Transaction{{TxType: TransactionType(txtype), Data: []byte(data)}})
}

func (bc *blockchain) addBlockchain(transaction []Transaction, send Sender) {
	preBlockbyte := bc.BD.View(bc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), database.BlockBucket)
	preBlock := Block{}
	preBlock.Deserialize(preBlockbyte)
	height := preBlock.Height + 1
	nb, err := mineBlock(transaction, bc.BD.View([]byte(LastBlockHashMapping), database.BlockBucket), height)
	if err != nil {
		fmt.Println(err)
		return
	}
	bc.AddBlock(nb)
	send.SendVersionToPeers(nb.Height)
}

// 发送交易
func (bc *blockchain) SendTx(tss []Transaction, send Sender) {
	bc.addBlockchain(tss, send)
}

// 添加区块信息到数据库，并更新lastHash
func (bc *blockchain) AddBlock(block *Block) {
	bc.BD.Put(block.Hash, block.Serialize(), database.BlockBucket)
	bci := NewBlockchainIterator(bc)
	currentBlock := bci.Next()
	if currentBlock == nil || currentBlock.Height < block.Height {
		bc.BD.Put([]byte(LastBlockHashMapping), block.Hash, database.BlockBucket)
	}
}

func (bc *blockchain) GetLastBlockHeight() int {
	bcl := NewBlockchainIterator(bc)
	lastblock := bcl.Next()
	if lastblock == nil {
		return 0
	}
	return lastblock.Height
}

func (bc *blockchain) GetBlockHashByHeight(height int) []byte {
	bcl := NewBlockchainIterator(bc)
	for {
		currentBlock := bcl.Next()
		if currentBlock == nil {
			return nil
		} else if currentBlock.Height == height {
			return currentBlock.Hash
		} else if isGenesisBlock(currentBlock) {
			return nil
		}
	}
}

func (bc *blockchain) GetBlockByHash(hash []byte) []byte {
	return bc.BD.View(hash, database.BlockBucket)
}

func (bc *blockchain) PrintAllBlockInfo() {
	blcIterator := NewBlockchainIterator(bc)
	for {
		block := blcIterator.Next()
		if block == nil {
			fmt.Println("genesis block not exist!")
			return
		}
		fmt.Println("======================================================================")
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println("------------------------------Txs Data------------------------------")
		for _, v := range block.Transactions {
			fmt.Printf("tx_type:  %s\n", v.TxType)
			fmt.Printf("tx_data:  %s\n", v.Data)
		}
		fmt.Println("--------------------------------------------------------------------")
		fmt.Printf("Timestamp: %s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev Hash: %x\n", block.PreHash)
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	fmt.Println("======================================================================")
}
