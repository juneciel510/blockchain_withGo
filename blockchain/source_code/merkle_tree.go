package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	RootNode *Node
	Leafs    []*Node
}

// Node represents a Merkle tree node
type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	Hash   []byte
}

const (
	leftNode = iota
	rightNode
)

// MerkleProof represents way to prove element inclusion on the merkle tree
type MerkleProof struct {
	proof [][]byte
	index []int64
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	if data==nil{
		panic("No merkle tree nodes")
	}
	leafs:=[]*Node{}
	var parents []*Node
	var parentNode *Node
	//initialize leaf nodes
	for _, value:= range data{
		hashValue:=sha256.Sum256(value)
		leafNode := &Node{Left:nil,Right:nil,Hash:hashValue[:]}
		leafs=append(leafs,leafNode)
	}
	//duplicate leaf nodes if the total number of leaves is odd
	if len(leafs)%2 != 0 && len(leafs)!=1 {
		leafs=append(leafs,leafs[len(leafs)-1])			
	}
	//generate parent nodes
	intermediate:=leafs
	tree:=[][]*Node{}
	tree = append(tree, leafs)
	for{
		if len(intermediate)>1{
			if len(intermediate)%2 != 0{
				intermediate=append(intermediate,intermediate[len(intermediate)-1])			
			}
			parents= []*Node{}
			for i := 0; i < len(intermediate); {
				hashValue:=sha256.Sum256(append(intermediate[i].Hash[:],intermediate[i+1].Hash[:]...))
				parentNode=&Node{Left:intermediate[i],Right:intermediate[i+1],Hash:hashValue[:]}
			
				parents=append(parents,parentNode)
				intermediate[i].Parent, intermediate[i+1].Parent = parentNode,parentNode
				i=i+2
			}
			intermediate=parents
			tree = append(tree, parents)						
		}else{
			rootNode:= intermediate[0]
			rootNode.Parent=nil
			return &MerkleTree{RootNode:rootNode,Leafs:leafs}
		}
	}
}

// NewMerkleNode creates a new Merkle tree node
func NewMerkleNode(left, right *Node, data []byte) *Node {
	h:=sha256.New()
	var node *Node
	if left == nil && right == nil {
		h.Write(data)
		node=&Node{Left:nil,Right:nil,Hash:h.Sum(nil)}
	}else {
		h.Write(append(left.Hash, right.Hash...))
		node=&Node{Left:left,Right:right,Hash:h.Sum(nil)}
		left.Parent=node
		right.Parent=node
	}	
	return node
}

// MerkleRootHash return the hash of the merkle root node
func (mt *MerkleTree) MerkleRootHash() []byte {
	return mt.RootNode.Hash
}

// MakeMerkleProof returns a list of hashes and indexes required to
// reconstruct the merkle path of a given hash
//
// @param hash represents the hashed data (e.g. transaction ID) stored on
// the leaf node
// @return the merkle proof (list of intermediate hashes), a list of indexes
// indicating the node location in relation with its parent (using the
// constants: leftNode or rightNode), and a possible error.
func (mt *MerkleTree) MakeMerkleProof(hash []byte) ([][]byte, []int64, error) {
	proof:= [][]byte{}
	indexes:= []int64{}
	var targetNode *Node
	//locate the node which has the given hash
	for _,leaf := range mt.Leafs{
		if bytes.Equal(hash,leaf.Hash)  {
			targetNode = leaf
			break
		}
	}
	if targetNode == nil {
		return [][]byte{}, []int64{}, fmt.Errorf("Node %x not found", hash)
	}
	for targetNode.Parent!=nil {
		if bytes.Equal(targetNode.Parent.Left.Hash,hash) {
			proof= append(proof,targetNode.Parent.Right.Hash)
			indexes = append(indexes, rightNode)
		}else{
			proof= append(proof,targetNode.Parent.Left.Hash)
			indexes = append(indexes, leftNode)
		}
		targetNode = targetNode.Parent
		hash=targetNode.Hash
	}
	return proof,indexes,nil
}

// VerifyProof verifies that the correct root hash can be retrieved by
// recreating the merkle path for the given hash and merkle proof.
//
// @param rootHash is the hash of the current root of the merkle tree
// @param hash represents the hash of the data (e.g. transaction ID)
// to be verified
// @param mProof is the merkle proof that contains the list of intermediate
// hashes and their location on the tree required to reconstruct
// the merkle path.
func VerifyProof(rootHash []byte, hash []byte, mProof MerkleProof) bool {
	intermediateHash:=hash
	for  i:=0;i<len(mProof.index);i++ {
		 h := sha256.New()
		if mProof.index[i]==leftNode{
			h.Write(append(mProof.proof[i],intermediateHash...))
		}else{
			h.Write(append(intermediateHash,mProof.proof[i]...))
		}
		intermediateHash=h.Sum(nil)
	}
	return bytes.Equal(intermediateHash,rootHash)
}
