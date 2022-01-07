/*

NOTE: 
- granted own class since to many functions 
*/
/*
 *  Brown University, CS138, Spring 2018
 *
 *  Purpose: Defines Backpointers struct and implements accessors and
 *  mutators for Backpointers objects.
 */

 package tapestry

 import (
	 "sync"
 )
 
 // Backpointers are stored by level, like the routing table
 // A backpointer at level n indicates that the backpointer shares a prefix of length n with this node
 // Access to the backpointers is managed by a lock
 type Backpointers struct {
	 local RemoteNode       // the local tapestry node
	 sets  [DIGITS]*NodeSet // backpointers
 }
 
 // Represents a set of nodes.
 // The implementation is just a wrapped map with a mutex.
 type NodeSet struct {
	 data  map[RemoteNode]bool
	 mutex sync.Mutex
 }
 
 // Creates and returns a new backpointer set.
 func NewBackpointers(me RemoteNode) *Backpointers {
	 b := new(Backpointers)
	 b.local = me
	 for i := 0; i < DIGITS; i++ {
		 b.sets[i] = NewNodeSet()
	 }
	 return b
 }

 
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

 // Add a backpointer for the provided node
 // Returns true if a new backpointer was added.
 func (b *Backpointers) Add(node RemoteNode) bool {
	 if b.local != node {
		 return b.level(node).Add(node)
	 }
	 return false // 
 }
 
 // Remove a backpointer for the provided node, if it existed
 // Returns true if the backpointer existed and was subsequently removed.
 func (b *Backpointers) Remove(node RemoteNode) bool {
	 if b.local != node {
		 return b.level(node).Remove(node)
	 }
	 return false
 }
 
 // Get all backpointers at the provided level.
 func (b *Backpointers) Get(level int) []RemoteNode {
	 if level >= DIGITS || level < 0 {
		 return []RemoteNode{}
	 }
	 return b.sets[level].Nodes()
 }
 
 // Gets the node set for the level that the specified node should occupy.
 func (b *Backpointers) level(node RemoteNode) *NodeSet {
	 return b.sets[SharedPrefixLength(b.local.Id, node.Id)]
 }
 
 // Create a new node set.
 func NewNodeSet() *NodeSet {
	 s := new(NodeSet)
	 s.data = make(map[RemoteNode]bool)
	 return s
 }
 
 // Add the given node to the node set if it isn't already in the set.
 // Returns true if the node was added; false if it already existed
 func (s *NodeSet) Add(n RemoteNode) bool {
	 s.mutex.Lock()
	 _, exists := s.data[n]
	 s.data[n] = true
	 s.mutex.Unlock()
	 return !exists
 }
 
 // Add all of the nodes to the node set.
 func (s *NodeSet) AddAll(nodes []RemoteNode) {
	 s.mutex.Lock()
	 for _, node := range nodes {
		 s.data[node] = true
	 }
	 s.mutex.Unlock()
 }
 
 // Remove the given node from the node set if it's currently in the set
 // Returns true if the node was removed; false if it was not in the set.
 func (s *NodeSet) Remove(n RemoteNode) bool {
	 s.mutex.Lock()
	 _, exists := s.data[n]
	 delete(s.data, n)
	 s.mutex.Unlock()
	 return exists
 }
 
 // Test whether the specified node is contained in the set
 func (s *NodeSet) Contains(n RemoteNode) (b bool) {
	 s.mutex.Lock()
	 b = s.data[n]
	 s.mutex.Unlock()
	 return
 }
 
 // Returns the size of the set
 func (s *NodeSet) Size() int {
	 s.mutex.Lock()
	 size := len(s.data)
	 s.mutex.Unlock()
	 return size
 }
 
 // Get all nodes in the set as a slice
 func (s *NodeSet) Nodes() []RemoteNode {
	 s.mutex.Lock()
	 nodes := make([]RemoteNode, 0, len(s.data))
	 for node := range s.data {
		 nodes = append(nodes, node)
	 }
	 s.mutex.Unlock()
	 return nodes
 }
 


/*

backpointers == 
doubly linked list: if i point at you, you point back at me too


find the closest set of nodes to fill the routing table with

*/
func TraverseBackpointers(){


}

