package merkle

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"sync"
)

const (
	// 账户树深度，与以太坊一致
	AccountTreeDepth = 160
)

// 默克尔树节点
type Node struct {
	Key    []byte
	Value  []byte
	Hash   []byte
	Left   *Node
	Right  *Node
	Parent *Node
	IsLeaf bool
}

// 稀疏默克尔树
type SparseMerkleTree struct {
	root       *Node
	depth      int
	mu         sync.RWMutex
	nullHashes [][]byte // 空节点的哈希缓存
}

// 创建新的稀疏默克尔树
func NewSparseMerkleTree(depth int) *SparseMerkleTree {
	tree := &SparseMerkleTree{
		depth:      depth,
		nullHashes: make([][]byte, depth+1),
	}

	// 初始化空节点哈希
	empty := sha256.Sum256([]byte{})
	tree.nullHashes[0] = empty[:]

	for i := 1; i <= depth; i++ {
		h := sha256.New()
		h.Write(tree.nullHashes[i-1])
		h.Write(tree.nullHashes[i-1])
		tree.nullHashes[i] = h.Sum(nil)
	}

	// 创建根节点
	tree.root = &Node{
		Hash:   tree.nullHashes[depth],
		IsLeaf: false,
	}

	return tree
}

// 获取根哈希
func (t *SparseMerkleTree) GetRootHash() []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.root.Hash
}

// 插入或更新键值对
func (t *SparseMerkleTree) Update(key, value []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 计算值的哈希
	valueHash := sha256.Sum256(value)

	// 插入或更新叶子节点
	path := t.keyToPath(key)
	t.updateNode(t.root, path, 0, key, value, valueHash[:])

	return nil
}

// 计算键的路径（二进制表示）
func (t *SparseMerkleTree) keyToPath(key []byte) []bool {
	// 将键转换为二进制路径，长度为树的深度
	path := make([]bool, t.depth)

	for i := 0; i < t.depth && i < len(key)*8; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)

		if byteIndex < len(key) {
			path[i] = (key[byteIndex] & (1 << uint(bitIndex))) != 0
		}
	}

	return path
}

// 更新节点
func (t *SparseMerkleTree) updateNode(node *Node, path []bool, depth int, key, value []byte, valueHash []byte) *Node {
	// 到达叶子节点
	if depth == len(path) {
		if node == nil || !node.IsLeaf {
			node = &Node{
				Key:    make([]byte, len(key)),
				Value:  make([]byte, len(value)),
				Hash:   make([]byte, len(valueHash)),
				IsLeaf: true,
			}
			copy(node.Key, key)
			copy(node.Value, value)
			copy(node.Hash, valueHash)
		} else {
			// 更新现有叶子节点
			copy(node.Value, value)
			copy(node.Hash, valueHash)
		}
		return node
	}

	// 内部节点
	if node == nil {
		node = &Node{
			IsLeaf: false,
			Hash:   t.nullHashes[t.depth-depth],
		}
	}

	// 根据路径决定更新左子树还是右子树
	if path[depth] {
		// 更新右子树
		node.Right = t.updateNode(node.Right, path, depth+1, key, value, valueHash)
		if node.Right != nil {
			node.Right.Parent = node
		}
	} else {
		// 更新左子树
		node.Left = t.updateNode(node.Left, path, depth+1, key, value, valueHash)
		if node.Left != nil {
			node.Left.Parent = node
		}
	}

	// 重新计算节点哈希
	t.recomputeHash(node, depth)

	return node
}

// 重新计算节点哈希
func (t *SparseMerkleTree) recomputeHash(node *Node, depth int) {
	if node.IsLeaf {
		return // 叶子节点的哈希在更新时已计算
	}

	h := sha256.New()

	// 如果左子节点存在，使用其哈希；否则使用空哈希
	leftHash := t.nullHashes[t.depth-depth-1]
	if node.Left != nil {
		leftHash = node.Left.Hash
	}

	// 如果右子节点存在，使用其哈希；否则使用空哈希
	rightHash := t.nullHashes[t.depth-depth-1]
	if node.Right != nil {
		rightHash = node.Right.Hash
	}

	h.Write(leftHash)
	h.Write(rightHash)
	node.Hash = h.Sum(nil)
}

// 获取值
func (t *SparseMerkleTree) Get(key []byte) ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	path := t.keyToPath(key)
	node := t.getNode(t.root, path, 0)

	if node == nil || !node.IsLeaf {
		return nil, errors.New("key not found")
	}

	return node.Value, nil
}

// 根据路径获取节点
func (t *SparseMerkleTree) getNode(node *Node, path []bool, depth int) *Node {
	if node == nil {
		return nil
	}

	// 到达叶子节点
	if node.IsLeaf {
		return node
	}

	// 到达路径末端
	if depth == len(path) {
		return node
	}

	// 根据路径选择子节点
	if path[depth] {
		return t.getNode(node.Right, path, depth+1)
	}

	return t.getNode(node.Left, path, depth+1)
}

// 生成默克尔证明
func (t *SparseMerkleTree) GenerateProof(key []byte) ([][]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	path := t.keyToPath(key)
	proof := make([][]byte, t.depth)

	// 收集证明路径
	node := t.root
	for i := 0; i < t.depth; i++ {
		if node == nil {
			// 使用空节点哈希
			proof[i] = t.nullHashes[t.depth-i-1]
			continue
		}

		if path[i] {
			// 右路径，提供左节点的哈希
			if node.Left != nil {
				proof[i] = node.Left.Hash
			} else {
				proof[i] = t.nullHashes[t.depth-i-1]
			}
			node = node.Right
		} else {
			// 左路径，提供右节点的哈希
			if node.Right != nil {
				proof[i] = node.Right.Hash
			} else {
				proof[i] = t.nullHashes[t.depth-i-1]
			}
			node = node.Left
		}
	}

	return proof, nil
}

// 验证默克尔证明
func VerifyProof(rootHash []byte, key, value []byte, proof [][]byte, depth int) bool {
	// 计算值的哈希
	valueHash := sha256.Sum256(value)
	currentHash := valueHash[:]

	// 计算键的路径
	path := make([]bool, depth)
	for i := 0; i < depth && i < len(key)*8; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)

		if byteIndex < len(key) {
			path[i] = (key[byteIndex] & (1 << uint(bitIndex))) != 0
		}
	}

	// 从叶子节点开始，向上计算哈希
	for i := depth - 1; i >= 0; i-- {
		h := sha256.New()

		if path[i] {
			// 当前节点在右边
			h.Write(proof[i])
			h.Write(currentHash)
		} else {
			// 当前节点在左边
			h.Write(currentHash)
			h.Write(proof[i])
		}

		currentHash = h.Sum(nil)
	}

	// 比较计算出的根哈希与给定的根哈希
	return bytes.Equal(currentHash, rootHash)
}
