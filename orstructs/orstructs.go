// Package orstructs defines various structs used throughout the service
package orstructs

import (
	"math/big"
)

// The Payload struct is used to recursively pack and encrypt a request and response
type Payload struct {
	NextNode string
	Payload  []byte
}

// The Node struct represents a node in the router
type Node struct {
	IP           string
	Port         string
	PubX         *big.Int `json:"X"`
	PubY         *big.Int `json:"Y"`
	SharedSecret []byte
}

// Address returns the socket address of node
func (node *Node) Address() string {
	return node.IP + ":" + node.Port
}

// The KeyResponse struct is used to pass a public key in an ECDH key exchange
type KeyResponse struct {
	X *big.Int `json:"X"`
	Y *big.Int `json:"Y"`
}
