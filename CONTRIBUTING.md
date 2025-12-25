# Contributing to LGH

[中文](CONTRIBUTING.zh-CN.md)

Thank you for your interest in LGH! We welcome all forms of contributions.

## 🚀 Getting Started

### 1. Fork and Clone

```bash
# Fork this repository, then clone your fork
git clone https://github.com/YOUR_USERNAME/lgh.git
cd lgh
```

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 3. Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o lgh ./cmd/lgh/

# Run tests
go test ./... -v
```

### 4. Commit Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add new tunnel method"
git commit -m "fix: handle empty repository list"
git commit -m "docs: update README with new examples"
git commit -m "security: fix password echo vulnerability"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## 📝 Code Standards

### Go Code Style

- Use `gofmt` to format code
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Add documentation comments for public functions and types

### Commit Convention

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation update
- `test:` Test related
- `refactor:` Code refactoring
- `security:` Security fix
- `chore:` Build/tool related

## 🔒 Security Considerations

When contributing to LGH, please pay special attention to:

- **Password Handling**: Passwords must use hidden input, never echo in terminal
- **Password Storage**: Only store salted hashes, never plaintext
- **Path Validation**: Validate path safety before file operations
- **Least Privilege**: Config file permissions should be 0600

If you discover a security vulnerability, please report privately, not publicly.

## 🧪 Testing

### Run Tests

```bash
# All tests
go test ./... -v

# Short tests (skip integration)
go test ./... -v -short

# With coverage
go test ./... -v -cover

# Security check
make security
```

### Test Requirements

- New features must include tests
- Bug fixes should include regression tests
- Security fixes must include test validation
- Maintain test coverage

## 🔄 CI/CD Pipeline

All Pull Requests must pass CI checks before merging. The CI pipeline includes three main stages:

### 1. Code Linting

Using `golangci-lint v1.60` for code quality checks:

```bash
# Run lint locally (recommended before committing)
make lint
# or
golangci-lint run
```

**Required checks:**
- ✅ `govet` - Official Go static analysis (including shadow detection)
- ✅ `staticcheck` - Advanced static analysis
- ✅ `gosimple` - Code simplification suggestions
- ✅ `unused` - Unused code detection
- ✅ `ineffassign` - Ineffective assignment detection
- ✅ `gosec` - Security scanning (G104, G204, G304, G115, etc.)
- ✅ `revive` - Code style checking
- ✅ `goimports` - Import sorting and formatting (with `github.com/JoeGlenn1213/lgh` as local-prefix)
- ✅ `gofmt` - Code formatting
- ✅ `typecheck` - Type checking

**Common Issue Fixes:**

1. **Unused Parameters** (revive):
   ```go
   // ❌ Wrong
   func runCommand(cmd *cobra.Command, args []string) error {
       // args unused
   }
   
   // ✅ Correct
   func runCommand(cmd *cobra.Command, _ []string) error {
       // Use _ to explicitly ignore
   }
   ```

2. **Variable Shadowing** (govet):
   ```go
   // ❌ Wrong
   err := doSomething()
   if err := doAnother(); err != nil {  // shadows outer err
       return err
   }
   
   // ✅ Correct
   err := doSomething()
   if err2 := doAnother(); err2 != nil {
       return err2
   }
   ```

3. **Unhandled Errors** (gosec G104):
   ```go
   // ❌ Wrong
   file.Close()
   
   // ✅ Correct
   _ = file.Close()  // Explicitly ignore
   // or
   if err := file.Close(); err != nil {
       log.Printf("failed to close: %v", err)
   }
   ```

4. **File Permissions** (gosec G306):
   ```go
   // ❌ Wrong
   os.WriteFile(path, data, 0644)
   
   // ✅ Correct (for sensitive files)
   os.WriteFile(path, data, 0600)
   ```

### 2. License Check

Using `addlicense` to ensure all source files contain MIT license headers:

```bash
# Add license headers
addlicense -l mit -c "JoeGlenn1213" **/*.go

# Check licenses
addlicense -check -l mit -c "JoeGlenn1213" .
```

**Requirements:**
- All `.go` files must include MIT license headers
- Exclude `dist/` and `vendor/` directories

### 3. Cross-Platform Testing

Running full test suite on three platforms:

**Test Platforms:**
- 🐧 Ubuntu (latest)
- 🍎 macOS (latest)
- 🪟 Windows (latest)

**Test Commands:**
```bash
# Linux: All tests + race detector
go test ./... -v -race

# macOS/Windows: All tests (without race detector)
go test ./... -v
```

**Test Requirements:**
- ✅ All integration tests must pass
- ✅ Tests must complete within 5 minutes
- ✅ Cross-platform compatibility (Unix/Windows path handling)
- ✅ Proper environment variables (`HOME` and `USERPROFILE`)

### CI Configuration

See complete CI config at [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

### Local Verification

Before submitting a PR, ensure all checks pass locally:

```bash
# 1. Format code
go fmt ./...
goimports -local github.com/JoeGlenn1213/lgh -w .

# 2. Run linter
golangci-lint run

# 3. Run tests
go test ./... -v

# 4. (Optional) Run race detector
go test ./... -v -race

# 5. Check licenses
addlicense -check -l mit -c "JoeGlenn1213" .
```

### CI Failure Handling

If CI fails:

1. **View Logs**: Click failed check to see detailed error messages
2. **Fix Locally**: Run the same command locally to reproduce
3. **Fix and Push**: Push fixes to the same PR branch, CI will re-run automatically
4. **Ask for Help**: Comment on PR if you need assistance

### Performance and Stability Requirements

- Tests should not have flaky behavior
- Integration tests should clean up temporary files
- Use unique temporary directories to avoid conflicts
- Windows path handling: use `filepath.ToSlash()` for conversion

## 📋 Issue Guidelines

### Reporting Bugs

Please include:

- LGH version (`lgh --version`)
- OS and version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant logs

### Reporting Security Issues

**DO NOT report security vulnerabilities in public issues!**

Email project maintainers with:
- Vulnerability description
- Steps to reproduce
- Potential impact

### Feature Requests

Please describe:

- Problem you want to solve
- Your proposed solution
- Possible alternatives

## 🏗️ Project Structure

```
lgh/
├── cmd/lgh/           # CLI commands
│   ├── main.go       # Entry point
│   ├── init.go       # lgh init
│   ├── serve.go      # lgh serve
│   ├── add.go        # lgh add
│   ├── list.go       # lgh list
│   ├── status.go     # lgh status
│   ├── remove.go     # lgh remove
│   ├── tunnel.go     # lgh tunnel
│   └── auth.go       # lgh auth (authentication)
├── internal/          # Internal packages
│   ├── config/       # Configuration management
│   ├── git/          # Git operations
│   ├── registry/     # Repository mapping
│   ├── server/       # HTTP server + auth middleware
│   ├── tunnel/       # Tunnel functionality
│   └── mdns/         # mDNS service
├── pkg/ui/           # Terminal UI
├── docs/             # Documentation
│   └── SECURITY.md   # Security guidelines
└── test/             # Integration tests
```

## ✅ PR Checklist

- [ ] Code passes `go fmt` and `go vet`
- [ ] All tests pass
- [ ] Added necessary tests
- [ ] Updated relevant documentation
- [ ] Commit messages follow convention
- [ ] Security-sensitive code reviewed

## 📄 License

By contributing code, you agree that your contributions will be licensed under the MIT License.

---

Thank you for your contribution! Feel free to open an issue if you have questions.
