package graph

import (
	"testing"
)

func TestNewModule(t *testing.T) {
	module := NewModule("main.go", "<#main.go>")

	if module.Path != "main.go" {
		t.Errorf("Path = %s, want main.go", module.Path)
	}

	if module.URI != "<#main.go>" {
		t.Errorf("URI = %s, want <#main.go>", module.URI)
	}

	if len(module.Dependencies) != 0 {
		t.Errorf("Expected empty dependencies")
	}
}

func TestModule_AddDependency(t *testing.T) {
	module := NewModule("main.go", "<#main.go>")

	module.AddDependency("utils.go")
	module.AddDependency("models.go")

	if len(module.Dependencies) != 2 {
		t.Errorf("Dependencies count = %d, want 2", len(module.Dependencies))
	}

	// Test duplicate prevention
	module.AddDependency("utils.go")
	if len(module.Dependencies) != 2 {
		t.Errorf("Dependencies count = %d, want 2 (duplicate should not be added)", len(module.Dependencies))
	}
}

func TestModule_AddDependent(t *testing.T) {
	module := NewModule("utils.go", "<#utils.go>")

	module.AddDependent("main.go")
	module.AddDependent("test.go")

	if len(module.Dependents) != 2 {
		t.Errorf("Dependents count = %d, want 2", len(module.Dependents))
	}
}

func TestModule_AddExport(t *testing.T) {
	module := NewModule("main.go", "<#main.go>")

	module.AddExport("main")
	module.AddExport("helper")

	if len(module.Exports) != 2 {
		t.Errorf("Exports count = %d, want 2", len(module.Exports))
	}

	// Test duplicate prevention
	module.AddExport("main")
	if len(module.Exports) != 2 {
		t.Errorf("Exports count = %d, want 2 (duplicate should not be added)", len(module.Exports))
	}
}

func TestModule_AddTag(t *testing.T) {
	module := NewModule("main.go", "<#main.go>")

	module.AddTag("entrypoint")
	module.AddTag("core")

	if len(module.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(module.Tags))
	}
}

func TestModule_AddProperty(t *testing.T) {
	module := NewModule("main.go", "<#main.go>")

	module.AddProperty("custom:property", "value1")
	module.AddProperty("custom:property", "value2")

	if len(module.Properties["custom:property"]) != 2 {
		t.Errorf("Property values count = %d, want 2", len(module.Properties["custom:property"]))
	}
}

func TestModule_HasCircularDependency(t *testing.T) {
	// Create a simple graph with circular dependency
	graph := NewGraph("/test", nil)

	moduleA := NewModule("a.go", "<#a.go>")
	moduleB := NewModule("b.go", "<#b.go>")
	moduleC := NewModule("c.go", "<#c.go>")

	// A -> B -> C -> A (circular)
	moduleA.AddDependency("<#b.go>")
	moduleB.AddDependency("<#c.go>")
	moduleC.AddDependency("<#a.go>")

	graph.AddModule(moduleA)
	graph.AddModule(moduleB)
	graph.AddModule(moduleC)

	// Check if adding C as dependency of A would create a cycle
	if !moduleA.HasCircularDependency("<#c.go>", graph) {
		t.Error("Expected circular dependency to be detected")
	}

	// Check non-circular case
	moduleD := NewModule("d.go", "<#d.go>")
	graph.AddModule(moduleD)

	if moduleD.HasCircularDependency("<#a.go>", graph) {
		t.Error("Expected no circular dependency")
	}
}
