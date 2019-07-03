// Pacakge art provides a golang implementation of Adaptive Radix Trees
package art

import (
	"bytes"
	"sort"
)

const (
	// From the specification: Radix trees consist of two types of nodes:
	// Inner nodes, which map partial keys to other nodes,
	// and leaf nodes, which store the values corresponding to the keys.
	NODE4 = iota
	NODE16
	NODE48
	NODE256
	LEAF

	// Inner nodes of type Node4 must have between 2 and 4 children.
	NODE4MIN = 2
	NODE4MAX = 4

	// Inner nodes of type Node16 must have between 5 and 16 children.
	NODE16MIN = 5
	NODE16MAX = 16

	// Inner nodes of type Node48 must have between 17 and 48 children.
	NODE48MIN = 17
	NODE48MAX = 48

	// Inner nodes of type Node256 must have between 49 and 256 children.
	NODE256MIN = 49
	NODE256MAX = 256

	MAX_PREFIX_LEN = 10
)

// Defines a single ArtNode and its attributes.
type ArtNode struct {
	// Internal Node Attributes
	keys      []byte
	children  []*ArtNode
	prefix    []byte
	prefixLen int
	size      uint8

	// Leaf Node Attributes
	key      []byte
	keySize  uint64
	value    interface{}
	nodeType uint8
}

func NewLeafNode(key []byte, value interface{}) *ArtNode {
	newKey := make([]byte, len(key))
	copy(newKey, key)
	l := &ArtNode{
		key:      newKey,
		value:    value,
		nodeType: LEAF,
	}

	return l
}

// From the specification: The smallest node type can store up to 4 child
// pointers and uses an array of length 4 for keys and another
// array of the same length for pointers. The keys and pointers
// are stored at corresponding positions and the keys are sorted.
func NewNode4() *ArtNode {
	return &ArtNode{keys: make([]byte, NODE4MAX), children: make([]*ArtNode, NODE4MAX), nodeType: NODE4, prefix: make([]byte, MAX_PREFIX_LEN)}
}

// From the specification: This node type is used for storing between 5 and
// 16 child pointers. Like the Node4, the keys and pointers
// are stored in separate arrays at corresponding positions, but
// both arrays have space for 16 entries. A key can be found
// efﬁciently with binary search or, on modern hardware, with
// parallel comparisons using SIMD instructions.
func NewNode16() *ArtNode {
	return &ArtNode{keys: make([]byte, NODE16MAX), children: make([]*ArtNode, NODE16MAX), nodeType: NODE16, prefix: make([]byte, MAX_PREFIX_LEN)}
}

// From the specification: As the number of entries in a node increases,
// searching the key array becomes expensive. Therefore, nodes
// with more than 16 pointers do not store the keys explicitly.
// Instead, a 256-element array is used, which can be indexed
// with key bytes directly. If a node has between 17 and 48 child
// pointers, this array stores indexes into a second array which
// contains up to 48 pointers.
func NewNode48() *ArtNode {
	return &ArtNode{keys: make([]byte, 256), children: make([]*ArtNode, NODE48MAX), nodeType: NODE48, prefix: make([]byte, MAX_PREFIX_LEN)}
}

// From the specification: The largest node type is simply an array of 256
// pointers and is used for storing between 49 and 256 entries.
// With this representation, the next node can be found very
// efﬁciently using a single lookup of the key byte in that array.
// No additional indirection is necessary. If most entries are not
// null, this representation is also very space efﬁcient because
// only pointers need to be stored.
func NewNode256() *ArtNode {
	return &ArtNode{children: make([]*ArtNode, NODE256MAX), nodeType: NODE256, prefix: make([]byte, MAX_PREFIX_LEN)}
}

// Returns whether or not this particular art node is full.
func (n *ArtNode) IsFull() bool { return uint16(n.size) == uint16(n.MaxSize()) }

// Returns whether or not this particular art node is a leaf node.
func (n *ArtNode) IsLeaf() bool { return n.nodeType == LEAF }

// Returns whether or not the key stored in the leaf matches the passed in key.
func (n *ArtNode) IsMatch(key []byte) bool {

	// Bail if user tries to compare  anything but a leaf node
	if n.nodeType != LEAF {
		return false
	}

	return bytes.Compare(n.key, key) == 0

}

