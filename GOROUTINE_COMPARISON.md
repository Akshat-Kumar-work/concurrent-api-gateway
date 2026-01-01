# Goroutine Pattern Comparison

## Overview
we 3 different handlers that all fetch data from 3 services concurrently, but each uses a different goroutine pattern.

---

## Version 1: WaitGroup + Mutex Pattern
**File:** `aggregate_wg.go`  
**Endpoint:** `/api/aggregate/wg`

### How it works:
1. Creates a `WaitGroup` to track when all goroutines finish
2. Uses a `Mutex` (lock) to safely write results to a shared map
3. Each goroutine fetches data and writes directly to the shared map (with lock protection)
4. Main thread waits for all goroutines to complete using `wg.Wait()`

### Benefits:
**Simple and straightforward** - Easy to understand for beginners  
**Guaranteed completion** - Waits for ALL goroutines to finish (no early exit)  
**Direct data access** - Results go directly into the map  
**No channel overhead** - Slightly more efficient (no channel operations)

### Drawbacks:
**No timeout protection** - If a service hangs, the whole request hangs forever  
**Mutex contention** - All goroutines compete for the same lock when writing  
**No cancellation** - Can't cancel slow requests  
**Risk of deadlock** - If a goroutine never calls `wg.Done()`, the program hangs

---

## Version 2: Channel Pattern
**File:** `aggregate_channel.go`  
**Endpoint:** `/api/aggregate/channel`

### How it works:
1. Creates a buffered channel to collect results
2. Each goroutine sends its result through the channel
3. Main thread reads from the channel (one result per service)
4. No explicit waiting - relies on channel blocking behavior

### Benefits:
**Clean separation** - Goroutines don't directly access shared data  
**Go-idiomatic** - Uses channels (the "Go way" of communication)  
**No mutex needed** - Channels handle synchronization automatically  
**Simpler than WaitGroup** - No need to track completion manually

### Drawbacks:
**No timeout protection** - Still can hang forever if a service is slow  
**Potential bug** - If you read fewer times than services, goroutines leak  
**No cancellation** - Can't stop slow requests  
**Order dependency** - Results come in random order (but that's usually fine)

---

## Version 3: Context + Timeout + WaitGroup + Channels (Hybrid)
**File:** `aggregate_context_timeout.go`  
**Endpoint:** `/api/aggregate/channel-with-context-timeout`

### How it works:
1. Creates a context with a 1-second timeout
2. Uses both WaitGroup AND channels
3. Each goroutine has an inner goroutine that does the actual fetch
4. Uses `select` to wait for either the result OR timeout cancellation
5. Closes the channel when all goroutines finish

### Benefits:
**Timeout protection** - Request automatically fails after 1 second (prevents hanging)  
**Cancellation support** - Can cancel slow requests via context  
**Production-ready** - Handles real-world failure scenarios  
**Resource cleanup** - Properly closes channels and cancels context  
**Error reporting** - Tells you which services timed out  
**Respects HTTP context** - Uses the request's context (important for HTTP servers)

### Drawbacks:
**More complex** - Harder to understand (nested goroutines, select statements)  
**More overhead** - Extra goroutines and channel operations  
**Slightly slower** - More synchronization overhead

---

## *** for Production: Version 3 (Context + Timeout)**

### Why Version 3 is Best for Production:

1. **Timeout Protection** - In production, services can hang or be slow. Without timeouts, your API can hang forever, consuming resources and making users wait.

2. **Proper Resource Management** - Uses context cancellation which is the standard way to handle timeouts in Go HTTP servers. This prevents resource leaks.

3. **Graceful Degradation** - If one service times out, you still get results from the other services. The API doesn't completely fail.

4. **HTTP Context Integration** - Uses `c.Request.Context()` which respects HTTP client disconnections. If a user closes their browser, the request is cancelled.

5. **Error Transparency** - Tells you exactly which services failed and why (timeout vs actual error).

### When to Use Each:

- **WaitGroup**: Learning, simple internal tools, when you're 100% sure services will respond quickly
- **Channels**: When you want idiomatic Go code but don't need timeouts (rare in production)
- **Context+Timeout**: **Always use this in production APIs** - it's the safest and most robust

---

## Recommendation

**For production, use Version 3** (`AggregateHandlerWithTimeout`). 

You might want to:
- Make the timeout configurable (not hardcoded to 1 second)
- Consider using `errgroup` package for even cleaner code
- Add metrics/logging for timeout events

The extra complexity is worth it for the reliability and user experience improvements.

