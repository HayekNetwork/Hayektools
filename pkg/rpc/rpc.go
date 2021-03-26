package rpc

// https://github.com/ethereum/wiki/wiki/JSON-RPC

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"xcoin/HayekTool/pkg/common"
	"xcoin/HayekTool/pkg/util"
)

var (
	CoinId  = "hyk"
	ChainID = big.NewInt(20210)
)

type RPCClient struct {
	sync.RWMutex
	sickRate         int64
	successRate      int64
	Accepts          int64
	Rejects          int64
	LastSubmissionAt int64
	FailsCount       int64
	Url              string
	login            string
	password         string
	Name             string
	sick             bool
	client           *http.Client
	info             atomic.Value
}

type GetBlockTemplateReply struct {
	Difficulty     int64  `json:"difficulty"`
	Height         int64  `json:"height"`
	Blob           string `json:"blocktemplate_blob"`
	ReservedOffset int    `json:"reserved_offset"`
	PrevHash       string `json:"prev_hash"`
}

type GetInfoReply struct {
	IncomingConnections int64  `json:"incoming_connections_count"`
	OutgoingConnections int64  `json:"outgoing_connections_count"`
	Height              int64  `json:"height"`
	TxPoolSize          int64  `json:"tx_pool_size"`
	Status              string `json:"status"`
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

type GetBlockReply struct {
	Number           string   `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Nonce            string   `json:"nonce"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	TransactionsRoot string   `json:"transactionsRoot"`
	StateRoot        string   `json:"stateRoot"`
	Miner            string   `json:"miner"`
	Difficulty       string   `json:"difficulty"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	ExtraData        string   `json:"extraData"`
	Size             string   `json:"size"`
	GasLimit         string   `json:"gasLimit"`
	GasUsed          string   `json:"gasUsed"`
	Timestamp        string   `json:"timestamp"`
	Transactions     []Tx     `json:"transactions"`
	Uncles           []string `json:"uncles"`
	SealFields       []string `json:"sealFields"`
}

const receiptStatusSuccessful = "0x1"

type TxReceipt struct {
	TxHash            string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	BlockNumber       string `json:"blockNumber"`
	BlockHash         string `json:"blockHash"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress"`
	LogsBloom         string `json:"logsBloom"`
	Status            string `json:"status"`
}

func (r *TxReceipt) Confirmed() bool {
	return len(r.BlockHash) > 0
}

// Use with previous method
func (r *TxReceipt) Successful() bool {
	if len(r.Status) > 0 {
		return r.Status == receiptStatusSuccessful
	}
	return true
}

type Tx struct {
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Hash     string `json:"hash"`
}

func NewRPCClient(name, url string, timeout time.Duration) (*RPCClient, error) {
	rpcClient := &RPCClient{Name: name, Url: url}
	rpcClient.client = &http.Client{}
	if timeout != 0 {
		rpcClient.client.Timeout = timeout
	}
	return rpcClient, nil
}

func (r *RPCClient) GetWork() ([]string, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_getWork", []string{})
	if err != nil {
		return nil, err
	}
	var reply []string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) GetPendingBlock(fullList bool) (*GetBlockReply, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_getBlockByNumber", []interface{}{"pending", fullList})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *GetBlockReply
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) GetLatestBlock(fullList bool) (*GetBlockReply, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_getBlockByNumber", []interface{}{"latest", fullList})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *GetBlockReply
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) GetBlockByHeight(height int64, fullList bool) (*GetBlockReply, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), fullList}
	return r.getBlockBy(CoinId+"_getBlockByNumber", params)
}

func (r *RPCClient) GetBlockByHash(hash string, fullList bool) (*GetBlockReply, error) {
	params := []interface{}{hash, fullList}
	return r.getBlockBy(CoinId+"_getBlockByHash", params)
}

func (r *RPCClient) GetUncleByBlockNumberAndIndex(height int64, index int) (*GetBlockReply, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), fmt.Sprintf("0x%x", index)}
	return r.getBlockBy(CoinId+"_getUncleByBlockNumberAndIndex", params)
}

func (r *RPCClient) getBlockBy(method string, params []interface{}) (*GetBlockReply, error) {
	rpcResp, err := r.doPost(r.Url, method, params)
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *GetBlockReply
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) GetTxReceipt(hash string) (*TxReceipt, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_getTransactionReceipt", []string{hash})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *TxReceipt
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply, err
	}
	return nil, nil
}

func (r *RPCClient) SubmitBlock(params []string) (bool, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_submitWork", params)
	if err != nil {
		return false, err
	}
	var reply bool
	err = json.Unmarshal(*rpcResp.Result, &reply)
	return reply, err
}

