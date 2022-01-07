/*

key: filename (NOT object hash)
value: real file node's ip
phyiscally: virtual root node


NOTE: 
- dedicated a class to implement timer 

*/

package tapestry

import (
	"sync"
	"time"
)

type LocationMap struct{
	// multimap: stores multiple nodes per key, and each node has a timeout
	objecthash_fileIPs_map map[string]map[RemoteNode]*timer.Timer
	sync.Mutex
}


func (lm *LocationMap) Get(filename string) ([]RemoteNode) {
	
	timer_map, exists := objecthash_fileIPs_map[filename]
	if exists {
		for k,v := range timer_map{
			file_nodes_ips = append(file_nodes_ips, k)
		}
		return file_nodes_ips
	} else{
		return nil
	}
}


func (lm *LocationMap) Add(filename string, real_file_node RemoteNode){

	lm.Lock()
	defer lm.Unlock()

	// 1. check if key already exist, if not malloc
	_, exists := ls.objecthash_fileIPs_map[filename]
	if !exists{
		ls.objecthash_fileIPs_map[filename] = make(map[RemoteNode]*timer.Timer)
	}
	
	// 2. add to map + create timer 
	timer, exists := ls.objecthash_fileIPs_map[filename][real_file_node]
	if !exists{
		ls.objecthash_fileIPs_map[filename][real_file_node] = lm.createTimer(key, real_file_node)
	}else{
		timer.Reset(TIMEOUT)
	}	
}



// delete map entry after 10s
func (lm *LocationMap) createTimer(key string, value RemoteNode, replica_real_node_ip RemoteNode, timeout time.Duration){
	
	timerDelete := func{
		lm.Lock()
		defer lm.Unlock()

		timer, exists := ls.objecthash_fileIPs_map[key][replica_real_node_ip] 		
		if exists{ // mapping might be deleted if file is deleted
			timer.Stop()
			delete(ls.objecthash_fileIPs_map[key], replica_real_node_ip)
		}
	}
	
	// THIS IS THE WHOLE REASON THERE IS A CLASS !!!!!!!!!!
	// https://www.geeksforgeeks.org/time-afterfunc-function-in-golang-with-examples/ 
	// wait for 10s, then calls timerDelete
	return time.AfterFunc(timeout, timerDelete)
}



// useless 
func (lm *LocationMap) Delete(key string, replica_real_node_ip RemoteNode){
	lm.Lock()
	defer lm.Unlock()

	delete(ls.objecthash_fileIPs_map[key], replica_real_node_ip)

}

