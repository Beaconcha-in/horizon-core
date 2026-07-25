package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

// Node گره درخت مرکل
type Node struct {
	Hash  string
	Left  *Node
	Right *Node
}

// Tree ساختار درخت مرکل
type Tree struct {
	Root  *Node
	Leaves []*Node
}

// NewTree ساخت درخت مرکل جدید از داده‌ها
func NewTree(data [][]byte) *Tree {
	if len(data) == 0 {
		return &Tree{}
	}

	var leaves []*Node
	for _, d := range data {
		hash := sha256.Sum256(d)
		leaves = append(leaves, &Node{
			Hash: hex.EncodeToString(hash[:]),
			Left:  nil,
			Right: nil,
		})
	}

	tree := &Tree{Leaves: leaves}
	tree.Root = tree.buildTree(leaves)
	return tree
}

// buildTree ساخت درخت به‌صورت بازگشتی
func (t *Tree) buildTree(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}

	var newLevel []*Node
	for i := 0; i < len(nodes); i += 2 {
		if i+1 < len(nodes) {
			combined := append([]byte(nodes[i].Hash), []byte(nodes[i+1].Hash)...)
			hash := sha256.Sum256(combined)
			newLevel = append(newLevel, &Node{
				Hash:  hex.EncodeToString(hash[:]),
				Left:  nodes[i],
				Right: nodes[i+1],
			})
		} else {
			newLevel = append(newLevel, nodes[i])
		}
	}
	return t.buildTree(newLevel)
}

// GetRootHash دریافت هش ریشه
func (t *Tree) GetRootHash() string {
	if t.Root == nil {
		return ""
	}
	return t.Root.Hash
}

// GenerateProof تولید اثبات برای یک برگ
func (t *Tree) GenerateProof(index int) []string {
	if index < 0 || index >= len(t.Leaves) {
		return nil
	}

	var proof []string
	current := t.Leaves
	currentIndex := index

	for len(current) > 1 {
		var nextLevel []*Node
		for i := 0; i < len(current); i += 2 {
			if i+1 < len(current) {
				combined := append([]byte(current[i].Hash), []byte(current[i+1].Hash)...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, &Node{
					Hash:  hex.EncodeToString(hash[:]),
					Left:  current[i],
					Right: current[i+1],
				})
			} else {
				nextLevel = append(nextLevel, current[i])
			}
		}

		// تعیین اینکه در کدام سمت قرار دارد
		if currentIndex%2 == 0 {
			// در سمت چپ هستیم
			if currentIndex+1 < len(current) {
				proof = append(proof, current[currentIndex+1].Hash)
			}
		} else {
			// در سمت راست هستیم
			proof = append(proof, current[currentIndex-1].Hash)
		}
		currentIndex = currentIndex / 2
		current = nextLevel
	}
	return proof
}

// VerifyProof راستی‌آزمایی اثبات
func VerifyProof(leafHash string, proof []string, rootHash string) bool {
	currentHash := leafHash
	for _, siblingHash := range proof {
		combined := append([]byte(currentHash), []byte(siblingHash)...)
		hash := sha256.Sum256(combined)
		currentHash = hex.EncodeToString(hash[:])
	}
	return currentHash == rootHash
}
