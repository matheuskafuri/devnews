package classify

import (
	"fmt"
	"strings"
	"unicode"
)

// Category represents an article classification.
type Category string

const (
	AIML               Category = "AI/ML"
	Infrastructure     Category = "Infrastructure"
	Databases          Category = "Databases"
	DistributedSystems Category = "Distributed Systems"
	Security           Category = "Security"
	DeveloperTools     Category = "Developer Tools"
	Platform           Category = "Platform"
)

// AllCategories returns all valid categories in canonical order.
func AllCategories() []Category {
	return []Category{AIML, Infrastructure, Databases, DistributedSystems, Security, DeveloperTools, Platform}
}

var categoryKeywords = map[Category][]string{
	AIML: {
		"machine learning", "deep learning", "neural", "llm", "gpt", "transformer",
		"inference", "training", "model", "embedding", "diffusion", "reinforcement",
		"classification", "nlp", "computer vision", "pytorch", "tensorflow",
	},
	Infrastructure: {
		"kubernetes", "docker", "container", "cloud", "aws", "gcp", "azure",
		"terraform", "infrastructure", "deploy", "cdn", "load balancer", "nginx",
		"networking", "dns", "edge", "proxy", "observability", "monitoring",
	},
	Databases: {
		"database", "sql", "nosql", "postgres", "postgresql", "mysql", "redis",
		"mongodb", "cassandra", "dynamodb", "indexing", "query", "schema",
		"migration", "replication", "sharding",
	},
	DistributedSystems: {
		"distributed", "consensus", "raft", "paxos", "microservice", "grpc",
		"message queue", "kafka", "event driven", "saga", "idempotent",
		"consistency", "partition", "replication", "failover", "circuit breaker",
	},
	Security: {
		"security", "vulnerability", "exploit", "authentication", "authorization",
		"encryption", "tls", "ssl", "certificate", "firewall", "zero trust",
		"oauth", "jwt", "xss", "csrf", "injection", "penetration",
	},
	DeveloperTools: {
		"developer", "tooling", "ide", "editor", "debugger", "profiler",
		"compiler", "linter", "formatter", "cli", "terminal", "git",
		"ci/cd", "pipeline", "build system", "package manager",
	},
	Platform: {
		"platform", "api", "sdk", "framework", "runtime", "language",
		"performance", "optimization", "architecture", "design", "engineering",
		"open source", "release", "announcement",
	},
}

// FocusAliases maps short CLI flags to full category names.
var FocusAliases = map[string]Category{
	"infra":       Infrastructure,
	"ai":          AIML,
	"db":          Databases,
	"distributed": DistributedSystems,
	"security":    Security,
	"tools":       DeveloperTools,
	"platform":    Platform,
}

// ResolveAlias maps a CLI alias to a Category.
func ResolveAlias(alias string) (Category, error) {
	alias = strings.ToLower(strings.TrimSpace(alias))
	if cat, ok := FocusAliases[alias]; ok {
		return cat, nil
	}
	// Also accept full category names (case-insensitive)
	for _, cat := range AllCategories() {
		if strings.EqualFold(string(cat), alias) {
			return cat, nil
		}
	}
	valid := make([]string, 0, len(FocusAliases))
	for k := range FocusAliases {
		valid = append(valid, k)
	}
	return "", fmt.Errorf("unknown focus %q (valid: %s)", alias, strings.Join(valid, ", "))
}

// Classify determines the category for an article based on title and description.
// Title keywords are weighted 2x. Returns Platform as default.
func Classify(title, description string) Category {
	titleTokens := tokenize(title)
	descTokens := tokenize(description)
	titleLower := strings.ToLower(title)
	descLower := strings.ToLower(description)

	var bestCat Category
	bestScore := 0

	for i, cat := range AllCategories() {
		score := 0
		keywords := categoryKeywords[cat]
		for _, kw := range keywords {
			if !strings.Contains(kw, " ") {
				// Single-word keyword
				for _, t := range titleTokens {
					if t == kw || strings.Contains(t, kw) {
						score += 2
					}
				}
				for _, t := range descTokens {
					if t == kw || strings.Contains(t, kw) {
						score++
					}
				}
			} else {
				// Multi-word keyword: check in pre-lowered text
				if strings.Contains(titleLower, kw) {
					score += 2
				}
				if strings.Contains(descLower, kw) {
					score++
				}
			}
		}
		if score > bestScore || (score == bestScore && score > 0 && i < categoryIndex(bestCat)) {
			bestScore = score
			bestCat = cat
		}
	}

	if bestScore == 0 {
		return Platform
	}
	return bestCat
}

func categoryIndex(cat Category) int {
	for i, c := range AllCategories() {
		if c == cat {
			return i
		}
	}
	return len(AllCategories())
}

func tokenize(s string) []string {
	var tokens []string
	for _, word := range strings.Fields(strings.ToLower(s)) {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if word != "" {
			tokens = append(tokens, word)
		}
	}
	return tokens
}
