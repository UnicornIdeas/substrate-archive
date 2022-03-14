package main

import (
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type MyEventRecords struct {
	types.EventRecords
	Scheduler_Scheduled []ParaInclusion_CandidateIncluded //nolint:stylecheck,golint
}

type ParaInclusion_CandidateIncluded struct {
	Phase types.Phase
	// snip
	Topics []types.Hash
}

func main() {
	api, _ := gsrpc.NewSubstrateAPI("wss://polkadot.api.onfinality.io/public-ws")
	hash, _ := api.RPC.Chain.GetBlockHash(100000)

	meta, _ := api.RPC.State.GetMetadata(hash)

	key, _ := types.CreateStorageKey(meta, "System", "Events", nil, nil)

	fmt.Println(string(key))

	// rawEvents, _ := api.RPC.State.GetStorageRaw(key, hash)

	// events := MyEventRecords{} // Make sure this struct contains all the events that you care about
	// err := types.EventRecordsRaw(*rawEvents).DecodeEventRecords(meta, &events)

	// fmt.Println(err)
	// fmt.Println(events)

	// block, _ := api.RPC.Chain.GetBlock(hash)

	// v, _ := api.RPC.State.GetRuntimeVersion(hash)

	// fmt.Println(v.SpecName)
	// fmt.Println(v.SpecVersion)

}
