# When NOT to Use Goroutines & Database Concurrency Best Practices

## When NOT to Use Goroutines

### 1. **Simple Sequential Operations**

**Don't use goroutines when:**
- Operations must happen in order
- Operations are very fast (< 1ms)
- The overhead of creating goroutines exceeds the benefit

**Example - DON'T:**
```go
// BAD: Unnecessary goroutines for simple operations
go func() {
    user.Name = "John"
}()
go func() {
    user.Email = "john@example.com"
}()
go func() {
    user.Age = 30
}()
// These are just variable assignments - no benefit from goroutines!
```

**Example - DO:**
```go
// GOOD: Simple sequential operations
user.Name = "John"
user.Email = "john@example.com"
user.Age = 30
```

---

### 2. **Operations That Must Be Synchronous**

**Don't use goroutines when:**
- You need the result immediately before proceeding
- The next operation depends on the previous one
- You're building a result step-by-step

**Example - DON'T:**
```go
// BAD: Need result before next step
var userID string
go func() {
    userID = createUser() // Need this before creating order!
}()
createOrder(userID) // Will fail - userID is empty!
```

**Example - DO:**
```go
// GOOD: Sequential when order matters
userID := createUser()
createOrder(userID)
```

---

### 3. **Single Database Transaction**

**Don't use goroutines when:**
- Operations must be in the same transaction
- You need ACID guarantees across operations
- Database operations are already fast

**Example - DON'T:**
```go
// BAD: These must be in same transaction
go func() {
    db.Exec("INSERT INTO users ...")
}()
go func() {
    db.Exec("INSERT INTO profiles ...") // Different transaction!
}()
// If one fails, the other might succeed - data inconsistency!
```

**Example - DO:**
```go
// GOOD: Single transaction
tx, _ := db.Begin()
tx.Exec("INSERT INTO users ...")
tx.Exec("INSERT INTO profiles ...")
tx.Commit() // Both succeed or both fail
```

---

### 4. **Very Small Workloads**

**Don't use goroutines when:**
- Processing a few items (< 10)
- Each operation is very fast
- The overhead isn't worth it

**Example - DON'T:**
```go
// BAD: Only 3 items, overhead > benefit
items := []string{"a", "b", "c"}
for _, item := range items {
    go processItem(item) // Overhead of 3 goroutines > processing time
}
```

**Example - DO:**
```go
// GOOD: Process sequentially for small batches
items := []string{"a", "b", "c"}
for _, item := range items {
    processItem(item)
}
```

---

### 5. **Shared State Without Proper Synchronization**

**Don't use goroutines when:**
- You're writing to shared variables without locks
- Race conditions are likely
- You can't properly synchronize access

**Example - DON'T:**
```go
// BAD: Race condition!
var count int
for i := 0; i < 1000; i++ {
    go func() {
        count++ // Multiple goroutines writing - race condition!
    }()
}
```

**Example - DO:**
```go
// GOOD: Use atomic operations or mutex
var count int64
var mu sync.Mutex
for i := 0; i < 1000; i++ {
    go func() {
        mu.Lock()
        count++
        mu.Unlock()
    }()
}
// OR use atomic
atomic.AddInt64(&count, 1)
```

---

## When to Use Goroutines

### ✅ **Good Use Cases:**

1. **Independent I/O Operations**
   - Multiple HTTP API calls
   - Multiple database queries that don't depend on each other
   - File operations that can run in parallel

2. **CPU-Intensive Work with Multiple Cores**
   - Image processing
   - Data transformation
   - Calculations that can be parallelized

3. **Background Tasks**
   - Sending emails
   - Logging
   - Cleanup jobs

---

## Database Operations with Goroutines

### ⚠️ **Critical Considerations:**

#### 1. **Connection Pool Exhaustion**

**Problem:** Database connections are limited. Too many concurrent goroutines can exhaust the pool.

**Example - BAD:**
```go
// BAD: 1000 goroutines = 1000 connections (pool exhausted!)
users := get1000Users()
for _, user := range users {
    go func(u User) {
        db.Exec("UPDATE users SET ... WHERE id = ?", u.ID)
    }(user)
}
```

**Solution: Use Semaphore to Limit Concurrency**

```go
// GOOD: Limit to 10 concurrent database operations
const maxConcurrency = 10
sem := make(chan struct{}, maxConcurrency) // Buffered channel = semaphore

users := get1000Users()
var wg sync.WaitGroup

for _, user := range users {
    wg.Add(1)
    go func(u User) {
        defer wg.Done()
        
        // Acquire semaphore (blocks if 10 already running)
        sem <- struct{}{}
        defer func() { <-sem }() // Release semaphore
        
        db.Exec("UPDATE users SET ... WHERE id = ?", u.ID)
    }(user)
}

wg.Wait()
```

---

#### 2. **When to Use Semaphores**

