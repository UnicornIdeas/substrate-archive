package connection

import (
	"github.com/itering/substrate-api-rpc/rpc"
)

func (c *WsClient) GetBlockHash(blockHeight int) {
	c.WriteMessage(1, rpc.ChainGetBlockHash(blockHeight, blockHeight))

	// response := <-c.receiversList[c.currentIndex]

	// fmt.Println(response)
	// blHash, err := response.ToString()
	// if err != nil {
	// 	return ""
	// }

	// return blHash
}

func (c *WsClient) GetBlockHashes(blockHeight int) {
	for i := 0; i < blockHeight; i++ {
		c.GetBlockHash(i)
	}
}
