package process

import (
	"go-dictionary/internal/connection"

	"github.com/itering/substrate-api-rpc/rpc"
)

func GetBlockHash(c *connection.WsClient, blockHeight int) {
	c.WriteMessage(1, rpc.ChainGetBlockHash(blockHeight, blockHeight))

	// response := <-c.receiversList[c.currentIndex]

	// fmt.Println(response)
	// blHash, err := response.ToString()
	// if err != nil {
	// 	return ""
	// }

	// return blHash
}

func GetBlockHashes(c *connection.WsClient, blockHeight int) {
	for i := 0; i < blockHeight; i++ {
		GetBlockHash(c, i)
	}
}
