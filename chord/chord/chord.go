/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Chord struct and related functions to create new nodes, etc.    */
/*                                                                           */

package chord

import (
	"../../cs138"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Number of bits (i.e. M value), assumes <= 128 and divisible by 8
const KEY_LENGTH = 8

// Turn debug-mode printing on/off
const DEBUG = false

/* Non-local node representation */
type RemoteNode struct {
	Id   []byte
	Addr string
}

/* 
Local node representation 

chord server's major struct 
*/
type Node struct {
	Id          []byte            /* Unique Node ID */
	Listener    net.Listener      /* Node listener socket */
	Addr        string            /* String of listener address */
	Successor   *RemoteNode       /* This Node's successor */
	Predecessor *RemoteNode       /* This Node's predecessor */
	RemoteSelf  *RemoteNode       /* Remote node of our self */
	IsShutdown  bool              /* Is node in process of shutting down? */
	FingerTable []FingerEntry     /* Finger table entries */
	ftLock      sync.RWMutex      /* RWLock for finger table */
	dataStore   map[string]string /* Local datastore for this node */
	dsLock      sync.RWMutex      /* RWLock for datastore */
}


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

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


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


/* Initailize a Chord node, start listener, rpc server, and go routines */
func (node *Node) init(parent *RemoteNode, definedId []byte) error {
	if KEY_LENGTH > 128 || KEY_LENGTH%8 != 0 {
		log.Fatal(fmt.Sprintf("KEY_LENGTH of %v is not supported! Must be <= 128 and divisible by 8", KEY_LENGTH))
	}

	listener, _, err := cs138.OpenListener()
	if err != nil {
		return err
	}

	node.Id = HashKey(listener.Addr().String())
	if definedId != nil {
		node.Id = definedId
	}

	node.Listener = listener
	node.Addr = listener.Addr().String()
	node.IsShutdown = false
	node.dataStore = make(map[string]string)

	// Populate RemoteNode that points to self
	node.RemoteSelf = new(RemoteNode)
	node.RemoteSelf.Id = node.Id
	node.RemoteSelf.Addr = node.Addr

	// Join this node to the same chord ring as parent
	err = node.join(parent)
	if err != nil {
		return err
	}

	// Populate finger table
	node.initFingerTable()

	// Thread 1: start RPC server on this connection
	rpc.RegisterName(node.Addr, node)
	go node.startRpcServer()

	// Thread 2: kick off timer to stabilize periodically
	ticker1 := time.NewTicker(time.Millisecond * 100) //freq
	go node.stabilize(ticker1)

	// Thread 3: kick off timer to fix finger table periodically
	ticker2 := time.NewTicker(time.Millisecond * 90) //freq
	go node.fixNextFinger(ticker2)

	return err
}


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

/* Go routine to accept and process RPC requests */
func (node *Node) startRpcServer() {
	for {
		if node.IsShutdown {
			fmt.Printf("[%v] Shutting down RPC server\n", HashStr(node.Id))
			return
		}
		if conn, err := node.Listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			go rpc.ServeConn(conn)
		}
	}
}

/* Shutdown a specified Chord node (gracefully) */
func ShutdownNode(node *Node) {
	node.IsShutdown = true
	// Wait for go routines to quit, should be enough time.
	time.Sleep(time.Millisecond * 2000)
	node.Listener.Close()

	//TODO students should modify this method to gracefully shutdown a node

}


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


/*
chord paper figure 6

this node trying to join remote node's existing ring



*/
func (node *Node) join(other *RemoteNode) error {

	// there is an existing ring 
	if (other) 
		// 1. init your own finger table 
		initFingerTable(node)
		
		// 2. update predecessors to see if new node is their new successor 
		update_others()


	// you are the only node 
	else
		// 3. all table entries are just yourself
		for (i = 1, i < KEY_LENGTH, i++) {
			node.FingerTable[0].Node.Id = node.Id 
			node.FingerTable[0].Node.Addr = node.Addr
		}
		
		node.predecessor.Id = node.Id 
		node.predecessor.Node.Addr = node.Addr
	
	return nil	
}


func (node *Node) update_others(){
	int new_node_id = node.id 
	for (i = 1, i < KEY_LENGTH, i++) {
		Node p = find_predecessor(new_node_id - 2^(i-1)); // ??	
		p.update_finger_table(new_node_id, i)
	}
}


func (node *Node) update_finger_table(int new_node_id , int entry_index){
	int current_node_id = node.id
	
	// if new node is new successor
	if( new_node_id < node.FingerTable[entry_index] && new_node_id > current_node_id){
		node.FingerTable[entry_index].Id = new_node_id
		
		// new node could be current node's predecessor's new successor
		Node p = node.predecessor
		p.update_finger_table(new_node_id, entry_index) 
	}
	
}

