package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents an RPC client
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new RPC client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetNewAddress generates a new address
func (c *Client) GetNewAddress() (string, error) {
	resp, err := c.get("/getnewaddress")
	if err != nil {
		return "", err
	}

	var result NewAddressResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Address, nil
}

// GetBalance retrieves wallet balance
func (c *Client) GetBalance() (int64, error) {
	resp, err := c.get("/getbalance")
	if err != nil {
		return 0, err
	}

	var result BalanceResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return 0, err
	}

	return result.Balance, nil
}

// SendToAddress sends coins to an address
func (c *Client) SendToAddress(address string, amount int64) (string, error) {
	reqBody := map[string]interface{}{
		"address": address,
		"amount":  amount,
	}

	resp, err := c.post("/sendtoaddress", reqBody)
	if err != nil {
		return "", err
	}

	var result SendResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return "", err
	}

	return result.TxHash, nil
}

// GetBlockCount retrieves current blockchain height
func (c *Client) GetBlockCount() (uint64, error) {
	resp, err := c.get("/getblockcount")
	if err != nil {
		return 0, err
	}

	var result BlockCountResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return 0, err
	}

	return result.Height, nil
}

// GetBlock retrieves block by height
func (c *Client) GetBlock(height uint64) (*BlockResponse, error) {
	url := fmt.Sprintf("/getblock?height=%d", height)
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result BlockResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTransaction retrieves transaction by hash
func (c *Client) GetTransaction(txHash string) (*TransactionResponse, error) {
	url := fmt.Sprintf("/gettransaction?txhash=%s", txHash)
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}

	var result TransactionResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListAddresses lists all wallet addresses
func (c *Client) ListAddresses() ([]string, error) {
	resp, err := c.get("/listaddresses")
	if err != nil {
		return nil, err
	}

	var result ListAddressesResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Addresses, nil
}

// Helper methods
func (c *Client) get(path string) (*http.Response, error) {
	url := c.baseURL + path
	return c.client.Get(url)
}

func (c *Client) post(path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}

func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var rpcResp Response
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return err
	}

	if rpcResp.Error != "" {
		return fmt.Errorf("RPC error: %s", rpcResp.Error)
	}

	// Convert result to JSON and back to parse into the target type
	resultJSON, err := json.Marshal(rpcResp.Result)
	if err != nil {
		return err
	}

	return json.Unmarshal(resultJSON, result)
}
