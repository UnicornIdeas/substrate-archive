package main

import (
	"bufio"
	"fmt"
	"go-dictionary/internal/connection"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/itering/substrate-api-rpc/websocket"
)

type SpecVersionClient struct {
	svRpcClient           *connection.SpecversionClient
	hashRpcClient         *connection.BlockHashClient
	specVersionConfigPath string
	knownSpecVersions     []float64
}

func main() {
	endpoint := "wss://polkadot.api.onfinality.io/public-ws"
	// metaFP := "./meta_files"
	svConfigFile := "./spec_version_files/config"

	// m, err := metadata.NewMetaClient(endpoint, metaFP)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// metaV := map[float64]int{0: 0, 1: 50000}
	// m.GenerateMetaFileForSpecVersions(metaV)

	// rv := &rpc.JsonRpcResult{}
	// _ = websocket.SendWsRequest(nil, rv, rpc.ChainGetRuntimeVersion(1, hash))
	// versionName := rv.Result.(map[string]interface{})["specVersion"].(float64)

	specVClient := SpecVersionClient{
		specVersionConfigPath: svConfigFile,
	}

	err := specVClient.init(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hash, err := specVClient.hashRpcClient.GetBlockHash(idx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(idx, hash)
		}(i)
	}

	wg.Wait()
	fmt.Println(time.Now().Sub(t))
	// fmt.Println(specVClient.knownSpecVersions)

	// specVClient.getFirstBlockForSpecVersion(specVClient.knownSpecVersions[0], 8999999, 9000000)
}

func (svc *SpecVersionClient) init(endpoint string) error {
	err := svc.initConnection(endpoint)
	if err != nil {
		return err
	}
	err = svc.loadConfigFromFile()
	if err != nil {
		return err
	}
	return nil
}

func (svc *SpecVersionClient) initConnection(endpoint string) error {
	svRpcClient, err := connection.NewSpecVersionClient(endpoint)
	if err != nil {
		return err
	}
	svc.svRpcClient = svRpcClient

	hashClient, err := connection.NewBlockHashClient(endpoint)
	if err != nil {
		return err
	}
	svc.hashRpcClient = hashClient

	return nil
}

// load all the spec versions from a file
func (svc *SpecVersionClient) loadConfigFromFile() error {
	file, err := os.Open(svc.specVersionConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return err
		}
		svc.knownSpecVersions = append(svc.knownSpecVersions, s)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

//search between start and end block heights the first node for a spec version
func (svc *SpecVersionClient) getFirstBlockForSpecVersion(specVersion float64, start, end int) (int, error) {
	s := start
	// e := end

	v := &rpc.JsonRpcResult{}

	err := websocket.SendWsRequest(nil, v, rpc.ChainGetBlockHash(1, s))
	if err != nil {
		fmt.Printf("Error getting block hash: [err: %v] [specV: %.0f] [node: %d]\n", err, specVersion, s)
		return 0, err
	}
	startHash, _ := v.ToString()

	err = websocket.SendWsRequest(nil, v, rpc.ChainGetRuntimeVersion(1, startHash))
	if err != nil {
		fmt.Printf("Error getting block runtime version: [err: %v] [specV: %.0f] [node: %d]\n", err, specVersion, s)
		return 0, err
	}
	startSpec := v.ToRuntimeVersion().SpecVersion

	fmt.Println(startSpec)

	return 0, nil
}
