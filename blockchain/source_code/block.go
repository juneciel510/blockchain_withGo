package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// Block keeps block information
type Block struct {
	Timestamp     int64          // the block creation timestamp
	Transactions  []*Transaction // The block transactions
	PrevBlockHash []byte         // the hash of the previous block
	Hash          []byte         // the hash of the block
	Nonce         int            // the nonce of the block
}

// NewBlock creates and returns a non-mined Block
func NewBlock(timestamp int64, transactions []*Transaction, prevBlockHash []byte) *Block {
	
	block := &Block{Timestamp:timestamp, Transactions: transactions, PrevBlockHash:prevBlockHash}
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(timestamp int64, tx *Transaction) *Block {
	return NewBlock(timestamp, []*Transaction{tx}, nil)
}

// Mine calculates and sets the block hash and nonce.
func (b *Block) Mine() {
	pow:=NewProofOfWork(b)
	nonce, hash:= pow.Run()
	b.Nonce =nonce
	b.Hash = hash
}

// HashTransactions returns a hash of the transactions in the block
// This function iterates over all transactions in a block, serialize them
// and make a merkle tree of it.
// It return the merkle root hash.
func (b *Block) HashTransactions() []byte {
	var merkleRoot []byte
	var parentHash [][]byte
	childHash:= [][]byte{}
	//take the hash value of all transactions as leaves of the merkle Tree
	for _, tx:= range b.Transactions{
		h:=sha256.New()
		h.Write(tx.Serialize())
		childHash=append(childHash,h.Sum(nil))
	}
	//append left node and right node, 
	//take the hash values of the appended value as parent node
	//repeat until one node left
	for{
		if len(childHash)>1{
			if len(childHash)%2 != 0{
				childHash=append(childHash,childHash[len(b.Transactions)-1])
			}
			parentHash=[][]byte{}
			for i := 0; i < len(childHash); {
				appendLeftRight:=append(childHash[i][:],childHash[i+1][:]...)
				h:=sha256.New()
				h.Write(appendLeftRight)
				parentHash=append(parentHash,h.Sum(nil))
				i=i+2
			}
			childHash=parentHash
		}else{
			merkleRoot= childHash[0]
			return merkleRoot[:]
		}
	}
}

// FindTransaction finds a transaction by its ID
func (b *Block) FindTransaction(ID []byte) (*Transaction, error) {
	for _, tx := range b.Transactions {
		if bytes.Equal(tx.ID,ID){
			return tx, nil
		}		
	}
	return nil, ErrTxNotFound
}

func (b *Block) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("============ Block %x ============", b.Hash))
	lines = append(lines, fmt.Sprintf("Prev. hash: %x", b.PrevBlockHash))
	lines = append(lines, fmt.Sprintf("Timestamp: %v", time.Unix(b.Timestamp, 0)))
	lines = append(lines, fmt.Sprintf("Nonce: %d", b.Nonce))
	lines = append(lines, fmt.Sprintf("Transactions:"))
	for i, tx := range b.Transactions {
		lines = append(lines, fmt.Sprintf("%d: %x", i, tx.ID))
	}
	return strings.Join(lines, "\n")
}

func (b *Block) StringDetail() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("============ Block %x ============", b.Hash))
	lines = append(lines, fmt.Sprintf("Prev. hash: %x", b.PrevBlockHash))
	lines = append(lines, fmt.Sprintf("Timestamp: %v", time.Unix(b.Timestamp, 0)))
	lines = append(lines, fmt.Sprintf("Nonce: %d", b.Nonce))
	lines = append(lines, fmt.Sprintf("Transactions:"))
	for i, tx := range b.Transactions {
		lines = append(lines, fmt.Sprintf("%d: %x", i, tx.ID))
		for j,input := range tx.Vin{
			lines = append(lines,fmt.Sprintf("input %d: ", j) )
			lines = append(lines,fmt.Sprintf("Txid: %x ", input.Txid) )
			lines = append(lines,fmt.Sprintf("OutIdx: %d ", input.OutIdx) )
		}
		for k,output := range tx.Vout{
			lines = append(lines,fmt.Sprintf("output %d: ", k))
			lines = append(lines,fmt.Sprintf("value: %d ", output.Value) )
		}
	}
	return strings.Join(lines, "\n")
}
