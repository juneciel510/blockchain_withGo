package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrNoFunds         = errors.New("not enough funds")
	ErrTxInputNotFound = errors.New("transaction input not found")
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data=RandomString(10)
	}
	tXInput :=TXInput{OutIdx:-1,PubKey:[]byte(data)}
	txOutput:=TXOutput{Value:BlockReward,PubKeyHash:GetPubKeyHashFromAddress(to)}
	tx:=&Transaction{ Vin:[]TXInput{tXInput}, Vout:[]TXOutput{txOutput}}
	tx.ID=tx.Hash()
	return tx,nil
}



// NewUTXOTransaction creates a new UTXO transaction
// NOTE: The returned tx is NOT signed!
func NewUTXOTransaction(pubKey []byte, to string, amount int, utxos UTXOSet) (*Transaction, error) {
	// 1) Find valid spendable outputs and the current balance of the sender
	// 2) The sender has sufficient funds? If not return the error:
	// "Not enough funds"
	// 3) Build a list of inputs based on the current valid outputs
	// 4) Build a list of new outputs, creating a "change" output if necessary
	// 5) Create a new transaction with the input and output list.
	pubKeyHashSender:=HashPubKey(pubKey)
	accumulatedBalance, spendableOutputs:=utxos.FindSpendableOutputs(pubKeyHashSender,amount)
	if accumulatedBalance<amount {
		return nil,ErrNoFunds
	}

	vin:=[]TXInput{}
	for id,val:=range spendableOutputs{
		for _,outIdx:=range val{
			tXInput:=TXInput{Txid:Hex2Bytes(id),OutIdx:outIdx,PubKey:pubKey}
			vin= append(vin,tXInput)
		}

	}
	PubKeyHashRecepient:=GetPubKeyHashFromAddress(to)
	tXOutputTo:=TXOutput{Value:amount,PubKeyHash:PubKeyHashRecepient}
	tXOutputFrom:=TXOutput{Value:accumulatedBalance-amount,PubKeyHash:pubKeyHashSender}
	vout:=[]TXOutput{tXOutputTo,tXOutputFrom}
	tx:=&Transaction{Vin:vin,Vout:vout}
	tx.ID=tx.Hash()
	return tx,nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	if tx.Vin[0].OutIdx==-1{
		return true
	}
	return false
}

// Equals checks if the given transaction ID matches the ID of tx
func (tx Transaction) Equals(ID []byte) bool {
	if bytes.Equal(tx.ID, ID){
		return true
	}
	return false
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var serializedTransaction bytes.Buffer 
	enc := gob.NewEncoder(&serializedTransaction) 
	enc.Encode(tx)
	return serializedTransaction.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var txCopy Transaction
	txCopy=*tx
	txCopy.ID=nil
	var serializedTransaction bytes.Buffer 
	enc := gob.NewEncoder(&serializedTransaction) 
	enc.Encode(txCopy)
	h:=sha256.New()
	h.Write(serializedTransaction.Bytes())
	return h.Sum(nil)
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx Transaction) TrimmedCopy() Transaction {
	// The fields Signature and PubKey of the input need to be nil
	// since they are not included in signature.
	vin:=[]TXInput{}
	for _,input:=range tx.Vin{
		input.Signature=nil
		input.PubKey=nil
		vin = append(vin, input)
	}
	txCopy:=tx
	txCopy.Vin=vin
	return txCopy
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]*Transaction) error {
	//1) coinbase transactions are not signed.
	if tx.IsCoinbase(){ return nil }
	//2) Throw a Panic in case of any prevTXs (used inputs) didn't exists
	for _, input := range tx.Vin{
		//if the input txID is not in the key of prevTXs, return error
		if prevTXs[Bytes2Hex(input.Txid)]==nil{
			return ErrTxInputNotFound
		}
		//check if the public key field of input and output match or not
		exsitFlag := false
		for _, output := range prevTXs[Bytes2Hex(input.Txid)].Vout{
			if(bytes.Equal(HashPubKey(input.PubKey),output.PubKeyHash)){
				exsitFlag=true
				break
			}
		}
		if !exsitFlag{
			return ErrTxInputNotFound
		} 
	}
	//3) Create a copy of the transaction to be signed
	txCopy:=tx.TrimmedCopy()
	//4) Sign all the previous TXInputs of the transaction tx using the
	// copy as the payload (serialized) to be signed in the ecdsa.Sig

	//calculate the signature
	payload:=txCopy.Serialize()
	r, s, serr := ecdsa.Sign(rand.Reader, &privKey, payload)
	if serr != nil {
		fmt.Println(serr)
		panic(serr)
	}
	signature :=append(r.Bytes(),s.Bytes()...)
	//assign the signature to each input of the transaction
	vin:=[]TXInput{}
	for _,input:=range tx.Vin{
		input.Signature=signature
		vin = append(vin, input)
	}
	tx.Vin=vin
	return nil	
}

// Verify verifies signatures of Transaction inputs
func (tx Transaction) Verify(prevTXs map[string]*Transaction) bool {
	//1) coinbase transactions are not signed.
	if tx.IsCoinbase(){ return true }
	//2) Throw a Panic in case of any prevTXs (used inputs) didn't exists
	for _, input := range tx.Vin{
		//if the input txID is not in the key of prevTXs, return error
		if prevTXs[Bytes2Hex(input.Txid)]==nil{
			return false
		}
		//check if the public key field of input and output match or not
		exsitFlag := false
		for _, output := range prevTXs[Bytes2Hex(input.Txid)].Vout{
			if(bytes.Equal(HashPubKey(input.PubKey),output.PubKeyHash)){
				exsitFlag=true
				break
			}
		}
		if !exsitFlag{
			return false
			//panic(ErrTxInputNotFound)
		} 
	}
	//3) Create the same copy of the transaction that was signed
	// and get the curve used for sign: P256
	txCopy:=tx.TrimmedCopy()
	curveForSign:= elliptic.P256()
	// 4) Doing the opposite operation of the signing, perform the
	// verification of the signature,
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
			return false
		}	
	}
	return true
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x :", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       OutIdx:    %d", input.OutIdx))
		lines = append(lines, fmt.Sprintf("       PubKey: %x", input.PubKey))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", HashPubKey(input.PubKey)))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
