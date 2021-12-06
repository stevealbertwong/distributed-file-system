/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: API and interal functions to interact with the Key-Value store  */
/*           that the Chord ring is providing.                               */
/*                                                                           */

package chord

import (
	"fmt"
)

/*                             */
/* client level library / External API Into Datastore */
/*                             */

/* Get a value in the datastore, provided an abitrary node in the ring */
func Get(node *Node, key string) (string, error) {
	
	dest_node, err := node.locate(key)
	string value = Get_RPC(dest_node, key)

	return value, nil
}

/* Put a key/value in the datastore, provided an abitrary node in the ring */
func Put(node *Node, key string, value string) error {

	Put_RPC(node, key, value)
	return nil
}


/* Internal helper method to find the appropriate node in the ring */
func (node *Node) locate(key string) (*RemoteNode, error) {

	// 1. hash key into object id 
	object_id = HashKey(key)

	// 2. successor of object_id == node that stores key
	node = FindSuccessor_RPC(object_id)

	return node, nil
}





////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


/* Print the contents of a node's data store */
func PrintDataStore(node *Node) {
	fmt.Printf("Node-%v datastore: %v\n", HashStr(node.Id), node.dataStore)
}



/* Creates a Chord node with a pre-defined ID (useful for testing) */
func CreateDefinedNode(parent *RemoteNode, definedId []byte) (*Node, error) {
	node := new(Node)
	err := node.init(parent, definedId)
	if err != nil {
		return nil, err
	}
	return node, err
}

/* Create Chord node with random ID based on listener address */
func CreateNode(parent *RemoteNode) (*Node, error) {
	node := new(Node)
	err := node.init(parent, nil)
	if err != nil {
		return nil, err
	}
	return node, err
}


/* Shutdown a specified Chord node (gracefully) */
func ShutdownNode(node *Node) {
	node.IsShutdown = true
	// Wait for go routines to quit, should be enough time.
	time.Sleep(time.Millisecond * 2000)
	node.Listener.Close()

	// 1. u() immediate predecessor's immediate successor n immediate successor's immediate predecessor
	SetSuccessorId_RPC(node.Predecessor, node.Successor)
	SetPredecessorId_RPC(node.Successor, node.Predecessor)

	// 2. us as predecessor transfer ALL our data to our successor 
	node.dsLock.Lock()
	for k,v := range node.dataStore {
		Put_RPC(node.succesor, k, v)
		// 3. delete kv map
		delete(node.dataStore, k)
	}
	node.dsLock.Unlock()
	

}

