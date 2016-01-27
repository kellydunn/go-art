// Pacakge art provides a golang implementation of Adaptive Radix Trees
package art

import (
	"bytes"
	_ "fmt"
	_ "math"
	_ "os"
)

type ArtTree struct {
	root *ArtNode
	size int64
}

// Creates and returns a new Art Tree with a nil root and a size of 0.
func NewArtTree() *ArtTree {
	return &ArtTree{root: nil, size: 0}
}

// Finds the starting node for a prefix search and returns an array of all the objects under it
func (t *ArtTree) PrefixSearch(key []byte) []interface{} {
	foundStart := t.searchHelper(t.root, key, 0)
	if foundStart != nil {
		ret := make([]interface{}, 0)
		t.eachHelper(foundStart, func(node *ArtNode) {
			if node.IsLeaf() {
				ret = append(ret, node.value)
			}
		})
		return ret
	}
	return nil
}

// Returns the node that contains the passed in key, or nil if not found.
func (t *ArtTree) Search(key []byte) interface{} {
	key = ensureNullTerminatedKey(key)
	foundNode := t.searchHelper(t.root, key, 0)
	if foundNode != nil {
		// i think with the null terminated key, the return is always a leaf, or nil?
		// if foundNode.IsLeaf() {
			return foundNode.value
		// }
	}
	return nil
}

// Recursive search helper function that traverses the tree.
// Returns the node that contains the passed in key, or nil if not found.
func (t *ArtTree) searchHelper(current *ArtNode, key []byte, depth int) *ArtNode {
	// While we have nodes to search
	if current != nil {

		// Check if the current is a match
		if current.IsLeaf() {
			if current.IsMatch(key) {
				return current
			}

			// Bail if no match
			return nil
		}

		// Check if our key mismatches the current compressed path
		if current.PrefixMismatch(key, depth) != current.prefixLen {
			// Bail if there's a mismatch during traversal.
			return nil
		} else {
			// Otherwise, increase depth accordingly.
			depth += current.prefixLen
			if depth > len(key)-1 {
				return current
			}
		}

		// Find the next node at the specified index, and update depth.
		return t.searchHelper(*(current.FindChild(key[depth])), key, depth+1)
	}

	return nil
}

// Inserts the passed in value that is indexed by the passed in key into the ArtTree.
func (t *ArtTree) Insert(key []byte, value interface{}) {
	key = ensureNullTerminatedKey(key)
	t.insertHelper(t.root, &t.root, key, value, 0)
}

// Recursive helper function that traverses the tree until an insertion point is found.
// There are four methods of insertion:
//
// If the current node is null, a new node is created with the passed in key-value pair
// and inserted at the current position.
//
// If the current node is a leaf node, it will expand to a new ArtNode of type NODE4
// to contain itself and a new leaf node containing the passed in key-value pair.
//
// If the current node's prefix differs from the key at a specified depth,
// a new ArtNode of type NODE4 is created to contain the current node and the new leaf node
// with an adjusted prefix to account for the mismatch.
//
// If there is no child at the specified key at the current depth of traversal, a new leaf node
// is created and inserted at this position.
func (t *ArtTree) insertHelper(current *ArtNode, currentRef **ArtNode, key []byte, value interface{}, depth int) {
	// @spec: Usually, the leaf can
	//        simply be inserted into an existing inner node, after growing
	//        it if necessary.
	if current == nil {
		*currentRef = NewLeafNode(key, value)
		t.size += 1
		return
	}

	// @spec: If, because of lazy expansion,
	//        an existing leaf is encountered, it is replaced by a new
	//        inner node storing the existing and the new leaf
	if current.IsLeaf() {

		// TODO Determine if we should overwrite keys if they are attempted to overwritten.
		//      Currently, we bail if the key matches.
		if current.IsMatch(key) {
			return
		}

		// Create a new Inner Node to contain the new Leaf and the current node.
		newNode4 := NewNode4()
		newLeafNode := NewLeafNode(key, value)

		// Determine the longest common prefix between our current node and the key
		limit := current.LongestCommonPrefix(newLeafNode, depth)

		newNode4.prefixLen = limit

		memcpy(newNode4.prefix, key[depth:], min(newNode4.prefixLen, MAX_PREFIX_LEN))

		*currentRef = newNode4

		// Add both children to the new Inner Node
		newNode4.AddChild(current.key[depth+newNode4.prefixLen], current)
		newNode4.AddChild(key[depth+newNode4.prefixLen], newLeafNode)

		t.size += 1
		return
	}

	// @spec: Another special case occurs if the key of the new leaf
	//        differs from a compressed path: A new inner node is created
	//        above the current node and the compressed paths are adjusted accordingly.
	if current.prefixLen != 0 {
		mismatch := current.PrefixMismatch(key, depth)

		// If the key differs from the compressed path
		if mismatch != current.prefixLen {

			// Create a new Inner Node that will contain the current node
			// and the desired insertion key
			newNode4 := NewNode4()
			*currentRef = newNode4
			newNode4.prefixLen = mismatch

			// Copy the mismatched prefix into the new inner node.
			memcpy(newNode4.prefix, current.prefix, mismatch)

			// Adjust prefixes so they fit underneath the new inner node
			if current.prefixLen < MAX_PREFIX_LEN {
				newNode4.AddChild(current.prefix[mismatch], current)
				current.prefixLen -= (mismatch + 1)
				memmove(current.prefix, current.prefix[mismatch+1:], min(current.prefixLen, MAX_PREFIX_LEN))
			} else {
				current.prefixLen -= (mismatch + 1)
				minKey := current.Minimum().key
				newNode4.AddChild(minKey[depth+mismatch], current)
				memmove(current.prefix, minKey[depth+mismatch+1:], min(current.prefixLen, MAX_PREFIX_LEN))
			}

			// Attach the desired insertion key
			newLeafNode := NewLeafNode(key, value)
			newNode4.AddChild(key[depth+mismatch], newLeafNode)

			t.size += 1
			return
		}

		depth += current.prefixLen
	}

	// Find the next child
	next := current.FindChild(key[depth])

	// If we found a child that matches the key at the current depth
	if *next != nil {

		// Recurse, and keep looking for an insertion point
		t.insertHelper(*next, next, key, value, depth+1)

	} else {
		// Otherwise, Add the child at the current position.
		current.AddChild(key[depth], NewLeafNode(key, value))
		t.size += 1
	}
}