**Use semaphores when:**
- You have many goroutines accessing a limited resource (DB, API, file handles)
- You want to control the maximum number of concurrent operations
- You need to prevent resource exhaustion

**Common Limits:**
- Database: 5-20 concurrent connections (depends on DB config)
- HTTP API: 10-50 concurrent requests (depends on rate limits)
- File operations: 5-10 concurrent file handles

---

#### 3. **Semaphore Pattern Examples**

**Pattern 1: Simple Semaphore with WaitGroup**
```go
const maxWorkers = 10
sem := make(chan struct{}, maxWorkers)
var wg sync.WaitGroup

for _, task := range tasks {
    wg.Add(1)
    go func(t Task) {
        defer wg.Done()
        
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        
        // Do database work
        processTask(t)
    }(task)
}

wg.Wait()
```

**Pattern 2: Worker Pool (Better for Large Batches)**
```go
const numWorkers = 10
jobs := make(chan Task, 100)
var wg sync.WaitGroup

// Start worker pool
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for task := range jobs {
            processTask(task) // Database operation
        }
    }()
}

// Send jobs
for _, task := range tasks {
    jobs <- task
}
close(jobs)

wg.Wait()
```

---

#### 4. **Database-Specific Best Practices**

**✅ DO:**
- Use connection pooling (most DB drivers do this automatically)
- Limit concurrent database operations with semaphores
- Use transactions for related operations
- Add timeouts to database operations
- Handle connection errors gracefully

**❌ DON'T:**
- Create unlimited goroutines for database operations
- Share database connections across goroutines without synchronization
- Ignore connection pool limits
- Forget to close rows/transactions

---

## Real-World Example: Processing User Reminders

### ❌ BAD (No Concurrency Control):
```go
func processReminders(reminders []Reminder) {
    for _, reminder := range reminders {
        go func(r Reminder) {
            // 1000 reminders = 1000 goroutines = DB pool exhausted!
            db.Exec("UPDATE reminders SET processed = true WHERE id = ?", r.ID)
        }(reminder)
    }
}
```

### ✅ GOOD (With Semaphore):
```go
func processReminders(reminders []Reminder) error {
    const maxConcurrency = 10 // Limit to 10 concurrent DB operations
    sem := make(chan struct{}, maxConcurrency)
    var wg sync.WaitGroup
    var mu sync.Mutex
    var errors []error
    
    for _, reminder := range reminders {
        wg.Add(1)
        go func(r Reminder) {
            defer wg.Done()
            
            // Acquire semaphore
            sem <- struct{}{}
            defer func() { <-sem }()
            
            // Database operation
            if err := db.Exec("UPDATE reminders SET processed = true WHERE id = ?", r.ID); err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
            }
        }(reminder)
    }
    
    wg.Wait()
    
    if len(errors) > 0 {
        return fmt.Errorf("failed to process %d reminders", len(errors))
    }
    return nil
}
```

---

## Decision Tree: Should I Use Goroutines?

```
Is the operation I/O-bound or CPU-intensive?
├─ NO → Don't use goroutines (use sequential processing)
└─ YES → Continue

Are operations independent (no dependencies)?
├─ NO → Don't use goroutines (use sequential processing)
└─ YES → Continue

Will I have many concurrent operations (> 10)?
├─ NO → Use goroutines (simple case)
└─ YES → Continue

Is the resource limited (DB connections, API rate limits)?
├─ NO → Use goroutines with WaitGroup
└─ YES → Use goroutines with SEMAPHORE + WaitGroup
```

---

## Summary

### When NOT to Use Goroutines:
1. Simple sequential operations
2. Operations that must be synchronous
3. Single database transaction
4. Very small workloads (< 10 items)
5. Shared state without proper synchronization

### When to Use Semaphores:
1. Many goroutines accessing limited resources (DB, API)
2. Need to control maximum concurrency
3. Prevent resource exhaustion
4. Database operations with > 10 concurrent requests

### Best Practices:
- **Database operations**: Always use semaphores (limit to 5-20 concurrent)
- **HTTP API calls**: Use semaphores if rate-limited (limit to 10-50)
- **File operations**: Use semaphores (limit to 5-10)
- **CPU-bound work**: Usually don't need semaphores (limited by CPU cores)

---

## Quick Reference

| Scenario | Use Goroutines? | Use Semaphore? |
|----------|----------------|-----------------|
| 3 independent API calls | ✅ Yes | ❌ No |
| 1000 database updates | ✅ Yes | ✅ Yes (limit 10-20) |
| Sequential data processing | ❌ No | ❌ No |
| Single transaction | ❌ No | ❌ No |
| Background email sending | ✅ Yes | ✅ Yes (limit 5-10) |
| 5 file operations | ✅ Yes | ❌ No (small batch) |
| 100 file operations | ✅ Yes | ✅ Yes (limit 10) |

