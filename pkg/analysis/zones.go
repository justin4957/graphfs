/*
# Module: pkg/analysis/zones.go
Security zone classification and detection.

Classifies modules into security zones based on tags, metadata, and path patterns.
Supports both predefined zones and custom zone definitions.

## Linked Modules
- [../graph](../graph/graph.go) - Graph data structure

## Tags
analysis, security, zones

## Exports
SecurityZone, ZoneClassifier, ClassifyZones

<!-- LinkedDoc RDF -->
@prefix code: <https://schema.codedoc.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<#zones.go> a code:Module ;
    code:name "pkg/analysis/zones.go" ;
    code:description "Security zone classification and detection" ;
    code:language "go" ;
    code:layer "analysis" ;
    code:linksTo <../graph/graph.go> ;
    code:exports <#SecurityZone>, <#ZoneClassifier>, <#ClassifyZones> ;
    code:tags "analysis", "security", "zones" .
<!-- End LinkedDoc RDF -->
*/

package analysis

import (
	"strings"

	"github.com/justin4957/graphfs/pkg/graph"
)

// SecurityZone represents a security boundary level
type SecurityZone string

const (
	ZonePublic   SecurityZone = "public"   // External APIs, public endpoints
	ZoneTrusted  SecurityZone = "trusted"  // Internal services, authenticated
	ZoneInternal SecurityZone = "internal" // Private modules, implementation
	ZoneAdmin    SecurityZone = "admin"    // Administrative functions
	ZoneData     SecurityZone = "data"     // Database and storage layer
	ZoneUnknown  SecurityZone = "unknown"  // Unclassified
)

// ZoneInfo contains information about a security zone
type ZoneInfo struct {
	Zone        SecurityZone
	Description string
	RiskLevel   int // 1 (lowest) to 5 (highest)
}

var zoneInfoMap = map[SecurityZone]ZoneInfo{
	ZonePublic: {
		Zone:        ZonePublic,
		Description: "External APIs and public endpoints",
		RiskLevel:   5, // Highest risk - exposed to outside
	},
	ZoneTrusted: {
		Zone:        ZoneTrusted,
		Description: "Internal services with authentication",
		RiskLevel:   3,
	},
	ZoneInternal: {
		Zone:        ZoneInternal,
		Description: "Private modules and implementation details",
		RiskLevel:   2,
	},
	ZoneAdmin: {
		Zone:        ZoneAdmin,
		Description: "Administrative and privileged functions",
		RiskLevel:   4,
	},
	ZoneData: {
		Zone:        ZoneData,
		Description: "Database and storage layer",
		RiskLevel:   4,
	},
	ZoneUnknown: {
		Zone:        ZoneUnknown,
		Description: "Unclassified modules",
		RiskLevel:   3,
	},
}

// ModuleZone represents a module's security zone classification
type ModuleZone struct {
	Module     *graph.Module
	Zone       SecurityZone
	Confidence float64 // 0.0-1.0
	Reason     string
}

// ZoneClassifier classifies modules into security zones
type ZoneClassifier struct {
	graph *graph.Graph
}

// NewZoneClassifier creates a new zone classifier
func NewZoneClassifier(g *graph.Graph) *ZoneClassifier {
	return &ZoneClassifier{
		graph: g,
	}
}

// ClassifyZones classifies all modules into security zones
func ClassifyZones(g *graph.Graph) map[SecurityZone][]*ModuleZone {
	classifier := NewZoneClassifier(g)
	return classifier.ClassifyAll()
}

// ClassifyAll classifies all modules
func (zc *ZoneClassifier) ClassifyAll() map[SecurityZone][]*ModuleZone {
	result := make(map[SecurityZone][]*ModuleZone)

	for _, module := range zc.graph.Modules {
		mz := zc.ClassifyModule(module)
		result[mz.Zone] = append(result[mz.Zone], mz)
	}

	return result
}

// ClassifyModule classifies a single module
func (zc *ZoneClassifier) ClassifyModule(module *graph.Module) *ModuleZone {
	// Try classification by tags first (highest confidence)
	if zone, confidence, reason := zc.classifyByTags(module); confidence > 0.7 {
		return &ModuleZone{
			Module:     module,
			Zone:       zone,
			Confidence: confidence,
			Reason:     reason,
		}
	}

	// Try classification by path patterns
	if zone, confidence, reason := zc.classifyByPath(module); confidence > 0.6 {
		return &ModuleZone{
			Module:     module,
			Zone:       zone,
			Confidence: confidence,
			Reason:     reason,
		}
	}

	// Try classification by layer
	if zone, confidence, reason := zc.classifyByLayer(module); confidence > 0.5 {
		return &ModuleZone{
			Module:     module,
			Zone:       zone,
			Confidence: confidence,
			Reason:     reason,
		}
	}

	// Default to unknown
	return &ModuleZone{
		Module:     module,
		Zone:       ZoneUnknown,
		Confidence: 0.3,
		Reason:     "No clear zone indicators found",
	}
}

