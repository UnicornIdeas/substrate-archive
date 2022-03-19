package internal

import (
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/rpc"
)

func ParseMetadata(rawMeta rpc.JsonRpcResult) {
	blockHeight := rawMeta.Id
	specVersion := GlobalStorage.GetSpecVersion(blockHeight)
	metaString, _ := rawMeta.ToString()
	rawMetaStruct := metadata.RuntimeRaw{Spec: specVersion, Raw: metaString}
	instant := metadata.Process(&rawMetaStruct)
	GlobalStorage.PutMetadata(blockHeight, instant)
}
