package myBlockchain

import (
	"log"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	blockchain *Blockchain
	nodeId     string
}

func NewHandler() Handler {
	handler := Handler{
		blockchain: NewBlockchain(),
		nodeId:     strings.Replace(uuid.New().String(), "-", "", -1),
	}
	return handler
}

func (handler *Handler) Mine() Block {
	log.Println("Mining")
	lastBlock := handler.blockchain.LastBlock()
	lastProof := lastBlock.Proof
	proof := handler.blockchain.ProofOfWork(lastProof)

	// We must receive a reward for finding the proof.
	// The sender is "0" to signify that this node has mined a new coin.
	newTransaction := Transaction{
		Sender:    "0",
		Recipient: handler.nodeId,
		Amount:    1,
	}
	handler.blockchain.NewTransaction(newTransaction)

	block := handler.blockchain.NewBlock(proof, "")
	return block
}

func (handler *Handler) AddTransaction(transaction Transaction) int64 {
	return handler.blockchain.NewTransaction(transaction)
}

func (handler *Handler) RegisterNode(node string) {
	handler.blockchain.RegisterNode(node)
}

func (handler *Handler) ResolveConflicts() bool {
	return handler.blockchain.ResolveConflicts()
}

func (handler *Handler) Blockchain() map[string]interface{} {
	return map[string]interface{}{"chain": handler.blockchain.chain, "length": len(handler.blockchain.chain)}
}
