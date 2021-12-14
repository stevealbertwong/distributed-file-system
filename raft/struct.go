type RaftNode struct {


	// core 
	kv_pairs
	state 


	// leader specific  



	// rpc channels, 1 for each type of packet


	// 



}



type RemoteNode struct {



}

type MockRaft struct {



}



/*
written to disk, persist between sessions 

*/
type PersistState struct {


	
}


/*
file wraps over fd that raft node writes its logs into 

*/
type FileData struct {}

