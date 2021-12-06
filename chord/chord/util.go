/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: Utility functions to help with dealing with ID hashes in Chord. */
/*                                                                           */

package chord

import (
	"bytes"
	"crypto/sha1"
	"math/big"
)

/* Hash a string to its appropriate size */
func HashKey(key string) []byte {
	h := sha1.New()
	h.Write([]byte(key))
	v := h.Sum(nil)
	return v[:KEY_LENGTH/8]
}

/* Convert a []byte to a big.Int string, useful for debugging/logging */
func HashStr(keyHash []byte) string {
	keyInt := big.Int{}
	keyInt.SetBytes(keyHash)
	return keyInt.String()
}

func EqualIds(a, b []byte) bool {
	return bytes.Equal(a, b)
}

/* Example of how to do math operations on []byte IDs, you may not need this function. */
func AddIds(a, b []byte) []byte {
	aInt := big.Int{}
	aInt.SetBytes(a)

	bInt := big.Int{}
	bInt.SetBytes(b)

	sum := big.Int{}
	sum.Add(&aInt, &bInt)
	return sum.Bytes()
}

/* On this crude ascii Chord ring, X is between (A : B)
   ___
  /   \-A
 |     |
B-\   /-X
   ---
*/
func Between(nodeX, nodeA, nodeB []byte) bool {

	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = false
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) == -1)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) == -1)
	}

	return result
}

/* Is X between (A : B] */
func BetweenRightIncl(nodeX, nodeA, nodeB []byte) bool {

	xInt := big.Int{}
	xInt.SetBytes(nodeX)

	aInt := big.Int{}
	aInt.SetBytes(nodeA)

	bInt := big.Int{}
	bInt.SetBytes(nodeB)

	var result bool
	if aInt.Cmp(&bInt) == 0 {
		result = true
	} else if aInt.Cmp(&bInt) < 0 {
		result = (xInt.Cmp(&aInt) == 1 && xInt.Cmp(&bInt) <= 0)
	} else {
		result = !(xInt.Cmp(&bInt) == 1 && xInt.Cmp(&aInt) <= 0)
	}

	return result
}


/* (n + 2^i) mod (2^m) */
func fingerMath(n []byte, i int, m int) []byte {
	two := &big.Int{}
	two.SetInt64(2)

	N := &big.Int{}
	N.SetBytes(n)

	// 2^i
	I := &big.Int{}
	I.SetInt64(int64(i))
	I.Exp(two, I, nil)

	// 2^m
	M := &big.Int{}
	M.SetInt64(int64(m))
	M.Exp(two, M, nil)

	result := &big.Int{}
	result.Add(N, I)
	result.Mod(result, M)

	// Big int gives an empty array if value is 0.
	// Here is a way for us to still return a 0 byte
	zero := &big.Int{}
	zero.SetInt64(0)
	if result.Cmp(zero) == 0 {
		return []byte{0}
	}

	return result.Bytes()
}

/* Print contents of a node's finger table */
func PrintFingerTable(node *Node) {
	fmt.Printf("[%v] FingerTable:\n", HashStr(node.Id))
	for _, val := range node.FingerTable {
		fmt.Printf("\t{start:%v\tnodeLoc:%v %v}\n",
			HashStr(val.Start), HashStr(val.Node.Id), val.Node.Addr)
	}
}
