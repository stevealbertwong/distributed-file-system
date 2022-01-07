 /*
Q: DHT/routing table is designed to store ONLY a portion of IPs ?? which portions ?? 
Q: will nodes disappear since routing table only store max 3 nodes ?? any heartbeat protocol ?? 
Q: how does each node's routing table get populated in the first place ??
Q: routing table == same if in same network w same root ?? 
Q: how to keep routing table up to date / maintain its freshness ??


Q: why DHT ?? why not a central routing table ??
A: 


Q: consistent hashing vs Chord vs Pastry vs Tapestry
A: 
consistent hashing / DHT
- load balancing using ring + range to the right node
- need to move files when a new node joins 

Chord 
- distributed recursion 
- mod node_id n finger table row no.
- improved consistent hashing to a virtual form where no files need to transfer to node joins/leave 
- just update distributed routing table of where the files are at 
- new node joins == register node's file in distributed finger table

Pastry 
- distributed breadth first search
- instead of finger table 
- hashed to GUID + prefix matching 

Tapestry 
- Pastry but with data very close by, geographically cluster the data 


Q: root node vs node id vs object id 
A: 
object id == hash of filename
node id == hash of ip 
root == another real node with node id the closest to object id BUT does NOT contain the file

e.g. object id 777 physically in node id 123 but virtually in node id 778
node id 778 == root node == does not contains file but contains routing entry of object 777 in node 123



Q: how does a node get its hashed ID ?? a node has it own root node ?? 
   not that any node could route to any other nodes ?? need-to-know nodes ?? 
A: need-to-know nodes == when a new node joins, multiple nodes (ala on same level) could have 
   potential file nodes mappings to send to new node. 
   == routed by diff nodes. no route table contains every root.
   resulting list of nodes == all need to know nodes == other potentially better root nodes 
   that the one you find, have file nodes mappings to send to new node



Q: what is root ?? why node_id n hashed filename could both hashed to root ?? 
A: root is the "closest GUID node"
   root also stores filename's filenode's IP (object store)
   

Q: node_id =/= filename GUID, then why your filename hashed to your node_id to get your own file ??
A: it WON'T since this is DHT's virtual DNS / overlay !!!!!! 
   file is in your node, but your node_id store ONLY mappings of other files from other file nodes
   so when you issue a query on your own file, it would map to other nodes that store your IP


Q: are there file transfer when a new node with files join ??
A: NO file transfer among any nodes when adding new node
   tapestry == file locality to client == file scattered all over the world to shorten distance
   filename == route to "closest GUID node" that knows which "file node" the file is at 
   BUT root DOES NOT store the file itself, it only stores file node's IP (object_store)
   THEREFORE: key == filename, value == nodes that contain the file content (NOT the file itself)
   new node is ONLY getting new file_nodes_IPs mapping


Q: how is which server become file node decided ?? 
A: when a new node joins, it registered all its files on root node !!!! 
   the new node == file node itself
   new node shoulder private ryan's enermies by routing filenode "mappings", NOT transfering files


Q: why there is replicas ?? 



Q: major functions ?? major struct ??
A: implementation of client lib
   1. start_this_tapestry_node()
   2. store_file_into_tapestry_cluster()
   3. query_filename()
   4. get_root_node_of_file()

   1. routing table
   2. object store == mapping of GUID to actual file nodes that contain replicates of the file
   3. back pointers 
   4. blob store == actual file chunks in disk 


Q: why timeout ??


Q: why backpointers ?? reverse reference ?? replicate of routetable ?? 
A: routetable == nodes that you point to 
   backpointers == routetable but nodes that points at you (you dont point to them)
   for notifying remote nodes when your node add / remove nodes in your routing table
   


Q: gossip membership, multicast snapshot vs tapestry ?? 





Q: what is publishing path ? why Node A publishes O to node E
publishing path == A -> B -> C -> D -> E
see: how to get A 


Q: root vs path vs neigbour vs need to know nodes 




















*/