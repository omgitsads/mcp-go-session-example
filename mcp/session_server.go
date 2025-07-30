package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SessionServer struct {
	MCPServer *mcp.Server
}

func NewSessionServer() *SessionServer {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-go-session-example",
		Version: "1.0.0",
	}, nil)

	ss := &SessionServer{
		MCPServer: server,
	}

	// Add the hello world tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello_world",
		Description: "A simple tool that outputs 'Hello world!'",
	}, ss.handleHelloWorldTool)

	return ss
}

type HelloWorldArgs struct {
	// No arguments needed for this simple tool
}

func (s *SessionServer) handleHelloWorldTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HelloWorldArgs]) (*mcp.CallToolResultFor[any], error) {
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello world!"},
		},
	}, nil
}