func (r *RPCClient) GetBalance(address string) (*big.Int, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_getBalance", []string{address, "latest"})
	if err != nil {
		return nil, err
	}
	var reply string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return nil, err
	}
	return util.String2Big(reply), err
}

func (r *RPCClient) Sign(from string, s string) (string, error) {
	hash := sha256.Sum256([]byte(s))
	rpcResp, err := r.doPost(r.Url, CoinId+"_sign", []string{from, common.ToHex(hash[:])})
	var reply string
	if err != nil {
		return reply, err
	}
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return reply, err
	}
	if util.IsZeroHash(reply) {
		err = errors.New("Can't sign message, perhaps account is locked")
	}
	return reply, err
}

func (r *RPCClient) GetPeerCount() (int64, error) {
	rpcResp, err := r.doPost(r.Url, "net_peerCount", nil)
	if err != nil {
		return 0, err
	}
	var reply string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.Replace(reply, "0x", "", -1), 16, 64)
}

func (r *RPCClient) SendTransaction(from, to, gas, gasPrice, value string, autoGas bool) (string, error) {
	params := map[string]string{
		"from":  from,
		"to":    to,
		"value": value,
	}
	if !autoGas {
		params["gas"] = gas
		params["gasPrice"] = gasPrice
	}
	rpcResp, err := r.doPost(r.Url, CoinId+"_sendTransaction", []interface{}{params})
	var reply string
	if err != nil {
		return reply, err
	}
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return reply, err
	}
	/* There is an inconsistence in a "standard". Galan returns error if it can't unlock signer account,
	 * but Parity returns zero hash 0x000... if it can't send tx, so we must handle this case.
	 * https://xcoin/HayekPool/wiki/JSON-RPC#returns-22
	 */
	if util.IsZeroHash(reply) {
		err = errors.New("transaction is not yet available")
	}
	return reply, err
}

func (r *RPCClient) GetTransactionCount(address, block string) (uint64, error) {
	// QUANTITY|TAG - integer block number, or the string "latest", "earliest" or "pending"

	rpcResp, err := r.doPost(r.Url, CoinId+"_getTransactionCount", []interface{}{
		address, block,
	})
	if err != nil {
		return 0, err
	}

	var reply string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return 0, err
	}

	return util.String2Big(reply).Uint64(), nil
}

func (r *RPCClient) NetVersion() (int, error) {
	rpcResp, err := r.doPost(r.Url, "net_version", []interface{}{})
	if err != nil {
		return 0, err
	}

	var reply string
	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return 0, err
	}

	return int(util.String2Big(reply).Int64()), nil
}

func (r *RPCClient) SendRawTransaction(hexData string) (string, error) {
	rpcResp, err := r.doPost(r.Url, CoinId+"_sendRawTransaction", []interface{}{hexData})
	var reply string
	if err != nil {
		fmt.Printf("RPCClient.SendRawTransaction: err = %v\n", err)
		return reply, err
	}

	err = json.Unmarshal(*rpcResp.Result, &reply)
	if err != nil {
		return reply, err
	}

	if util.IsZeroHash(reply) {
		err = errors.New("transaction is not yet available")
	}
	return reply, err
}

func (r *RPCClient) doPost(url, method string, params interface{}) (*JSONRpcResp, error) {
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "id": 0, "method": method, "params": params}
	data, _ := json.Marshal(jsonReq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(r.login, r.password)
	resp, err := r.client.Do(req)
	if err != nil {
		r.markSick()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}

	var rpcResp *JSONRpcResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		r.markSick()
		return nil, err
	}
	if rpcResp.Error != nil {
		r.markSick()
		return nil, errors.New(rpcResp.Error["message"].(string))
	}
	return rpcResp, err
}

func (r *RPCClient) Check() (bool, error) {
	_, err := r.GetWork()
	if err != nil {
		return false, err
	}
	r.markAlive()
	return !r.Sick(), nil
}

func (r *RPCClient) Sick() bool {
	r.RLock()
	defer r.RUnlock()
	return r.sick
}

func (r *RPCClient) markSick() {
	r.Lock()
	if !r.sick {
		atomic.AddInt64(&r.FailsCount, 1)
	}
	r.sickRate++
	r.successRate = 0
	if r.sickRate >= 5 {
		r.sick = true
	}
	r.Unlock()
}

func (r *RPCClient) markAlive() {
	r.Lock()
	r.successRate++
	if r.successRate >= 5 {
		r.sick = false
		r.sickRate = 0
		r.successRate = 0
	}
	r.Unlock()
}
