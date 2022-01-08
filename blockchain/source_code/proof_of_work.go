package main

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64

// TARGETBITS define the mining difficulty
const TARGETBITS = 8

// ProofOfWork represents a block mined with a target difficulty
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds a ProofOfWork
func NewProofOfWork(block *Block) *ProofOfWork {
	targetDifficulty:=new(big.Int).Exp(big.NewInt(2), big.NewInt(256-TARGETBITS), nil)
	return &ProofOfWork{block:block, target:targetDifficulty}
}

// setupHeader prepare the header of the block
func (pow *ProofOfWork) setupHeader() []byte {
	block:=pow.block
	slice:=[][]byte{block.PrevBlockHash,block.HashTransactions(),IntToHex(block.Timestamp),IntToHex(TARGETBITS)}
	header:=[]byte{}
	for _,value := range slice {
		header=append(header,value...)
	}
	return header
}

// addNonce adds a nonce to the header
func addNonce(nonce int, header []byte) []byte {
	nonceBytes:=[]byte(IntToHex(int64(nonce)))
	return append(header, nonceBytes...)
}

// Run performs the proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	header:=pow.setupHeader()
	nonce:=0
	for {
		h:=sha256.New()
		h.Write(addNonce(nonce,header))
		hash:=h.Sum(nil)	
		hashBigInt:= new(big.Int).SetBytes(hash[:])
		if hashBigInt.Cmp(pow.target)==-1{
			return nonce, hash
		}
		nonce++
	}
}

// Validate validates block's Proof-Of-Work
// This function just validates if the block header hash
// is less than the target AND equals to the mined block hash.
func (pow *ProofOfWork) Validate() bool {
	header:=pow.setupHeader()
	h:=sha256.New()
	h.Write(addNonce(pow.block.Nonce,header))
	hash:=h.Sum(nil)
	hashBigInt:= new(big.Int).SetBytes(hash[:])
	if hashBigInt.Cmp(pow.target)==-1 && bytes.Equal(hash,pow.block.Hash){
		return true
	}else{
		return false
	}
}
