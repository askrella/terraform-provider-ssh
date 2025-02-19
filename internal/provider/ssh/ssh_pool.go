package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// SSHPool manages a pool of SSH connections
type SSHPool struct {
	mu       sync.RWMutex
	clients  map[string]*pooledClient
	logger   *logrus.Logger
	maxIdle  time.Duration
	maxConns int
}

type pooledClient struct {
	client    *SSHClient
	lastUsed  time.Time
	inUse     bool
	closeOnce sync.Once
}

// PoolConfig holds configuration for the SSH connection pool
type PoolConfig struct {
	MaxIdleTime time.Duration // Maximum time a connection can be idle before being closed
	MaxConns    int           // Maximum number of connections in the pool
	Logger      *logrus.Logger
}

// NewSSHPool creates a new SSH connection pool
func NewSSHPool(config PoolConfig) *SSHPool {
	if config.MaxIdleTime == 0 {
		config.MaxIdleTime = 5 * time.Minute
	}
	if config.MaxConns == 0 {
		config.MaxConns = 10
	}
	if config.Logger == nil {
		config.Logger = logrus.New()
	}

	pool := &SSHPool{
		clients:  make(map[string]*pooledClient),
		logger:   config.Logger,
		maxIdle:  config.MaxIdleTime,
		maxConns: config.MaxConns,
	}

	// Start cleanup goroutine
	go pool.cleanup()

	return pool
}

// GetClient gets or creates a client for the given configuration
func (p *SSHPool) GetClient(ctx context.Context, config SSHConfig) (*SSHClient, error) {
	ctx, span := otel.Tracer("ssh-provider").Start(ctx, "SSHPool.GetClient")
	defer span.End()

	key := p.configKey(config)

	// Try to get an existing client
	p.mu.Lock()
	defer p.mu.Unlock()

	if pc, exists := p.clients[key]; exists && !pc.inUse {
		// Test if the connection is still alive
		if err := pc.client.sshClient.Conn.Wait(); err == nil {
			pc.inUse = true
			pc.lastUsed = time.Now()
			return pc.client, nil
		}
		// Connection is dead, remove it and create a new one
		delete(p.clients, key)
	}

	// Check if we're at capacity
	if len(p.clients) >= p.maxConns {
		return nil, fmt.Errorf("connection pool is at capacity (max %d connections)", p.maxConns)
	}

	// Create a new client
	client, err := NewSSHClient(ctx, config)
	if err != nil {
		return nil, err
	}

	p.clients[key] = &pooledClient{
		client:   client,
		lastUsed: time.Now(),
		inUse:    true,
	}

	return client, nil
}

// ReleaseClient marks a client as no longer in use
func (p *SSHPool) ReleaseClient(config SSHConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.configKey(config)
	if pc, exists := p.clients[key]; exists {
		pc.inUse = false
		pc.lastUsed = time.Now()
	}
}

// Close closes all connections in the pool
func (p *SSHPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pc := range p.clients {
		pc.closeOnce.Do(func() {
			if err := pc.client.Close(); err != nil {
				p.logger.WithError(err).Error("Failed to close SSH client")
			}
		})
		delete(p.clients, key)
	}
}

// cleanup periodically removes idle connections
func (p *SSHPool) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		now := time.Now()
		for key, pc := range p.clients {
			if !pc.inUse && now.Sub(pc.lastUsed) > p.maxIdle {
				pc.closeOnce.Do(func() {
					if err := pc.client.Close(); err != nil {
						p.logger.WithError(err).Error("Failed to close idle SSH client")
					}
				})
				delete(p.clients, key)
			}
		}
		p.mu.Unlock()
	}
}

// configKey generates a unique key for an SSH configuration
func (p *SSHPool) configKey(config SSHConfig) string {
	return fmt.Sprintf("%s:%d:%s", config.Host, config.Port, config.Username)
}
