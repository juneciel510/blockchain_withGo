package main

type Users struct {
	UsersMap map[string]Account
}

func CopyBlockchain(bc *Blockchain) *Blockchain {
	blocksCopy := []*Block{}
	for _, block := range bc.blocks {
		blockCopy := *block
		blocksCopy = append(blocksCopy, &blockCopy)
	}
	return &Blockchain{blocks: blocksCopy}
}

func NewUsers() *Users {
	userA := NewAccount("a")
	userB := NewAccount("b")
	userC := NewAccount("c")
	userA.AddressMap = map[string]string{"b": userB.Address, "c": userC.Address}
	userB.AddressMap = map[string]string{"a": userA.Address, "c": userC.Address}
	userC.AddressMap = map[string]string{"b": userB.Address, "a": userA.Address}
	userA.ChannelMap = map[string]chan *Block{"b": userB.BlockIn, "c": userC.BlockIn}
	userB.ChannelMap = map[string]chan *Block{"a": userA.BlockIn, "c": userC.BlockIn}
	userC.ChannelMap = map[string]chan *Block{"b": userB.BlockIn, "a": userA.BlockIn}
	genesisBC, err := NewBlockchain(userA.Address)
	PrintErr(err)

	userA.Blockchain = CopyBlockchain(genesisBC)
	userB.Blockchain = CopyBlockchain(genesisBC)
	userC.Blockchain = CopyBlockchain(genesisBC)
	return &Users{
		UsersMap: map[string]Account{"a": userA, "b": userB, "c": userC},
	}
}

//the sender produce transfer tx, miner mine it,
//add to its own blockchain and broadcast to other users
func (u Users) Transfer(from string, to string, miner string, amount int) error {
	tx, err := u.UsersMap[from].ProduceTransferTx(u.UsersMap[to].Address, amount)
	if err != nil {
		return err
	}
	minedBlock := u.UsersMap[miner].MineTransaction(tx)
	u.UsersMap[miner].BroadcastBlock(minedBlock)
	return nil
}

func (u *Users) HandleChannel() {
	for {
		select {
		case minedBlock := <-u.UsersMap["a"].BlockIn:
			err := u.UsersMap["a"].HandleMinedBlockIn(minedBlock)
			PrintErr(err)
		case minedBlock := <-u.UsersMap["b"].BlockIn:
			err := u.UsersMap["b"].HandleMinedBlockIn(minedBlock)
			PrintErr(err)
		case minedBlock := <-u.UsersMap["c"].BlockIn:
			err := u.UsersMap["c"].HandleMinedBlockIn(minedBlock)
			PrintErr(err)
		}
	}
}