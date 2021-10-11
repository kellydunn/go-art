package art

import (
	"bufio"
	"bytes"
	"encoding/binary"
	_ "fmt"
	_ "log"
	"math/rand"
	"os"
	"testing"
)

// @spec: After a single insert operation, the tree should have a size of 1
//        and the root should be a leaf.
func TestArtTreeInsert(t *testing.T) {
	tree := NewArtTree()
	tree.Insert([]byte("hello"), "world")
	if tree.root == nil {
		t.Error("Tree root should not be nil after insterting.")
	}

	if tree.size != 1 {
		t.Error("Unexpected size after inserting.")
	}

	if tree.root.nodeType != LEAF {
		t.Error("Unexpected node type for root after a single insert.")
	}
}

// @spec: After a single insert operation, the tree should be able
//        to retrieve there term it had inserted earlier
func TestArtTreeInsertAndSearch(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("hello"), "world")
	res := tree.Search([]byte("hello"))

	if res != "world" {
		t.Error("Unexpected search result.")
	}
}

// @spec: After Inserting twice and causing the root node to grow,
//        The tree should be able to successfully retrieve any of
//        the previous inserted values
func TestArtTreeInsert2AndSearch(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("hello"), "world")
	tree.Insert([]byte("yo"), "earth")

	res := tree.Search([]byte("yo"))
	if res == nil {
		t.Error("Could not find Leaf Node with expected key: 'yo'")

	} else {

		if res != "earth" {
			t.Error("Unexpected search result.")
		}
	}

	res2 := tree.Search([]byte("hello"))
	if res2 == nil {
		t.Error("Could not find Leaf Node with expected key: 'hello'")

	} else {

		if res2 != "world" {

			t.Error("Unexpected search result.")
		}
	}
}

// An Art Node with a similar prefix should be split into new nodes accordingly
// And should be searchable as intended.
func TestArtTreeInsert2WithSimilarPrefix(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("a"), "a")
	tree.Insert([]byte("aa"), "aa")

	res := tree.Search([]byte("aa"))
	if res == nil {
		t.Error("Could not find Leaf Node with expected key: 'aa'")
	} else {
		if res != "aa" {
			t.Error("Unexpected search result.")
		}
	}
}

// An Art Node with a similar prefix should be split into new nodes accordingly
// And should be searchable as intended.
func TestArtTreeInsert3AndSearchWords(t *testing.T) {
	tree := NewArtTree()

	searchTerms := []string{"A", "a", "aa"}

	for i := range searchTerms {
		tree.Insert([]byte(searchTerms[i]), searchTerms[i])
	}

	for i := range searchTerms {
		res := tree.Search([]byte(searchTerms[i]))
		if res == nil {
			t.Error("Could not find Leaf Node with expected key.")
		} else {
			if res != searchTerms[i] {
				t.Error("Unexpected search result.")
			}
		}
	}
}

// An ArtNode of type NODE4 should expand to NODE16, and attached to the tree accordingly.
func TestArtTreeInsert5AndRootShouldBeNode16(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 5; i++ {
		tree.Insert([]byte{byte(i)}, "data")
	}

	if tree.root.nodeType != NODE16 {
		t.Error("Unexpected root value after inserting past Node4 Maximum")
	}
}

// An ArtNode of type NODE16 should expand to NODE48, and attached to the tree accordingly.
func TestArtTreeInsert17AndRootShouldBeNode48(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 17; i++ {
		tree.Insert([]byte{byte(i)}, "data")
	}

	if tree.root.nodeType != NODE48 {
		t.Error("Unexpected root value after inserting past Node16 Maximum")
	}
}

// An ArtNode of type NODE48 should expand to NODE256, and attached to the tree accordingly.
func TestArtTreeInsert49AndRootShouldBeNode256(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 49; i++ {
		tree.Insert([]byte{byte(i)}, "data")
	}

	if tree.root.nodeType != NODE256 {
		t.Error("Unexpected root value after inserting past Node16 Maximum")
	}
}

