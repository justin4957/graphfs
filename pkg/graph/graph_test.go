package graph

import (
	"testing"

	"github.com/justin4957/graphfs/internal/store"
)

func TestNewGraph(t *testing.T) {
	ts := store.NewTripleStore()
	graph := NewGraph("/test/root", ts)

	if graph.Root != "/test/root" {
		t.Errorf("Root = %s, want /test/root", graph.Root)
	}

	if graph.Store != ts {
		t.Error("Store not set correctly")
	}

	if len(graph.Modules) != 0 {
		t.Error("Expected empty modules map")
	}
}

func TestGraph_AddModule(t *testing.T) {
	graph := NewGraph("/test", nil)

	module1 := NewModule("main.go", "<#main.go>")
	module1.Language = "go"
	module1.Layer = "app"

	graph.AddModule(module1)

	if graph.Statistics.TotalModules != 1 {
		t.Errorf("TotalModules = %d, want 1", graph.Statistics.TotalModules)
	}

	if graph.Statistics.ModulesByLanguage["go"] != 1 {
		t.Errorf("ModulesByLanguage[go] = %d, want 1", graph.Statistics.ModulesByLanguage["go"])
	}

	if graph.Statistics.ModulesByLayer["app"] != 1 {
		t.Errorf("ModulesByLayer[app] = %d, want 1", graph.Statistics.ModulesByLayer["app"])
	}
}

func TestGraph_GetModule(t *testing.T) {
	graph := NewGraph("/test", nil)

	module := NewModule("main.go", "<#main.go>")
	graph.AddModule(module)

	retrieved := graph.GetModule("main.go")
	if retrieved == nil {
		t.Fatal("Expected to retrieve module")
	}

	if retrieved.Path != "main.go" {
		t.Errorf("Retrieved module path = %s, want main.go", retrieved.Path)
	}

	// Test non-existent module
	if graph.GetModule("nonexistent.go") != nil {
		t.Error("Expected nil for non-existent module")
	}
}

func TestGraph_GetModulesByLanguage(t *testing.T) {
	graph := NewGraph("/test", nil)

	goModule1 := NewModule("main.go", "<#main.go>")
	goModule1.Language = "go"

	goModule2 := NewModule("utils.go", "<#utils.go>")
	goModule2.Language = "go"

	pyModule := NewModule("script.py", "<#script.py>")
	pyModule.Language = "python"

	graph.AddModule(goModule1)
	graph.AddModule(goModule2)
	graph.AddModule(pyModule)

	goModules := graph.GetModulesByLanguage("go")
	if len(goModules) != 2 {
		t.Errorf("GetModulesByLanguage(go) = %d modules, want 2", len(goModules))
	}

	pyModules := graph.GetModulesByLanguage("python")
	if len(pyModules) != 1 {
		t.Errorf("GetModulesByLanguage(python) = %d modules, want 1", len(pyModules))
	}
}

func TestGraph_GetModulesByLayer(t *testing.T) {
	graph := NewGraph("/test", nil)

	serviceModule1 := NewModule("auth.go", "<#auth.go>")
	serviceModule1.Layer = "services"

	serviceModule2 := NewModule("user.go", "<#user.go>")
	serviceModule2.Layer = "services"

	utilModule := NewModule("logger.go", "<#logger.go>")
	utilModule.Layer = "utils"

	graph.AddModule(serviceModule1)
	graph.AddModule(serviceModule2)
	graph.AddModule(utilModule)

	serviceModules := graph.GetModulesByLayer("services")
	if len(serviceModules) != 2 {
		t.Errorf("GetModulesByLayer(services) = %d modules, want 2", len(serviceModules))
	}

	utilModules := graph.GetModulesByLayer("utils")
	if len(utilModules) != 1 {
		t.Errorf("GetModulesByLayer(utils) = %d modules, want 1", len(utilModules))
	}
}

func TestGraph_GetModulesByTag(t *testing.T) {
	graph := NewGraph("/test", nil)

	module1 := NewModule("main.go", "<#main.go>")
	module1.AddTag("entrypoint")
	module1.AddTag("core")

	module2 := NewModule("cli.go", "<#cli.go>")
	module2.AddTag("entrypoint")

	module3 := NewModule("utils.go", "<#utils.go>")
	module3.AddTag("helper")

	graph.AddModule(module1)
	graph.AddModule(module2)
	graph.AddModule(module3)

	entrypointModules := graph.GetModulesByTag("entrypoint")
	if len(entrypointModules) != 2 {
		t.Errorf("GetModulesByTag(entrypoint) = %d modules, want 2", len(entrypointModules))
	}

	coreModules := graph.GetModulesByTag("core")
	if len(coreModules) != 1 {
		t.Errorf("GetModulesByTag(core) = %d modules, want 1", len(coreModules))
	}
}

func TestGraph_GetTransitiveDependencies(t *testing.T) {
	graph := NewGraph("/test", nil)

	// Create dependency chain: A -> B -> C
	moduleA := NewModule("a.go", "<#a.go>")
	moduleA.AddDependency("b.go")

	moduleB := NewModule("b.go", "<#b.go>")
	moduleB.AddDependency("c.go")

	moduleC := NewModule("c.go", "<#c.go>")

	graph.AddModule(moduleA)
	graph.AddModule(moduleB)
	graph.AddModule(moduleC)

	deps := graph.GetTransitiveDependencies("a.go")

	// Should include B and C (transitive)
	if len(deps) != 2 {
		t.Errorf("GetTransitiveDependencies = %d, want 2", len(deps))
	}
}

func TestGraph_GetDirectDependencies(t *testing.T) {
	graph := NewGraph("/test", nil)

	module := NewModule("main.go", "<#main.go>")
	module.AddDependency("utils.go")
	module.AddDependency("models.go")

	graph.AddModule(module)

	deps := graph.GetDirectDependencies("main.go")
	if len(deps) != 2 {
		t.Errorf("GetDirectDependencies = %d, want 2", len(deps))
	}
}

func TestGraph_GetDependents(t *testing.T) {
	graph := NewGraph("/test", nil)

	moduleA := NewModule("a.go", "<#a.go>")
	moduleA.AddDependent("b.go")
	moduleA.AddDependent("c.go")

	graph.AddModule(moduleA)

	dependents := graph.GetDependents("a.go")
	if len(dependents) != 2 {
		t.Errorf("GetDependents = %d, want 2", len(dependents))
	}
}
