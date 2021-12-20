/*
client level lib 
- start Raft and Tapestry cluster 
- implemented with a mix of zookeeper, raft, tapestry APIs
- 1. service registration
- 2. service discovery


Zookeeper usage:
1. hierachy file system mapped to AGUID
2. group membership / service discovery 
3. distributed lock


Q: for zookeeper, 10 path under multilevel == 10 nodes ?? 
Q: zk membership == service discovery, since theres ip id mapping ?? if not what else need to be done ?? 

what is watch event, how to register watch events in go ?? exist() ?? getChildren() ??
oreiily common pattern exam !!!
leader election 



Q: why zookeeper not raft ?
A: zk has 3 new features to build 2 main types of primatives

3 features
- children hierarchy file system
- event notificaiton
- ephemeral n sequential znodes

2 primatives
- distributed struct
   - queue, group membership, config management, DNS, service discovery 
- distributed sync
   - barrier, lock, 2PC, leader election 


https://github.com/apache/zookeeper/blob/master/zookeeper-docs/src/main/resources/markdown/zookeeperTutorial.md 

barrier example:
0. 2+ clients, 1 zk cluster 
1. any 1 client creates persistent root znode
2. every client creates 1 ephemeral child znode under root 
	- empty data
	- znode == not an actual node == "file-like" map item replicated in each cluster node 
	- ephemeral == will timeout
3. every client for loop + mutex blocked on getChildren(true)
	- getChildren() returns all child znodes under root znode 
	- getChildren() is non blocking, therefore needs mutex blocking
	- "watch event" is true == cluster will constantly rpc "event" packets whenever cluster var is changed 
	- mutex unblocks for loop once client receives "event" packet
	- then blocks again, until all children have exited or time out 
4. sync 2 times
	- 1st on all clients create child
	- 2nd on all clients delete child


queue example:
0. 2+ clients, 1 zk cluster 
1. any 1 client creates persistent root znode
2. every client creates any no. of persistent + sequential child znode under /root/element 
	- with any data by client 
	- NOTICE it is /root/element NOT /root
	- sequence flag lets the new znode get a name like queue-N, where N is a monotonically increasing number
3. every client for loop + mutex blocked on getChildren(true)
	- getChildren() returns znode ordered sequentially
	- getData() from FIFO znode
	- delete znode from cluster 



packt book

lock example:
0. each client create a euphumeral + sequential znode /lock/lock-
	- clients create /lock/lock-001,  /lock/lock-002, /lock/lock-003 ...
	- 001, 002 etc is auto assigned since sequential znode
1. client checks if it has lowest seq no.
	- every client calls getChildren("/_locknode_/lock-", false) to check if it holds cluster's lowest seq no znode
		- false == avoids herd effect (but how ??)
	- client that created /lock/lock-001 has the lock
2. if client is not lowest
	- client that created /lock/lock-002 watches /lock/lock-001
	- /lock/lock-003 watches /lock/lock-002
	- exists("/lock/lock-002", True)
3. release lock
	- client holding the lock deletes the node, thereby triggering the next client in line to acquire the lock
	- client that created the next higher sequence node will be notified and get the lock
4. bug
	- https://stackoverflow.com/questions/14275613/concerns-about-zookeepers-lock-recipe?rq=1 
	- packt book
	- If there was a partial failure in the creation of znode due to connection loss, it's
	possible that the client won't be able to correctly determine whether it successfully
	created the child znode. To resolve such a situation, the client can store its session ID
	in the znode data field or even as a part of the znode name itself. As a client retains
	the same session ID after a reconnect, it can easily determine whether the child znode
	was created by it by looking at the session ID.



group membership example: (simpler version of barrier)
0. 2+ clients, 1 zk cluster 
1. any 1 client creates persistent root znode
2. every client creates 1 ephemeral child znode under root 
3. every client for loop + mutex blocked on getChildren(true)
4. when other client joins or leaves, all members are notified 
	- leaves by deleting or timeout ephemeral znode 




https://github.com/apache/zookeeper/blob/master/zookeeper-recipes/zookeeper-recipes-lock/src/main/java/org/apache/zookeeper/recipes/lock/WriteLock.java 
https://github.com/apache/zookeeper/blob/master/zookeeper-recipes/zookeeper-recipes-queue/src/main/java/org/apache/zookeeper/recipes/queue/DistributedQueue.java 
https://github.com/apache/zookeeper/blob/master/zookeeper-recipes/zookeeper-recipes-election/src/main/java/org/apache/zookeeper/recipes/leader/LeaderElectionSupport.java 


NOTE:
tapesrry, raft, oceanstore's structs are all meshed into 1 ps 
tapesrry, raft, oceanstore client == running in same ps as tapesrry, raft, oceanstore server
znode == zk "file-like thing" so they can have children 
*/


////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Puddlestore node related APIs 

/*


*/
func PuddleStoreCreateCluster(){





}





////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////


// Raft node related APIs 
// tapestry n raft seem to have different ways of new node joining
func CreateRaftNode(){


	// 1. connect to zookeeper to update DNS


	// 2. creates raft root dir in zk root znode 

	

	// 3. each node "races" to be zookeeper's "1st raft node" 
	// 8 roles in 1 ps == both client and server of puddle, raft, tapestry, zk 
	// 1st raft node == every newly joined node connects to this node
	// 1st raft node == leader node ?? 

	// 3.1 if your node is 1st one, creates 1st znode in zookeeper w your addr n id under "/1st_raft"
	// then creates raft node, then creates new znode in zookeeper w your addr n id under "/root"


	// 3.2 if some other ps already 1st node 
	// then query its addr, id. then creates raft node connected to that 1st node, 
	// then creates new znode in zookeeper w your addr n id under "/root"


	// 4. register signal handler to close zookeeper connection on SIGINT
	// why ?? where does SIGINT come in ?? cli ?? 


}






/*
query zookeeper DNS to return a Raft client connected to a random Raft node


NOTE: 


*/
func getRaftClient(){

	// 1. rpc zookeeper to get introducer Raft node ip
	
	// 1.1 rpc zookeeper to get all its children path
	
	// 1.2 rpc zookeeper about any random Raft node of its data (addr, id mapping)




	// 2. Raft client connects to the random Raft node 



}




////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

// Tapestry node related APIs 


func CreateTapestryNode(){

	// 1. connect to zookeeper for DNS



	// 2. creates zookeeper's root node 
	// but why here also spwan multiple nodes 



	// 3. look up addr of any random zookeeper in DNS


	// 4. create 1 tapestry node 


	// 5. if 1st tapestry node, register it in zookeeper DNS of tapestry cluster 
	// if already existing tapestry, connect to any random node in the cluster



	// 6. register signal handler to close zookeeper connection on SIGINT
	// why ?? where does SIGINT come in ?? cli ?? 


}






/*
query zookeeper DNS to return a tapestry client connected to a random tapestry node


NOTE: 


zkGetTapestryClient()

*/
func getTapestryClient(){

	// 1. rpc zookeeper to get introducer tapestry node ip
	
	// 1.1 rpc zookeeper to get all its children path
	
	// 1.2 rpc zookeeper about any random tapestry node of its data (addr, id mapping)




	// 2. tapestry client connects to the random tapestry node 



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
private DNS APIs implemented with Zookeeper APIs
https://pkg.go.dev/github.com/go-zookeeper/zk#section-readme 


*/