// After inserting many words into the tree, we should be able to successfully retreive all of them
// To ensure their presence in the tree.
func TestInsertManyWordsAndEnsureSearchResultAndMinimumMaximum(t *testing.T) {
	tree := NewArtTree()

	file, err := os.Open("test/assets/words.txt")
	if err != nil {
		t.Error("Couldn't open words.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Insert(line, line)
		}
	}

	file.Seek(int64(os.SEEK_SET), 0)

	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else {
			res := tree.Search(line)

			if res == nil {
				t.Error("Unexpected nil value for search result")
			}

			if res == nil {
				t.Error("Expected payload for element in tree")
			}

			if bytes.Compare(res.([]byte), line) != 0 {
				t.Errorf("Incorrect value for node %v.", line)
			}
		}
	}

	// TODO find a better way of testing the words without slurping up the newline character
	minimum := tree.root.Minimum()
	if bytes.Compare(minimum.Value().([]byte), []byte("A\n")) != 0 {
		t.Error("Unexpected Minimum node.")
	}

	maximum := tree.root.Maximum()
	if bytes.Compare(maximum.Value().([]byte), []byte("zythum\n")) != 0 {
		t.Error("Unexpected Maximum node.")
	}
}

// After inserting many random UUIDs into the tree, we should be able to successfully retreive all of them
// To ensure their presence in the tree.
func TestInsertManyUUIDsAndEnsureSearchAndMinimumMaximum(t *testing.T) {
	tree := NewArtTree()

	file, err := os.Open("test/assets/uuid.txt")
	if err != nil {
		t.Error("Couldn't open uuid.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Insert(line, line)
		}
	}

	file.Seek(int64(os.SEEK_SET), 0)

	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else {
			res := tree.Search(line)

			if res == nil {
				t.Error("Unexpected nil value for search result")
			}

			if res == nil {
				t.Error("Expected payload for element in tree")
			}

			if bytes.Compare(res.([]byte), line) != 0 {
				t.Errorf("Incorrect value for node %v.", line)
			}
		}
	}

	// TODO find a better way of testing the words without slurping up the newline character
	minimum := tree.root.Minimum()
	if bytes.Compare(minimum.Value().([]byte), []byte("00026bda-e0ea-4cda-8245-522764e9f325\n")) != 0 {
		t.Error("Unexpected Minimum node.")
	}

	maximum := tree.root.Maximum()
	if bytes.Compare(maximum.Value().([]byte), []byte("ffffcb46-a92e-4822-82af-a7190f9c1ec5\n")) != 0 {
		t.Error("Unexpected Maximum node.")
	}
}

