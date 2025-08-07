package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
)

// RedisSessionStore implements StreamableHTTPSessionStore using Redis as the backend
type RedisSessionStore struct {
	client          redis.Client
	prefix          string
	ttl             time.Duration
	server          *mcp.Server                  // Reference to the MCP server for connecting sessions
	activeSessions  map[string]*mcp.SessionState // Active sessions by ID
	activeSessionMu sync.RWMutex
}

var _ mcp.SessionStore = (*RedisSessionStore)(nil)

// RedisSessionStoreConfig holds configuration for the Redis session store
type RedisSessionStoreConfig struct {
	Addr     string        // Redis server address (default: "localhost:6379")
	Password string        // Redis password (default: "")
	DB       int           // Redis database number (default: 0)
	Prefix   string        // Key prefix for session storage (default: "mcp:session:")
	TTL      time.Duration // Session TTL (default: 1 hour)
	Server   *mcp.Server   // Reference to MCP server for connecting sessions
}

// NewRedisSessionStore creates a new Redis-backed session store
func NewRedisSessionStore(config RedisSessionStoreConfig) (*RedisSessionStore, error) {
	// Set defaults
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}
	if config.Prefix == "" {
		config.Prefix = "mcp:session:"
	}
	if config.TTL == 0 {
		config.TTL = time.Hour
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if config.Server == nil {
		return nil, fmt.Errorf("MCP server reference is required")
	}

	store := &RedisSessionStore{
		client:         *client,
		prefix:         config.Prefix,
		ttl:            config.TTL,
		server:         config.Server,
		activeSessions: make(map[string]*mcp.SessionState),
	}

	return store, nil
}

// Get retrieves a session from Redis
func (r *RedisSessionStore) Load(ctx context.Context, sessionID string) (*mcp.SessionState, error) {
	r.activeSessionMu.RLock()
	if session, ok := r.activeSessions[sessionID]; ok {
		r.activeSessionMu.RUnlock()
		return session, nil
	}
	r.activeSessionMu.RUnlock()

	key := r.getKey(sessionID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fs.ErrNotExist // Session not found
		}
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var sessionState mcp.SessionState
	if err := json.Unmarshal([]byte(data), &sessionState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	r.activeSessions[sessionID] = &sessionState

	return &sessionState, nil
}

// Set stores a session in Redis
func (r *RedisSessionStore) Store(ctx context.Context, sessionID string, sessionState *mcp.SessionState) error {
	key := r.getKey(sessionID)

	data, err := json.Marshal(sessionState)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}

	// Store the transport in the active sessions map
	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	r.activeSessions[sessionID] = sessionState

	return nil
}

// Delete removes a session from Redis
func (r *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	key := r.getKey(sessionID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	// Delete from active sessions map
	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	delete(r.activeSessions, sessionID)

	return nil
}

// Close closes the Redis connection
func (r *RedisSessionStore) Close() error {
	return r.client.Close()
}

// getKey generates a Redis key for a session ID
func (r *RedisSessionStore) getKey(sessionID string) string {
	return r.prefix + sessionID
}

// Health checks the health of the Redis connection
func (r *RedisSessionStore) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