// classifyByTags classifies based on module tags
func (zc *ZoneClassifier) classifyByTags(module *graph.Module) (SecurityZone, float64, string) {
	for _, tag := range module.Tags {
		tagLower := strings.ToLower(tag)

		// Public zone indicators
		publicTags := []string{"public", "api", "external", "endpoint", "http", "rest", "graphql"}
		for _, pt := range publicTags {
			if strings.Contains(tagLower, pt) {
				return ZonePublic, 0.9, "Tagged as " + tag
			}
		}

		// Admin zone indicators
		adminTags := []string{"admin", "privileged", "superuser", "root"}
		for _, at := range adminTags {
			if strings.Contains(tagLower, at) {
				return ZoneAdmin, 0.95, "Tagged as " + tag
			}
		}

		// Data zone indicators
		dataTags := []string{"database", "storage", "persistence", "sql", "db", "store"}
		for _, dt := range dataTags {
			if strings.Contains(tagLower, dt) {
				return ZoneData, 0.9, "Tagged as " + tag
			}
		}

		// Internal zone indicators
		internalTags := []string{"internal", "private", "impl", "implementation"}
		for _, it := range internalTags {
			if strings.Contains(tagLower, it) {
				return ZoneInternal, 0.85, "Tagged as " + tag
			}
		}

		// Trusted zone indicators
		trustedTags := []string{"service", "business", "logic", "auth"}
		for _, tt := range trustedTags {
			if strings.Contains(tagLower, tt) {
				return ZoneTrusted, 0.8, "Tagged as " + tag
			}
		}
	}

	return ZoneUnknown, 0.0, ""
}

// classifyByPath classifies based on file path
func (zc *ZoneClassifier) classifyByPath(module *graph.Module) (SecurityZone, float64, string) {
	pathLower := strings.ToLower(module.Path)

	// Check for public paths
	if strings.Contains(pathLower, "/api/") || strings.Contains(pathLower, "/public/") ||
		strings.Contains(pathLower, "/handlers/") || strings.Contains(pathLower, "/endpoints/") {
		return ZonePublic, 0.75, "Path indicates public API"
	}

	// Check for admin paths
	if strings.Contains(pathLower, "/admin/") || strings.Contains(pathLower, "/privileged/") {
		return ZoneAdmin, 0.8, "Path indicates admin functionality"
	}

	// Check for data paths
	if strings.Contains(pathLower, "/database/") || strings.Contains(pathLower, "/storage/") ||
		strings.Contains(pathLower, "/persistence/") || strings.Contains(pathLower, "/store/") {
		return ZoneData, 0.75, "Path indicates data layer"
	}

	// Check for internal paths
	if strings.Contains(pathLower, "/internal/") || strings.Contains(pathLower, "/impl/") {
		return ZoneInternal, 0.7, "Path indicates internal module"
	}

	// Check for service paths
	if strings.Contains(pathLower, "/service/") || strings.Contains(pathLower, "/business/") {
		return ZoneTrusted, 0.65, "Path indicates service layer"
	}

	return ZoneUnknown, 0.0, ""
}

// classifyByLayer classifies based on architectural layer
func (zc *ZoneClassifier) classifyByLayer(module *graph.Module) (SecurityZone, float64, string) {
	layerLower := strings.ToLower(module.Layer)

	layerMap := map[string]SecurityZone{
		"api":         ZonePublic,
		"public":      ZonePublic,
		"handlers":    ZonePublic,
		"admin":       ZoneAdmin,
		"data":        ZoneData,
		"database":    ZoneData,
		"storage":     ZoneData,
		"internal":    ZoneInternal,
		"services":    ZoneTrusted,
		"business":    ZoneTrusted,
		"application": ZoneTrusted,
	}

	if zone, ok := layerMap[layerLower]; ok {
		return zone, 0.6, "Layer '" + module.Layer + "' indicates " + string(zone) + " zone"
	}

	return ZoneUnknown, 0.0, ""
}

// GetZoneInfo returns information about a security zone
func GetZoneInfo(zone SecurityZone) ZoneInfo {
	if info, ok := zoneInfoMap[zone]; ok {
		return info
	}
	return zoneInfoMap[ZoneUnknown]
}

// GetZoneRiskLevel returns the risk level for a zone (1-5)
func GetZoneRiskLevel(zone SecurityZone) int {
	return GetZoneInfo(zone).RiskLevel
}

// GetZoneDescription returns the description for a zone
func GetZoneDescription(zone SecurityZone) string {
	return GetZoneInfo(zone).Description
}
