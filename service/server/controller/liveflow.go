package controller

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/server/jwt"
	"github.com/v2rayA/v2rayA/server/service"
)

var liveFlowUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return isSameOriginOrNonBrowser(r)
	},
	EnableCompression: true,
}

func isSameOriginOrNonBrowser(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser diagnostic clients do not send Origin.
		return true
	}
	originURL, err := url.Parse(origin)
	return err == nil && strings.EqualFold(originURL.Host, r.Host)
}

// RateLimiter limits WebSocket connections per client
type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
}

// ClientInfo stores client connection info
type ClientInfo struct {
	Connections int
	LastConnect time.Time
}

const (
	rateLimitWindow    = time.Minute
	rateLimitRetention = 5 * time.Minute
	rateLimitMaximum   = 5
)

var rateLimiter = &RateLimiter{
	clients: make(map[string]*ClientInfo),
}

// IsAllowed checks if a client is allowed to connect
func (r *RateLimiter) IsAllowed(clientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, client := range r.clients {
		if now.Sub(client.LastConnect) > rateLimitRetention {
			delete(r.clients, id)
		}
	}

	client, exists := r.clients[clientID]
	if !exists {
		r.clients[clientID] = &ClientInfo{
			Connections: 1,
			LastConnect: now,
		}
		return true
	}

	// Reset counter if more than 1 minute has passed
	if now.Sub(client.LastConnect) > rateLimitWindow {
		client.Connections = 1
		client.LastConnect = now
		return true
	}

	// Limit to 5 connections per minute
	if client.Connections >= rateLimitMaximum {
		return false
	}

	client.Connections++
	client.LastConnect = now
	return true
}

// WsLiveFlow handles WebSocket connection for live flow visualization
func WsLiveFlow(ctx *gin.Context) {
	logInfo("[LiveFlow] WebSocket connection request from %s", ctx.ClientIP())

	token := ctx.Query("token")
	if token == "" {
		token = ctx.Query("Authorization")
	}
	if token == "" {
		token = ctx.GetHeader("Authorization")
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}
	}

	logInfo("[LiveFlow] Token source: %s, token present: %v",
		func() string {
			if ctx.Query("token") != "" {
				return "query:token"
			}
			if ctx.Query("Authorization") != "" {
				return "query:Authorization"
			}
			if ctx.GetHeader("Authorization") != "" {
				return "header:Authorization"
			}
			return "none"
		}(), token != "")

	// Validate token
	if token == "" {
		logInfo("[LiveFlow] Rejected: no token provided")
		common.Response(ctx, common.UNAUTHORIZED, "unauthorized: no token")
		return
	}
	if !jwt.ValidateToken(token) {
		logInfo("[LiveFlow] Rejected: invalid token")
		common.Response(ctx, common.UNAUTHORIZED, "unauthorized: invalid token")
		return
	}

	// Rate-limit without retaining or logging the bearer token itself.
	digest := sha256.Sum256([]byte(token))
	clientID := fmt.Sprintf("%s:%x", ctx.ClientIP(), digest[:8])
	if !rateLimiter.IsAllowed(clientID) {
		logInfo("[LiveFlow] Rejected: rate limit exceeded for %s", ctx.ClientIP())
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"code":    "FAIL",
			"message": "rate limit exceeded",
			"data":    nil,
		})
		return
	}

	logInfo("[LiveFlow] Upgrading to WebSocket...")
	conn, err := liveFlowUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logInfo("[LiveFlow] WebSocket upgrade failed: %v", err)
		logError(err)
		return
	}

	logInfo("[LiveFlow] WebSocket connected, starting handler")
	h := service.NewLiveFlowHandler(conn)
	h.Start()
	defer h.Stop()

	// Keep connection alive until client disconnects
	<-h.Done
	logInfo("[LiveFlow] WebSocket disconnected for %s", ctx.ClientIP())
}
