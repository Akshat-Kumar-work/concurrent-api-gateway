# How to Use Go Packages

## Steps to Use a Downloaded Package

### 1. **Import the Package**

Add the import statement at the top of your Go file:

```go
package main

import (
    "fmt"
    "github.com/Akshat-Kumar-work/golang-rest-api"  // Your package
    // Or with an alias:
    restapi "github.com/Akshat-Kumar-work/golang-rest-api"
)
```

### 2. **Check What the Package Exports**

To see what functions/types the package provides:

```bash
# See package documentation
go doc github.com/Akshat-Kumar-work/golang-rest-api

# See all exported symbols
go doc -all github.com/Akshat-Kumar-work/golang-rest-api

# See specific function/type
go doc github.com/Akshat-Kumar-work/golang-rest-api FunctionName
```

### 3. **Use the Package in Your Code**

Once you know what the package exports, use it:

```go
package main

import (
    "fmt"
    "github.com/Akshat-Kumar-work/golang-rest-api"
)

func main() {
    // Example usage (adjust based on actual package API)
    result := restapi.SomeFunction()
    fmt.Println(result)
}
```

---

## For Your Specific Package

Since `golang-rest-api` is currently marked as "indirect", you need to:

### Step 1: Make it a Direct Dependency

Remove the `// indirect` comment by actually importing it in your code:

```go
// In any .go file (e.g., main.go or a handler)
import (
    "github.com/Akshat-Kumar-work/golang-rest-api"
)
```

Then run:
```bash
go mod tidy
```

This will make it a direct dependency.

### Step 2: Discover the Package API

**Option A: Check GitHub Repository**
```bash
# Visit the repository
# https://github.com/Akshat-Kumar-work/golang-rest-api
```

**Option B: Check Local Package**
```bash
# Find where it's cached
go env GOMODCACHE

# List files in the package
ls -la $(go env GOMODCACHE)/github.com/akshat-kumar-work/golang-rest-api@v0.0.0-20251231190755-b18719a24e9e
```

**Option C: Use Go Doc**
```bash
go doc github.com/Akshat-Kumar-work/golang-rest-api
```

### Step 3: Use It in Your Code

Based on the package name (`golang-rest-api`), it likely provides REST API utilities. Here's a general example:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/Akshat-Kumar-work/golang-rest-api"
    "github.com/gin-gonic/gin"
)

func main() {
    // Example: If package provides a client
    client := restapi.NewClient("http://api.example.com")
    
    // Example: If package provides utilities
    router := gin.Default()
    router.GET("/api/data", func(c *gin.Context) {
        // Use package functions here
        data, err := restapi.FetchData()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, data)
    })
    
    router.Run(":8080")
}
```

---

## Common Package Usage Patterns

### Pattern 1: Client/Service Package
```go
import "github.com/Akshat-Kumar-work/golang-rest-api"

// Usually provides a client
client := restapi.NewClient(config)
result, err := client.Get("/endpoint")
```

### Pattern 2: Utility Functions
```go
import "github.com/Akshat-Kumar-work/golang-rest-api"

// Usually provides helper functions
data := restapi.ParseJSON(jsonString)
response := restapi.BuildResponse(data)
```

### Pattern 3: Middleware/Handlers
```go
import "github.com/Akshat-Kumar-work/golang-rest-api"

// Usually provides middleware
router.Use(restapi.AuthMiddleware())
router.Use(restapi.LoggingMiddleware())
```

---

## Using Private Repositories

If the package repository is **private**, you need to configure authentication:

### Method 1: Using SSH (Recommended for GitHub/GitLab)

**Step 1: Configure Git to use SSH for the domain**

```bash
# For GitHub
git config --global url."git@github.com:".insteadOf "https://github.com/"

