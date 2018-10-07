package myBlockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Blockchain struct {
	chain               []Block
	currentTransactions []Transaction
	nodes               map[string]bool
}

type Block struct {
	Index        int64         `json:"index"`
	TimeStamp    int64         `json:"timeStamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int64         `json:"proof"`
	PreviousHash string        `json:"prevoius_hash"`
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

type blockchainInfo struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

func (blockchain *Blockchain) NewBlock(proof int64, previousHash string) Block {
	if previousHash == "" {
		previousBlock := blockchain.chain[len(blockchain.chain)-1]
		previousHash = hash(previousBlock)
	}

	block := Block{
		Index:        int64(len(blockchain.chain) + 1),
		TimeStamp:    time.Now().UnixNano(),
		Transactions: blockchain.currentTransactions,
		Proof:        proof,
		PreviousHash: previousHash,
	}

	blockchain.currentTransactions = nil
	blockchain.chain = append(blockchain.chain, block)
	return block
}

func (blockchain *Blockchain) NewTransaction(transaction Transaction) int64 {
	blockchain.currentTransactions = append(blockchain.currentTransactions, transaction)
	return blockchain.LastBlock().Index + 1
}

func (blockchain *Blockchain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !blockchain.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (blockchain *Blockchain) ValidProof(lastProof, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	guessHash := fmt.Sprintf("%s", sha256.Sum256([]byte(guess)))
	return guessHash[:2] == "00"
}

func hash(block Block) string {
	var buf bytes.Buffer
	jsonblock, marshalErr := json.Marshal(block)
	if marshalErr != nil {
		log.Fatalf("Could not marshal block: %s", marshalErr.Error())
	}
	hashingErr := binary.Write(&buf, binary.BigEndian, jsonblock)
	if hashingErr != nil {
		log.Fatalf("Could not hash block: %s", hashingErr.Error())
	}
	return fmt.Sprintf("%s", sha256.Sum256(buf.Bytes()))
}

func (blockchain *Blockchain) LastBlock() Block {
	return blockchain.chain[len(blockchain.chain)-1]
}

func (blockchain *Blockchain) ValidateChain(chain *[]Block) bool {
	lastBlock := (*chain)[0]
	currentIndex := 1
	for currentIndex < len(*chain) {
		block := (*chain)[currentIndex]
		if block.PreviousHash != hash(lastBlock) {
			return false
		}
		if !blockchain.ValidProof(lastBlock.Proof, block.Proof) {
			return false
		}
		lastBlock = block
		currentIndex += 1
	}
	return true
}

func (blockchain *Blockchain) ResolveConflicts() bool {
	neighbours := blockchain.nodes
	newChain := make([]Block, 0)

	maxLength := len(blockchain.chain)
	for k, _ := range neighbours {
		response, err := http.Get(fmt.Sprintf("http://%s/chain", k))
		defer response.Body.Close()
		var bi blockchainInfo
		if err == nil && response.StatusCode == http.StatusOK {
			body, _ := ioutil.ReadAll(response.Body)
			parseErr := json.Unmarshal(body, &bi)
			if parseErr != nil {
				continue
			}
		}
		if bi.Length > maxLength && blockchain.ValidateChain(&bi.Chain) {
			maxLength = bi.Length
			newChain = bi.Chain
		}
	}
	if len(newChain) > 0 {
		blockchain.chain = newChain
		return true
	}
	return false
}

func (blockchain *Blockchain) RegisterNode(address string) bool {
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	_, find := blockchain.nodes[u.Host]
	blockchain.nodes[u.Host] = true
	return !find
}

func NewBlockchain() *Blockchain {
	newBlockchain := &Blockchain{
		chain:               make([]Block, 0),
		currentTransactions: make([]Transaction, 0),
		nodes:               make(map[string]bool),
	}
	newBlockchain.NewBlock(100, "1")
	return newBlockchain
}
