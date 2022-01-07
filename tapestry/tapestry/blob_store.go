/*
Blob Store == storing actual file in RAM
the thing that ate Cincinnati, Cleveland, or whatever

A binary large object (BLOB) data type was introduced to describe data 
not originally defined in traditional computer database systems, 
particularly because it was too large to store practically at the time 
the field of database systems was first being defined in the 1970s and 1980s. 
The data type became practical when disk space became cheap


NOTE: 
- granted its own class since has channel implemented in it
channel == once file deleted in file node, location mapping to be delete in virtual root node

- if published key has been timed out, WILL NOT invalidize file !!!!
it only means query routed to that replica will not find {object hash : real node ip} atm
but will be available later since key is republished periodically

*/

// filename or object hash ?? 
type BlobStore struct {

	filename_file_map map[string]Blob // {k: filename, v: file}
	sync.RWMutex

}

type Blob struct {

	file_bytes []byte		
	
	// length == no. of replicas	
	deleted_replicas []chan bool // WHOLE REASON DEDICATED A CLASS to a map
}


func (bs *BlobStore) Get(key string) ([]byte, bool) {

	blob, mapping_exists := bs.filename_file_map[key]
	if mapping_exists {
		return blob.file_bytes, true // true == mapping exist
	}else{
		return nil, false
	}
}


func (bs *BlobStore) Add(key string, value []byte, new_replicas_delete_channels []chan bool){
	bs.Lock()
	defer bs.Unlock()
	
	// 1. 
	blob, mapping_exists := bs.filename_file_map[key]
	
	// 2. broadcast "delete file" event to subscriber
	// subscriber == root node's {object hash:real node} mapping
	if mapping_exists { 		
		for _ , deleted_replica := range(blob.deleted_replicas) { 
			// once file deleted in file node, location mapping to be delete in virtual root node
			deleted_replica <- true 
		}
	}
	// 3. 
	bs.filename_file_map[key] = Blob(value, new_replicas_delete_channels)
}


func (bs *BlobStore) Delete(){
	// 1. 
	blob, mapping_exists := bs.filename_file_map[key]

	// 2. 
	if mapping_exists { 		
		for _ , deleted_replica := range(blob.deleted_replicas) { 
			deleted_replica <- true 
		}
	}
	// 3. 
	delete(bs.filename_file_map, key) // golang map lib
}


func DeleteAll(){

	for k,v := range bs.filename_file_map{
		for _ , deleted_replica := range(v.deleted_replicas) { 
			deleted_replica <- true 
		}
		delete(bs.filename_file_map, k)
	}
}