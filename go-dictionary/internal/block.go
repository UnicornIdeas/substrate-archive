package internal

import (
	"go-dictionary/internal/connection"
	"go-dictionary/models"
	"log"
	"strconv"

	"github.com/itering/substrate-api-rpc"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/rpc"
)

const (
	CidAura = 0x61727561
	CidBabe = 0x45424142
)

func GetBlockData(c *connection.WsClient, blockInfo *rpc.JsonRpcResult) {
	c.WriteMessage(1, rpc.ChainGetBlock(blockInfo.Id, blockInfo.Result.(string)))
}

func ExtractBlockData(c *connection.WsClient, blockData *rpc.JsonRpcResult) {
	jsonBlock := blockData.Result.(map[string]interface{})["block"].(map[string]interface{})

	extrinsics := []string{}
	jsonExtrinsics := jsonBlock["extrinsics"].([]interface{})
	for _, jsonExtrinsic := range jsonExtrinsics {
		extrinsics = append(extrinsics, jsonExtrinsic.(string))
	}
	ParseExtrinsics(extrinsics)

	logs := []string{}
	jsonLogs := jsonBlock["header"].(map[string]interface{})["digest"].(map[string]interface{})["logs"].([]interface{})
	for _, jsonLog := range jsonLogs {
		logs = append(logs, jsonLog.(string))
	}
	ParseLogs(logs)
}

func ParseExtrinsics(extrinsics []string) {
	extrinsicsResult := []models.Extrinsic{}

	var blockHeight int

	var specVersion int
	meta := &rpc.JsonRpcResult{}

	metaString, err := meta.ToString()
	if err != nil {
		log.Println(err)
		// Set error somehow
	}
	rawMeta := metadata.RuntimeRaw{Spec: 14, Raw: metaString}
	instant := metadata.Process(&rawMeta)

	decodedExtrinsics, err := substrate.DecodeExtrinsic(extrinsics, instant, specVersion)
	if err != nil {
		log.Println(err)
		// Set error somehow
	}

	for extrinsicPosition, decodedExtrinsic := range decodedExtrinsics {
		extrinsicsResult[extrinsicPosition].Id = strconv.Itoa(blockHeight) + "-" + strconv.Itoa(extrinsicPosition)
		extrinsicsResult[extrinsicPosition].TxHash = decodedExtrinsic["extrinsic_hash"].(string)
		extrinsicsResult[extrinsicPosition].Module = decodedExtrinsic["call_module"].(string)
		extrinsicsResult[extrinsicPosition].Call = decodedExtrinsic["call_module_function"].(string)
		extrinsicsResult[extrinsicPosition].BlockHeight = blockHeight
		// extrinsicsResult[extrinsicPosition].Success = ???? missing success but I don't fucking know how to get it:)
		extrinsicsResult[extrinsicPosition].IsSigned = decodedExtrinsic["extrinsic_hash"] != nil
	}

}

func ParseLogs(logs []string) {
	// logsResults := models.EvmLog{}
}
