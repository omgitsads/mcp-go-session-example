package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
)

// RedisSessionStore implements StreamableHTTPSessionStore using Redis as the backend
type RedisSessionStore[T mcp.Transport] struct {
	client          redis.Client
	prefix          string
	ttl             time.Duration
	server          *mcp.Server  // Reference to the MCP server for connecting sessions
	activeSessions  map[string]T // Active sessions by ID
	activeSessionMu sync.RWMutex
}

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
func NewRedisSessionStore[T mcp.Transport](config RedisSessionStoreConfig) (*RedisSessionStore[T], error) {
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

	store := &RedisSessionStore[T]{
		client:         *client,
		prefix:         config.Prefix,
		ttl:            config.TTL,
		server:         config.Server,
		activeSessions: make(map[string]T),
	}

	return store, nil
}

// sessionData represents the serializable data for a session
type sessionData struct {
	SessionID string `json:"session_id"`
}

// Get retrieves a session from Redis
func (r *RedisSessionStore[T]) Get(sessionID string) (T, error) {
	var zero T // Zero value for T
	// Check active sessions first
	r.activeSessionMu.RLock()
	if transport, ok := r.activeSessions[sessionID]; ok {
		r.activeSessionMu.RUnlock()
		return transport, nil
	}
	r.activeSessionMu.RUnlock()

	key := r.getKey(sessionID)

	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return zero, nil // Session not found
		}
		return zero, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var sessionData sessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return zero, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	transport := mcp.NewStreamableServerTransport(sessionData.SessionID, nil)
	result, ok := any(transport).(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast transport to expected type: %T", transport)
	}

	// Connect the transport to the MCP server
	if r.server == nil {
		return zero, fmt.Errorf("MCP server reference is nil - this should not happen")
	}

	_, err = r.server.Connect(context.Background(), transport)
	if err != nil {
		return zero, fmt.Errorf("failed to connect session to server: %w", err)
	}

	// Store the transport in the active sessions map
	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	r.activeSessions[sessionID] = result

	return result, nil
}

// Set stores a session in Redis
func (r *RedisSessionStore[T]) Set(sessionID string, transport T) error {
	ctx := context.Background()
	key := r.getKey(sessionID)

	// NOTE: This is a simplified serialization. In a real implementation,
	// you would need to serialize the actual session state properly.
	// The StreamableServerTransport might need additional methods to support
	// serialization, or you might need to store only the essential state.
	sessionData := sessionData{
		SessionID: sessionID,
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}

	// Store the transport in the active sessions map
	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	r.activeSessions[sessionID] = transport

	return nil
}

// Delete removes a session from Redis
func (r *RedisSessionStore[T]) Delete(sessionID string) error {
	ctx := context.Background()
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

func (r *RedisSessionStore[T]) All() (iter.Seq[T], error) {
	return func(yield func(T) bool) {
		r.activeSessionMu.RLock()
		defer r.activeSessionMu.RUnlock()
		for _, session := range r.activeSessions {
			if !yield(session) {
				return
			}
		}
	}, nil
}

func (r *RedisSessionStore[T]) Reset() error {
	r.activeSessionMu.Lock()
	defer r.activeSessionMu.Unlock()
	r.activeSessions = make(map[string]T)
	return nil
}

// Close closes the Redis connection
func (r *RedisSessionStore[T]) Close() error {
	return r.client.Close()
}

// getKey generates a Redis key for a session ID
func (r *RedisSessionStore[T]) getKey(sessionID string) string {
	return r.prefix + sessionID
}

// Health checks the health of the Redis connection
func (r *RedisSessionStore[T]) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
