package ethutil

import (
	"bytes"
	"context"
	"deployer/internal/logger"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	URL       string
	EthClient *ethclient.Client
	RpcClient *rpc.Client
}

// NewEthClient creates and returns a wrapped Ethereum client with chain ID.
func NewEthClient(ctx context.Context, url string) (*Client, *big.Int, error) {
	// Dial the low-level RPC client once
	rpcClient, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial RPC: %w", err)
	}
	// Create a new Ethereum client using the RPC client
	// This allows us to use the RPC client for other purposes if needed
	ethClient := ethclient.NewClient(rpcClient)

	// Check if the RPC client is connected
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		ethClient.Close()
		return nil, nil, fmt.Errorf("failed to retrieve chainID: %w", err)
	}

	logger.Logger.Info().Str("chainId", chainID.String()).Str("url", url).Msg("Connected to Ethereum node")
	return &Client{URL: url, EthClient: ethClient, RpcClient: rpcClient}, chainID, nil
}

func (c *Client) Close() {
	// Close the RPC client connection
	if c.RpcClient != nil {
		c.RpcClient.Close()
		logger.Logger.Info().Str("url", c.URL).Msg("RPC client connection closed")
	}
	// Close the Ethereum client connection
	if c.EthClient != nil {
		c.EthClient.Close()
		logger.Logger.Info().Str("url", c.URL).Msg("Ethereum client connection closed")
	}
}

// GetRevertReason simulates a failed transaction and extracts the revert reason.
func (c *Client) GetRevertReason(ctx context.Context, from common.Address, txHash common.Hash) ([]byte, error) {
	// Get transaction by hash
	tx, isPending, err := c.EthClient.TransactionByHash(ctx, txHash)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	if isPending {
		return []byte{}, fmt.Errorf("transaction is still pending")
	}

	// Build call message
	callMsg := map[string]interface{}{
		"from": from,
		"to":   tx.To().Hex(),
		"data": fmt.Sprintf("0x%x", tx.Data()),
	}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_call",
		"params":  []interface{}{callMsg, "latest"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return []byte{}, err
	}

	// Make the HTTP POST request to the eth_call endpoint
	resp, err := http.Post("http://localhost:22000", "application/json", bytes.NewReader(body))
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("failed to call eth_call: %s", resp.Status)
	}

	var result struct {
		Result string `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data,omitempty"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []byte{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Error.Message) == 0 && len(result.Error.Data) == 0 {
		if strings.HasPrefix(result.Result, "0x") {
			result.Result = result.Result[2:]
		}
		rawBytes, err := hex.DecodeString(result.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to decode result hex: %w", err)
		}
		return rawBytes, nil
	}

	// Decode the result
	data, err := hexutil.Decode(result.Error.Data)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decode result: %w", err)
	}
	// If the data length is at least 4, try decoding as a custom error first.
	if len(data) >= 4 {
		signature, args, err := DecodeCustomError(data)
		if err == nil {
			if len(args) > 0 {
				argsStr := make([]string, len(args))
				for i, arg := range args {
					argsStr[i] = fmt.Sprintf("%v", arg)
				}
				return []byte(fmt.Sprintf("%s(%s)", signature, strings.Join(argsStr, ", "))), nil
			}
			return []byte(signature), nil
		}
		// If decode failed, fall through to try standard error string below
	}

	// If the data length is less than 64 bytes (and custom error decode failed or length < 4), try parsing as standard error string
	if len(data) < 64 {
		reason, err := parseStandardErrorString(data)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to parse standard error string: %w", err)
		}
		return []byte(reason), nil
	}

	// If the data length is exactly 64 bytes, try parsing as standard error string
	if len(data) == 64 {
		reason, err := parseStandardErrorString(data)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to parse standard error string: %w", err)
		}
		return []byte(reason), nil
	}
	return data, nil
}

func parseStandardErrorString(data []byte) (string, error) {
	if len(data) < 64 {
		return "", errors.New("invalid revert reason data")
	}
	strLen := new(big.Int).SetBytes(data[32:64]).Int64()
	if int64(len(data)) < 64+strLen {
		return "", errors.New("revert reason string too short")
	}
	return string(data[64 : 64+strLen]), nil
}

// DecodeCustomError decodes revert data using selector-to-signature mapping
func DecodeCustomError(revertData []byte) (string, []interface{}, error) {
	if len(revertData) < 4 {
		return "", nil, fmt.Errorf("invalid revert data: too short")
	}

	selector := hexutil.Encode(revertData[:4])
	signature, ok := errorSelectorToSignature[selector]
	if !ok {
		return "", nil, fmt.Errorf("unknown error selector: %s", selector)
	}

	// Extract the parameter types from the signature string
	paramTypes, err := parseParamTypes(signature)
	if err != nil {
		return signature, nil, fmt.Errorf("failed to parse param types: %w", err)
	}

	// Build dynamic ABI arguments
	args := abi.Arguments{}
	for _, typ := range paramTypes {
		abiType, err := abi.NewType(typ, "", nil)
		if err != nil {
			return signature, nil, fmt.Errorf("invalid ABI type %q: %w", typ, err)
		}
		args = append(args, abi.Argument{Type: abiType})
	}

	// Decode the data (after the selector)
	data := revertData[4:]
	values, err := args.Unpack(data)
	if err != nil {
		return signature, nil, fmt.Errorf("failed to unpack error args: %w", err)
	}

	return signature, values, nil
}

// parseParamTypes parses the types from an error signature like "ErrorName(type1,type2)"
func parseParamTypes(signature string) ([]string, error) {
	re := regexp.MustCompile(`\w+\((.*)\)`)
	matches := re.FindStringSubmatch(signature)
	if len(matches) != 2 {
		return nil, fmt.Errorf("invalid signature format: %s", signature)
	}

	paramStr := matches[1]
	if strings.TrimSpace(paramStr) == "" {
		return []string{}, nil // no parameters
	}
	paramTypes := strings.Split(paramStr, ",")
	for i := range paramTypes {
		paramTypes[i] = strings.TrimSpace(paramTypes[i])
	}
	return paramTypes, nil
}
