package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	random "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"reflect"
)

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

type KeyResponse struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func (node *Node) address() string {
	return node.IP + ":" + node.Port
}

var nodes []Node
var routerKey *ecdsa.PrivateKey

func main() {

	// Set handlers
	http.HandleFunc("/connect", connectNode)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../html/index.html")
		err := t.Execute(w, nil)
		check(err)
	} else {
		err := r.ParseForm()
		check(err)

		linkAddress := r.Form["code"][0]

		resp := sendThroughNodes("http://" + linkAddress)

		// Print to client
		defer func() {
			err = resp.Body.Close()
			check(err)
		}()
		_, err = io.Copy(w, resp.Body)
		check(err)
	}
}

func connectNode(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		routerKey, _ = ecdsa.GenerateKey(elliptic.P256(), random.Reader)
		// Extract IP-address and key
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		var node Node
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&node)
		check(err)
		node.IP = ip
		node.SharedSecret = establishSharedSecret(node)
		fmt.Println(node.SharedSecret)
		//fmt.Printf("\n IP: %x \n Port: %x \n PublicKey: (%x,%x) \n Shared Secret: %x", ip, node.Port, node.PublicKeyX, node.PublicKeyY, node.SharedSecret)

		x := routerKey.X.Text(10)
		y := routerKey.Y.Text(10)
		response := KeyResponse{X: x, Y: y}
		jsonDetails, err := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err = w.Write(jsonDetails)
		check(err)

		// Add to available nodes
		nodes = append(nodes, node)
	}
}

func sendThroughNodes(url string) *http.Response {
	// Select random nodes and pack payload in encrypted layers
	payload, err := selectAndPack(url)

	// Create request
	request, err := http.NewRequest("POST", "http://"+string(payload.NextNode), bytes.NewBuffer(payload.Payload))
	check(err)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Send request and collect response
	client := http.Client{}
	resp, err := client.Do(request)
	check(err)

	return resp
}

func selectAndPack(url string) (Payload, error) {
	var selectedNodes [3]Node

	// Ensure there are at least three nodes to traverse
	if len(nodes) < 3 {
		return Payload{}, errors.New("there has to be at least three nodes connected")
	}

	// Randomly select three unique nodes
	for i := 0; i < 3; i++ {
		currentNode := nodes[rand.Intn(len(nodes))]
		for _, prevNode := range selectedNodes {
			if reflect.DeepEqual(currentNode, prevNode) {
				i--
				continue
			}
			selectedNodes[i] = currentNode
		}
	}

	// Recursively pack payload
	currentPayload := Payload{[]byte(""), []byte(url)}
	for i := 0; i < 2; i++ {
		// Convert previous payload to a JSON string and pack into new payload
		jsonPayload, err := json.Marshal(currentPayload)
		check(err)
		encryptedPayload, _ := encrypt(selectedNodes[i].SharedSecret[:], jsonPayload)
		currentPayload = Payload{[]byte(selectedNodes[i].address()), encryptedPayload}
	}

	// Pack final payload to be sent from this entity
	jsonFinal, err := json.Marshal(currentPayload)
	check(err)

	return Payload{[]byte(selectedNodes[2].address()), jsonFinal}, nil
}

func establishSharedSecret(node Node) [32]byte {
	x, _ := new(big.Int).SetString(node.PublicKeyX, 10)
	y, _ := new(big.Int).SetString(node.PublicKeyY, 10)
	a, _ := routerKey.PublicKey.Curve.ScalarMult(x, y, routerKey.D.Bytes())
	sharedSecret := sha256.Sum256(a.Bytes())

	return sharedSecret
}
func encode(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

func encrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cfb := cipher.NewCFBEncrypter(block, in)
	cipherText := make([]byte, len(in))
	cfb.XORKeyStream(cipherText, in)
	return []byte(encode(cipherText)), nil
}

func decode(in []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(in))
	check(err)
	return decoded
}

func decrypt(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	check(err)
	cipherText := decode(in)
	cfb := cipher.NewCFBEncrypter(block, in)
	text := make([]byte, len(cipherText))
	cfb.XORKeyStream(text, cipherText)
	return text, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
