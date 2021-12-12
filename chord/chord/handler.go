/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: RPC API implementation, these are the functions that actually   */
/*           get executed on a destination Chord node when a *_RPC()         */
/*           function is called.                                             */
/*                                                                           */

/*
receive, parse packets, return reply
then call chord distributed algo 

*/

package chord

import (
	"bytes"
	"errors"
	"fmt"
)

/* Validate that we're executing this RPC on the intended node */
func validateRpc(node *Node, reqId []byte) error {
	if !bytes.Equal(node.Id, reqId) {
		errStr := fmt.Sprintf("Node ids do not match %v, %v", node.Id, reqId)
		return errors.New(errStr)
	}
	return nil
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
RPC handler 

1. parse req
2. return reply
*/
func (node *Node) GetPredecessorId_Handler(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		reply.Ok = false
		return err
	}
	// Predecessor may be nil, which is okay.
	if node.Predecessor == nil {
		reply.Id = nil
		reply.Addr = ""
		reply.Valid = false
	} else {
		reply.Id = node.Predecessor.Id
		reply.Addr = node.Predecessor.Addr
		reply.Valid = true
	}
	return nil
}



/* 
RPC receiving end handler 

return immedaite successor 

1. parse req
2. return reply
*/
func (node *Node) GetSuccessorId_Handler(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
		reply.Ok = false
		return err
	}
	if node.Successor == nil {
		reply.Id = nil
		reply.Addr = ""
		reply.Valid = false
	} else {
		reply.Id = node.Successor.Id
		reply.Addr = node.Successor.Addr
		reply.Valid = true
	}
	return nil
}


func (node *Node) SetPredecessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		reply.Ok = false
		return err
	}
	node.Predecessor.Id = req.UpdateId
	node.Predecessor.Addr = req.UpdateAddr
	reply.Ok = true
	return nil
}


func (node *Node) SetSuccessorId(req *UpdateReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.FromId); err != nil {
		reply.Ok = false
		return err
	}
	node.Successor.Id = req.UpdateId
	node.Successor.Addr = req.UpdateAddr
	reply.Ok = true
	return nil
}


func (node *Node) Notify_Handler(req *NotifyReq, reply *RpcOkay) error {
	if err := validateRpc(node, query.FromId); err != nil {
		reply.Ok = false
		return err
	}

	predecessor := new(RemoteNode)
	predecessor.Id = req.UpdateId
	predecessor.Addr = req.UpdateAddr
	node.notify(predecessor) // handler implementation 
	reply.Ok = true

	return nil
}


/* 
RPC handler

us successor transfer data to predecessor 

1. loop thru kv map to find data belongs to predecessor 
2. rpc data to successor
3. delete from kv map 

*/
func (node *Node) TransferKeys_Handler(req *TransferReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		reply.Ok = false
		return err
	}
	for k,v := range node.dataStore{
		hashed_key := HashKey(k)
		if BetweenRightIncl(hashed_key, req.PredId, req.FromId){ // FromId == successor 
			Put_RPC(node.Predecessor, k, v)
			delete(node.dataStore, k)
		}
	}
	
	reply.Ok = true

	return nil
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
for get(key) n new_node.join()
*/

/* 
RPC handler

1. parse req
2. recursion: find the 1st successor of local finger table 
2. return reply
*/
func (node *Node) FindSuccessor_Handler(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		reply.Ok = false
		return err
	}
	// check break condition, if not then recurse again
	id = find_closest_successor(query.id) 
	if node.Predecessor == nil {
		reply.Id = nil
		reply.Addr = ""
		reply.Valid = false
	} else {
		reply.Id = node.FingerTable[id].Node.id 
		reply.Addr = node.FingerTable[id].Node.Addr
		reply.Valid = true
	}

	return nil
}


/* 
RPC handler

from down to up loop thru finger table 
-> return biggest successor that is smaller than new node 

1. parse req
2. return reply
*/
func (node *Node) ClosestPrecedingFinger_Handler(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		reply.Ok = false
		return err
	}
	
	for i := KEY_LENGTH; i > 0; i-=1 {		
		if BetweenRightIncl(node.FingerTable[i].Node.Id, node.Id, query.Id) {
			if node.Predecessor == nil { // found new node's closest predecessor 
				reply.Id = nil
				reply.Addr = ""
				reply.Valid = false
			} else { // biggest successor node that is smaller than new node == NOT closest predecessor 
				reply.Id = node.FingerTable[i].Node.id 
				reply.Addr = node.FingerTable[i].Node.Addr
				reply.Valid = true
			}
		}
	}



	return nil
}
