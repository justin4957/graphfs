/*
# Module: utils/crypto.go
Cryptographic utilities for secure operations.

Provides password hashing and token validation utilities.
Used by authentication service for security-critical operations.

## Linked Modules
None (utility module with no dependencies)

## Tags
utility, security, cryptography

## Exports
CryptoHelper, NewCryptoHelper

## Security Boundaries
This module handles sensitive cryptographic operations.
Must use secure algorithms and never expose raw credentials.

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix sec: <https://schema.codedoc.org/security/> .

<#crypto.go> a code:Module ;
    code:name "utils/crypto.go" ;
    code:description "Cryptographic utilities for secure operations" ;
    code:language "go" ;
    code:layer "utility" ;
    code:exports <#CryptoHelper>, <#NewCryptoHelper> ;
    code:tags "utility", "security", "cryptography" ;
    code:isLeaf true ;
    sec:securityBoundary true ;
    sec:handlesSensitiveData true .

<#CryptoHelper> a code:Type ;
    code:name "CryptoHelper" ;
    code:kind "struct" ;
    code:description "Helper for cryptographic operations" ;
    code:hasMethod <#CryptoHelper.HashPassword>, <#CryptoHelper.ValidateHash>, <#CryptoHelper.GenerateToken> .

<#CryptoHelper.HashPassword> a code:Method ;
    code:name "HashPassword" ;
    code:description "Hashes password securely" ;
    sec:securityCritical true .

<#CryptoHelper.ValidateHash> a code:Method ;
    code:name "ValidateHash" ;
    code:description "Validates hashed value" ;
    sec:securityCritical true .

<#CryptoHelper.GenerateToken> a code:Method ;
    code:name "GenerateToken" ;
    code:description "Generates secure token" ;
    sec:securityCritical true .

<#NewCryptoHelper> a code:Function ;
    code:name "NewCryptoHelper" ;
    code:description "Creates new crypto helper instance" ;
    code:returns <#CryptoHelper> .
<!-- End LinkedDoc RDF -->
*/

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CryptoHelper provides cryptographic utilities
type CryptoHelper struct {
	salt string
}

// NewCryptoHelper creates a new crypto helper
func NewCryptoHelper() *CryptoHelper {
	return &CryptoHelper{
		salt: "example-salt-123", // In real implementation, use secure random salt
	}
}

// HashPassword hashes a password securely
func (c *CryptoHelper) HashPassword(password string) string {
	// Stubbed hashing - in real implementation use bcrypt or argon2
	hasher := sha256.New()
	hasher.Write([]byte(password + c.salt))
	return hex.EncodeToString(hasher.Sum(nil))
}

// ValidateHash validates a hash
func (c *CryptoHelper) ValidateHash(hash string) bool {
	// Stubbed validation logic
	return len(hash) > 0
}

// GenerateToken generates a secure token
func (c *CryptoHelper) GenerateToken(userID string) string {
	// Stubbed token generation - in real implementation use JWT or similar
	hasher := sha256.New()
	hasher.Write([]byte(userID + c.salt))
	return fmt.Sprintf("token_%s", hex.EncodeToString(hasher.Sum(nil))[:16])
}
