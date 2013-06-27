```
                                        __      
                                       /\ \__   
   __     ___               __     _ __\ \ ,_\  
 /'_ `\  / __`\  _______  /'__`\  /\`'__\ \ \/  
/\ \L\ \/\ \L\ \/\______\/\ \L\.\_\ \ \/ \ \ \_ 
\ \____ \ \____/\/______/\ \__/.\_\\ \_\  \ \__\
 \/___L\ \/___/           \/__/\/_/ \/_/   \/__/
   /\____/                                      
   \_/__/                                       

an adaptive radix tree implementation in go
```
# what

An Adaptive Radix Tree is an indexing data structure similar to traditional radix trees, but uses internal nodes that grow and shrink intelligently with consecutive inserts and removals.

Adaptive Radix Trees have many interesting attributes that could be seen as improvements on other indexing data structures like Hash Maps or other Prefix Trees, such as:

  - Worst-case search complexity of O(k), where `k` is the length of the key.
  - They don't have to be rebuilt due to excessive inserts.
  - The structure of their inner nodes is space-efficent when compared to traditional Prefix Trees.
  - They provide prefix compression, a technique where each inner node specifies how far to 'fast-forward' in the search key before traversing to the next child. 

# usage

Include `go-art` in your go pacakages with:

```
import( "github.com/kellydunn/go-art" )
```

Go nuts:

```
// Make an ART Tree
tree := art.NewTree()

// Insert some stuff
tree.Insert([]byte("art trees"), []byte("are rad"))

// Search for a key, and get the resultant value
res := tree.Search([]byte("art trees"))

// Inspect your result!
fmt.Printf("%s\n", res) // "are rad"
```

# documentation

Check out the documentation on godoc.org: http://godoc.org/github.com/kellydunn/go-art

# implementation details

  - It's currently unclear if golang supports SIMD instructions, so Node16s make use of Binary Search for lookups instead of the originally specified manner.
  - Search is currently implemented in the pessimistic variation as described in the specification linked below.  

# performance

Worst-case scenarios for basic operations are: 

| Search | Insert | Removal |
| ------ |:------:| -------:|
|  O(k)  | O(k)+c | O(k)+c  |

  - `k` is the length of the key that we wish to insert.  With prefix compression, this can be faster than Hashing functions, since hashing functions are O(k) operations.
  - `c` is the number of children at the parent node of insertion or removal. This accounts for the growing or shrinking of the inner node.  At the worst case, this is number is 48; the maximum number of children to move when transitioning between the biggest types of inner nodes.

# releated works

  - http://www-db.in.tum.de/~leis/papers/ART.pdf (Specification)
  - http://www-db.in.tum.de/~leis/index/ART.tgz (C++ reference Implementation)
  - https://github.com/armon/libart (an ANSI C Implementation)