package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
)


type Account struct {
	Name        string 
	PrivateKey  ecdsa.PrivateKey
	PubKeyBytes []byte
	Address     string
	Balance     int
	Blockchain 	*Blockchain
	AddressMap 	map[string]string
	BlockIn chan *Block
	ChannelMap 	map[string]chan *Block
}

func PrintErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func NewAccount(name string) Account {
	privateKey, pubKeyBytes:=newKeyPair()
	addressBytes:=GetAddress(pubKeyBytes)
	addressString:=GetStringAddress(addressBytes)
	return Account{
		Name:name, 
		PrivateKey:privateKey,
		PubKeyBytes:pubKeyBytes,
		Address:addressString,
		BlockIn: make(chan *Block, 8),
		ChannelMap:make(map[string]chan *Block),		
	}
}





//sender create the transactions and sign it
func (acc Account) ProduceTransferTx(to string, amount int) (*Transaction,error){
	utxos:=acc.Blockchain.FindUTXOSet()
	tx,err:=NewUTXOTransaction(acc.PubKeyBytes,to,amount,utxos)
	if err != nil{
		return nil,err
	}
	
	err=acc.Blockchain.SignTransaction(tx,acc.PrivateKey)
	if err != nil{
		return nil,err
	}

	return tx,nil
}

//mine a block with the given transaction and get reward 
//and add it to the mined block
func (acc Account) MineTransaction(tx *Transaction) *Block{
	if !acc.Blockchain.VerifyTransaction(tx){
		return nil
	}
	coinbaseTX,err:=NewCoinbaseTX(acc.Address,"")
	PrintErr(err)
	minedBlock,err:=acc.Blockchain.MineBlock([]*Transaction{coinbaseTX,tx})
	PrintErr(err)
	return minedBlock
}

func (acc Account) BroadcastBlock(block *Block) {
	for _, channel := range acc.ChannelMap{
		channel <-block
	}	
}

//if the mined Block is valid, add to the block chain and return true
func (acc Account) HandleMinedBlockIn(minedBlock *Block) error{
	//verify the pow of the block
	if acc.Blockchain.ValidateBlock(minedBlock)==false{
		return ErrInvalidBlock
	}
	//verify the signature for each transaction in that block
	for _,tx := range minedBlock.Transactions{
		if acc.Blockchain.VerifyTransaction(tx)==false{
			return errors.New("Signature not valid")
		}
	}
	//add to blockchain if all valid
	acc.Blockchain.blocks = append(acc.Blockchain.blocks,minedBlock)
	return nil
}

//get balance for a certain public key
func (acc Account) GetBalance() int{
	//all unspent transaction outputs in the blockchain 
	u:=acc.Blockchain.FindUTXOSet()
	accumulatedBalance:=0
	for _, mapTXOutput := range u{	
		for _, txOutput := range mapTXOutput{	
			if  txOutput.IsLockedWithKey(HashPubKey(acc.PubKeyBytes)) {
				accumulatedBalance=accumulatedBalance+txOutput.Value
			}
		}
	}
	return accumulatedBalance
}




