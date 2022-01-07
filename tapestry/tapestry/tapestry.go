/*
handler implementation

tapestry == shared prefix routing 

height: 1 level == each digit of node id 
width: base-16 number == 16 slots for each level 

e.g. 3f93 (base 16)
0th level: slot 3
1st level: slot f (16)
2nd level: slot 9
3rd level: slot 3

level: object hash's shared prefix with entry

key APIs:
- Add()
- Publish()
- Get()
- Store()
- Join()
*/

type TapestryNode struct {
	
	local RemoteNode	

	// 4 major data struct == constitutes tapestry
	routing_table 		RoutingTable 
	backpointer_table 	BackpointerTable // point back, doubly linked list 
	location_map 		LocationMap // virtual real ip mapping
	blob_store 			Blobstore // file

	// rpc channels, 1 for each type of packet
}


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Register() handler implementation

/*
add {object hash : real node ip} to root n non-root nodes for path caching

why non root node can also store real virtual map ?
- along the path of routing 
- client each hop check location map even if not virutal root node


NOTE: 
- key: filename
- value: real file node's ip
- we dont send RPC == we are advertisers == root + non-root nodes that received rpc from file node
- file node needs to keep "advertising it is alive" to us 
- kv delete itself if times out, timer logic is implemented in location map class 
*/
func (tap *TapestryNode) Register(filename string, real_file_node RemoteNode) root bool{

	root = false 

	// 1. write to our local real virtual node map
	// kick off a timer to remove the node if it's not advertised again after a set amount of time	
	tap.location_map.Add(filename, real_file_node)

	
	// 2. whether we are virtual root, as non root node also store real virtual map
	if tap.routing_table.GetClosestPrefixMatchesNode(utils.Hash(filename)) == tap.node {
		root = true 
	}

	return
}


/*
a sender file node rpc you all of files it has

*/
func (tap *TapestryNode) RegisterAll(sender_file_node RemoteNode, files map[string][]RemoteNode){

	// 1. call Add() multiple times 
	for k,v := range files{
		tap.Add(k, v)
	}

	// 2. add sender to our local routinig table 
	tap.routing_table.Add(sender_file_node)

}

////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Publish() implementation

/*

we, as filenode, publish what files we have to tapestry cluster 
== register IPs of real nodes that contain the files in root virtual node 
+ register IPs of real nodes on all nodes' routing table along the path 


key: object hash (salted key/file name)
value: real file node's ip (returned by virtual root node)

*/
func (tap *TapestryNode) Publish(filename string){

	// 0. build a func
	threaded_publish := func(filename string) {
		// 1. for loop rpc nodes in cluster until get to key's root node
		virtual_root_node, path := tap.findRoot(utils.Hash(filename))


		// 2. rpc virtual root node + each node along the path  	
		// to register map {file name : real file node ip} in their location map
		// i.e. rpc node 4377 {4378's original file name : 4228's original ip}
		// 4378 == hash of original file name, 4228 == hash of original ip
		virtual_root_node.RegisterRPC(filename, tap.local)
		for node := range path{
			node.RegisterRPC(filename, tap.local)
		}
	}


	// 3. periodically repeat step 1 n 2
	// "republish" the key since location map == timer map to maintain freshness	
	kill_publish_key := make(chan bool)
	ticker := time.NewTicker(REPUBLISH_INTERVAL)
	
	go func(){
		for {
			select{
			case <-ticker.C:{
				threaded_publish(filename)
			}
			case <-kill_publish_key:{ // for client to kill this forever running thread
				return
			}						
			}
			
		}
	}
	return
}



/*
we, as filenode, for loop rpc nodes in cluster until get to key's root node

e.g. 
we are 4228 file node 
1st rpc: 4228 -> 43FE
2nd rpc: 43FE -> 4228
3rd rpc: 4228 -> 437A
4th rpc: 437A -> 4228
5th rpc: 4228 -> 4377


root node == node that stores addr of an object 
object hash == hash of filename path n client id

picking node 70f5 as root node for object w a hash of 60f4 == closest in value 

NOTE: 
- carved out so could be heavily tested 
- guaranteed globally there must be 1 closest node, as long as all routing tables are globally accurate

findRoot()
*/
func (tap *TapestryNode) findRoot(starting_node RemoteNode, objectHash string) (node RemoteNode, visited_nodes []RemoteNode) {
		
	
	// 1. for loop rpc node by node til hit the closest root node
	// TEST for root node: if object hash closest match is remote node itself 
	// TODO: bad nodes == err from rpc calls
	for {
		node, is_virtual_root_node := curr_node.GetClosestPrefixMatchesNodeRPC(objectHash)				
		
		if is_virtual_root_node { // found virtual root node !!!!!!!
			return 
		}
		
		visited_nodes := append(visited_nodes, node)
		curr_node = node
	}
	
	return 
}





////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Store() implementation



