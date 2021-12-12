/*
Brown University, CS138

chord distributed algo implementation


FAQ:
1. only join() and stabilize() could update your node's pred n succ ?? 

2. why only transfer data when "notified" i am your new predecessor 
but not during join() or "stabilize" i am your new successor  ??

3. why no networking call + thread + channel pattern ??
A: implemented in rpc lib, could config to same/diff packet send many times, many rpc wait for same replies
*/
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

/* A single finger table entry */
type FingerEntry struct {
	Start []byte       /* ID hash of (n + 2^i) mod (2^m)  */
	Node  *RemoteNode  /* RemoteNode that Start points to */
}

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

/*
creating and shutting down nodes

*/


/* 

Initailize a Chord node, start listener, rpc server, and go routines 
*/
func (node *Node) init(parent *RemoteNode, definedId []byte) error {
	if KEY_LENGTH > 128 || KEY_LENGTH%8 != 0 {
		log.Fatal(fmt.Sprintf("KEY_LENGTH of %v is not supported! Must be <= 128 and divisible by 8", KEY_LENGTH))
	}

	// 1. init chord node 
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
	node.RemoteSelf = new(RemoteNode) // RemoteNode that points to yourself
	node.RemoteSelf.Id = node.Id
	node.RemoteSelf.Addr = node.Addr	
	err = node.join(parent) // "join" packet == Join this node to the same chord ring as parent
	if err != nil {
		return err
	}
	node.initFingerTable()

	// 2. 3 threads == all run periodically 
	// Thread 1: listening for all types of incoming packets 
	rpc.RegisterName(node.Addr, node)
	go node.startRpcServer()

	// Thread 2: "stabilize/notify" packet == fresh immediate predecessor n successor 
	ticker1 := time.NewTicker(time.Millisecond * 100)
	go node.stabilize(ticker1)

	// Thread 3: "find immediate successor for every entry" packet == fresh finger table
	ticker2 := time.NewTicker(time.Millisecond * 90)
	go node.fixNextFinger(ticker2)

	return err
}


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


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


func (node *Node) initFingerTable() {
	
	node.FingerTable = make([]FingerEntry, KEY_LENGTH) 

	// all entries init to point to itself
	for i := 1; i < KEY_LENGTH; i+=1 {
		node.ftLock.Lock()
		node.FingerTable[i].Start = fingerMath(node.Id, i, KEY_LENGTH)
		node.FingerTable[i].Node = node.RemoteSelf
		node.ftLock.Unlock()
	}
	node.Successor = node.RemoteSelf
	
}


