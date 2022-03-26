package connection

import "github.com/itering/substrate-api-rpc/rpc"

type SpecversionClient struct {
	wsclient *WsClient
	receiver chan *rpc.JsonRpcResult
	caller   map[int]chan *rpc.JsonRpcResult //each caller will have a receiving channel
}

func NewSpecVersionClient(endpoint string) (*SpecversionClient, error) {
	chanSize := 100
	receiver := make(chan *rpc.JsonRpcResult, chanSize)
	wsClient, err := InitWSClient(endpoint, receiver)
	if err != nil {
		return nil, err
	}
	caller := make(map[int]chan *rpc.JsonRpcResult, 100)

	svc := SpecversionClient{
		wsclient: wsClient,
		receiver: receiver,
		caller:   caller,
	}
	go svc.readSpecVersions()
	return &svc, nil
}

func (svc *SpecversionClient) readSpecVersions() {
	for {
		rawSpecV := <-svc.receiver
		svc.caller[rawSpecV.Id] <- rawSpecV
	}
}

func (svc *SpecversionClient) GetBlockSpecVersion(blockHeight int, blockHash string) (int, error) {
	responseChan := make(chan *rpc.JsonRpcResult)
	svc.caller[blockHeight] = responseChan
	defer delete(svc.caller, blockHeight)

	svc.wsclient.SendMessage(rpc.ChainGetRuntimeVersion(blockHeight, blockHash))

	specV := <-responseChan
	return specV.ToRuntimeVersion().SpecVersion, nil
}
