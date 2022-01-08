package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
)

var (
	ErrTxNotFound    = errors.New("transaction not found")
	ErrNoValidTx     = errors.New("there is no valid transaction")
	ErrBlockNotFound = errors.New("block not found")
	ErrInvalidBlock  = errors.New("block is not valid")
)

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	blocks []*Block
}

// NewBlockchain creates a new blockchain with genesis Block
func NewBlockchain(address string) (*Blockchain, error) {
	timeStamp:=time.Now().Unix()
	genesisTransaction,err:=NewCoinbaseTX(address,"")
	genesisBlock:=NewGenesisBlock(timeStamp,genesisTransaction)
	genesisBlock.Mine()
	return &Blockchain{blocks:[]*Block{genesisBlock}},err
}

// addBlock saves the block into the blockchain
func (bc *Blockchain) addBlock(block *Block) error {
	if bc.ValidateBlock(block) { 
		bc.blocks = append(bc.blocks,block)
		return nil
	}
	return ErrInvalidBlock
}

// GetGenesisBlock returns the Genesis Block
func (bc Blockchain) GetGenesisBlock() *Block {
	return bc.blocks[0]
}

// CurrentBlock returns the last block
func (bc Blockchain) CurrentBlock() *Block {
	return bc.blocks[len(bc.blocks)-1]
}

// GetBlock returns the block of a given hash
func (bc Blockchain) GetBlock(hash []byte) (*Block, error) {
	for _, block := range bc.blocks {
		if bytes.Equal(block.Hash, hash)  {
			return block, nil
		}
	}
	return nil, ErrBlockNotFound
}

// ValidateBlock validates the block before adding it to the blockchain
func (bc *Blockchain) ValidateBlock(block *Block) bool {
	pow :=NewProofOfWork(block)
	if block!=nil&& len(block.Transactions)!=0 && pow.Validate()  { 
		return true
	}
	return false
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	// 1) Verify the existence of transactions inputs and discard invalid transactions that make reference to unknown inputs
	// 2) Add a block if there is a list of valid transactions
	validTx:=[]*Transaction{}
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx){
			validTx=append(validTx,tx)
		}
	}
	if len(validTx)==0{
		return nil, ErrNoValidTx
	}

	//newblock with the valid transactions, mine and add to blockchain
	timestamp:=time.Now().Unix()
	currentBlock :=bc.CurrentBlock()
	prevBlockHash:=currentBlock.Hash
	block:=NewBlock(timestamp,validTx,prevBlockHash)
	block.Mine()
	bc.blocks= append(bc.blocks,block)
	return block, nil
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	
	//1)extract all unspent outputs to build UTXOset according to the blockchain state.
	u:=bc.FindUTXOSet()
	//2)
	//check if it is a coinbase transaction, return true
	if tx.IsCoinbase(){			
		return true
	}
	//check if it is not in UTXOset, return false
	for _, txInput := range tx.Vin{

		txInputTxidString:=fmt.Sprintf("%x", txInput.Txid)
		if reflect.DeepEqual(u[txInputTxidString][txInput.OutIdx], TXOutput{}){		
			fmt.Println("-----not in UTXOset")
			return false
		}
	}
	//3)verify the signature of the given transaction
	txCopy:=tx.TrimmedCopy()
	curveForSign:= elliptic.P256()

	for _,input:=range tx.Vin{
		//recovering the R and S byte fields of the Signature
		length:=len(input.Signature)
		r:= new(big.Int).SetBytes(input.Signature[:length/2])
		s:= new(big.Int).SetBytes(input.Signature[length/2:])
		//recovering X and Y fields of the PubKey from the inputs of tx
		length=len(input.PubKey)
		x:= new(big.Int).SetBytes(input.PubKey[:length/2])
		y:= new(big.Int).SetBytes(input.PubKey[length/2:])

		ecdsaPubKey:=ecdsa.PublicKey{curveForSign, x, y}
		if !ecdsa.Verify(&ecdsaPubKey, txCopy.Serialize(), r, s) {
			fmt.Println("-----signature not correct")
			return false
		}	
	}
	
	return true
}

// FindTransaction finds a transaction by its ID in the whole blockchain
func (bc Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	var tx *Transaction
	var err error
	for _,block := range bc.blocks {
		tx, err =block.FindTransaction(ID)
		if err==nil{
			return tx, nil
		}
	}
	return nil, ErrTxNotFound	
}

// FindUTXOSet finds and returns all unspent transaction outputs
func (bc Blockchain) FindUTXOSet() UTXOSet {
	uTXOSet:=make(UTXOSet)
	for _, block := range bc.blocks {
		uTXOSet.Update(block.Transactions)
	}
	return uTXOSet
}

// GetInputTXsOf returns a map index by the ID,
// of all transactions used as inputs in the given transaction
func (bc *Blockchain) GetInputTXsOf(tx *Transaction) (map[string]*Transaction, error) {
	txMap := make(map[string]*Transaction)
	for _,input := range tx.Vin{
		tx,err :=bc.FindTransaction(input.Txid)
		if err==nil{
			txMap[Bytes2Hex(input.Txid)]=tx
		}
	}
	if len(txMap)==0{
		return nil,ErrTxNotFound
	}
	return txMap, nil
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs,errFindPrevTXs:=bc.GetInputTXsOf(tx)
	if errFindPrevTXs != nil{
		return errFindPrevTXs
	}
	errSign:=tx.Sign(privKey, prevTXs)
	if errSign != nil{
		return errSign
	}
	return nil
}

func (bc Blockchain) String() string {
	var lines []string
	for _, block := range bc.blocks {
		lines = append(lines, fmt.Sprintf("%v", block))
	}
	return strings.Join(lines, "\n")
}