// Returns the relative index of the first byte that doesn't match
// between key and the current node's prefix, starting at depth.
// Ex: if the depth is 3 and the current prefix is 'baz',
//     for key "foobar" the result is 2, for "foobaz", 3, and for
//     "fooquux" 0.
func (n *ArtNode) PrefixMismatch(key []byte, depth int) int {
	index := 0
	prefix := n.prefix

	for ; index < n.prefixLen && depth+index < len(key) && key[depth+index] == prefix[index]; index++ {
		if index == MAX_PREFIX_LEN-1 {
			// Once we get past MAX_PREFIX_LEN, the rest of the prefix isn't stored.
			// So grab the first child of this node; the first n.prefixLen bytes of
			// its key are the full prefix.
			prefix = n.Minimum().key[depth:]
		}
	}

	return index
}

func (n *ArtNode) Index(key byte) int {
	switch n.nodeType {
	case NODE4:
		// ArtNodes of type NODE4 have a relatively simple lookup algorithm since
		// they are of very small size:  Simply iterate over all keys and check to see if they match.
		for i := uint8(0); i < n.size; i++ {
			if n.keys[i] == key {
				return int(i)
			}
		}
		return -1
	case NODE16:
		// From the specification: First, the searched key is replicated and then compared to the
		// 16 keys stored in the inner node. In the next step, a
		// mask is created, because the node may have less than
		// 16 valid entries. The result of the comparison is converted to
		// a bit ﬁeld and the mask is applied. Finally, the bit
		// ﬁeld is converted to an index using the count trailing zero
		// instruction. Alternatively, binary search can be used
		// if SIMD instructions are not available.
		//
		// TODO It is currently unclear if golang has intentions of supporting SIMD instructions
		//      So until then, go-art will opt for Binary Search
		index := sort.Search(int(n.size), func(i int) bool { return n.keys[uint8(i)] >= key })
		if index < len(n.keys) && n.keys[index] == key {
			return index
		}

		return -1
	case NODE48:
		// ArtNodes of type NODE48 store the indicies in which to access their children
		// in the keys array which are byte-accessible by the desired key.
		// However, when this key array initialized, it contains many 0 value indicies.
		// In order to distinguish if a child actually exists, we increment this value
		// during insertion and decrease it during retrieval.
		index := int(n.keys[key])
		if index > 0 {
			return int(index) - 1
		}

		return -1
	case NODE256:
		// ArtNodes of type NODE256 possibly have the simplest lookup algorithm.
		// Since all of their keys are byte-addressable, we can simply index to the specific child with the key.
		return int(key)
	default:
		return -1
	}

	return -1
}

// Returns a pointer to the child that matches the passed in key,
// or nil if not present.
func (n *ArtNode) FindChild(key byte) **ArtNode {
	var nullNode *ArtNode = nil

	if n == nil {
		return &nullNode
	}

	switch n.nodeType {
	case NODE4, NODE16, NODE48:
		index := n.Index(key)
		if index >= 0 {
			return &n.children[index]
		}

		return &nullNode

	case NODE256:
		// NODE256 Types directly address their children with bytes
		child := n.children[key]
		if child != nil {
			return &n.children[key]
		}

		return &nullNode

	default:
	}

	return &nullNode
}

// Adds the passed in node to the current ArtNode's children at the specified key.
// The current node will grow if necessary in order for the insertion to take place.
func (n *ArtNode) AddChild(key byte, node *ArtNode) {
	switch n.nodeType {
	case NODE4:
		if !n.IsFull() {
			index := uint8(0)
			for ; index < n.size; index++ {
				if key < n.keys[index] {
					break
				}
			}

			for i := n.size; i > index; i-- {
				if n.keys[i-1] > key {
					n.keys[i] = n.keys[i-1]
					n.children[i] = n.children[i-1]
				}
			}

			n.keys[index] = key
			n.children[index] = node
			n.size += 1
		} else {
			n.grow()
			n.AddChild(key, node)
		}

	case NODE16:
		if !n.IsFull() {
			index := uint8(sort.Search(int(n.size), func(i int) bool { return n.keys[byte(i)] >= key }))

			for i := n.size; i > index; i-- {
				if n.keys[i-1] > key {
					n.keys[i] = n.keys[i-1]
					n.children[i] = n.children[i-1]
				}
			}

			n.keys[index] = key
			n.children[index] = node
			n.size += 1
		} else {
			n.grow()
			n.AddChild(key, node)
		}

	case NODE48:
		if !n.IsFull() {
			index := 0

			for i := 0; i < len(n.children); i++ {
				if n.children[index] != nil {
					index++
				}
			}

			n.children[index] = node
			n.keys[key] = byte(index + 1)
			n.size += 1
		} else {
			n.grow()
			n.AddChild(key, node)
		}

	case NODE256:
		if !n.IsFull() {
			n.children[key] = node

			n.size += 1
		}
	default:
	}
}

