package controller

import (
	"net/http"
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
		return true
	},
	EnableCompression: true,
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

var rateLimiter = &RateLimiter{
	clients: make(map[string]*ClientInfo),
}

// IsAllowed checks if a client is allowed to connect
func (r *RateLimiter) IsAllowed(clientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.clients[clientID]
	if !exists {
		r.clients[clientID] = &ClientInfo{
			Connections: 1,
			LastConnect: time.Now(),
		}
		return true
	}

	// Reset counter if more than 1 minute has passed
	if time.Since(client.LastConnect) > time.Minute {
		client.Connections = 1
		client.LastConnect = time.Now()
		return true
	}

	// Limit to 5 connections per minute
	if client.Connections >= 5 {
		return false
	}

	client.Connections++
	client.LastConnect = time.Now()
	return true
}

// WsLiveFlow handles WebSocket connection for live flow visualization
func WsLiveFlow(ctx *gin.Context) {
	logInfo("[LiveFlow] WebSocket connection request from %s", ctx.ClientIP())

	// Get client identifier (IP + token)
	clientID := ctx.ClientIP()
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

	// Rate limiting
	clientID += ":" + token[:8] // Use first 8 chars of token
	if !rateLimiter.IsAllowed(clientID) {
		logInfo("[LiveFlow] Rejected: rate limit exceeded for %s", clientID)
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

	// Send initial batch state
	h.SendBatchState()

	// Keep connection alive until client disconnects
	<-h.Done
	logInfo("[LiveFlow] WebSocket disconnected for %s", ctx.ClientIP())
}
