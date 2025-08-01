module github.com/omgitsads/mcp-go-session-example

go 1.24.5

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/modelcontextprotocol/go-sdk v0.2.0
	github.com/redis/go-redis/v9 v9.0.5
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
)

// replace github.com/modelcontextprotocol/go-sdk => ../go-sdk
// replace github.com/modelcontextprotocol/go-sdk => github.com/omgitsads/go-sdk v0.0.0-20250731090223-ccbedcf20bab
replace github.com/modelcontextprotocol/go-sdk => github.com/joshwlewis/go-sdk v0.0.0-20250731172750-95cb06c54eee
