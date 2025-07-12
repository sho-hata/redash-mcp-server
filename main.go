// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
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

// Function to fetch Redash queries
func fetchRedashQueries() ([]RedashQuery, error) {
	baseURL := os.Getenv("REDASH_BASE_URL")
	apiKey := os.Getenv("REDASH_API_KEY")
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("REDASH_BASE_URL or REDASH_API_KEY is not set")
	}

	endpoint := baseURL + "/api/queries"
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Key "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Redash API request failed: %s", resp.Status)
	}

	var result RedashQueryListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Results, nil
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
	queries, err := fetchRedashQueries()
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

func main() {
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{Name: "greeter"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "list_queries", Description: "Get a list of Redash queries"}, ListQueries)

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