// Inserting a single value into the tree and removing it should result in a nil tree root.
func TestInsertAndRemove1(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("test"), []byte("data"))

	tree.Remove([]byte("test"))

	if tree.size != 0 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root != nil {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting Two values into the tree and removing one of them
// should result in a tree root of type LEAF
func TestInsert2AndRemove1AndRootShouldBeLeafNode(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("test"), []byte("data"))
	tree.Insert([]byte("test2"), []byte("data"))

	tree.Remove([]byte("test"))

	if tree.size != 1 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root == nil || tree.root.nodeType != LEAF {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting Two values into a tree and deleting them both
// should result in a nil tree root
// This tests the expansion of the root into a NODE4 and
// successfully collapsing into a LEAF and then nil upon successive removals
func TestInsert2AndRemove2AndRootShouldBeNil(t *testing.T) {
	tree := NewArtTree()

	tree.Insert([]byte("test"), []byte("data"))
	tree.Insert([]byte("test2"), []byte("data"))

	tree.Remove([]byte("test"))
	tree.Remove([]byte("test2"))

	if tree.size != 0 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root != nil {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting Five values into a tree and deleting one of them
// should result in a tree root of type NODE4
// This tests the expansion of the root into a NODE16 and
// successfully collapsing into a NODE4 upon successive removals
func TestInsert5AndRemove1AndRootShouldBeNode4(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 5; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	tree.Remove([]byte{1})
	res := *(tree.root.FindChild(byte(1)))
	if res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 4 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root == nil || tree.root.nodeType != NODE4 {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting Five values into a tree and deleting all of them
// should result in a tree root of type nil
// This tests the expansion of the root into a NODE16 and
// successfully collapsing into a NODE4, LEAF, then nil
func TestInsert5AndRemove5AndRootShouldBeNil(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 5; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	for i := 0; i < 5; i++ {
		tree.Remove([]byte{byte(i)})
	}

	res := tree.root.FindChild(byte(1))
	if res != nil && *res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 0 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root != nil {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting 17 values into a tree and deleting one of them should
// result in a tree root of type NODE16
// This tests the expansion of the root into a NODE48, and
// successfully collapsing into a NODE16
func TestInsert17AndRemove1AndRootShouldBeNode16(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 17; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	tree.Remove([]byte{2})
	res := *(tree.root.FindChild(byte(2)))
	if res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 16 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root == nil || tree.root.nodeType != NODE16 {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting 17 values into a tree and removing them all should
// result in a tree of root type nil
// This tests the expansion of the root into a NODE48, and
// successfully collapsing into a NODE16, NODE4, LEAF, and then nil
func TestInsert17AndRemove17AndRootShouldBeNil(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 17; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	for i := 0; i < 17; i++ {
		tree.Remove([]byte{byte(i)})
	}

	res := tree.root.FindChild(byte(1))
	if res != nil && *res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 0 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root != nil {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting 49 values into a tree and removing one of them should
// result in a tree root of type NODE48
// This tests the expansion of the root into a NODE256, and
// successfully collapasing into a NODE48
func TestInsert49AndRemove1AndRootShouldBeNode48(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 49; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	tree.Remove([]byte{2})
	res := *(tree.root.FindChild(byte(2)))
	if res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 48 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root == nil || tree.root.nodeType != NODE48 {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// Inserting 49 values into a tree and removing all of them should
// result in a nil tree root
// This tests the expansion of the root into a NODE256, and
// successfully collapsing into a NODE48, NODE16, NODE4, LEAF, and finally nil
func TestInsert49AndRemove49AndRootShouldBeNil(t *testing.T) {
	tree := NewArtTree()

	for i := 0; i < 49; i++ {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	for i := 0; i < 49; i++ {
		tree.Remove([]byte{byte(i)})
	}

	res := tree.root.FindChild(byte(1))
	if res != nil && *res != nil {
		t.Error("Did not expect to find child after removal")
	}

	if tree.size != 0 {
		t.Error("Unexpected tree size after inserting and removing")
	}

	if tree.root != nil {
		t.Error("Unexpected root node after inserting and removing")
	}
}

// A traversal of the tree should be in preorder
func TestEachPreOrderness(t *testing.T) {
	tree := NewArtTree()
	tree.Insert([]byte("1"), []byte("1"))
	tree.Insert([]byte("2"), []byte("2"))

	var traversal []*ArtNode

	tree.Each(func(node *ArtNode) {
		traversal = append(traversal, node)
	})

	// Order should be Node4, 1, 2
	if traversal[0] != tree.root || traversal[0].nodeType != NODE4 {
		t.Error("Unexpected node at begining of traversal")
	}

	if bytes.Compare(traversal[1].key, append([]byte("1"), 0)) != 0 || traversal[1].nodeType != LEAF {
		t.Error("Unexpected node at second element of traversal")
	}

	if bytes.Compare(traversal[2].key, append([]byte("2"), 0)) != 0 || traversal[2].nodeType != LEAF {
		t.Error("Unexpected node at third element of traversal")
	}
}

// A traversal of a Node48 node should preserve order
// And traverse in the same way for all other nodes.
// Node48s do not store their children in order, and require different logic to traverse them
// so we must test that logic seperately.
func TestEachNode48(t *testing.T) {
	tree := NewArtTree()

	for i := 48; i > 0; i-- {
		tree.Insert([]byte{byte(i)}, []byte{byte(i)})
	}

	var traversal []*ArtNode

	tree.Each(func(node *ArtNode) {
		traversal = append(traversal, node)
	})

	// Order should be Node48, then the rest of the keys in sorted order
	if traversal[0] != tree.root || traversal[0].nodeType != NODE48 {
		t.Error("Unexpected node at begining of traversal")
	}

	for i := 1; i < 48; i++ {
		if bytes.Compare(traversal[i].key, append([]byte{byte(i)}, 0)) != 0 || traversal[i].nodeType != LEAF {
			t.Error("Unexpected node at second element of traversal")
		}
	}
}

// After inserting many values into the tree, we should be able to iterate through all of them
// And get the expected number of nodes.
func TestEachFullIterationExpectCountOfAllTypes(t *testing.T) {
	tree := NewArtTree()

	file, err := os.Open("test/assets/words.txt")
	if err != nil {
		t.Error("Couldn't open words.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Insert(line, line)
		}
	}

	var leafCount int = 0
	var node4Count int = 0
	var node16Count int = 0
	var node48Count int = 0
	var node256Count int = 0

	tree.Each(func(node *ArtNode) {
		switch node.nodeType {
		case NODE4:
			node4Count++
		case NODE16:
			node16Count++
		case NODE48:
			node48Count++
		case NODE256:
			node256Count++
		case LEAF:
			leafCount++
		default:
		}
	})

	if leafCount != 235886 {
		t.Error("Did not correctly count all leaf nodes during traversal")
	}

	if node4Count != 111616 {
		t.Error("Did not correctly count all node4 nodes during traversal")
	}

	if node16Count != 12181 {
		t.Error("Did not correctly count all node16 nodes during traversal")
	}

	if node48Count != 458 {
		t.Error("Did not correctly count all node48 nodes during traversal")
	}

	if node256Count != 1 {
		t.Error("Did not correctly count all node256 nodes during traversal")
	}
}

// After Inserting many values into the tree, we should be able to remove them all
// And expect nothing to exist in the tree.
func TestInsertManyWordsAndRemoveThemAll(t *testing.T) {
	tree := NewArtTree()

	file, err := os.Open("test/assets/words.txt")
	if err != nil {
		t.Error("Couldn't open words.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Insert(line, line)
		}
	}

	file.Seek(int64(os.SEEK_SET), 0)

	numFound := 0

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Remove(line)

			dblCheck := tree.Search(line)
			if dblCheck != nil {
				numFound += 1
			}
		}
	}

	if tree.size != 0 {
		t.Error("Tree is not empty after adding and removing many words")
	}

	if tree.root != nil {
		t.Error("Tree is expected to be nil after removing many words")
	}
}

// After Inserting many values into the tree, we should be able to remove them all
// And expect nothing to exist in the tree.
func TestInsertManyUUIDsAndRemoveThemAll(t *testing.T) {
	tree := NewArtTree()

	file, err := os.Open("test/assets/uuid.txt")
	if err != nil {
		t.Error("Couldn't open uuid.txt")
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Insert(line, line)
		}
	}

	file.Seek(int64(os.SEEK_SET), 0)

	numFound := 0

	for {
		if line, err := reader.ReadBytes('\n'); err != nil {
			break
		} else {
			tree.Remove(line)

			dblCheck := tree.Search(line)
			if dblCheck != nil {
				numFound += 1
			}
		}
	}

	if tree.size != 0 {
		t.Error("Tree is not empty after adding and removing many uuids")
	}

	if tree.root != nil {
		t.Error("Tree is expected to be nil after removing many uuids")
	}
}

// Regression test for issue/2
func TestInsertWithSameByteSliceAddress(t *testing.T) {
	rand.Seed(42)
	key := make([]byte, 8)
	tree := NewArtTree()

	// Keep track of what we inserted
	keys := make(map[string]bool)

	for i := 0; i < 135; i++ {
		binary.BigEndian.PutUint64(key, uint64(rand.Int63()))
		tree.Insert(key, key)

		// Ensure that we can search these records later
		keys[string(key)] = true
	}

	if tree.size != int64(len(keys)) {
		t.Errorf("Mismatched size of tree and expected values.  Expected: %d.  Actual: %d\n", len(keys), tree.size)
	}

	for k := range keys {
		n := tree.Search([]byte(k))
		if n == nil {
			t.Errorf("Did not find entry for key: %v\n", []byte(k))
		}
	}
}

func TestPrefixSearch(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"abcd", "abde", "abfg", "abgh",
		"abcfgh", "abezyx",
		"bcdef", "bcdi", "bcdgh", "abef",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	res := tree.Search([]byte("abezyx"))
	if res != "abezyx" {
		t.Error("Unexpected search result.")
	}

	rr := tree.PrefixSearch([]byte("x"))
	if rr == nil {
		t.Error("empty results empty arrays")
	} else if len(rr) > 0 {
		t.Error("shouldn't be found", rr)
	}

	rr = tree.PrefixSearch([]byte("ax"))
	if rr == nil {
		t.Error("empty results empty arrays")
	} else if len(rr) > 0 {
		t.Error("shouldn't be found", rr)
	}

	rs := ""
	for res := range tree.PrefixSearchChan([]byte("abc")) {
		rs += res.Value.(string) + ","
	}
	if rs != "abcd,abcfgh," {
		t.Error("array didn't match, got", rs)
	}

	rr = tree.PrefixSearch([]byte("ab"))
	if rr == nil {
		t.Error("something should have been found for abc")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "abcd,abcfgh,abde,abef,abezyx,abfg,abgh," {
			t.Error("array didn't match, got", ss)
		}
	}

	rr = tree.PrefixSearch([]byte("bcd"))
	if rr == nil {
		t.Error("something should have been found for abc")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "bcdef,bcdgh,bcdi," {
			t.Error("array didn't match, got", ss)
		}
	}

	rr = tree.PrefixSearch([]byte("a"))
	if rr == nil {
		t.Error("something should have been found for a")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "abcd,abcfgh,abde,abef,abezyx,abfg,abgh," {
			t.Error("array didn't match, got", ss)
		}
	}
}

func TestPrefixSearch2(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"ab", "abc",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	rr := tree.PrefixSearch([]byte("a"))
	if rr == nil {
		t.Error("something should have been found for abc")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "ab,abc," {
			t.Error("array didn't match, got", ss)
		}
	}
}

func TestPrefixSearch3(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"foo:bar", "a", "foo:baz",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	rr := tree.PrefixSearch([]byte("foo:b"))
	if rr == nil {
		t.Error("something should have been found for foo:b")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "foo:bar,foo:baz," {
			t.Error("array didn't match, got", ss)
		}
	}
}

func TestPrefixSearch4(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"a",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	rr := tree.PrefixSearch([]byte(""))
	if rr == nil {
		t.Error("something should have been found for ''")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "a," {
			t.Error("array didn't match, got", ss)
		}
	}
	rr = tree.PrefixSearch([]byte("x"))
	if rr == nil || len(rr) > 1 {
		t.Error("Shouldn't have gotten results for x")

	}
}

func TestPrefixSearch5(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"foot", "food",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	rr := tree.PrefixSearch([]byte("for"))
	if len(rr) > 0 {
		t.Error("should get no results for for")
	}

	rr = tree.PrefixSearch([]byte("fo"))
	if rr == nil {
		t.Error("something should have been found for fo")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "food,foot," {
			t.Error("array didn't match, got", ss)
		}
	}
}

func TestPrefixSearchWithLongCommonPrefix(t *testing.T) {
	tree := NewArtTree()

	searchWords := []string{
		"full-name:abc", "full-name:abc1",
	}

	for _, s := range searchWords {
		tree.Insert([]byte(s), s)
	}
	rr := tree.PrefixSearch([]byte("full-name:ax"))
	if len(rr) > 0 {
		t.Error("should get no results for for")
	}

	rr = tree.PrefixSearch([]byte("full-name:a"))
	if rr == nil {
		t.Error("something should have been found for fo")
	} else {
		ss := ""
		for _, s := range rr {
			ss += s.(string) + ","
		}
		if ss != "full-name:abc,full-name:abc1," {
			t.Error("array didn't match, got", ss)
		}
	}
}
