package metadata

import (
	"fmt"
	"os"
	"path"

	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/itering/substrate-api-rpc/websocket"
)

type MetaClient struct {
	endpoint      string
	metaFilesPath string
}

func NewMetaClient(endpoint, metaFilesPath string) (*MetaClient, error) {
	m := MetaClient{
		endpoint:      endpoint,
		metaFilesPath: metaFilesPath,
	}
	err := m.init()
	return &m, err
}

func (m *MetaClient) init() error {
	err := os.MkdirAll(m.metaFilesPath, os.ModePerm)
	if err != nil {
		return err
	}
	m.initConnection()
	return nil
}

func (m *MetaClient) initConnection() {
	websocket.SetEndpoint(m.endpoint)
	//TODO: set bigger timeout
}

//generates metadata file for a specific block height
func (m *MetaClient) generateMetaFileForBlock(specVersion float64, blockNumber int) {
	v := &rpc.JsonRpcResult{}
	err := websocket.SendWsRequest(nil, v, rpc.ChainGetBlockHash(1, blockNumber))
	if err != nil {
		fmt.Printf("Error getting block hash: [err: %v] [specV: %.0f]\n", err, specVersion)
		return
	}
	hash, _ := v.ToString()

	err = websocket.SendWsRequest(nil, v, rpc.StateGetMetadata(1, hash))
	if err != nil {
		fmt.Printf("Error getting block meta: [err: %v] [specV: %.0f]\n", err, specVersion)
		return
	}
	metaString, _ := v.ToString()

	fullPath := path.Join(m.metaFilesPath, fmt.Sprintf("%.0f", specVersion))
	f, err := os.Create(fullPath)
	if err != nil {
		fmt.Printf("Error creating meta file: [err: %v] [specV: %.0f]\n", err, specVersion)
		return
	}
	f.WriteString(metaString)
	f.Close()
}

//generates meta files for all known spec versions
func (m *MetaClient) GenerateMetaFileForSpecVersions(versionsFirstBlock map[float64]int) {
	for specV, blockH := range versionsFirstBlock {
		m.generateMetaFileForBlock(specV, blockH)
	}
}