// The child indexed by the passed in key is removed if found
// and the current ArtNode is shrunk if it falls below its minimum size.
func (n *ArtNode) RemoveChild(key byte) {
	switch n.nodeType {
	case NODE4, NODE16:
		idx := n.Index(key)

		n.keys[idx] = 0
		n.children[idx] = nil

		if idx >= 0 {
			for i := uint8(idx); i < n.size-1; i++ {
				n.keys[i] = n.keys[i+1]
				n.children[i] = n.children[i+1]
			}

		}

		n.keys[n.size-1] = 0
		n.children[n.size-1] = nil

		n.size -= 1

	case NODE48:
		idx := n.Index(key)

		if idx >= 0 {
			child := n.children[idx]
			if child != nil {
				n.children[idx] = nil
				n.keys[key] = 0
				n.size -= 1
			}
		}

	case NODE256:
		idx := n.Index(key)

		child := n.children[idx]
		if child != nil {
			n.children[idx] = nil
			n.size -= 1
		}

	default:
	}

	if int(n.size) < n.MinSize() {
		n.shrink()
	}
}

// Grows the current ArtNode to the next biggest size.
// ArtNodes of type NODE4 will grow to NODE16
// ArtNodes of type NODE16 will grow to NODE48.
// ArtNodes of type NODE48 will grow to NODE256.
// ArtNodes of type NODE256 will not grow, as they are the biggest type of ArtNodes
func (n *ArtNode) grow() {
	switch n.nodeType {
	case NODE4:
		other := NewNode16()
		other.copyMeta(n)
		for i := 0; i < int(n.size); i++ {
			other.keys[i] = n.keys[i]
			other.children[i] = n.children[i]
		}

		n.replaceWith(other)

	case NODE16:
		other := NewNode48()
		other.copyMeta(n)
		for i := 0; i < int(n.size); i++ {
			child := n.children[i]
			if child != nil {
				index := 0

				for j := 0; j < len(other.children); j++ {
					if other.children[index] != nil {
						index++
					}
				}

				other.children[index] = child
				other.keys[n.keys[i]] = byte(index + 1)
			}
		}

		n.replaceWith(other)

	case NODE48:
		other := NewNode256()
		other.copyMeta(n)
		for i := 0; i < len(n.keys); i++ {
			child := *(n.FindChild(byte(i)))
			if child != nil {
				other.children[byte(i)] = child
			}
		}

		n.replaceWith(other)

	case NODE256:
		// Can't get no bigger (⊙ ロ  ⊙;)
	default:
	}
}

// Shrinks the current ArtNode to the next smallest size.
// ArtNodes of type NODE256 will grow to NODE48
// ArtNodes of type NODE48 will grow to NODE16.
// ArtNodes of type NODE16 will grow to NODE4.
// ArtNodes of type NODE4 will collapse into its first child.
// If that child is not a leaf, it will concatenate its current prefix with that of its childs
// before replacing itself.
func (n *ArtNode) shrink() {
	switch n.nodeType {
	case NODE4:
		// From the specification: If that node now has only one child, it is replaced by its child
		// and the compressed path is adjusted.
		other := n.children[0]

		if !other.IsLeaf() {
			currentPrefixLen := n.prefixLen

			if currentPrefixLen < MAX_PREFIX_LEN {
				n.prefix[currentPrefixLen] = n.keys[0]
				currentPrefixLen++
			}

			if currentPrefixLen < MAX_PREFIX_LEN {
				childPrefixLen := min(other.prefixLen, MAX_PREFIX_LEN-currentPrefixLen)
				memcpy(n.prefix[currentPrefixLen:], other.prefix, childPrefixLen)
				currentPrefixLen += childPrefixLen
			}

			memcpy(other.prefix, n.prefix, min(currentPrefixLen, MAX_PREFIX_LEN))
			other.prefixLen += n.prefixLen + 1
		} else {
			other.copyMeta(n)
		}

		n.replaceWith(other)

	case NODE16:
		other := NewNode4()
		other.copyMeta(n)
		other.size = 0

		for i := 0; i < len(other.keys); i++ {
			other.keys[i] = n.keys[i]
			other.children[i] = n.children[i]
			other.size++
		}

		n.replaceWith(other)

	case NODE48:
		other := NewNode16()
		other.copyMeta(n)
		other.size = 0

		for i := 0; i < len(n.keys); i++ {
			idx := n.keys[byte(i)]
			if idx > 0 {
				child := n.children[idx-1]
				if child != nil {
					other.children[other.size] = child
					other.keys[other.size] = byte(i)
					other.size++
				}
			}
		}

		n.replaceWith(other)

	case NODE256:
		other := NewNode48()
		other.copyMeta(n)
		other.size = 0

		for i := 0; i < len(n.children); i++ {
			child := n.children[byte(i)]
			if child != nil {
				other.children[other.size] = child
				other.keys[byte(i)] = byte(other.size + 1)
				other.size++
			}
		}

		n.replaceWith(other)

	default:
	}
}

