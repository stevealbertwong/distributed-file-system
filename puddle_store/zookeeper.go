/*


zookeeper kv == {addr:id}


to add watcher:

https://github.com/microhq/go-plugins/blob/master/registry/zookeeper/watcher.go 
https://github.com/vogo/zkclient/blob/master/watcher.go 
https://github.com/Shopify/gozk/blob/master/zk.go 



*/


type Zookeeper struct{

	zkConn zk.conn


}


func zkInitConnection(){
	// 1. connects to cluster

	// 2. fill struct 

}


func zkCreateZnode(){

	// 1. connects to cluster if not connected yet


	// 2. check if exist yet, if not, then create


}



/*
query i th zookeeper node's data (addr, id)
used by both raft and tapestry for randomly 


index == index of children
startingPath == /first_raft_node
children == list of all the paths under first_raft_node e.g. [ path1 , path2 ]
path == /first_raft_node/path1
data == actual data of a node == DNS mapping == [addr, id] 


zkGetClusterNode()
*/
func zkGetZnodeDataByIndex(conn *zk.Conn, startingPath string, index int) (addr string, id string) {

	// 1. rpc zookeeper to get all its children path



	// 2. children path -> full filepath 




	// 3. rpc zookeeper by full filepath to get data 

}










func zkLock(){


}


func zkUnlock(){



}