/*
store a kv pair into tapestry cluster 

1 file, 3 file nodes, 9 object hashes. (1 file node == 3 object hashes)

replicate kv inputs to ourself + 3 random nodes in our routing table
store a blob on the local node and publish the key to the tapestry


salts == if the root crashes, the item is still accessible.
replicas == 3 random nodes in routing table

1. storing replicas
key: file name
value: file bytes

2. publishing key
key: salt(file name) == object hash
value: real file node's ip (returned by virtual root node)


e.g. node AA93 to store phil's book (4378)
1. AA93 randomly select from its routing table replica node 4228
2. AA93 rpc 4228 n itself to publish object hash + store file locally

NOTE: 
- call by client, who is also original file node
*/
func (tap *TapestryNode) Store(filename string, file_bytes []byte){


	// 1. randomly select 3 nodes in our routing table to create 3 replicas in cluster
	random_nodes := tap.routing_table.GetRandomNodesForReplicas()
	
	// 2 rpc these 3 nodes n ourself to store replicas of kv + publish keys
	random_nodes := append(random_nodes, tap.local)
	go func(){
		for _, node := range random_nodes{
			node.StoreRPC(filename, file_bytes)
		}
	}

}

// called by replica file node
func (tap *TapestryNode) store_handler(filename string, file_bytes []byte){

	// 2.1 
	object_hashes := salts(filename) // 1 file node == 3 object hashes

	// 2.2 publish object hash == update cluster node's routing tables 
	// not threaded even RPC ==> 
	// publish file node ip in root n path nodes before storing file in this file node
	for object_hash := range object_hashes{
		tap.Publish(object_hash)
	}
	
	// 2.3 stores file locally 
	tap.blob_store.Add(filename, file_bytes)
}





////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Get() implementation


/*
get == lookup virtual node + find virtual real file node mapping + file node sending file 

key: filename ?? 


why bad nodes ??

surrogate node == closest digit in routing table
i.e. not exact match but distributedly closest in whole cluster 
there might not be exact match since not every 

replicas == ip of real file nodes that also contains the file
i.e. 4228, AA93 


NOTE:
- 1 filename, 3 file nodes, 3 objects hashes, 3 root nodes
- 3 versions of {filename:file node} at each root node
- 3 versions x 3 root nodes = 9 copies of {filename:file node}
- each root node contains 3 replicas file nodes' ip
- salts == 3 object hashes of 1 file 
*/
func (local *TapestryNode) Get(filename string) file_bytes[] {

	
	// 1. salt the key to get all object hashes
	salts := utils.salts(filename)
	
	// 2. "lookup" virtual root node of the object hash
	// salt -> for loop hop -> replicas	
	for salt := range salts{
		root_node := tap.findRoot(tap.local, salt)
		file_nodes_ips := root_node.GetLocationMapRPC(salt)
		for file_node := range file_nodes_ips{
			file_bytes := file_node.GetFileRPC()
			return
		}
	}	

	// 2.1 for loop hop node by node 

	
	// 2.2 query location mapping at every hop 
	// since 9 copies at 3 root nodes
	

	// 3. contact replica 1 by 1 as soon as anyone returns file 


}


/*
for loop hop til hit root node of object hash
then query that node's virtual real file node mapping 
then return replicas (multiple real file nodes's ips)


NOTE: for loop hop NOT recusrively hop == rpc back result locally first
there wont be exact match 
*/
func (local *TapestryNode) forLoopHopGetReplicas(){

	// 2. for loop hop virtual root node of the object hash	
	// salt -> for loop hop -> replicas	



	// 2.1 for loop hop node by node 
	// rpc closest node on routing table

	
	// 2.2 query that node's virtual real file node mapping 
	// salt -> replicas of real file nodes ip	
	// keep for loop until another's routing table returns itself being closest (surrogate node)



	// 3. return replicas (multiple real file nodes's ips)
}



////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Join() implementation


