/*
client lib to connect and CRUD a running raft cluster 
client is a seperate ps on a seperate cpu 



Q&A:

Q: SequenceNum for vector clock ?
A: No, this is not multicast where multiple nodes msgs need to be synced 
just to avoid duplicated request 

*/




type Client struct {
	Id          uint64           // Client ID, determined by the Raft node
	Leader      *raft.RemoteNode // Raft node we're connected to (also last known leader)
	SequenceNum uint64           // Sequence number of the latest request sent by the client
}


/*
client ps connects to an existing raft cluster 

*/
func Connect(){

	// 1. send "connect" request to any node



	// 2. reply logic 


}





/*
used by client to CRUD raft cluster 
sent to last known raft leader 

*/
func SendRequest(command uint64, data []byte){

	// 1. send all types of request w data to leader 




	// 2. reply logic 
	// deal w changed leader 



}


