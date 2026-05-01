package mongodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"mongoDB_visual/backend/internal/config"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DefaultHost       = "127.0.0.1"
	DefaultPort       = "27017"
	DefaultDatabase   = "admin"
	DefaultAuthSource = "admin"
)

type ConnectionConfig struct {
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
	AuthSource string
}

func DefaultConnectionConfig(cfg *config.Config) ConnectionConfig {
	if cfg == nil {
		return ConnectionConfig{
			Host:       DefaultHost,
			Port:       DefaultPort,
			Database:   DefaultDatabase,
			AuthSource: DefaultAuthSource,
		}
	}

	if cfg.MongoURI != "" {
		if parsed, err := ConnectionConfigFromURI(cfg.MongoURI, cfg.MongoDatabase); err == nil {
			return parsed.Normalize()
		}
	}

	return ConnectionConfig{
		Host:       DefaultHost,
		Port:       DefaultPort,
		Database:   valueOrFallback(cfg.MongoDatabase, DefaultDatabase),
		AuthSource: DefaultAuthSource,
	}
}

func ConnectionConfigFromURI(uri string, database string) (ConnectionConfig, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("parse mongo uri: %w", err)
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if host == "" {
		host = DefaultHost
	}
	if port == "" {
		port = DefaultPort
	}

	username := ""
	password := ""
	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}

	authSource := parsed.Query().Get("authSource")
	pathDatabase := strings.TrimPrefix(parsed.Path, "/")
	if database == "" {
		database = pathDatabase
	}

	return ConnectionConfig{
		Host:       host,
		Port:       port,
		Database:   database,
		Username:   username,
		Password:   password,
		AuthSource: authSource,
	}, nil
}

func (c ConnectionConfig) Normalize() ConnectionConfig {
	return ConnectionConfig{
		Host:       valueOrFallback(strings.TrimSpace(c.Host), DefaultHost),
		Port:       valueOrFallback(strings.TrimSpace(c.Port), DefaultPort),
		Database:   valueOrFallback(strings.TrimSpace(c.Database), DefaultDatabase),
		Username:   strings.TrimSpace(c.Username),
		Password:   c.Password,
		AuthSource: valueOrFallback(strings.TrimSpace(c.AuthSource), DefaultAuthSource),
	}
}

func (c ConnectionConfig) Validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %q", c.Port)
	}
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("database is required")
	}
	return nil
}

func (c ConnectionConfig) URI() string {
	normalized := c.Normalize()

	target := normalized.Host + ":" + normalized.Port
	path := "/" + normalized.Database

	if normalized.Username != "" {
		queryValues := url.Values{}
		queryValues.Set("authSource", normalized.AuthSource)
		return (&url.URL{
			Scheme:   "mongodb",
			User:     url.UserPassword(normalized.Username, normalized.Password),
			Host:     target,
			Path:     path,
			RawQuery: queryValues.Encode(),
		}).String()
	}

	return (&url.URL{
		Scheme: "mongodb",
		Host:   target,
		Path:   path,
	}).String()
}

func (c ConnectionConfig) RedactedURI() string {
	uri := c.URI()
	if !strings.Contains(uri, "@") {
		return uri
	}

	parsed, err := url.Parse(uri)
	if err != nil || parsed.User == nil {
		return uri
	}

	parsed.User = url.UserPassword("***", "***")
	return parsed.String()
}

func (c ConnectionConfig) CacheKey() string {
	normalized := c.Normalize()
	sum := sha256.Sum256([]byte(strings.Join([]string{
		normalized.Host,
		normalized.Port,
		normalized.Database,
		normalized.Username,
		normalized.Password,
		normalized.AuthSource,
	}, "|")))
	return hex.EncodeToString(sum[:])
}

type Client struct {
	raw  *mongo.Client
	conn ConnectionConfig
	mu   sync.Mutex
}

func NewClient(cfg *config.Config) (*Client, error) {
	return NewClientWithConnection(DefaultConnectionConfig(cfg))
}

func NewClientWithConnection(conn ConnectionConfig) (*Client, error) {
	client := &Client{conn: conn.Normalize()}
	if err := client.reconnect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) reconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	raw, err := mongo.Connect(options.Client().ApplyURI(c.conn.URI()))
	if err != nil {
		return err
	}

	if err := raw.Ping(ctx, nil); err != nil {
		_ = raw.Disconnect(context.Background())
		return err
	}

	if c.raw != nil {
		_ = c.raw.Disconnect(context.Background())
	}
	c.raw = raw
	return nil
}

func (c *Client) ensureConnected() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.raw == nil {
		return c.reconnect()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.raw.Ping(ctx, nil); err == nil {
		return nil
	}

	return c.reconnect()
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.raw == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.raw.Disconnect(ctx)
}

func (c *Client) Database(name string) *mongo.Database {
	_ = c.ensureConnected()
	return c.raw.Database(name)
}

func (c *Client) Raw() *mongo.Client {
	_ = c.ensureConnected()
	return c.raw
}

func (c *Client) ConnectionConfig() ConnectionConfig {
	return c.conn
}

func (c *Client) HealthInfo(ctx context.Context) (map[string]any, time.Duration, error) {
	start := time.Now()

	if err := c.ensureConnected(); err != nil {
		return nil, 0, err
	}

	if err := c.raw.Ping(ctx, nil); err != nil {
		return nil, 0, err
	}

	var result bson.M
	err := c.Database(c.conn.Database).RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, 0, err
	}

	return map[string]any{
		"version":  fmt.Sprintf("%v", result["version"]),
		"database": c.conn.Database,
		"server":   c.conn.RedactedURI(),
	}, time.Since(start), nil
}

type Pool struct {
	defaultConn ConnectionConfig
	clients     map[string]*Client
	mu          sync.RWMutex
}

func NewPool(defaultConn ConnectionConfig) *Pool {
	return &Pool{
		defaultConn: defaultConn.Normalize(),
		clients:     make(map[string]*Client),
	}
}

func (p *Pool) DefaultConnection() ConnectionConfig {
	return p.defaultConn
}

func (p *Pool) Get(conn ConnectionConfig) (*Client, error) {
	normalized := conn.Normalize()
	if err := normalized.Validate(); err != nil {
		return nil, err
	}

	key := normalized.CacheKey()

	p.mu.RLock()
	existing := p.clients[key]
	p.mu.RUnlock()
	if existing != nil {
		return existing, nil
	}

	client, err := NewClientWithConnection(normalized)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if existing = p.clients[key]; existing != nil {
		_ = client.Disconnect()
		return existing, nil
	}
	p.clients[key] = client
	return client, nil
}

func (p *Pool) DisconnectAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for key, client := range p.clients {
		if err := client.Disconnect(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(p.clients, key)
	}
	return firstErr
}

func valueOrFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
