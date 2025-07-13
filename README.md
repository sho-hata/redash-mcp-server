# redash-mcp-server
redash-mcp-server is a server that makes the redash API available via the Model Context Protocol (MCP)

## Usage

```json
{
  "mcpServers": {
    "redash-mcp-server": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "REDASH_BASE_URL",
        "-e",
        "REDASH_API_KEY",
        "redash-mcp-server"
      ],
      "env": {
        "REDASH_BASE_URL": "your redash base url here",
        "REDASH_API_KEY": "your redash api key here"
      }
    }
  }
}
```

## Available Tools
- list-queries: List all available queries in Redash
- get-query: Get details of a specific query
- create-query: Create a new query in Redash
- update-query: Update an existing query in Redash
- archive-query: Archive (soft-delete) a query
- list-data-sources: List all available data sources
- execute-query: Execute a query and return results

## Development

```bash
docker build -t redash-mcp-server .
```

### inspect
```bash
npx @modelcontextprotocol/inspector docker run -i --rm redash-mcp-server
```

And set the environment variables (REDASH_BASE_URL, REDASH_API_KEY) in the Inspector.

## License
MIT