////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

/*
Thread 2: Psuedocode from figure 7 of chord paper
(runs on client node)

stabilization == ensures immediate successor n predecessor of every node is accurate 

immediate predecessor asks immediate successor for its predecessor 
to see if there's new node between them

*/
func (node *Node) stabilize(ticker *time.Ticker) {
	for _ = range ticker.C {
		if node.IsShutdown {
			fmt.Printf("[%v-stabilize] Shutting down stabilize timer\n", HashStr(node.Id))
			ticker.Stop()
			return
		}

		// 1. i ask my immediate successor for its immediate predecessor to see if its still me
		// my_successor_predecessor == might be new node in btwn me n my successor == my new successor 
		my_successor_predecessor, err := GetPredecessorId_RPC(node.Successor)
		if err != nil {
			log.Fatal("stabilize() GetPredecessorId_RPC() error: " + err.Error())
		}

		// 2. if my immediate successor there is new node, its predecessor is no longer you
		// then new node is your new successor
		int current_node_id = node.id
		if(my_successor_predecessor > my_successor && my_successor_predecessor < current_node_id){
			node.successor = my_successor_predecessor
			node.FingerTable[0].Node = my_successor_predecessor
		}
		
		// 3. notify new node "i am your father"
		// notify my new successor i am its new predecessor		
		if !EqualIds(node.Successor.Id, node.Id) { // if you are your own successor, do not notify yourself
			node.Successor.notify(current_node_id)
		}
	}
}

// update remoteNode to be my new predecessor 
func (node *Node) notify(remoteNode *RemoteNode) {

	int current_node_id = node.id
	// if remoteNode is in between me and my predecessor
	if( remoteNode.id > current_node_id && remoteNode.id < node.predecessor ) or node.predecessor == NULL {
		node.predecessor = remoteNode
	}
}


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


/*
Psuedocode from figure 4 of chord paper


Q: can there be overshoot if hopping finger table ??
A: yes finger table might not be most updated, eventual consistency 
   "start" is diff in diff nodes, and thus its successor node

Q: what if overshoots so much that predecessor's sucessor is still new node's predecessor ??
A: it will hop clockwise back node by node, thru immediate successor 
   so as long as immediate successor is accurate, can reach real successor 

id == new node id / object id
node == any random node 
n == best guess of predecessor node e.g. random node 's best guess of predecessor of new node 

intuition:
- recursively hopping finger tables via succusser nodes to look for closest predecessor to new node
- new node's successor == new node's predecessor's successor 

chord paper's example: 
- node 6 starts at any arbitrary node, recursively hop its biggest successors
- keeps hopping where biggest successors is still smaller than node 6 
- til hit node 3 whose biggest successors 0 is just bigger than node 6 
- node 3's finger table == where every entry's successor is 0, bigger than 6
- i.e. node 3 is node 6's closest predecessor 
- i.e. node 6 is in btwn node 0 n 3 

chord paper figure 5a: when node 6 joins, random node == node 1 
== recursively hop from node 1 -> node 0 -> node 3
*/
func (node *Node) find_closest_successor(id []byte) (*RemoteNode, error) {
	
	// 1. break condition of distributed recursion -> 
	// this runs on immediate_predecessor node
	if BetweenRightIncl(id, node.Id, node.Successor.Id) ||
		EqualIds(node.Successor.Id, node.Id) {
		return node.Successor, nil
	}

	// 2. recursively hop finger table until just overshoots
	// this runs on client node
	immediate_predecessor = node.find_closest_predecessor(id)

	// 3. rpc immediate_predecessor to check break condition
	// then rpc us back 
	// this runs on client node
	return FindSuccessor_RPC(immediate_predecessor, object_id)

}

/*
runs on client node 

confusing name == since hopping successors to find new node's closest predecessor
id == new node id / object id
*/
func (node *Node) find_closest_predecessor(id []byte) (*RemoteNode, error) {
	
	int current_node_id = node.id
	immediate_succesor, err := GetSuccessorId_RPC(current_node_id)

	// break when target id == current node's immediate successor 
	// then current node == closest_predecessor
	while( !Between(id, current_node_id, immediate_succesor.Id) && !EqualIds(current_node_id, immediate_succesor.Id)){
		current_node_id, err = ClosestPrecedingFinger_RPC(current_node_id, id)
		immediate_succesor, err := GetSuccessorId_RPC(current_node_id)
	}
	return current_node_id // found new node's closest predecessor
}


