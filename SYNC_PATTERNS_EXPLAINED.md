# Why Each Version Uses Different Synchronization Patterns

## Version 1: WaitGroup Only (No Channels)

**Pattern:** `WaitGroup + Mutex`

```go
var wg sync.WaitGroup
var mu sync.Mutex
results := make(map[string]interface{})

// Goroutines write directly to shared map
go func() {
    data, err := fetcher(userID)
    mu.Lock()           // Lock for safe access
    results[name] = data // Write directly to map
    mu.Unlock()
    wg.Done()
}()

wg.Wait() // Wait for all to finish
```

**Why WaitGroup only?**
- ✅ **Simple pattern**: Direct shared memory access
- ✅ **No data transfer needed**: Results go directly into the map
- ✅ **Straightforward**: Wait for completion, then use results
- ❌ **Requires mutex**: Need to protect shared map from race conditions
- ❌ **No timeout**: Can't cancel or timeout

**Use case:** Simple scenarios where you just need to wait for all goroutines to finish and collect results in a shared structure.

---

## Version 2: Channels Only (No WaitGroup)

**Pattern:** `Channel blocking for synchronization`

```go
resultChan := make(chan result, 3)

// Goroutines send results through channel
go func() {
    data, err := fetcher(userID)
    resultChan <- result{...} // Send through channel
}()

// Main thread blocks on receives
for range servicesToCall {
    res := <-resultChan  // Blocks until result arrives
    results[res.service] = res.data
}
```

**Why Channels only?**
- ✅ **Go-idiomatic**: Channels are the "Go way" of communication
- ✅ **No mutex needed**: Channels handle synchronization automatically
- ✅ **Clean separation**: Goroutines don't access shared memory
- ✅ **Implicit synchronization**: Blocking receives naturally wait for all results
- ❌ **No timeout**: Still can't cancel or timeout
- ⚠️ **Must match counts**: Loop must read exactly as many times as goroutines send

**Use case:** When you want idiomatic Go code and don't need timeout/cancellation. The blocking receive acts as implicit synchronization.

---

## Version 3: Both WaitGroup AND Channels

**Pattern:** `WaitGroup + Channels + Context`

```go
var wg sync.WaitGroup
resultChan := make(chan result, 3)

// Goroutines send through channel
go func() {
    defer wg.Done()
    // ... fetch with timeout ...
    resultChan <- result{...}
}()

// Separate goroutine waits and closes channel
go func() {
    wg.Wait()        // Wait for all workers
    close(resultChan) // Close channel
}()

// Main thread reads until channel closes
for res := range resultChan {
    // Process results
}
```

**Why BOTH WaitGroup AND Channels?**

### The Problem:
- We need **timeout support** (context cancellation)
- We need to know when **all goroutines are done** (to close channel)
- We need to **collect results** as they arrive (channels)
- We can't know **how many results** we'll get (some might timeout)

### Why WaitGroup?
- ✅ **Tracks completion**: Knows when all 3 worker goroutines finish
- ✅ **Closes channel safely**: Only close after all workers are done
- ✅ **Handles variable results**: Some might timeout, some might succeed

### Why Channels?
- ✅ **Collect results**: Receive results as they arrive
- ✅ **Range loop**: `for res := range resultChan` exits when channel closes
- ✅ **Timeout handling**: Can send timeout errors through channel

### Why NOT just one?

**Can't use only WaitGroup:**
- ❌ How do you know when to stop reading? (Can't use `range` without closing)
- ❌ How do you handle timeouts? (Need channels for `select` with context)

**Can't use only Channels:**
- ❌ How do you know when to close the channel? (Don't know how many results to expect)
- ❌ If some timeout, you might close too early or wait forever

**Solution: Use BOTH:**
1. **Channels** → Collect results and handle timeouts
2. **WaitGroup** → Track when all workers finish
3. **WaitGroup goroutine** → Closes channel when all done
4. **Range loop** → Exits when channel closes

---

## Summary Table

| Version | WaitGroup | Channels | Why? |
|---------|-----------|----------|------|
| **Version 1** | ✅ | ❌ | Simple: Direct shared memory, just need to wait |
| **Version 2** | ❌ | ✅ | Idiomatic: Channels provide sync + communication |
| **Version 3** | ✅ | ✅ | Production: Need both for timeout + variable results |

---

## Key Insight

**Version 3 needs BOTH because:**
- **Channels** handle the "what" (collecting results, timeout handling)
- **WaitGroup** handles the "when" (knowing when all workers are done to close channel)

They solve different problems:
- **WaitGroup** = "Are all workers finished?"
- **Channels** = "How do I collect results and handle timeouts?"

In Version 3, you need to answer BOTH questions, so you need BOTH tools!

