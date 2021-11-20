/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: RPC API implementation, these are the functions that actually   */
/*           get executed on a destination Chord node when a *_RPC()         */
/*           function is called.                                             */
/*                                                                           */

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
RPC receiving end handler 

1. parse req
2. return reply
*/
func (node *Node) GetPredecessorId_Handler(req *RemoteId, reply *IdReply) error {
	if err := validateRpc(node, req.Id); err != nil {
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


/* 
RPC receiving end handler 

1. parse req
2. return reply
*/
func (node *Node) Notify_Handler(remoteNode *RemoteNode, reply *RpcOkay) error {
	//TODO students should implement this method

	
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
RPC receiving end handler (runs on local node)

1. parse req
2. recursion: find the 1st successor of local finger table 
2. return reply
*/
func (node *Node) FindSuccessor_Handler(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
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
RPC receiving end handler (runs on local node)

from down to up loop thru local finger table 
-> closest entry ("start" just smaller than "id", while "successor" is bigger)



1. parse req
2. return reply
*/
func (node *Node) ClosestPrecedingFinger_Handler(query *RemoteQuery, reply *IdReply) error {
	if err := validateRpc(node, query.FromId); err != nil {
		return err
	}
	
	int current_node_id = node.id
	for (i = KEY_LENGTH, i > 0 , i--) {
		if(node.FingerTable[i].Node.id < query.id  && node.FingerTable[i].Node.id >= current_node_id ){
			// found succusser node that is new node's closest predecessor 
			if node.Predecessor == nil {
				reply.Id = nil
				reply.Addr = ""
				reply.Valid = false
			} else {
				reply.Id = node.FingerTable[i].Node.id 
				reply.Addr = node.FingerTable[i].Node.Addr
				reply.Valid = true
			}
		}
	}



	return nil
}
