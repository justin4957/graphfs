# Contributing to GraphFS

Thank you for your interest in contributing to GraphFS! This document provides guidelines and information for contributors.

## 🎯 Project Goals

GraphFS aims to transform codebases into queryable knowledge graphs by:
1. Parsing LinkedDoc+RDF metadata from source code
2. Building semantic relationships between modules
3. Enabling intelligent code navigation and analysis
4. Providing actionable insights for developers

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- Git
- Basic understanding of RDF/Turtle (helpful but not required)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/justin4957/graphfs
cd graphfs

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the CLI
go build -o bin/graphfs ./cmd/graphfs
```

## 📋 How to Contribute

### 1. Find an Issue
- Check [open issues](https://github.com/justin4957/graphfs/issues)
- Look for issues tagged `good-first-issue` or `help-wanted`
- Comment on the issue to express interest

### 2. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 3. Write Code
- Follow the Go style guidelines (see below)
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 4. Test Your Changes
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter (if installed)
golangci-lint run
```

### 5. Submit a Pull Request
- Push your branch to your fork
- Create a pull request against `main`
- Fill out the PR template
- Link related issues

## 🎨 Code Style Guidelines

### Go Code
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions under 50 lines when possible
- Write descriptive variable names
- Add comments for exported functions

### Example
```go
// ParseLinkedDocHeader extracts RDF triples from LinkedDoc comment blocks.
// It searches for <!-- LinkedDoc RDF --> markers and parses the enclosed
// Turtle syntax into a structured representation.
func ParseLinkedDocHeader(source string) (*LinkedDocHeader, error) {
    // Implementation...
}
```

### LinkedDoc Headers
All Go packages should include LinkedDoc headers:

```go
/*
# Module: pkg/parser/turtle.go
RDF/Turtle parser for LinkedDoc metadata extraction.

## Linked Modules
- [scanner](../scanner/scanner.go) - File scanning utilities

## Tags
parser, rdf, turtle

## Exports
TurtleParser, ParseTriples, ValidateSyntax

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<this> a code:Module ;
    code:name "pkg/parser/turtle.go" ;
    code:description "RDF/Turtle parser for LinkedDoc metadata extraction" ;
    code:tags "parser", "rdf", "turtle" .
<!-- End LinkedDoc RDF -->
*/
package parser
```

## 🧪 Testing Guidelines

### Unit Tests
- Test files should be named `*_test.go`
- Use table-driven tests for multiple cases
- Aim for 80%+ code coverage

### Example Test
```go
func TestParseTriples(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    []Triple
        wantErr bool
    }{
        {
            name:  "simple triple",
            input: "<s> <p> <o> .",
            want:  []Triple{{Subject: "s", Predicate: "p", Object: "o"}},
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseTriples(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseTriples() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseTriples() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 📚 Documentation

### Code Documentation
- Add godoc comments for all exported types and functions
- Include usage examples in comments when helpful
- Keep comments up-to-date with code changes

### User Documentation
- Update README.md for user-facing features
- Add examples to `examples/` directory
- Create guides in `docs/` for complex features

## 🐛 Reporting Bugs

### Before Reporting
- Search existing issues to avoid duplicates
- Try to reproduce with the latest version
- Gather relevant information (OS, Go version, etc.)

### Bug Report Template
```markdown
**Describe the bug**
A clear description of the bug.

**To Reproduce**
Steps to reproduce:
1. Run command '...'
2. See error '...'

**Expected behavior**
What you expected to happen.

**Environment**
- OS: [e.g., macOS 14.0]
- Go version: [e.g., 1.21.0]
- GraphFS version: [e.g., 0.1.0]

**Additional context**
Any other relevant information.
```

## 💡 Feature Requests

We welcome feature requests! Please:
1. Check if the feature aligns with project goals
2. Describe the use case clearly
3. Provide examples of how it would be used
4. Consider implementation complexity

## 🔍 Code Review Process

### What We Look For
- **Correctness**: Does the code work as intended?
- **Tests**: Are there adequate tests?
- **Documentation**: Is it well-documented?
- **Style**: Does it follow project conventions?
- **Performance**: Are there obvious inefficiencies?

### Review Timeline
- Initial feedback: Within 2-3 days
- Follow-up reviews: Within 1-2 days
- Merging: After approval from maintainers

## 📅 Development Phases

We're currently in **Phase 1** (Core Infrastructure). Check [ROADMAP.md](ROADMAP.md) for details on upcoming phases and how you can contribute.

## 🏆 Recognition

Contributors will be:
- Listed in the README
- Credited in release notes
- Eligible for maintainer status (after consistent contributions)

## 📞 Communication

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas, general discussion
- **Pull Requests**: Code contributions

## ⚖️ Code of Conduct

### Our Pledge
We are committed to providing a welcoming and inclusive environment for all contributors, regardless of background or identity.

### Expected Behavior
- Be respectful and considerate
- Welcome newcomers
- Provide constructive feedback
- Focus on what's best for the community

### Unacceptable Behavior
- Harassment or discrimination
- Trolling or insulting comments
- Publishing others' private information
- Other conduct that would be inappropriate in a professional setting

## 📄 License

By contributing to GraphFS, you agree that your contributions will be licensed under the MIT License.

---

**Questions?** Feel free to open a discussion or reach out to the maintainers!

Thank you for contributing to GraphFS! 🚀
