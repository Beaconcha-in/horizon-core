package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type ExplorerClient interface {
	GetBalance(address string) (float64, error)
	GetTransactions(address string) ([]Transaction, error)
	GetPrice(symbol string) (float64, error)
	GetNetworkName() string
}

type Transaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type BaseExplorerClient struct {
	APIKey   string
	BaseURL  string
	Network  string
	client   *http.Client
	mu       sync.Mutex
	lastCall time.Time
}

func NewBaseExplorerClient(baseURL, apiKey, network string) *BaseExplorerClient {
	return &BaseExplorerClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Network: network,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *BaseExplorerClient) RateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	elapsed := time.Since(c.lastCall)
	if elapsed < 200*time.Millisecond {
		time.Sleep(200*time.Millisecond - elapsed)
	}
	c.lastCall = time.Now()
}

func (c *BaseExplorerClient) doRequest(url string) ([]byte, error) {
	c.RateLimit()
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *BaseExplorerClient) GetNetworkName() string {
	return c.Network
}

// Etherscan
type EtherscanClient struct {
	*BaseExplorerClient
}

func NewEtherscanClient() *EtherscanClient {
	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	return &EtherscanClient{
		BaseExplorerClient: NewBaseExplorerClient(
			"https://api.etherscan.io/api",
			apiKey,
			"Ethereum",
		),
	}
}

func (c *EtherscanClient) GetBalance(address string) (float64, error) {
	url := fmt.Sprintf("%s?module=account&action=balance&address=%s&tag=latest&apikey=%s",
		c.BaseURL, address, c.APIKey)
	data, err := c.doRequest(url)
	if err != nil {
		return 0, err
	}
	var result struct {
		Status  string `json:"status"`
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, err
	}
	if result.Status != "1" {
		return 0, fmt.Errorf("etherscan error: %s", result.Message)
	}
	balance, _ := strconv.ParseFloat(result.Result, 64)
	return balance / 1e18, nil
}

func (c *EtherscanClient) GetTransactions(address string) ([]Transaction, error) {
	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=desc&apikey=%s",
		c.BaseURL, address, c.APIKey)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	var result struct {
		Status  string `json:"status"`
		Result  []struct {
			Hash      string `json:"hash"`
			From      string `json:"from"`
			To        string `json:"to"`
			Value     string `json:"value"`
			TimeStamp string `json:"timeStamp"`
		} `json:"result"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("etherscan error: %s", result.Message)
	}
	var txs []Transaction
	for _, tx := range result.Result {
		value, _ := strconv.ParseFloat(tx.Value, 64)
		timestamp, _ := strconv.ParseInt(tx.TimeStamp, 10, 64)
		txs = append(txs, Transaction{
			Hash:      tx.Hash,
			From:      tx.From,
			To:        tx.To,
			Value:     value / 1e18,
			Timestamp: time.Unix(timestamp, 0),
		})
	}
	return txs, nil
}

func (c *EtherscanClient) GetPrice(symbol string) (float64, error) {
	return 0, fmt.Errorf("price not supported by Etherscan")
}

// BscScan, SnowScan, Arbiscan, Optimism را هم می‌توانید به همین الگو اضافه کنید.
// برای اختصار، فقط Etherscan را آوردم.