/* 
periodically update finger table entries to keep it always fresh

1. compute 1st column "start" in finger table
2. rpc recursion to find successor node for key "start"
3. fill 3rd column "successor" in finger table

*/
func (node *Node) fixNextFinger(ticker *time.Ticker) {

	for _ = range ticker.C {
		// 1. find every finger table entry's successor node
		for i := 0; i < KEY_LENGTH; i+=1 {
			// // 1.1 optimized: if start smaller than previous entry's successor
			// // 2 table entries share same successor == no need to rpc
			// if node.FingerTable[i].Start < node.FingerTable[i-1].node {
			// 	continue
			
			// 1.2 normal case: recursively
			// }else{

			node.ftLock.Lock()
			each_entry_successor, err := node.find_closest_successor(node.FingerTable[i].Start)
			node.FingerTable[i].Node = each_entry_successor
			node.ftLock.Unlock()
			
		}
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

runs on every node peridically:

1. check if there is new node, if yes update new node as my new immediate successor
2. notify successor (might or might not be new node) to transfer my data to me predecessor

NOTE:
- stabilize == notify
- notify == being stabilized == "stabilize" packets handler 
*/
func (node *Node) stabilize(ticker *time.Ticker) {
	for _ = range ticker.C {
		if node.IsShutdown {
			fmt.Printf("[%v-stabilize] Shutting down stabilize timer\n", HashStr(node.Id))
			ticker.Stop()
			return
		}

		// 1. check if there is new node, if yes update new node as my new immediate successor
		// i ask my immediate successor for its immediate predecessor to see if its still me
		// my_successor_predecessor == might be new node in btwn me n my successor == my new successor 
		my_successor_predecessor, err := GetPredecessorId_RPC(node.Successor)
		if err != nil {
			log.Fatal("stabilize() GetPredecessorId_RPC() error: " + err.Error())
		}
			
		if BetweenRightIncl(my_successor_predecessor.Id, node.Id, node.Successor.Id) && my_successor_predecessor != nil {
			node.ftLock.Lock()
			node.Successor = my_successor_predecessor
			node.FingerTable[0].Node = my_successor_predecessor
			node.ftLock.Unlock()
		}
		
		// 2. notify successor (might or might not be new node) to transfer my data to me predecessor
		// if there is NO successor == nothing 
		// if YES successor but NOT new node == transfer data 
		// if YES successor and YES new node == transfer data + "i am your father"
		if !EqualIds(node.Successor.Id, node.Id) { // if you are your own successor, do not notify yourself
			Notify_RPC(node.Successor, node.RemoteSelf)
		}
	}
}


/*
implementation of handler: 
successor being asked by predecessor to transfer data

scenario 1: new node notified by original predecessor == new node -> original predecessor (dataflow)
scenario 2: original succesor notified by new node == original succesor -> new node

1. update predecessor remote node to be my new predecessor (if we are new node successor)
2. we as successor transfer predecessor data belongs to him (successor -> predecessor)


shouldn't new node ask your successor for data when joining ??
not predecessor when stabilizing ?? 

remoteNode == original predecessor or new node 
*/
func (node *Node) notify(new_predecessor *RemoteNode) {

	// 1. update predecessor remote node to be my new predecessor (if we are new node successor)
	// - if predecessor who sent me this is in between me and my existing predecessor 
	// - or i dont even have a predecessor (new node)
	if Between(new_predecessor.Id, node.Predecessor.Id, node.Id) || node.Predecessor == nil {
		node.Predecessor = new_predecessor		
	}
	
	// 2. we as successor transfer predecessor data belongs to him (successor -> predecessor)
	// node.RemoteSelf == where to rpc to 
	// new_predecessor == where to send data to 
	// == where to send data to 
	TransferKeys_RPC(node.RemoteSelf, node.RemoteSelf, new_predecessor.Id)
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

when new node first join an existing ring 
only need any random node to hop its his finger table to find immediate successor 

NOTE:
- when a new node join, it only needs to find its immediate successor 
	- only need successor to start "stabilize" 
	- could have empty finger table, empty predecessor, empty data 
- then uses "stabilize" + "notify" + "fixNextFinger" to seed its predecessor + data + finger table
*/
func (node *Node) join(random_node *RemoteNode) error {

	
	if random_node != nil { // there is any random node in an existing ring 
		// 1. ask any random node to hop its his finger table to find new node's immediate successor 
		succ, err := FindSuccessor_RPC(random_node, node.Id)
		
		// 2. u() new node's immediate successor 
		node.ftLock.Lock()
		node.Successor = succ
		node.FingerTable[0].Node = succ
		node.ftLock.Unlock()
	}else{ // you are the only node 
		return nil 
	}

		// // 1. init your own finger table 
		// initFingerTable(node)

		// // 2. update predecessors to see if new node is their new successor 
		// update_others()

		// // 3. all table entries are just yourself
		// for (i = 1, i < KEY_LENGTH, i++) {
		// 	node.FingerTable[0].Node.Id = node.Id 
		// 	node.FingerTable[0].Node.Addr = node.Addr
		// }
		
		// node.predecessor.Id = node.Id 
		// node.predecessor.Node.Addr = node.Addr
	return nil
}


// func (node *Node) update_others(){
// 	int new_node_id = node.id 
// 	for (i = 1, i < KEY_LENGTH, i++) {
// 		Node p = find_predecessor(new_node_id - 2^(i-1)); // ??	
// 		p.update_finger_table(new_node_id, i)
// 	}
// }


// func (node *Node) update_finger_table(int new_node_id , int entry_index){
// 	int current_node_id = node.id
	
// 	// if new node is new successor
// 	if( new_node_id < node.FingerTable[entry_index] && new_node_id > current_node_id){
// 		node.FingerTable[entry_index].Id = new_node_id
		
// 		// new node could be current node's predecessor's new successor
// 		Node p = node.predecessor
// 		p.update_finger_table(new_node_id, entry_index) 
// 	}
	
// }



/*
Psuedocode from figure 4 of chord paper

runs on multiple nodes 



id == new node id / object id
node == any random node 
n == best guess of predecessor node e.g. random node 's best guess of predecessor of new node 

intuition:
- recursively hopping finger tables via succusser nodes to look for closest predecessor to new node
- by looking at the biggest successor node that is smaller than new node
- keeps hopping as long as ANY successor is still smaller than new node
- until hit "closest predecessor node" itself smaller than new node, BUT ALL successors are bigger than new node
	- this is true as long as all finger table entries' successor is correct 
	- with stablization, "eventually" all entries will be correct, just not at all time
- new node's closest predecessor's immediate successor == new node's immediate successor

chord paper figure 5a example: 
- new node 6 wants to join 
- 6 starts at node 0 or node 1, both table's biggest successor that is smaller than 6 is node 3
- hop to node 3 
- node 3 itself smaller than node 6, but all successors (all 0) are bigger than node 6
- node 3 == node 6's closest predecessor
- node 3's immediate successor node 0 == node 6's immediate successor
- just to verify node 6 is in btwn node 0 n 3 

TLDR: when node 6 joins, random node == node 0, recursively hop from node 0 -> node 3
TLDR: when node 6 joins, random node == node 1, recursively hop from node 1 -> node 3

*/
func (node *Node) find_closest_successor(id []byte) (*RemoteNode, error) {
	
	// 1. break condition of distributed recursion -> 
	// this condition is true only on immediate_predecessor node
	if BetweenRightIncl(id, node.Id, node.Successor.Id) ||
		EqualIds(node.Successor.Id, node.Id) {

		// immediate_predecessor's old immediate_successor == new node's successor	
		return node.Successor, nil
	}

	// 2. recursively hop finger table
	// this runs on client node
	immediate_predecessor, err := node.find_closest_predecessor(id)

	// 3. rpc immediate_predecessor to check break condition
	// then rpc us back 
	// this runs on client node
	return FindSuccessor_RPC(immediate_predecessor, node.Id)

}

/*
runs on client node ONLY
"hopping" is done thru only rpc calls 

recursively hop finger tables as long as any of its successor is smaller than new node / object id


confusing name == since hopping successors to find new node's closest predecessor
id == new node id / object id
*/
func (node *Node) find_closest_predecessor(id []byte) (*RemoteNode, error) {
	
	// maybe random starting node happens to be 
	hopped_node := node.RemoteSelf // var hopped_node_id int = node.id
	immediate_succesor, err := GetSuccessorId_RPC(hopped_node)

	// TEST: if ANY of hopped node's successor is smaller than new node, hopped node is NOT closest_predecessor
	// BREAK: if hopped node's immediate(smallest) successor is bigger than new node
	// == hopped node is closest_predecessor == NONE of hopped node's successor is smaller than new node
	for !Between(id, hopped_node.Id, immediate_succesor.Id) && !EqualIds(hopped_node.Id, immediate_succesor.Id) {
		hopped_node, err := ClosestPrecedingFinger_RPC(hopped_node, id)
		immediate_succesor, err = GetSuccessorId_RPC(hopped_node)
	}
	return hopped_node, err // found new node's closest predecessor
}