// Returns the longest number of bytes that match between the current node's prefix
// and the passed in node at the specified depth.
func (n *ArtNode) LongestCommonPrefix(other *ArtNode, depth int) int {
	limit := min(len(n.key), len(other.key)) - depth

	i := 0
	for ; i < limit; i++ {
		if n.key[depth+i] != other.key[depth+i] {
			return i
		}
	}
	return i
}

// Returns the minimum number of children for the current node.
func (n *ArtNode) MinSize() int {
	switch n.nodeType {
	case NODE4:
		return NODE4MIN
	case NODE16:
		return NODE16MIN
	case NODE48:
		return NODE48MIN
	case NODE256:
		return NODE256MIN
	default:
	}
	return 0
}

// Returns the maximum number of children for the current node.
func (n *ArtNode) MaxSize() int {
	switch n.nodeType {
	case NODE4:
		return NODE4MAX
	case NODE16:
		return NODE16MAX
	case NODE48:
		return NODE48MAX
	case NODE256:
		return NODE256MAX
	default:
	}
	return 0
}

// Returns the Minimum child at the current node.
// The minimum child is determined by recursively traversing down the tree
// by selecting the smallest possible byte in each child
// until a leaf has been reached.
func (n *ArtNode) Minimum() *ArtNode {
	if n == nil {
		return nil
	}

	switch n.nodeType {
	case LEAF:
		return n

	case NODE4, NODE16:
		return n.children[0].Minimum()

	case NODE48:
		i := 0

		for n.keys[i] == 0 {
			i++
		}

		child := n.children[n.keys[i]-1]

		return child.Minimum()

	case NODE256:
		i := 0
		for n.children[i] == nil {
			i++
		}
		return n.children[i].Minimum()

	default:
	}

	return n
}

// Returns the Maximum child at the current node.
// The maximum child is determined by recursively traversing down the tree
// by selecting the biggest possible byte in each child
// until a leaf has been reached.
func (n *ArtNode) Maximum() *ArtNode {
	if n == nil {
		return nil
	}

	switch n.nodeType {
	case LEAF:
		return n

	case NODE4, NODE16:
		return n.children[n.size-1].Maximum()

	case NODE48:
		i := len(n.keys) - 1
		for n.keys[i] == 0 {
			i--
		}

		child := n.children[n.keys[i]-1]
		return child.Maximum()

	case NODE256:
		i := len(n.children) - 1
		for i > 0 && n.children[byte(i)] == nil {
			i--
		}

		return n.children[i].Maximum()

	default:
	}

	return n
}

// Replaces the current node with the passed in ArtNode.
func (n *ArtNode) replaceWith(other *ArtNode) {
	*n = *other
}

// Copies the prefix and size metadata from the passed in ArtNode
// to the current node.
func (n *ArtNode) copyMeta(other *ArtNode) {
	n.size = other.size
	n.prefix = other.prefix
	n.prefixLen = other.prefixLen
}

// Returns the value of the given node, or nil if it is not a leaf.
func (n *ArtNode) Value() interface{} {
	if n.nodeType != LEAF {
		return nil
	}

	return n.value
}

// Returns the smallest of the two passed in integers.
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
