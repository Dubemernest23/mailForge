package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"mailForgeApi/internal/config"
	redisclient "mailForgeApi/internal/redisClient"

	"github.com/joho/godotenv"
)

func main() {
	// TODO 1: load .env, same warn-don't-fail pattern as cmd/keycheck/main.go
	// (godotenv.Load(), log.Println a [WARN] if it errors, don't Fatalf here)
	err := godotenv.Load()
	if err != nil {
		log.Println("[warn] Error loading .env")
	}
	// TODO 2: build real config via config.NewInitConfig()

	cfg := config.NewInitConfig()

	// TODO 3: get a real *redis.Client via redisclient.NewRedisClient(cfg)
	// — if this errors, log.Fatalf immediately. Remember NewRedisClient already
	// does a Ping() internally, so an error here means Redis is genuinely unreachable.
	c, err := redisclient.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failure occurred due to: %v", err)
	}

	fmt.Println("[OK] connected to redis")

	// TODO 4: build a short-lived context for the round-trip commands below.
	// Think back to the ctx-cancellation conversation from D3 — this is a
	// one-shot script, not a request handler, so where does this context come from?
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testKey := "redischeck:ping"
	testValue := "pong"

	// TODO 5: SET testKey to testValue, with a short expiry (e.g. 10*time.Second).
	// go-redis's client.Set(ctx, key, value, expiration) — check the method signature.
	// On error, log.Fatalf.
	err = c.Set(ctx, testKey, testValue, 5*time.Second).Err()
	if err != nil {
		log.Fatalf("Failure occurred due to: %v", err)
	}
	fmt.Println("[OK] SET succeeded")

	// TODO 6: GET testKey back and confirm it equals testValue exactly.
	// client.Get(ctx, key) returns (string, error) — a *redis.Nil error specifically
	// means "key doesn't exist", which would itself be a bug worth failing loudly on here.
	// If the returned value != testValue, log.Fatalf — that's a real correctness bug,
	// not just a connectivity problem.
	gotVal, err := c.Get(ctx, testKey).Result()
	if err != nil {
		log.Fatalf("Failure occurred due to: %v", err)
	}
	if gotVal != testValue {
		log.Fatalf("GET returned %q, want %q", gotVal, testValue)
	}

	fmt.Println("[OK] GET succeeded, value matches")

	// TODO 7: DEL testKey to clean up after yourself — don't leave test keys
	// sitting in the real Redis instance after this script exits.
	// On error, log.Fatalf (a cleanup failure here is still worth knowing about).

	err = c.Del(ctx, testKey).Err()
	if err != nil {
		log.Fatalf("Failure occurred due to: %v", err)
	}

	fmt.Println("[OK] DEL succeeded")
	fmt.Println("[OK] full SET/GET/DEL round trip verified against real redis")

	_ = log.Default // placeholder so `log` import isn't unused before you fill in TODO 1/3
}