/*
we, as a new node, try to join an existing tapestry cluster 
called only by new node

- Find the root for our node's ID
- Call AddNode on our root to initiate the multicast and receive our initial neighbor set
- Iteratively get backpointers from the neighbor set and populate routing table

entire or just part of cluster ?? what is the reasoning ??

*/
func (tap *TapestryNode) Join(introducer_node RemoteNode){

	// 1. for loop rpc to get to the "root node of new node" (not object hash)
	// "root node of new node" == transfer new node {real virtual node mapping} 
	// e.g. 437A joins, routed to root node 4377 
	root_node := tap.findRoot(introducer_node, tap.local)


	// 2. recursively hop routing table "prefix matching rows" to find all neighbour nodes
	// 2.1 root node + neighbour nodes add new node's ip to their routing table	
	// 2.2 root node gathered all neigbour nodes, then send back to new node 
	// 2.3 root node + neighbour nodes transfer location mapping to new node	
	// e.g. object hash 437B+ used to routed to 4377, now will be routed to 437A
	// e.g. 437A has map of {object hash 437C : real file node's ip}
	multicast_starting_row := SharedPrefixLength(root_node, tap.local)
	// not threaded since step 3 needs neigbour_nodes
	neigbour_nodes := root_node.AddNodeMulticastRPC(tap.local, multicast_starting_row)
	

	// 3. seed new node's routing table
	// 3.1 add() ALL neigbours to your routing table 	
	for node := range neigbour_nodes{
		tap.routing_table.Add(node)
	}


	// 3.2 add() ONLY K CLOSEST neigbour's backpointers to your routing table (iterative nearest neighbor search)
	// distance == measured by no. prefix matches
	// intuition: neigbour can route to you, backpointers of neigbours can route to backpointers
	// now you can route to neigbour's backpointers 
	// optimal way to seed your routing table with closest set of nodes

	// 3.2.0 repeat this shared prefix no. of times
	num_iterations := utils.SharedPrefixLength(root.Id, local.node.Id)
	for ; num_iterations < 0 ; num_iterations--{

		// 3.2.1 reduce to NUM_CLOSEST_NEIGBOURS closest neigbours (after backpointers added)
		// https://stackoverflow.com/questions/23330781/collect-values-in-order-each-containing-a-map
		
		// construct []{distance:neighbour_node}
		for neigbour := range len(neigbour_nodes){
			distance := SharedPrefixLength(neigbour.Id, tap.local.node.Id)
			
			distance_neigbour_map := map[int] RemoteNode
			distance_neigbour_map[distance] = neigbour

			keys_to_be_sorted := make([]int, len(neigbour_nodes))
			for i, n := range neigbour_nodes {
				keys_to_be_sorted[i] = distance
			}		
		}

		// slice only top NUM_CLOSEST_NEIGBOURS left
		sort.Ints(keys_to_be_sorted)
		neigbour_nodes := []
		for i=0 ; i< NUM_CLOSEST_NEIGBOURS ; i++{
			neigbour_nodes = append(neigbour_nodes, distance_neigbour_map[keys_to_be_sorted[i]])
		}


		// 3.2.2 new node for loop rpc neighbors for their backpointers
		for node := range neigbour_nodes{
			go func(){
				backpointers := node.GetBackpointersRPC()
				backpointers_return <- backpointers
			}
		}
	
		// 3.2.3 add neigbour's backpointers to your routing table n neigbours nodes
		for backpointers := range <-backpointers_return{
			neigbour_nodes = append(neigbour_nodes, backpointers)
			for backpointer := range backpointers{
				tap.routing_table.Add(backpointer)
			}
		}
	}
}


/*
handler run by "need to know nodes" 
need-to-know node of new node == root node + neighbour non root nodes 
neighbour nodes == non root nodes == recusively hop routing table to find nodes w many prefix matches with new node

called only by a need-to-know node 


NOTE: 
- recursive, yes rpc 
- this multicast seems will make every node check object hash registered with itself 
- then transfer its real virtual node mapping to new node

AddNodeMulticast()
*/
func (tap *TapestryNode) AddNodeMulticast(new_node RemoteNode, multicast_starting_row int) neigbour_nodes RemoteNode {

	lower_layer_return := make(chan []RemoteNode)

	// 1. each neigbour node to add routing table + transfer keys to new node
	// 1.1 break condition: bottom row of routing table == contains only ourself	
	if multicast_starting_row == DIGITS{ // when each neigbour node hit bottom level of routing table
		
		go func(){
			// 1.2. add new node to routing table + transfer keys
			tap.routing_table.Add(new_node)
			new_node.TransferKeyRPC(KeysBelongToNewNode(new_node))
		}
		
		neigbour_nodes = append(neigbour_nodes, tap.local)
	
	// 2. only root rpc a list of all neigbour nodes to new node
	} else {
	
		// 2.1 if not level DIGITS, rpc to every nodes 1 level down (neighbour nodes)
		// starts at root's routing table, then multicast to node w same level of prefix matches 
		// (likely only multicast 1 time == since starting at no. of prefix matches row of root node)
		// channel == upper layer needs to wait for all lower layer rpc back, before rpc to upper upper layer		
		neigbour_nodes := tap.routing_table.GetSameLevelNodes(multicast_starting_row)
		for node := range neigbour_nodes{
			go func(){
				neigbour_nodes := node.AddNodeMulticastRPC(new_node, multicast_starting_row + 1)
				lower_layer_return <- neigbour_nodes
			}
		}		
	
		// 2.2 locally recusrively go down 1 level of local routing table til the end
		neigbour_nodes := tap.local.AddNodeMulticast(new_node, multicast_starting_row + 1)
		
		// upper level blocks until lower level returns 
		neigbour_nodes = append(neigbour_nodes, (<-lower_layer_return)...)
	}
	
	// if root node, return == rpc to new node
	// if neigbour node, return == rpc to previous layer neigbour node
	return 
}


func (tap *TapestryNode) KeysBelongToNewNode(new_node RemoteNode){
	
	transfer_to_new_node := make(map[string][]RemoteNode)
	for key, value := range tap.location_map{
		if Hash(key).BetterChoice(remote.Id, local.Id) {
			transfer_to_new_node[key] = slice(values)
			delete(tap.location_map, key)
		}
	}

	return transfer_to_new_node
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
handler run by neigbours node 

new node rpc neigbours to add new node in their backpointers table
*/
func neigboursBackpointsBackNewnode(){

}









func (local *TapestryNode) Unpublish(){

}





func (local *TapestryNode) RouteToObject(){

}

func (local *TapestryNode) RouteToNode(){

}
