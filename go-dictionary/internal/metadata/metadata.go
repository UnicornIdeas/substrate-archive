package metadata

import (
	"github.com/itering/substrate-api-rpc/metadata"
)

func ParseMetadata(metaString string, specVersion int) *metadata.Instant {
	rawMetaStruct := metadata.RuntimeRaw{Spec: specVersion, Raw: metaString}
	instant := metadata.Process(&rawMetaStruct)
	return instant
}
