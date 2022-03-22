package orstructs

type Payload struct {
	NextNode []byte
	Payload  []byte
}

type Node struct {
	IP           string
	Port         string
	PublicKeyX   string
	PublicKeyY   string
	SharedSecret [32]byte
}
