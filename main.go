// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")

// Struct for Redash query list

type RedashQuery struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Add other fields as needed
}

type RedashQueryListResponse struct {
	Results []RedashQuery `json:"results"`
	// Add pagination fields if needed
}

// Struct for Redash query detail (simplified, add more fields as needed)
type RedashQueryDetail struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Query string `json:"query"`
	// Add more fields as needed
}

// Redash API client struct

type RedashClient struct {
	BaseURL string
	APIKey  string
}

func NewRedashClientFromEnv() (*RedashClient, error) {
	baseURL := os.Getenv("REDASH_BASE_URL")
	apiKey := os.Getenv("REDASH_API_KEY")
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("REDASH_BASE_URL or REDASH_API_KEY is not set")
	}
	return &RedashClient{BaseURL: baseURL, APIKey: apiKey}, nil
}

func (c *RedashClient) get(endpoint string, out interface{}) error {
	url := c.BaseURL + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Redash API request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Fetch all queries
func (c *RedashClient) GetQueries() ([]RedashQuery, error) {
	var result RedashQueryListResponse
	err := c.get("/api/queries", &result)
	if err != nil {
		return nil, err
	}
	return result.Results, nil
}

// Fetch a specific query by ID
func (c *RedashClient) GetQueryByID(id int) (*RedashQueryDetail, error) {
	var result RedashQueryDetail
	endpoint := fmt.Sprintf("/api/queries/%d", id)
	err := c.get(endpoint, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Args and result for create_query

type CreateQueryArgs struct {
	Name         string `json:"name"`
	Query        string `json:"query"`
	DataSourceID int    `json:"data_source_id"`
}

type CreateQueryResult struct {
	Query *RedashQueryDetail `json:"query"`
}

// RedashClient: create a new query
func (c *RedashClient) CreateQuery(args CreateQueryArgs) (*RedashQueryDetail, error) {
	endpoint := c.BaseURL + "/api/queries"
	body, err := json.Marshal(map[string]interface{}{
		"name":           args.Name,
		"query":          args.Query,
		"data_source_id": args.DataSourceID,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Redash API request failed: %s", resp.Status)
	}
	var result RedashQueryDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Args and result for execute_query

type ExecuteQueryArgs struct {
	ID int `json:"id"`
}

type ExecuteQueryResult struct {
	QueryResult interface{} `json:"query_result"`
}

// RedashClient: execute a query and get result
func (c *RedashClient) ExecuteQuery(id int) (interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/queries/%d/results", c.BaseURL, id)
	body := bytes.NewReader([]byte(`{}`))
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Redash API request failed: %s", resp.Status)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	// Return the query_result field if present
	if qr, ok := result["query_result"]; ok {
		return qr, nil
	}
	return result, nil
}

// Args and result for update_query

type UpdateQueryArgs struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Query        string `json:"query"`
	DataSourceID int    `json:"data_source_id"`
}

type UpdateQueryResult struct {
	Query *RedashQueryDetail `json:"query"`
}

// RedashClient: update an existing query
func (c *RedashClient) UpdateQuery(args UpdateQueryArgs) (*RedashQueryDetail, error) {
	endpoint := fmt.Sprintf("%s/api/queries/%d", c.BaseURL, args.ID)
	body, err := json.Marshal(map[string]interface{}{
		"name":           args.Name,
		"query":          args.Query,
		"data_source_id": args.DataSourceID,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Key "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Redash API request failed: %s", resp.Status)
	}
	var result RedashQueryDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MCP tool to fetch Redash query list

type ListQueriesArgs struct{}

type ListQueriesResult struct {
	Queries []RedashQuery `json:"queries"`
}

func ListQueries(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ListQueriesArgs],
) (*mcp.CallToolResultFor[ListQueriesResult], error) {
	client, err := NewRedashClientFromEnv()
	if err != nil {
		return &mcp.CallToolResultFor[ListQueriesResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create Redash client: %v", err)},
				&mcp.TextContent{Text: `{"queries":[]}`},
			},
		}, nil
	}
	queries, err := client.GetQueries()
	if err != nil {
		return &mcp.CallToolResultFor[ListQueriesResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to fetch queries: %v", err)},
				&mcp.TextContent{Text: `{"queries":[]}`},
			},
		}, nil
	}
	jsonBytes, err := json.Marshal(ListQueriesResult{Queries: queries})
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[ListQueriesResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Fetched %d queries.", len(queries))},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// MCP tool: get_query

type GetQueryArgs struct {
	ID int `json:"id"`
}
type GetQueryResult struct {
	Query *RedashQueryDetail `json:"query"`
}

func GetQuery(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[GetQueryArgs],
) (*mcp.CallToolResultFor[GetQueryResult], error) {
	client, err := NewRedashClientFromEnv()
	if err != nil {
		return &mcp.CallToolResultFor[GetQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create Redash client: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	query, err := client.GetQueryByID(params.Arguments.ID)
	if err != nil {
		return &mcp.CallToolResultFor[GetQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to fetch query: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	jsonBytes, err := json.Marshal(GetQueryResult{Query: query})
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[GetQueryResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Fetched query details."},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// MCP tool: create_query
func CreateQuery(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[CreateQueryArgs],
) (*mcp.CallToolResultFor[CreateQueryResult], error) {
	client, err := NewRedashClientFromEnv()
	if err != nil {
		return &mcp.CallToolResultFor[CreateQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create Redash client: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	query, err := client.CreateQuery(params.Arguments)
	if err != nil {
		return &mcp.CallToolResultFor[CreateQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create query: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	jsonBytes, err := json.Marshal(CreateQueryResult{Query: query})
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[CreateQueryResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Created new query."},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// MCP tool: execute_query
func ExecuteQuery(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ExecuteQueryArgs],
) (*mcp.CallToolResultFor[ExecuteQueryResult], error) {
	client, err := NewRedashClientFromEnv()
	if err != nil {
		return &mcp.CallToolResultFor[ExecuteQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create Redash client: %v", err)},
				&mcp.TextContent{Text: `{"query_result":null}`},
			},
		}, nil
	}
	qr, err := client.ExecuteQuery(params.Arguments.ID)
	if err != nil {
		return &mcp.CallToolResultFor[ExecuteQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to execute query: %v", err)},
				&mcp.TextContent{Text: `{"query_result":null}`},
			},
		}, nil
	}
	jsonBytes, err := json.Marshal(ExecuteQueryResult{QueryResult: qr})
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[ExecuteQueryResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Executed query and fetched result."},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// MCP tool: update_query
func UpdateQuery(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.CallToolParamsFor[UpdateQueryArgs],
) (*mcp.CallToolResultFor[UpdateQueryResult], error) {
	client, err := NewRedashClientFromEnv()
	if err != nil {
		return &mcp.CallToolResultFor[UpdateQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to create Redash client: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	query, err := client.UpdateQuery(params.Arguments)
	if err != nil {
		return &mcp.CallToolResultFor[UpdateQueryResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to update query: %v", err)},
				&mcp.TextContent{Text: `{"query":null}`},
			},
		}, nil
	}
	jsonBytes, err := json.Marshal(UpdateQueryResult{Query: query})
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResultFor[UpdateQueryResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Updated query."},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

func main() {
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{Name: "greeter"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "list_queries", Description: "Get a list of Redash queries"}, ListQueries)
	mcp.AddTool(server, &mcp.Tool{Name: "get_query", Description: "Get details of a specific Redash query"}, GetQuery)
	mcp.AddTool(server, &mcp.Tool{Name: "create_query", Description: "Create a new Redash query"}, CreateQuery)
	mcp.AddTool(server, &mcp.Tool{Name: "execute_query", Description: "Execute a Redash query and return the result"}, ExecuteQuery)
	mcp.AddTool(server, &mcp.Tool{Name: "update_query", Description: "Update an existing Redash query"}, UpdateQuery)

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)

		log.Printf("MCP handler listening at %s", *httpAddr)
		http.ListenAndServe(*httpAddr, handler)
	} else {
		t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
		if err := server.Run(context.Background(), t); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}
}