# For GitLab
git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/"
```

**Step 2: Set GOPRIVATE environment variable**

```bash
# Set for current session
export GOPRIVATE=github.com/Akshat-Kumar-work/*

# Or for all private repos under your org
export GOPRIVATE=github.com/Akshat-Kumar-work/*,gitlab.com/your-org/*

# Make it permanent (add to ~/.zshrc or ~/.bashrc)
echo 'export GOPRIVATE=github.com/Akshat-Kumar-work/*' >> ~/.zshrc
source ~/.zshrc
```

**Step 3: Ensure SSH key is set up**

```bash
# Check if SSH key exists
ls -la ~/.ssh/id_rsa.pub

# If not, generate one
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# Add to GitHub/GitLab (copy public key)
cat ~/.ssh/id_rsa.pub
# Then add it to your GitHub/GitLab account settings
```

**Step 4: Test and use**

```bash
# Test SSH connection
ssh -T git@github.com

# Now you can use go get
go get github.com/Akshat-Kumar-work/golang-rest-api
```

---

### Method 2: Using Personal Access Token (HTTPS)

**Step 1: Create a Personal Access Token**

- **GitHub**: Settings → Developer settings → Personal access tokens → Generate new token
- **GitLab**: Preferences → Access Tokens → Create personal access token

**Step 2: Configure Git credentials**

```bash
# Option A: Use credential helper (macOS)
git config --global credential.helper osxkeychain

# Option B: Store in .netrc file
echo "machine github.com login YOUR_USERNAME password YOUR_TOKEN" >> ~/.netrc
chmod 600 ~/.netrc
```

**Step 3: Set GOPRIVATE**

```bash
export GOPRIVATE=github.com/Akshat-Kumar-work/*
```

**Step 4: Use with token in URL (one-time)**

```bash
go get github.com/YOUR_USERNAME:YOUR_TOKEN@github.com/Akshat-Kumar-work/golang-rest-api.git
```

---

### Method 3: Using .netrc File (Simple)

Create/edit `~/.netrc`:

```
machine github.com
login YOUR_USERNAME
password YOUR_PERSONAL_ACCESS_TOKEN

machine gitlab.com
login YOUR_USERNAME
password YOUR_PERSONAL_ACCESS_TOKEN
```

Then:
```bash
chmod 600 ~/.netrc
export GOPRIVATE=github.com/Akshat-Kumar-work/*
go get github.com/Akshat-Kumar-work/golang-rest-api
```

---

### Method 4: Using GOPROXY with Direct (Recommended for Teams)

**Configure go.mod to use direct mode for private repos:**

```bash
# Set GOPRIVATE
export GOPRIVATE=github.com/Akshat-Kumar-work/*

# Configure GOPROXY to skip proxy for private repos
export GOPROXY=direct

# Or use both proxy and direct
export GOPROXY=https://proxy.golang.org,direct
```

Then use SSH or HTTPS as configured above.

---

### Verify Private Repo Access

```bash
# Check if GOPRIVATE is set
go env GOPRIVATE

# Try to download
go get github.com/Akshat-Kumar-work/golang-rest-api

# If it works, you'll see it in go.mod
cat go.mod | grep golang-rest-api
```

---

### Common Issues with Private Repos

**Issue: "fatal: could not read Username"**

**Solution:**
```bash
# Use SSH instead
git config --global url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE=github.com/Akshat-Kumar-work/*
```

**Issue: "authentication required"**

**Solution:**
- Ensure SSH key is added to GitHub/GitLab
- Or use Personal Access Token with HTTPS
- Check `GOPRIVATE` is set correctly

**Issue: "module declares its path as: X but was required as: Y"**

**Solution:**
- Ensure the module path in the private repo's `go.mod` matches the import path
- Check repository visibility settings

---

## Troubleshooting

### Issue: Package marked as "indirect"

**Solution:** Import it in your code and run `go mod tidy`

```bash
# 1. Add import to any .go file
# 2. Run:
go mod tidy
```

### Issue: "cannot find package"

**Solution:** Make sure the package is downloaded:

```bash
go get github.com/Akshat-Kumar-work/golang-rest-api
go mod download
```

**If it's a private repo:**
```bash
# Set GOPRIVATE first
export GOPRIVATE=github.com/Akshat-Kumar-work/*
go get github.com/Akshat-Kumar-work/golang-rest-api
```

### Issue: "package not used"

**Solution:** Either:
1. Use the package in your code
2. Or remove it: `go mod tidy` (will remove unused packages)

---

## Quick Reference

```bash
# Download/Update package
go get github.com/Akshat-Kumar-work/golang-rest-api

# Download specific version
go get github.com/Akshat-Kumar-work/golang-rest-api@v1.0.0

# Update to latest
go get -u github.com/Akshat-Kumar-work/golang-rest-api

# Remove unused packages
go mod tidy

# See package documentation
go doc github.com/Akshat-Kumar-work/golang-rest-api

# List all dependencies
go list -m all
```

---

## Next Steps for Your Package

1. **Check the package documentation:**
   ```bash
   go doc github.com/Akshat-Kumar-work/golang-rest-api
   ```

2. **Visit the GitHub repository** to see examples and README

3. **Import and use it** in your code to make it a direct dependency

4. **Run `go mod tidy`** to clean up dependencies
