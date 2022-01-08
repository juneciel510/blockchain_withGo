package main

import (
	"fmt"
	"reflect"
	"strings"
)

// UTXOSet represents a set of UTXO as an in-memory cache
// The key of the most external map is the transaction ID
// (encoded as string) that contains these outputs
// {map of transaction ID -> {map of TXOutput Index -> TXOutput}}
type UTXOSet map[string]map[int]TXOutput

// FindSpendableOutputs finds and returns unspent outputs in the UTXO Set
// to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	spendableOutputs:=make(map[string][]int)
	accumulatedBalance:=0
	for txID, mapTXOutput := range u{	
		for outIdx, txOutput := range mapTXOutput{	
			if  txOutput.IsLockedWithKey(pubKeyHash) {
				accumulatedBalance=accumulatedBalance+txOutput.Value
				spendableOutputs[txID]=append(spendableOutputs[txID],outIdx)
			}
		}
	}
	return accumulatedBalance, spendableOutputs
}

// FindUTXO finds all UTXO in the UTXO Set for a given unlockingData key (e.g., address)
// This function ignores the index of each output and returns
// a list of all outputs in the UTXO Set that can be unlocked by the user
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput
	for _, mapTXOutput := range u{	
		for _, txOutput := range mapTXOutput{	
			if  txOutput.IsLockedWithKey(pubKeyHash) {			
				UTXO=append(UTXO,txOutput)
			}
		}
	}
	return UTXO
}

// CountUTXOs returns the number of transactions outputs in the UTXO set
func (u UTXOSet) CountUTXOs() int {
	count := 0
	for _, mapTXOutput := range u{	
			count=count+len(mapTXOutput)					
	}

	return count
}

// Update updates the UTXO Set with the new set of transactions
func (u UTXOSet) Update(transactions []*Transaction) {
	for _, tx := range transactions {
		//if a entry is in the input of the new set of transactions
		//means it is consumed, need to be deleted from the the UTXO Set
		for _, txInput := range tx.Vin{
			if tx.IsCoinbase(){
				continue 
			}
			txInputTxidString:=fmt.Sprintf("%x", txInput.Txid)
			if !reflect.DeepEqual(u[txInputTxidString][txInput.OutIdx], TXOutput{}){
				delete(u[txInputTxidString],txInput.OutIdx)
				if len(u[txInputTxidString])==0 {
					delete(u,txInputTxidString)
				}
			}
		}
		//add new output from the new set of transactions
		for index, txOutput := range tx.Vout{
			txIDString:=fmt.Sprintf("%x", tx.ID)
			if u[txIDString]==nil{
				m := make(map[int]TXOutput)
				u[txIDString]=m
			}
			u[txIDString][index] =txOutput
		}
	}
}

func (u UTXOSet) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- UTXO SET:"))
	for txid, outputs := range u {
		lines = append(lines, fmt.Sprintf("     TxID: %s", txid))
		for i, out := range outputs {
			lines = append(lines, fmt.Sprintf("           Output %d: %v", i, out))
		}
	}

	return strings.Join(lines, "\n")
}
