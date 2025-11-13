# GraphFS - Semantic Code Filesystem Toolkit

> Transform your codebase into a queryable knowledge graph using LinkedDoc+RDF metadata

GraphFS is a filesystem-aware toolkit that parses RDF/Turtle documentation embedded in source code and exposes it as a semantic graph queryable via SPARQL and GraphQL. It turns implicit code relationships into explicit, navigable, and analyzable knowledge.

## 🎯 Vision

**From file-based thinking → concept-based thinking**

Instead of mentally tracking which files depend on what, GraphFS builds a living semantic model of your codebase that understands:
- Module dependencies and relationships
- Data flow and transformations
- API contracts and interfaces
- Security boundaries and constraints
- Team ownership and responsibilities

## 🚀 Quick Start

```bash
# Install GraphFS
go install github.com/justin4957/graphfs@latest

# Initialize in your project
cd /path/to/your/project
graphfs init

# Start query endpoints
graphfs serve
# 🌐 GraphQL endpoint: http://localhost:7681/graphql
# 🔍 SPARQL endpoint: http://localhost:7681/sparql

# Query your codebase semantically
graphfs query "show modules tagged security-critical"
```

## 📚 Core Concepts

### LinkedDoc+RDF Format

GraphFS parses embedded RDF/Turtle metadata from source code comments:

```go
/*
# Module: services/auth.go
Authentication and authorization service.

## Linked Modules
- [crypto](../utils/crypto.go) - Encryption utilities
- [session](./session.go) - Session management

## Tags
security, authentication, api

## Exports
AuthService, ValidateToken, CreateSession

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
<this> a code:Module ;
    code:name "services/auth.go" ;
    code:description "Authentication and authorization service" ;
    code:linksTo [
        code:name "crypto" ;
        code:path "../utils/crypto.go" ;
        code:relationship "Encryption utilities"
    ] ;
    code:exports :AuthService, :ValidateToken, :CreateSession ;
    code:tags "security", "authentication", "api" .
<!-- End LinkedDoc RDF -->
*/
package services
```

### Automatic Graph Construction

GraphFS scans your codebase and:
1. **Extracts** RDF triples from LinkedDoc headers
2. **Validates** semantic relationships and links
3. **Builds** an in-memory knowledge graph
4. **Indexes** modules, exports, dependencies, and tags
5. **Exposes** query interfaces (SPARQL, GraphQL, REST)

## 💡 Key Use Cases

### 1. Intelligent Code Navigation

```sparql
# Find all modules that depend on the auth service
PREFIX code: <https://schema.codedoc.org/>

SELECT ?module ?relationship WHERE {
  ?module code:linksTo ?link .
  ?link code:path "services/auth.go" ;
        code:relationship ?relationship .
}
```

### 2. Impact Analysis Before Refactoring

```bash
$ graphfs impact --module services/payment.go

🔍 Impact Analysis: services/payment.go

Direct Dependencies (3):
  • handlers/checkout.go - "Payment processing endpoint"
  • services/billing.go - "Invoice generation"
  • workers/subscription.go - "Recurring payments"

Transitive Impact (12 modules):
  • 8 handler modules
  • 3 background workers
  • 1 external webhook

⚠️  Risk Level: HIGH
💡 Recommendation: Coordinate with billing-team, payments-team
```

### 3. Architecture Validation

```bash
$ graphfs validate --rules .graphfs-rules.yml

✅ No UI modules directly access database (0 violations)
❌ Found 2 security boundary leaks:
   • internal/admin.go → public/api.go (line 47)
   • internal/config.go → external/webhook.go (line 112)
⚠️  3 modules missing LinkedDoc headers
```

### 4. AI-Powered Development Context

```bash
$ graphfs context --task "add rate limiting to API endpoints"

🤖 Gathering context for: add rate limiting to API endpoints

📦 Relevant Modules:
  • middleware/ratelimit.go - Existing rate limiting (Redis-based)
  • handlers/api.go - API endpoint handlers
  • config/limits.go - Rate limit configuration

🔗 Related Patterns:
  • services/cache.go uses similar Redis client
  • middleware/auth.go shows middleware integration pattern

📋 Suggested Approach:
  1. Extend middleware/ratelimit.go with per-endpoint limits
  2. Update handlers/api.go to apply middleware
  3. Add configuration to config/limits.go

📖 References:
  • docs/architecture/middleware.md
  • examples/rate-limiting-pattern.md
```

### 5. Automated Documentation Generation

```bash
$ graphfs docs generate --format interactive

📖 Generated Documentation:

📊 Architecture Overview:
  • Module dependency graph: docs/generated/dependencies.svg
  • Data flow diagram: docs/generated/dataflow.svg
  • Security boundaries: docs/generated/security-zones.svg

📚 Module Documentation:
  • API Reference: docs/generated/api/
  • Service Documentation: docs/generated/services/
  • Integration Guides: docs/generated/integrations/

🌐 Interactive Explorer:
  http://localhost:8080/explorer
```

## 🏗️ Architecture

```
graphfs/
├── cmd/
│   ├── graphfs/           # Main CLI tool
│   ├── server/            # Query server (SPARQL/GraphQL)
│   └── validate/          # Validation tool
├── pkg/
│   ├── parser/            # RDF/Turtle parser
│   ├── scanner/           # Filesystem scanner
│   ├── graph/             # Knowledge graph builder
│   ├── query/             # SPARQL/GraphQL engines
│   ├── rules/             # Architecture rule engine
│   └── codegen/           # Documentation generator
├── internal/
│   ├── store/             # Triple store implementation
│   ├── cache/             # Query result caching
│   └── index/             # Search indexing
└── examples/
    ├── go-project/        # Example Go project
    ├── python-project/    # Example Python project
    └── polyglot/          # Multi-language example
```