// Removes the child that is accessed by the passed in key.
func (t *ArtTree) Remove(key []byte) {
	key = ensureNullTerminatedKey(key)
	t.removeHelper(t.root, &t.root, key, 0)
}

// Recursive helper for Removing child nodes.
// There are two methods for removal:
//
// If the current node is a leaf and matches the specified key, remove it.
//
// If the next child at the specifed key and depth matches,
// the current node shall remove it accordingly.
func (t *ArtTree) removeHelper(current *ArtNode, currentRef **ArtNode, key []byte, depth int) {
	// Bail early if we are at a nil node.
	if current == nil {
		return
	}

	// If the current node matches, remove it.
	if current.IsLeaf() {
		if current.IsMatch(key) {
			*currentRef = nil
			t.size -= 1
			return
		}
	}

	// If the current node contains a prefix length
	if current.prefixLen != 0 {

		// Bail out if we encounter a mismatch
		mismatch := current.PrefixMismatch(key, depth)
		if mismatch != current.prefixLen {
			return
		}

		// Increase traversal depth
		depth += current.prefixLen
	}

	// Find the next child
	next := current.FindChild(key[depth])

	// Let the Inner Node handle the removal logic if the child is a match
	if *next != nil && (*next).IsLeaf() && (*next).IsMatch(key) {
		current.RemoveChild(key[depth])
		t.size -= 1
		// Otherwise, recurse.	t.size -= 1
	} else {
		t.removeHelper(*next, next, key, depth+1)
	}
}

// Convenience method for EachPreorder
func (t *ArtTree) Each(callback func(*ArtNode)) {
	t.eachHelper(t.root, callback)
}

// Recursive helper for iterative over the ArtTree.  Iterates over all nodes in the tree,
// executing the passed in callback as specified by the passed in traversal type.
func (t *ArtTree) eachHelper(current *ArtNode, callback func(*ArtNode)) {
	// Bail early if there's no node to iterate over
	if current == nil {
		return
	}

	callback(current)

	// Art Nodes of type NODE48 do not necessarily store their children in sorted order.
	// So we must instead iterate over their keys, acccess the children, and iterate properly.
	if current.nodeType == NODE48 {
		for i := 0; i < len(current.keys); i++ {
			index := current.keys[byte(i)]
			if index > 0 {
				next := current.children[index-1]

				if next != nil {

					// Recurse
					t.eachHelper(next, callback)
				}
			}
		}

		// Art Nodes of type NODE4, NODE16, and NODE256 keep their children in order,
		// So we can access them iteratively.
	} else {

		for i := 0; i < len(current.children); i++ {
			next := current.children[i]

			if next != nil {

				// Recurse
				t.eachHelper(next, callback)
			}
		}
	}
}

func memcpy(dest []byte, src []byte, numBytes int) {
	for i := 0; i < numBytes && i < len(src) && i < len(dest); i++ {
		dest[i] = src[i]
	}
}

func memmove(dest []byte, src []byte, numBytes int) {
	for i := 0; i < numBytes; i++ {
		dest[i] = src[i]
	}
}

// Returns the passed in key as a null terminated byte array
// if it is not already null terminated.
func ensureNullTerminatedKey(key []byte) []byte {
	index := bytes.Index(key, []byte{0})

	// Is there a null terminated character?
	if index < 0 {

		// Append one.
		key = append(key, byte(0))

	}

	return key
}
