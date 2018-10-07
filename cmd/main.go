package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/myBlockchain"
)

func main() {
	serverPort := flag.String("port", "8000", "http port number where server will run")
	flag.Parse()
	handler := myBlockchain.NewHandler()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/mine", func(c *gin.Context) {
		minedBlock := handler.Mine()
		c.JSON(200, gin.H{
			"message": "New Block Forged",
			"block":   minedBlock,
		})
	})

	r.POST("/transaction", func(c *gin.Context) {
		var transaction myBlockchain.Transaction
		c.ShouldBindJSON(&transaction)
		index := handler.AddTransaction(transaction)
		c.JSON(201, gin.H{
			"message": fmt.Sprintf("Transaction will be added to Block %d", index),
		})
	})

	r.POST("/nodes", func(c *gin.Context) {
		type PayLoad struct {
			Nodes []string `json:"nodes"`
		}
		var data PayLoad
		c.BindJSON(&data)
		for _, node := range data.Nodes {
			handler.RegisterNode(node)
		}
		c.JSON(201, gin.H{
			"message": "New nodes have been added",
		})
	})

	r.GET("/nodes/resolve", func(c *gin.Context) {
		replaced := handler.ResolveConflicts()
		if replaced {
			c.JSON(200, gin.H{
				"message": "Our chain was replaced",
			})
		} else {
			c.JSON(200, gin.H{
				"message": "Our chain is authoritative",
			})
		}
	})

	r.GET("/chain", func(c *gin.Context) {
		chain := handler.Blockchain()
		c.JSON(200, gin.H{
			"blockchain": chain,
		})
	})

	r.Run(":" + *serverPort)
}
