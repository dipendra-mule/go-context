package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Context keys should be unexpected types to avoid collisions

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
)

func main() {
	// Simulate incoming HTTP request
	handleRequest()
}

func handleRequest() {
	// Create context with request metadata
	ctx := context.WithValue(context.Background(), requestIDKey, generateRequestID())
	ctx = context.WithValue(ctx, userIDKey, "user-123")

	// Process order with context carrying request info
	processOrder(ctx, "order-456")

}

func processOrder(ctx context.Context, orderId string) {
	// Log with request context
	requestID := ctx.Value(requestIDKey).(string)
	userID := ctx.Value(userIDKey).(string)

	fmt.Printf("[%s] Processing order %s for user %s\n", requestID, orderId, userID)

	// pass context to downstream services

	validatePayment(ctx, orderId)
	updateInventory(ctx, orderId)
}

func validatePayment(ctx context.Context, orderId string) {
	requestID := ctx.Value(requestIDKey).(string)

	// simulate payment validation
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("[%s] Payment validation for order %s\n", requestID, orderId)
	case <-ctx.Done():
		fmt.Printf("[%s] Payment validation cancelled for order %s\n", requestID, ctx.Err())
	}
}

func updateInventory(ctx context.Context, orderId string) {
	requestID := ctx.Value(requestIDKey).(string)

	// simulate inventory update
	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Printf("[%s] Inventory update for order %s\n", requestID, orderId)
	case <-ctx.Done():
		fmt.Printf("[%s] Inventory update cancelled for order %s\n", requestID, ctx.Err())
	}
}

func generateRequestID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(1000))
}
