# Minimal GraphFS Example Application

This is a minimal example application demonstrating the **LinkedDoc+RDF** metadata format used by GraphFS. All code is intentionally stubbed to focus on the documentation format, but the LinkedDoc headers are fully functional and can be parsed by GraphFS.

## Purpose

This example serves as:
1. **Reference implementation** - Shows correct LinkedDoc format for developers
2. **Test fixture** - Can be used to test GraphFS parser and query functionality
3. **Documentation** - Demonstrates semantic relationships in a real codebase

## Application Structure

```
minimal-app/
├── main.go                   # Entry point, coordinates services
├── models/
│   └── user.go              # User data model with validation
├── services/
│   ├── auth.go              # Authentication service
│   └── user.go              # User management service (CRUD)
└── utils/
    ├── logger.go            # Logging utility
    ├── crypto.go            # Cryptographic utilities
    └── validator.go         # Data validation utilities
```

## Semantic Graph

The LinkedDoc headers create a semantic knowledge graph with the following relationships:

### Module Dependencies
- `main.go` → depends on → `services/user.go`, `services/auth.go`, `utils/logger.go`
- `services/user.go` → depends on → `services/auth.go`, `models/user.go`, `utils/logger.go`
- `services/auth.go` → depends on → `utils/crypto.go`, `utils/logger.go`, `models/user.go`
- `models/user.go` → depends on → `utils/validator.go`
- All `utils/*` modules are leaf nodes (no dependencies)

### Layered Architecture
- **Entrypoint**: `main.go`
- **Service Layer**: `services/auth.go`, `services/user.go`
- **Model Layer**: `models/user.go`
- **Utility Layer**: `utils/logger.go`, `utils/crypto.go`, `utils/validator.go`

### Security Boundaries
- `services/auth.go` - Handles credentials (security-critical)
- `utils/crypto.go` - Cryptographic operations (security-critical)

### Architecture Rules
- `services/user.go` must use `AuthService` for permission checks
- `services/user.go` must validate data using `models.ValidateUser`
- All password operations must use `utils/crypto.go`

## Example GraphFS Queries

Once GraphFS is implemented, you can query this codebase:

### Find all security-critical modules
```sparql
PREFIX code: <https://schema.codedoc.org/>
PREFIX sec: <https://schema.codedoc.org/security/>

SELECT ?module ?description
WHERE {
  ?module a code:Module ;
          sec:securityCritical true ;
          code:description ?description .
}
```

### Find all modules that depend on auth.go
```sparql
PREFIX code: <https://schema.codedoc.org/>

SELECT ?module ?name
WHERE {
  ?module a code:Module ;
          code:name ?name ;
          code:linksTo <services/auth.go> .
}
```

### Find architecture rule violations
```sparql
PREFIX code: <https://schema.codedoc.org/>
PREFIX arch: <https://schema.codedoc.org/architecture/>

SELECT ?rule ?description ?applies
WHERE {
  ?rule a arch:Rule ;
        arch:description ?description ;
        arch:applies ?applies .
}
```

### Impact analysis - what breaks if I change crypto.go?
```sparql
PREFIX code: <https://schema.codedoc.org/>

SELECT ?module ?name
WHERE {
  ?module a code:Module ;
          code:name ?name ;
          code:linksTo+ <utils/crypto.go> .
}
```

## LinkedDoc Format

Each file includes a LinkedDoc header with:

1. **Markdown documentation** - Human-readable description
2. **Linked Modules** - Dependencies on other files
3. **Tags** - Categorization keywords
4. **Exports** - Public API surface
5. **RDF/Turtle metadata** - Machine-readable semantic triples

Example:
```go
/*
# Module: services/auth.go
Authentication and authorization service.

## Linked Modules
- [crypto](../utils/crypto.go) - Cryptographic utilities
- [logger](../utils/logger.go) - Logging utilities

## Tags
security, authentication, service

## Exports
AuthService, NewAuthService

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .

<#auth.go> a code:Module ;
    code:name "services/auth.go" ;
    code:description "Authentication service" ;
    code:linksTo <../utils/crypto.go> ;
    code:exports <#AuthService> .
<!-- End LinkedDoc RDF -->
*/
```

## Running the Example

```bash
cd examples/minimal-app
go run main.go
```

Expected output:
```
[2024-01-13 10:30:00] INFO: Starting minimal example application
Minimal GraphFS Example App
Auth Service: &{...}
User Service: &{...}
[2024-01-13 10:30:00] INFO: Application initialized successfully
```

## Using with GraphFS

Once GraphFS is implemented:

```bash
# Initialize GraphFS for this codebase
graphfs init

# Scan the codebase and build the knowledge graph
graphfs scan .

# Query the graph
graphfs query "SELECT * WHERE { ?s ?p ?o } LIMIT 10"

# Find all security-critical modules
graphfs query --file queries/security-critical.sparql

# Generate architecture diagram
graphfs diagram --output architecture.svg
```

## Development

This example is intentionally minimal and stubbed. It focuses on:
- ✅ Correct LinkedDoc format
- ✅ Rich semantic metadata
- ✅ Realistic module relationships
- ✅ Security boundaries and architecture rules
- ❌ Actual business logic (stubbed)
- ❌ Database integration (stubbed)
- ❌ Error handling (minimal)

Use this as a reference when adding LinkedDoc headers to your own codebase.