## 🗺️ Roadmap

### Phase 1: Core Infrastructure (Weeks 1-2)
- [x] Project initialization
- [ ] RDF/Turtle parser for LinkedDoc headers
- [ ] Filesystem scanner with language detection
- [ ] In-memory triple store
- [ ] Basic SPARQL query engine
- [ ] CLI tool skeleton (`graphfs init`, `graphfs scan`, `graphfs query`)

### Phase 2: Query Interfaces (Weeks 3-4)
- [ ] SPARQL HTTP endpoint
- [ ] GraphQL schema generation
- [ ] GraphQL server implementation
- [ ] REST API for common queries
- [ ] Query result caching
- [ ] CLI interactive mode

### Phase 3: Analysis Tools (Weeks 5-6)
- [ ] Dependency graph analysis
- [ ] Impact analysis engine
- [ ] Architecture rule validation
- [ ] Dead code detection
- [ ] Circular dependency detection
- [ ] Security boundary analysis

### Phase 4: Documentation Generation (Weeks 7-8)
- [ ] Markdown documentation generator
- [ ] SVG/GraphViz diagram generation
- [ ] Mermaid diagram support
- [ ] Interactive HTML explorer
- [ ] API reference generator
- [ ] Integration guide templates

### Phase 5: Advanced Features (Weeks 9-12)
- [ ] Multi-language support (Go, Python, JavaScript, Rust, Java)
- [ ] Git integration for historical analysis
- [ ] Team ownership mapping
- [ ] CI/CD integration hooks
- [ ] VS Code extension
- [ ] AI agent integration SDK

### Phase 6: Production Readiness (Weeks 13-16)
- [ ] Performance optimization (large codebases 10k+ files)
- [ ] Persistent storage options (RocksDB, PostgreSQL)
- [ ] Distributed graph support
- [ ] Real-time updates via file watching
- [ ] Cloud-native deployment (Docker, Kubernetes)
- [ ] Metrics and observability

## 📊 Expected Benefits

### Developer Productivity
- **Onboarding**: 3 months → 2 weeks (understanding codebase structure)
- **Impact Analysis**: Days of manual tracing → seconds
- **Code Navigation**: Text search → semantic relationship traversal
- **Debugging**: "Who calls this?" → Full call graph visualization

### Code Quality
- **Architecture Enforcement**: Automated validation of design rules
- **Technical Debt**: Quantified, prioritized, tracked over time
- **Documentation**: Always up-to-date, generated from code metadata
- **Dependency Management**: Visual boundaries, circular dependency prevention

### Team Collaboration
- **Knowledge Sharing**: Institutional knowledge preserved as metadata
- **Cross-Team Coordination**: Understand downstream impacts
- **Code Reviews**: Semantic context automatically provided
- **Refactoring Safety**: Data-driven risk assessment

### AI/LLM Integration
- **Context-Aware Coding**: AI sees full semantic relationships
- **Intelligent Suggestions**: Based on actual usage patterns
- **Automated Refactoring**: With dependency awareness
- **Documentation Generation**: Semantic understanding of code purpose

## 🎯 Target Audiences

### Individual Developers
- Navigate unfamiliar codebases quickly
- Understand impact of changes before making them
- Generate documentation automatically

### Engineering Teams
- Enforce architectural boundaries
- Coordinate large-scale refactorings
- Maintain living documentation

### Engineering Managers
- Visualize system architecture
- Track technical debt quantitatively
- Assess refactoring costs and risks

### AI/LLM Tool Builders
- Provide semantic code context to AI models
- Enable intelligent code generation
- Build code analysis and transformation tools

## 🔧 Installation

### From Source
```bash
git clone https://github.com/justin4957/graphfs
cd graphfs
go install ./cmd/graphfs
```

### From Binary
```bash
# Coming soon: Pre-built binaries for Linux, macOS, Windows
```

### Docker
```bash
# Coming soon: Docker images
docker run -v $(pwd):/workspace graphfs/graphfs:latest init
```

## 📖 Documentation

- [Getting Started Guide](docs/getting-started.md) *(coming soon)*
- [LinkedDoc Format Specification](docs/linkedoc-format.md) *(coming soon)*
- [SPARQL Query Examples](docs/sparql-examples.md) *(coming soon)*
- [Architecture Rules](docs/architecture-rules.md) *(coming soon)*
- [API Reference](docs/api-reference.md) *(coming soon)*
- [Integration Guides](docs/integrations/) *(coming soon)*

## 🤝 Contributing

GraphFS is in early development. Contributions are welcome!

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🌟 Inspiration

GraphFS builds on ideas from:
- **LinkedDoc** - Human + AI readable documentation format
- **RDF/Semantic Web** - Knowledge representation standards
- **Language Server Protocol** - Code intelligence infrastructure
- **Sourcetrail** - Visual code navigation (discontinued)
- **Kythe** - Google's code indexing system
- **Sourcegraph** - Code search and intelligence platform

## 🔗 Related Projects

- **ec2-test-apps** - Original LinkedDoc implementation and validation
- **tools/linkedoc_build.go** - LinkedDoc parser and validator
- **docs/REFACTORING_GUIDE.md** - LinkedDoc+RDF documentation standard

---

**Status**: 🚧 Early Development - Phase 1 in progress

**Maintained by**: [@justin4957](https://github.com/justin4957)

**Questions?** Open an issue or start a discussion!
