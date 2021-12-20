/*
client level APIs
client == puddlestore sender to rpc puddlestore receiver  


Q: everytime rpc or cache in local struct 

*/

type Client struct {

	// filesystem state e.g. current inode, current file/dir path, root


}




////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////// 



// cd, ls, cp, mkdir, touch, rm, cat


func (c *Client) ls() ([]string, error) {
	
	// 1. rpc zookeeper DNS to get raft, tapestry client

	// 2. rpc raft + tapestry w current dir's inode aguid to get/refresh its children inode addr + typecast as dir (list of aguids)


	// 3. for each dir's aguid, rpc to get each inode_disk, typecast as list of dir_entries 


	// 4. return each dir_entry's name 
	

}



func touch(){
		// 0. check if file already exist 

	// 1. rpc zookeeper DNS to get raft, tapestry client


	// 2. rpc raft + tapestry to add nodes in both clusters 


	// 3. rpc raft + tapestry to add aguid, vguid in parent 


	// 4. fill file n inode struct w aguid, vguid etc. 

}




func mkdir(){


	// 0. check if dir already exist 

	// 1. rpc zookeeper DNS to get raft, tapestry client


	// 2. rpc raft + tapestry to add nodes in both clusters 


	// 3. rpc raft + tapestry to add aguid, vguid in parent 


	// 4. fill dir n inode struct w aguid, vguid etc. 





}





func rm(){

	// 1. rpc zookeeper DNS to get raft, tapestry client

	// 2. rpc raft + tapestry to remove nodes in both clusters 


	// 3. rpc raft + tapestry to remove aguid, vguid in parent 

}



func cd(){

	// 0. rpc zookeeper DNS to get raft, tapestry client

	// 1. absolute path

	// 1.1 parse path 

	// 1.2 recusively raft + tapestry til fill last dir inode



	// 2. relative path 

	// 2.1 raft + tapestry to fill dir inode


}





