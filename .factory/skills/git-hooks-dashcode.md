# Skill: Git Hooks & Dashcode Integration

## Overview
This skill covers the git hooks system and Dashcode integration for commit tracking.

## Hook Files

```
.githooks/
├── pre-commit      # Runs before commit: linting, security checks
├── commit-msg      # Validates commit message format
└── post-commit     # Reports commit to Dashcode after success
```

## Setup

```bash
# Configure git to use custom hooks directory
git config core.hooksPath .githooks

# Make hooks executable
chmod +x .githooks/*

# Or use the Makefile
make setup-hooks
```

## Pre-commit Hook

### What it checks:
1. **Security**
   - Secrets detection (API keys, tokens, passwords)
   - Sensitive file types (.env, .pem, .key)
   - Large files (>1MB)

2. **Go Backend**
   - `gofmt` formatting
   - `go vet` static analysis
   - Build verification

3. **General**
   - Merge conflict markers
   - Trailing whitespace
   - Debugger statements

4. **Dashcode Reporting**
   - Reports pre-commit results to Dashcode

### Output Example:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  SE Notes Pre-commit Checks
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  1. Security Checks
  ✓ No secrets detected
  ✓ No sensitive file types
  
  2. Go Backend Checks
  ✓ Go formatting OK
  ✓ Go vet passed
  ✓ Go build successful
  
  Summary
  ✓ All checks passed!
```

### Bypassing (not recommended):
```bash
git commit --no-verify -m "message"
```

## Commit-msg Hook

### Format Required:
```
type(scope): description

[optional body]
```

### Valid Types:
| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Code style (formatting) |
| `refactor` | Code refactoring |
| `test` | Adding tests |
| `chore` | Maintenance tasks |
| `init` | Initial commit |
| `perf` | Performance improvement |
| `ci` | CI/CD changes |
| `build` | Build system changes |
| `revert` | Revert previous commit |

### Examples:
```bash
# Good
git commit -m "feat(notes): add rich text editor"
git commit -m "fix(api): handle null account_id"
git commit -m "docs: update README"

# Bad (will be rejected)
git commit -m "fixed stuff"
git commit -m "WIP"
```

## Post-commit Hook

### What it does:
- Collects commit metadata
- Posts to Dashcode API
- Non-blocking (errors don't affect commit)

### Payload sent to Dashcode:
```json
{
  "repoName": "notes-droid",
  "commitHash": "abc123...",
  "branch": "main",
  "trigger": "post-commit",
  "status": "success",
  "durationMs": 100,
  "meta": {
    "author": "John Doe",
    "email": "john@example.com",
    "message": "feat: add feature",
    "shortHash": "abc123",
    "filesChanged": 3
  },
  "results": [
    {
      "type": "commit",
      "status": "pass",
      "output": "Commit abc123 by John Doe",
      "durationMs": 100,
      "exitCode": 0
    }
  ]
}
```

### Dashcode Endpoint:
```
POST http://localhost:3001/api/hooks/report
```

## Dashcode Configuration

The hooks use a hardcoded repo ID for Dashcode:
```bash
REPO_ID="e8450e26-c573-4aee-af1b-248405af0acc"
DASHCODE_URL="http://localhost:3001"
```

To change, edit `.githooks/pre-commit` and `.githooks/post-commit`.

## Modifying Hooks

### Adding a new check to pre-commit:
```bash
# In .githooks/pre-commit, add before Summary section:

print_header "5. My New Check"
print_info "Running my check..."

if my_check_command; then
    print_success "My check passed"
else
    print_error "My check failed"
fi
```

### Changing Dashcode payload:
```bash
# In .githooks/post-commit, modify the jq command:

JSON_PAYLOAD=$(jq -n \
    --arg repoName "$REPO_NAME" \
    --arg myNewField "value" \
    '{
        repoName: $repoName,
        myNewField: $myNewField,
        # ... rest of payload
    }')
```

## Troubleshooting

### Hooks not running
```bash
# Check hooks path
git config core.hooksPath
# Should output: .githooks

# Check permissions
ls -la .githooks/
# All hooks should be executable (-rwxr-xr-x)
```

### Dashcode not receiving data
```bash
# Test endpoint manually
curl -X POST http://localhost:3001/api/hooks/report \
  -H "Content-Type: application/json" \
  -d '{"repoName":"test","commitHash":"abc","branch":"main","trigger":"test","status":"success","durationMs":0,"meta":{},"results":[]}'

# Check if Dashcode is running
curl http://localhost:3001
```

### Pre-commit failing unexpectedly
```bash
# Run the hook manually to see full output
.githooks/pre-commit

# Check for bash syntax errors
bash -n .githooks/pre-commit
```

### jq not found
```bash
# Install jq
brew install jq  # macOS
apt install jq   # Ubuntu/Debian
```

## Disabling Hooks Temporarily

```bash
# Skip all hooks for one commit
git commit --no-verify -m "message"

# Disable hooks globally (not recommended)
git config core.hooksPath /dev/null

# Re-enable
git config core.hooksPath .githooks
```
