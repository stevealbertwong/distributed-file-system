/*


APIs (getter, setter):

AddEntry()
RemoveEntry()
GetClosestPrefixMatchesNode()

*/



type RoutingTable struct {
	local RemoteNode
	table [DIGITS][BASE]*[]RemoteNode // height, width, no. nodes each slot == 3D list
	sync.mutex
}




////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// setter


/*
add 1 node to a slot on routing table


node == virtual root node that contains the mapping of real file node 
we are on the path of real file node registering its file

NOTE: 
- threaded rpc 
- addRoute()
*/
func (rt *RoutingTable) Add(node RemoteNode) added bool {
	
	added = false // rpc added node to point back at us
	replaced = false // rpc replaced node to stop pointing back at us 

	// 0. check if table already has entry, or adding ourself 
	if rt.local.Id == node.Id{
		return 
	}

	// 1. find node's level n slot of routing table 
	// compare root node ip w ourself's ip == to get level (no. of matches)
	// last matching digit == slot 
	row := SharedPrefixLength(rt.local.Id, node.Id)
	column := node.Id[row]

	// 2. add or replace (if 3 copies all full) slot 
	
	// 2.1 if node already existed
	slot := rt.table[row][column] 
	for _, slot := range *slot { 
		if slot == node { 
			return
		}
	}
	
	// 2.2 if slot is full
	if len(*slot) == NUM_SLOT{
		// swap the most distant slot
		farthest := node
		for i, slot_node := range *slot {
			if utils.Closer(farthest.id, slot_node.id){
				tmp := *slot[i]
				*slot[i] := farthest 
				farthest := tmp
			}
		}
		if farthest != node{
			added = true // might not be added if new node is the furthest in slot
			replaced = true
		}

	// 2.3 if slot is not full
	}else{
		*slot = append(*slot, node)
		added = true
	}

	// 3. threaded rpc added node + replaced node "banckpointers" 
	// TODO: if rpc failed, add back deleted node
	go func(){

		// 3.1 rpc replaced node to remove backpointer
		if replaced {
			farthest.RemoveBackpointerRPC(local.node)
		}

		// 3.2 rpc added node to add us on their backpointer table == point back at us
		if added {
			node.AddBackpointerRPC(local.node)		
		}
	}

	return
	
}


/*
???????

*/
func (rt *RoutingTable) Delete(node RemoteNode){

	// 0. check if deleting ourself 
	if rt.local.Id == node.Id{
		return 
	}	
	
	// 1. find node to be removed in routing table + remove
	row := utils.SharedPrefixLength(rt.local.Id, node.Id)
	column := node.Id[row]


	// 2. threaded rpc removed node to remove "backpointer" 
	// TODO:

	
}



////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////



// getter 



/*
find the node w most n cloest prefix matches from local routing table

e.g. 
node_id == 3543
rt.table == cs138's example routing table
node w closest prefix matches == 362d
r == row / height / level 
c == column / width / slot


NOTE: 
- no rpc, all happens locally
- can return the local node itself, in such case 
*/
func (rt *RoutingTable) GetClosestPrefixMatchesNode(object_hash [DIGITS]uint8) closest_node RemoteNode, i_am_object_hash_root_node bool {

	rt.Lock() // lock everything even just read since nodes might be updated 
	defer rt.Unlock()

	i_am_object_hash_root_node = false

	// 1. from up to down, traverse to deepest level 
	for r := 0; r < DIGITS; r++ { 

		column := object_hash[r] // 1st digit of object hash is 3

		// 2. from left to right, if no exact match, find the next one to the right
		for c := 0; c < BASE; c++ { 
			
			slot := rt.table[r][column] 
			
			// if empty slot, +1 to the next column, mod so will wrap around
			if len(*slot) == 0 { 
				column := object_hash[ (c + 1) % BASE ] 
				continue
			}

			// 3. pick the closest nodes from the slot
			closest_node = *slot[0]
			for _, node := range *slot{
				if id.Closer(node.Id, closest.id){
					closest_node = node
				} 				
			}
		}

		// 4. at each row, as soon as object hash's digit does not match, cannot proceed to next row
		if object_hash[r] != rt.local.id[r] {
			return closest_node
		}
	}

	// 5. if get to here, closest match must be the owner of routing table itself 
	// == break condition == deepest level == object hash's virtual root node
	i_am_object_hash_root_node = true
	if closest_node != rt.local{
		Debug.Printf("GetClosestPrefixMatchesNode did not return object hash's virtual root node \n")
	}
	

	return 
}


/*
get random nodes from routing table as replicas to store copy of the file 

NOTE: 
- no rpc, all happens locally
- replicas == same file, same object hash, stored in diff real file nodes 
*/
func (rt *RoutingTable) GetRandomNodesForReplicas() []RemoteNode {
	
	// 1. 3D arrays -> 1D array
	for i := range rt.table{
		nodes := rt.GetSameLevelNodes(i)
		all_nodes := append(all_nodes, nodes)
	}
	
	// 2. not enough nodes on routing table
	if len(all_nodes) <= NUM_REPLICATION_NODES{
		return all_nodes
	}

	// 3. randomly pick nodes as replicas 
	random_indices_list := randomIndicesList(len(allNodes)-1)
	for i := 0; i < NUM_REPLICATION_NODES; i++ {
		random_nodes := append(random_nodes, all_nodes[random_indices_list[i]])
	}

	return random_nodes
}


/*
get all the nodes from the same level 

NOTE: 
- no rpc, all happens locally
*/
func (rt *RoutingTable) GetSameLevelNodes(level int) []RemoteNode{
	
	// 1. 
	if level < 0 || level > DIGITS{
		return nil
	}	
	
	// 2. 
	rt.Lock()
	defer rt.Unlock()
	for _, slot := range rt.table[level] {
		for _, node : range *slot { // 3 nodes each slot 
			if node != rt.local{
				nodes := append(nodes, node)
			}
		}  
	}
}










