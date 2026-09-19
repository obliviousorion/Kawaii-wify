package ipc

import (
	"errors"
	"fmt"
	"net/rpc"
)

type Client struct {
	rpcClient *rpc.Client
}

func NewClient() (*Client, error) {
	conn, err := Dial()
	if err != nil {
		return nil, errors.New("daemon is not running, start it with 'kawaii-wify daemon'")
	}
	return &Client{
		rpcClient: rpc.NewClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.rpcClient != nil {
		return c.rpcClient.Close()
	}
	return nil
}

func (c *Client) GetStatus() (*StatusResponse, error) {
	var resp StatusResponse
	err := c.rpcClient.Call("Daemon.GetStatus", StatusRequest{}, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get daemon status: %w", err)
	}
	return &resp, nil
}

func (c *Client) Connect() (*ActionResponse, error) {
	var resp ActionResponse
	err := c.rpcClient.Call("Daemon.Connect", ActionRequest{}, &resp)
	if err != nil {
		return nil, fmt.Errorf("connect request failed: %w", err)
	}
	return &resp, nil
}

func (c *Client) Disconnect() (*ActionResponse, error) {
	var resp ActionResponse
	err := c.rpcClient.Call("Daemon.Disconnect", ActionRequest{}, &resp)
	if err != nil {
		return nil, fmt.Errorf("disconnect request failed: %w", err)
	}
	return &resp, nil
}

func (c *Client) Stop() (*ActionResponse, error) {
	var resp ActionResponse
	err := c.rpcClient.Call("Daemon.Stop", ActionRequest{}, &resp)
	if err != nil {
		return nil, fmt.Errorf("stop request failed: %w", err)
	}
	return &resp, nil
}
