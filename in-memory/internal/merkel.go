package mvcc

import (
	"bytes"
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Hash  []byte
	Left  *MerkleNode
	Right *MerkleNode
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return nil
	}

	var nodes []*MerkleNode
	for _, datum := range data {
		sh := sha256.Sum256(datum)
		h := make([]byte, len(sh))
		copy(h, sh[:])
		node := &MerkleNode{
			Hash: h,
		}
		nodes = append(nodes, node)
	}

	for len(nodes) > 1 {
		newNodes := []*MerkleNode{}
		for i := 0; i < len(nodes); i += 2 {
			right := nodes[i+1]
			node := &MerkleNode{
				Left:  nodes[i],
				Right: right,
				Hash:  hashNodes(nodes[i].Hash, right.Hash),
			}
			newNodes = append(newNodes, node)
		}
		nodes = newNodes
	}

	return &MerkleTree{
		RootNode: nodes[0],
	}
}

func hashNodes(left, right []byte) []byte {
	h := sha256.New()
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}

func (mt *MerkleTree) GetRootHash() []byte {
	return mt.RootNode.Hash
}

func (mt *MerkleTree) GetProof(index int) [][]byte {
	node := mt.RootNode
	proof := [][]byte{}
	for node.Left != nil || node.Right != nil {
		if index%2 == 0 {
			proof = append(proof, node.Right.Hash)
			node = node.Left
		} else {
			proof = append(proof, node.Left.Hash)
			node = node.Right
		}
		index /= 2
	}
	return proof
}

func (mt *MerkleTree) VerifyProof(index int, proof [][]byte, targetHash []byte) bool {
	node := mt.RootNode
	for _, hash := range proof {
		if index%2 == 0 {
			node = &MerkleNode{
				Left:  node,
				Right: &MerkleNode{Hash: hash},
			}
		} else {
			node = &MerkleNode{
				Left:  &MerkleNode{Hash: hash},
				Right: node,
			}
		}
		index /= 2
	}
	return bytes.Equal(node.Hash, targetHash)
}
