package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpserver "github.com/omgitsads/mcp-go-session-example/mcp"
	"github.com/omgitsads/mcp-go-session-example/storage"
	"github.com/spf13/cobra"
)

// Config holds all configuration for the server
type Config struct {
	// HTTP Server configuration
	Host string `env:"MCP_HOST" envDefault:"localhost"`
	Port int    `env:"MCP_PORT" envDefault:"8080"`

	// Redis configuration
	RedisAddr     string        `env:"REDIS_ADDR"`
	RedisPassword string        `env:"REDIS_PASSWORD"`
	RedisDB       int           `env:"REDIS_DB" envDefault:"0"`
	RedisPrefix   string        `env:"REDIS_PREFIX" envDefault:"mcp:session:"`
	RedisTTL      time.Duration `env:"REDIS_TTL" envDefault:"1h"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the MCP HTTP server with Redis session storage",
	Long: `Start an HTTP server that implements the Model Context Protocol (MCP).
The server uses Redis for session storage to support multi-instance deployments and session persistence.
Redis connection is required - configure via REDIS_ADDR environment variable or --redis-addr flag.`,
	Run: runServer,
}

func init() {
	// HTTP server flags
	serverCmd.Flags().StringP("host", "H", "", "Host to bind to (default from MCP_HOST env or 'localhost')")
	serverCmd.Flags().IntP("port", "p", 0, "Port to listen on (default from MCP_PORT env or 8080)")

	// Redis session storage flags (required)
	serverCmd.Flags().String("redis-addr", "", "Redis address (REQUIRED - default from REDIS_ADDR env)")
	serverCmd.Flags().String("redis-password", "", "Redis password (default from REDIS_PASSWORD env)")
	serverCmd.Flags().Int("redis-db", -1, "Redis database number (default from REDIS_DB env or 0)")
	serverCmd.Flags().String("redis-prefix", "", "Redis key prefix for sessions (default from REDIS_PREFIX env or 'mcp:session:')")
	serverCmd.Flags().Duration("redis-ttl", 0, "Redis session TTL (default from REDIS_TTL env or 1h)")
}

func parseConfig(cmd *cobra.Command) (*Config, error) {
	// Load configuration from environment variables
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	// Override with command line flags if provided
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		cfg.Host = host
	}
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		cfg.Port = port
	}
	if addr, _ := cmd.Flags().GetString("redis-addr"); addr != "" {
		cfg.RedisAddr = addr
	}
	if password, _ := cmd.Flags().GetString("redis-password"); password != "" {
		cfg.RedisPassword = password
	}
	if db, _ := cmd.Flags().GetInt("redis-db"); db >= 0 {
		cfg.RedisDB = db
	}
	if prefix, _ := cmd.Flags().GetString("redis-prefix"); prefix != "" {
		cfg.RedisPrefix = prefix
	}
	if ttl, _ := cmd.Flags().GetDuration("redis-ttl"); ttl != 0 {
		cfg.RedisTTL = ttl
	}

	return &cfg, nil
}

func runServer(cmd *cobra.Command, args []string) {
	// Parse configuration from environment variables and flags
	cfg, err := parseConfig(cmd)
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Validate that Redis is configured
	if cfg.RedisAddr == "" {
		log.Fatal("Redis address is required. Set REDIS_ADDR environment variable or use --redis-addr flag")
	}

	// Create the MCP server instance that will be shared
	sessionServer := mcpserver.NewSessionServer()

	// Configure Redis session storage
	log.Printf("Configuring Redis session storage at %s", cfg.RedisAddr)
	redisStore, err := storage.NewRedisSessionStore(storage.RedisSessionStoreConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		Prefix:   cfg.RedisPrefix,
		TTL:      cfg.RedisTTL,
		Server:   sessionServer.MCPServer,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis session store: %v", err)
	}

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return sessionServer.MCPServer
	}, &mcp.StreamableHTTPOptions{
		SessionStore: redisStore,
	})

	svr := http.Server{
		Addr:    cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Handler: handler,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := svr.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server stopped")
}
