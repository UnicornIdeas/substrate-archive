package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-dictionary/internal"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/itering/substrate-api-rpc/rpc"
)

type SpecVersionClient struct {
	specVersionConfigPath string
	knownSpecVersions     []int
	rocksdbClient         internal.RockClient
	endpoint              string
}

// func main() {
// 	nr := 10000
// 	workers := 500
// 	endpoint := "http://127.0.0.1:9933"
// 	msg := make(chan bool)
// 	msg2 := make(chan bool)

// 	t := time.Now()
// 	for i := 0; i < workers; i++ {
// 		go sendRPC(endpoint, msg, msg2)
// 	}

// 	go func() {
// 		idx := 0
// 		for {
// 			<-msg2
// 			idx++
// 			if idx == nr {
// 				fmt.Println(time.Now().Sub(t))
// 			}
// 		}
// 	}()

// 	v := true
// 	for i := 0; i < nr; i++ {
// 		msg <- v
// 	}

// 	for {
// 	}
// }

// func sendRPC(endpoint string, msg chan bool, msg2 chan bool) {
// 	for {
// 		<-msg
// 		reqBody := bytes.NewBuffer([]byte(`{"id":4834,"method":"chain_getRuntimeVersion","params":["0x43bdf907fc14b6a7e4b9f5914b24a414075e71cc5deb0a2b830725d7fdabf418"],"jsonrpc":"2.0"}`))
// 		resp, err := http.Post(endpoint, "application/json", reqBody)
// 		if err != nil {
// 			fmt.Println("An Error Occured: ", err)
// 			return
// 		}
// 		//Read the response body
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			fmt.Println("Eroare raspuns:", err)
// 			return
// 		}
// 		sb := string(body)
// 		log.Printf(sb)
// 		resp.Body.Close()
// 		msg2 <- true
// 	}
// }

func main() {
	// endpoint := "wss://polkadot.api.onfinality.io/public-ws"
	// endpoint := "ws://localhost:9944"
	// metaFP := "./meta_files"
	endpoint := "http://127.0.0.1:9933"
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

	rdbClient, err := internal.OpenRocksdb("/tmp/rocksdb-polkadot/chains/polkadot/db/full")
	if err != nil {
		fmt.Println("Error opening rocksdb:", err)
		return
	}

	specVClient := SpecVersionClient{
		specVersionConfigPath: svConfigFile,
		rocksdbClient:         rdbClient,
		endpoint:              endpoint,
	}

	err = specVClient.init(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(specVClient.knownSpecVersions)

	r, err := specVClient.getFirstBlockForSpecVersion(specVClient.knownSpecVersions[7], 0, 3917385)
	fmt.Println(r, err)
}

func (svc *SpecVersionClient) init(endpoint string) error {
	err := svc.loadConfigFromFile()
	if err != nil {
		return err
	}
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
		s, err := strconv.Atoi(scanner.Text())
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
func (svc *SpecVersionClient) getFirstBlockForSpecVersion(specVersion int, start, end int) (int, error) {
	s := start
	e := end

	for {
		mid := (s + (e - 1)) / 2
		fmt.Println(s, e, mid) //dbg

		if e-1 == s {
			spec, err := svc.getSpecVersion(s)
			if err != nil {
				return -1, err
			}

			if spec == specVersion {
				return s, nil
			}

			return e, nil
		}

		mSpec, err := svc.getSpecVersion(mid)
		if err != nil {
			return -1, err
		}

		if mSpec > specVersion {
			e = mid - 1
			continue
		}

		if mSpec == specVersion {
			beforeMSpec, err := svc.getSpecVersion(mid - 1)
			if err != nil {
				return -1, err
			}

			if beforeMSpec < specVersion {
				return mid, nil
			}

			e = mid - 1
			continue
		}

		if mSpec < specVersion {
			s = mid + 1
		}
	}
}

func (svc *SpecVersionClient) getSpecVersion(height int) (int, error) {
	hash, err := svc.getBlockHash(height)
	if err != nil {
		return -1, err
	}

	msg := fmt.Sprintf(`{"id":1,"method":"chain_getRuntimeVersion","params":["%s"],"jsonrpc":"2.0"}`, hash)
	reqBody := bytes.NewBuffer([]byte(msg))

	resp, err := http.Post(svc.endpoint, "application/json", reqBody)
	if err != nil {
		return 0, err
	}

	v := &rpc.JsonRpcResult{}
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return 0, err
	}

	return v.ToRuntimeVersion().SpecVersion, nil
}

func (svc *SpecVersionClient) getBlockHash(height int) (string, error) {
	lk, err := svc.rocksdbClient.GetLookupKeyForBlockHeight(height)
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(lk[4:])
	return hash, nil
}
