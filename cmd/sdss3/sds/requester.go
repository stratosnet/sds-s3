package sds

import (
	"bytes"
	"encoding/json"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/rpc"
	"io"
	nethttp "net/http"
)

type Requester func(param interface{}, res any, rpcCmd string) error

func GetHttpRequester(rpcUrl string) Requester {
	return func(param interface{}, res any, rpcCmd string) error {
		var params []interface{}
		params = append(params, param)
		pm, err := json.Marshal(params)
		if err != nil {
			utils.ErrorLog("failed marshal param for " + rpcCmd)
			return nil
		}

		// wrap to the json-rpc message
		method := HttpRpcNamespace + "_" + rpcCmd
		request := wrapJsonRpc(method, pm)

		if len(request) < 300 {
			utils.DebugLog("--> ", string(request))
		} else {
			utils.DebugLog("--> ", string(request[:230]), "... \"}]}")
		}

		// http post
		req, err := nethttp.NewRequest("POST", rpcUrl, bytes.NewBuffer(request))
		if err != nil {
			return err
		}
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")

		client := &nethttp.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) < 300 {
			utils.DebugLog("<-- ", string(body))
		} else {
			utils.DebugLog("<-- ", string(body[:230]), "... \"}]}")
		}

		resp.Body.Close()

		if body == nil {
			utils.ErrorLog("json marshal error")
			return err
		}

		// handle rsp
		var rsp jsonrpcMessage
		err = json.Unmarshal(body, &rsp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(rsp.Result, res)
		if err != nil {
			utils.ErrorLog("unmarshal failed")
			return err
		}
		return nil
	}
}

func GetIpcRequester(rpcClient *rpc.Client) Requester {
	return func(params interface{}, res any, ipcCmd string) error {
		method := IpcNamespace + "_" + ipcCmd
		err := rpcClient.Call(res, method, params)
		if err != nil {
			return err
		}
		return nil
	}
}

func wrapJsonRpc(method string, param []byte) []byte {
	// compose json-rpc request
	request := &jsonrpcMessage{
		Version: "2.0",
		ID:      1,
		Method:  method,
		Params:  json.RawMessage(param),
	}
	r, e := json.Marshal(request)
	if e != nil {
		utils.ErrorLog("json marshal error", e)
		return nil
	}
	return r
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      int             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}
