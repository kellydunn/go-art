package art

import (
	"bytes"
	_ "fmt"
	_ "sort"
	"testing"
)

// A Leaf Node should be able to correctly determine if it is a match or not
func TestIsMatch(t *testing.T) {
	leaf := &ArtNode{key: []byte("test"), nodeType: LEAF}
	if !leaf.IsMatch([]byte("test")) {
		t.Error("Unexpected match for leaf node")
	}

	leaf2 := &ArtNode{key: []byte("test2"), nodeType: LEAF}
	if leaf2.IsMatch([]byte("test")) {
		t.Error("Unexpected match for leaf2 node")
	}
}

// An ArtNode should be able to determine if it is a leaf or not
func TestIsLeaf(t *testing.T) {
	leaf := &ArtNode{nodeType: LEAF}

	if !leaf.IsLeaf() {
		t.Error("Unable to successfully classify leaf")
	}

	innerNodes := []*ArtNode{NewNode4(), NewNode16(), NewNode48(), NewNode256()}

	for node := range innerNodes {
		if innerNodes[node].IsLeaf() {
			t.Error("Incorrectly classified inner node as leaf")
		}
	}
}

// A Leaf Node should be able to retreive its value
func TestValue(t *testing.T) {
	leaf := &ArtNode{nodeType: LEAF, value: "foo"}

	if leaf.Value() != "foo" {
		t.Error("Unexpected value for leaf node")
	}

}

// An ArtNode4 should be able to find the expected child element
func TestAddChildAndFindChildForAllNodeTypes(t *testing.T) {
	nodes := []*ArtNode{NewNode4(), NewNode16(), NewNode48(), NewNode256()}

	// For each different type of node
	for node := range nodes {
		n := nodes[node]

		// Fill it up
		for i := 0; i < n.MaxSize(); i++ {
			newChild := &ArtNode{value: byte(i)}
			n.AddChild(byte(i), newChild)
		}

		// Expect to find all children for that paticular type of node
		for i := 0; i < n.MaxSize(); i++ {
			x := *(n.FindChild(byte(i)))

			if x == nil {
				t.Error("Could not find child as expected")
			}

			if x.value.(byte) != byte(i) {
				t.Error("Child value does not match as expected")
			}
		}
	}
}

// Index should be able to return the correct location of the child
// at the specfied key for all inner node types
func TestIndexForAllNodeTypes(t *testing.T) {
	nodes := []*ArtNode{NewNode4(), NewNode16(), NewNode48(), NewNode256()}

	// For each different type of node
	for node := range nodes {
		n := nodes[node]

		// Fill it up
		for i := 0; i < n.MaxSize(); i++ {
			newChild := &ArtNode{value: byte(i)}
			n.AddChild(byte(i), newChild)
		}

		for i := 0; i < n.MaxSize(); i++ {
			if n.Index(byte(i)) != i {
				t.Error("Unexpected value for Index function")
			}
		}
	}
}

// An ArtNode4 should be able to add a child, and then return the expected child reference.
func TestArtNode4AddChild1AndFindChild(t *testing.T) {
	n := NewNode4()
	n2 := NewNode4()
	n.AddChild('a', n2)

	if n.size < 1 {
		t.Error("Size is incorrect after adding one child to empty Node4")
	}

	x := *(n.FindChild('a'))
	if x != n2 {
		t.Error("Unexpected child reference")
	}
}

// An ArtNode4 should be able to add two child elements with differing prefixes
// And preserve the sorted order of the keys.
func TestArtNode4AddChildTwicePreserveSorted(t *testing.T) {
	n := NewNode4()
	n2 := NewNode4()
	n3 := NewNode4()
	n.AddChild('b', n2)
	n.AddChild('a', n3)

	if n.size < 2 {
		t.Error("Size is incorrect after adding one child to empty Node4")
	}

	if n.keys[0] != 'a' {
		t.Error("Unexpected key value for first key index")
	}

	if n.keys[1] != 'b' {
		t.Error("Unexpected key value for second key index")
	}
}

// An ArtNode4 should be able to add 4 child elements with different prefixes
// And preserve the sorted order of the keys.
func TestArtNode4AddChild4PreserveSorted(t *testing.T) {
	n := NewNode4()

	for i := 4; i > 0; i-- {
		n.AddChild(byte(i), NewNode4())
	}

	if n.size < 4 {
		t.Error("Size is incorrect after adding one child to empty Node4")
	}

	expectedKeys := []byte{1, 2, 3, 4}
	if bytes.Compare(n.keys, expectedKeys) != 0 {
		t.Error("Unexpected key sequence")
	}
}

// An ArtNode16 should be able to add 16 children elements and preserve their sorted order
func TestArtNode16AddChild16PreserveSorted(t *testing.T) {
	n := NewNode16()
	for i := 16; i > 0; i-- {
		n.AddChild(byte(i), NewNode4())
	}

	if n.size < 16 {
		t.Error("Size is incorrect after adding one child to empty Node4")
	}

	for i := 0; i < 16; i++ {
		if n.keys[i] != byte(i+1) {
			t.Error("Unexpected key sequence")
		}
	}
}

// Art Nodes of all types should grow to the next biggest size in sequence
func TestGrow(t *testing.T) {
	nodes := []*ArtNode{NewNode4(), NewNode16(), NewNode48()}
	expectedTypes := []uint8{NODE16, NODE48, NODE256}

	for i := range nodes {
		node := nodes[i]

		node.grow()
		if node.nodeType != expectedTypes[i] {
			t.Error("Unexpected node type after growing")
		}
	}
}

// Art Nodes of all types should next smallest size in sequence
func TestShrink(t *testing.T) {
	nodes := []*ArtNode{NewNode256(), NewNode48(), NewNode16(), NewNode4()}
	expectedTypes := []uint8{NODE48, NODE16, NODE4, LEAF}

	for i := range nodes {
		node := nodes[i]

		for j := 0; j < node.MinSize(); j++ {
			if node.nodeType != NODE4 {
				node.AddChild(byte(i), NewNode4())
			} else {
				// We want to test that the Node4 reduces itself to
				// A LEAF if its only child is a leaf
				node.AddChild(byte(i), &ArtNode{nodeType: LEAF})
			}
		}

		node.shrink()
		if node.nodeType != expectedTypes[i] {
			t.Error("Unexpected node type after shrinking")
		}
	}
}

func TestNewLeafNode(t *testing.T) {
	key := []byte{'a', 'r', 't'}
	value := "tree"
	l := NewLeafNode(key, value)

	if &l.key == &key {
		t.Errorf("Address of key byte slices should not match.")
	}

	if bytes.Compare(l.key, key) != 0 {
		t.Errorf("Expected key value to match the one supplied")
	}

	if l.value != value {
		t.Errorf("Expected initial value to match the one supplied")
	}

	if l.nodeType != LEAF {
		t.Errorf("Expected Leaf node to be of LEAF type")
	}
}
