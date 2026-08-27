# Contributing to v2rayA

Thank you for your interest in contributing to v2rayA! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help maintain a welcoming community
- Follow project conventions

## Getting Started

### Prerequisites

**Backend (Go):**
- Go 1.21 or later
- GCC (for CGO/SQLite)
- Git

**Frontend (Vue.js):**
- Node.js 16+ 
- Yarn package manager
- Git

**Platform Tools:**
- Linux: `build-essential`, `libsqlite3-dev`
- Windows: MinGW-w64 or MSYS2
- macOS: Xcode Command Line Tools

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/v2rayA.git
cd v2rayA
git remote add upstream https://github.com/v2rayA/v2rayA.git
```

## Development Setup

### Backend (Go)

```bash
cd service

# Download dependencies
go mod download

# Build for development
go build -o v2raya-dev

# Run with development flags
./v2raya-dev --lite --config ./dev-config
```

### Frontend (Vue.js)

```bash
cd gui

# Install dependencies
yarn install

# Start development server
yarn serve

# Build for production
yarn build

# Lint code
yarn lint
```

### Database Setup

The database is automatically created on first run. For development:

```bash
# Database location
ls ~/.config/v2raya/db.sqlite

# Reset database (for testing)
rm ~/.config/v2raya/db.sqlite
```

## Project Structure

```
v2rayA/
├── core/                    # V2Ray/XRay core binary
│   ├── main/               # Entry point
│   ├── xray/               # XRay extensions
│   ├── dns/                # DNS handling
│   └── hint/               # Protocol hints
├── service/                # Backend service
│   ├── main.go             # Service entry point
│   ├── server/             # HTTP server
│   │   ├── router/         # Route definitions
│   │   ├── controller/     # API controllers
│   │   └── service/        # Business logic
│   ├── kernel/             # Core management
│   │   ├── v2ray/          # V2Ray integration
│   │   ├── iptables/       # Firewall rules
│   │   └── ipforward/      # IP forwarding
│   ├── db/                 # Database layer
│   │   └── configure/      # Configuration queries
│   ├── common/             # Shared utilities
│   └── pkg/                # Reusable packages
├── gui/                    # Frontend application
│   ├── src/                # Vue source code
│   │   ├── components/     # UI components
│   │   ├── store/          # Vuex state
│   │   ├── locales/        # Translations
│   │   └── plugins/        # Vue plugins
│   └── public/             # Static assets
├── install/                # Installation scripts
│   ├── aur/                # Arch Linux
│   ├── docker/             # Docker files
│   └── universal/          # Cross-platform
└── docs/                   # Documentation
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature
# or
git checkout -b fix/your-fix
```

### 2. Make Changes

- Follow coding standards
- Add tests for new features
- Update documentation if needed

### 3. Test Locally

```bash
# Backend tests
cd service
go test ./...

# Frontend lint
cd gui
yarn lint

# Manual testing
# Start both backend and frontend
```

### 4. Commit Changes

```bash
git add .
git commit -m "feat: add new feature"
# or
git commit -m "fix: resolve issue"
```

**Commit Message Format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `style:` Formatting
- `refactor:` Code restructuring
- `test:` Tests
- `chore:` Maintenance

## Coding Standards

### Go (Backend)

1. **Formatting:**
   - Use `gofmt` or `goimports`
   - Maximum line length: 120 characters

2. **Naming:**
   - Exported: PascalCase
   - Unexported: camelCase
   - Interfaces: `-er` suffix

3. **Comments:**
   - Package-level comments required
   - Exported functions documented
   - Complex logic explained

4. **Error Handling:**
   - Always handle errors
   - Use `fmt.Errorf` for wrapping
   - Log errors appropriately

### JavaScript/Vue.js (Frontend)

1. **Formatting:**
   - Use ESLint with project config
   - 2-space indentation
   - Single quotes

2. **Components:**
   - Single File Components (SFC)
   - PascalCase for components
   - Props validation required

3. **State Management:**
   - Use Vuex for global state
   - Keep components stateless when possible

4. **Imports:**
   - Group imports: external, internal, components
   - Absolute imports with `@/`

## Testing

### Backend Tests

```bash
cd service

# Run all tests
go test ./...

# Run specific test
go test ./db/...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Frontend Tests

```bash
cd gui

# Lint check
yarn lint

# Build check
yarn build
```

### Integration Testing

1. Start backend service
2. Start frontend dev server
3. Test API endpoints
4. Verify UI functionality
5. Check browser console for errors

## Pull Request Process

### 1. Prepare PR

- Update your fork
- Rebase on main if needed
- Ensure all tests pass
- Update documentation

### 2. Create PR

- Use descriptive title
- Reference related issues
- Provide detailed description
- Include screenshots for UI changes

### 3. PR Template

```markdown
## Description
[Describe your changes]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
[Describe tests performed]

## Checklist
- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### 4. Review Process

- Address reviewer feedback
- Keep PR focused and small
- Respond to comments promptly
- Squash commits if requested

## Reporting Issues

### Bug Reports

Include:
- v2rayA version
- Operating system and version
- Steps to reproduce
- Expected vs actual behavior
- Logs (if applicable)

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternatives considered
- Implementation ideas

## Communication

- **Issues:** GitHub Issues
- **Discussions:** GitHub Discussions
- **Chat:** Community channels (see README)

## License

By contributing, you agree that your contributions will be licensed under the AGPL-3.0 License.

## Thank You!

Your contributions help make v2rayA better for everyone. We appreciate your time and effort!
