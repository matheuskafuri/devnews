package classify

import "testing"

func TestClassifyAIML(t *testing.T) {
	cat := Classify("Training Large Language Models at Scale", "How we optimized transformer inference pipelines")
	if cat != AIML {
		t.Errorf("expected AI/ML, got %s", cat)
	}
}

func TestClassifyInfrastructure(t *testing.T) {
	cat := Classify("Building Our Kubernetes Platform", "How we deployed containers across multiple cloud regions")
	if cat != Infrastructure {
		t.Errorf("expected Infrastructure, got %s", cat)
	}
}

func TestClassifyDatabases(t *testing.T) {
	cat := Classify("Scaling PostgreSQL to 10TB", "Database sharding and replication strategies")
	if cat != Databases {
		t.Errorf("expected Databases, got %s", cat)
	}
}

func TestClassifyDistributedSystems(t *testing.T) {
	cat := Classify("Durable Idempotency in Distributed Services", "Implementing consensus across microservices with Kafka")
	if cat != DistributedSystems {
		t.Errorf("expected Distributed Systems, got %s", cat)
	}
}

func TestClassifySecurity(t *testing.T) {
	cat := Classify("Zero Trust Authentication at Scale", "How we implemented TLS encryption and OAuth across services")
	if cat != Security {
		t.Errorf("expected Security, got %s", cat)
	}
}

func TestClassifyDeveloperTools(t *testing.T) {
	cat := Classify("Building a Better CLI Developer Experience", "Our new debugger and profiler tooling")
	if cat != DeveloperTools {
		t.Errorf("expected Developer Tools, got %s", cat)
	}
}

func TestClassifyEmptyInput(t *testing.T) {
	cat := Classify("", "")
	if cat != Platform {
		t.Errorf("expected Platform for empty input, got %s", cat)
	}
}

func TestClassifyDefaultsToPlatform(t *testing.T) {
	cat := Classify("Our Year in Review", "A look back at what we accomplished")
	if cat != Platform {
		t.Errorf("expected Platform for generic content, got %s", cat)
	}
}

func TestClassifyTitleWeightedHigher(t *testing.T) {
	// "kubernetes" in title should weight Infrastructure over description-only matches
	cat := Classify("Kubernetes in Production", "")
	if cat != Infrastructure {
		t.Errorf("expected Infrastructure from title keyword, got %s", cat)
	}
}

func TestResolveAlias(t *testing.T) {
	tests := []struct {
		alias    string
		expected Category
		wantErr  bool
	}{
		{"infra", Infrastructure, false},
		{"ai", AIML, false},
		{"db", Databases, false},
		{"distributed", DistributedSystems, false},
		{"security", Security, false},
		{"tools", DeveloperTools, false},
		{"platform", Platform, false},
		{"AI/ML", AIML, false},           // full name
		{"Infrastructure", Infrastructure, false}, // full name
		{"bogus", "", true},
	}

	for _, tt := range tests {
		got, err := ResolveAlias(tt.alias)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ResolveAlias(%q): expected error", tt.alias)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveAlias(%q): unexpected error: %v", tt.alias, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("ResolveAlias(%q) = %q, want %q", tt.alias, got, tt.expected)
		}
	}
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	if len(cats) != 7 {
		t.Errorf("expected 7 categories, got %d", len(cats))
	}
}
