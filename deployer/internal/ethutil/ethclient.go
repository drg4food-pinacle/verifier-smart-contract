package ethutil

import (
	"context"
	"deployer/internal/logger"
	"fmt"
	"math/big"

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
