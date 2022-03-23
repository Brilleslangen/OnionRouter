package orstructs

import (
	"math/big"
)

type Payload struct {
	NextNode string
	Payload  []byte
}

type Node struct {
	IP           string
	Port         string
	PubX         *big.Int `json:"X"`
	PubY         *big.Int `json:"Y"`
	SharedSecret []byte
}

func (node *Node) Address() string {
	return node.IP + ":" + node.Port
}

type KeyResponse struct {
	X *big.Int `json:"X"`
	Y *big.Int `json:"Y"`
}
