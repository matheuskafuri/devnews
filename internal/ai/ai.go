package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/matheuskafuri/devnews/internal/config"
)

// Result holds the output from an LLM summarization call.
type Result struct {
	Summary string
	Tags    []string
}

// ArticleSummary holds minimal article data for theme detection.
type ArticleSummary struct {
	Title    string
	Category string
}

// Summarizer generates summaries and tags for articles.
type Summarizer interface {
	Summarize(ctx context.Context, title, description string) (Result, error)
	Brief(ctx context.Context, titles []string) (string, error)
	WhyItMatters(ctx context.Context, title, description string) (string, error)
	Themes(ctx context.Context, articles []ArticleSummary) ([]string, error)
}

// New creates a Summarizer from the given AI config.
func New(cfg *config.AIConfig, apiKey string) (Summarizer, error) {
	if cfg == nil || apiKey == "" {
		return nil, fmt.Errorf("AI not configured")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	switch cfg.Provider {
	case "claude":
		model := cfg.Model
		if model == "" {
			model = "claude-haiku-4-5-20251001"
		}
		return &claudeProvider{apiKey: apiKey, model: model, client: client}, nil
	case "openai":
		model := cfg.Model
		if model == "" {
			model = "gpt-4o-mini"
		}
		return &openaiProvider{apiKey: apiKey, model: model, client: client}, nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %q (valid: claude, openai)", cfg.Provider)
	}
}

const summarizePrompt = `Summarize this engineering blog post in one sentence (max 120 chars) and provide up to 3 topic tags (single words like: infrastructure, rust, performance, scaling, databases, security, frontend, api, mobile, devops).

Format your response EXACTLY like this:
SUMMARY: <one sentence summary>
TAGS: tag1, tag2, tag3

Title: %s
Description: %s`

const briefPrompt = `In one sentence (max 150 chars), summarize the main themes across these %d engineering blog posts:

%s`

const whyItMattersPrompt = `You are a senior engineering analyst. Given this blog post title and description, write a measured, technical "Why it matters" statement. Be precise and analytical. No hype or exclamation marks. 2-3 sentences, max 200 characters total.

Title: %s
Description: %s

Respond with ONLY the "why it matters" text, nothing else.`

const themesPrompt = `Given these engineering blog posts, identify 2-4 overarching technical themes. Be analytical and specific. Each theme should be under 60 characters.

Articles:
%s

Respond with one theme per line. No bullets, numbers, or other formatting.`

func parseSummaryResponse(text string) Result {
	var r Result
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SUMMARY:") {
			r.Summary = strings.TrimSpace(strings.TrimPrefix(line, "SUMMARY:"))
		} else if strings.HasPrefix(line, "TAGS:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "TAGS:"))
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(strings.ToLower(t))
				if t != "" {
					r.Tags = append(r.Tags, t)
				}
			}
			if len(r.Tags) > 3 {
				r.Tags = r.Tags[:3]
			}
		}
	}
	return r
}

func parseThemes(text string) []string {
	var themes []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		// Strip bullets and numbering
		line = strings.TrimLeft(line, "•-*")
		line = strings.TrimSpace(line)
		if len(line) > 2 && line[0] >= '0' && line[0] <= '9' {
			// Strip "1. " or "1) " style numbering
			for i, c := range line {
				if c == '.' || c == ')' {
					line = strings.TrimSpace(line[i+1:])
					break
				}
				if c < '0' || c > '9' {
					break
				}
			}
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 60 {
			line = line[:60]
		}
		themes = append(themes, line)
		if len(themes) >= 4 {
			break
		}
	}
	return themes
}

func formatArticlesForThemes(articles []ArticleSummary) string {
	var sb strings.Builder
	for _, a := range articles {
		sb.WriteString("- ")
		sb.WriteString(a.Title)
		if a.Category != "" {
			sb.WriteString(" [")
			sb.WriteString(a.Category)
			sb.WriteString("]")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// --- Claude provider ---

type claudeProvider struct {
	apiKey string
	model  string
	client *http.Client
}

type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []claudeMessage  `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *claudeProvider) Summarize(ctx context.Context, title, description string) (Result, error) {
	prompt := fmt.Sprintf(summarizePrompt, title, description)
	text, err := c.call(ctx, prompt)
	if err != nil {
		return Result{}, err
	}
	return parseSummaryResponse(text), nil
}

func (c *claudeProvider) Brief(ctx context.Context, titles []string) (string, error) {
	prompt := fmt.Sprintf(briefPrompt, len(titles), strings.Join(titles, "\n"))
	return c.call(ctx, prompt)
}

func (c *claudeProvider) WhyItMatters(ctx context.Context, title, description string) (string, error) {
	prompt := fmt.Sprintf(whyItMattersPrompt, title, description)
	text, err := c.call(ctx, prompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (c *claudeProvider) Themes(ctx context.Context, articles []ArticleSummary) ([]string, error) {
	prompt := fmt.Sprintf(themesPrompt, formatArticlesForThemes(articles))
	text, err := c.call(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return parseThemes(text), nil
}

func (c *claudeProvider) call(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(claudeRequest{
		Model:     c.model,
		MaxTokens: 256,
		Messages:  []claudeMessage{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("claude API %d: %s", resp.StatusCode, string(b))
	}

	var cr claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Content) == 0 {
		return "", fmt.Errorf("empty claude response")
	}
	return cr.Content[0].Text, nil
}

// --- OpenAI provider ---

type openaiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (o *openaiProvider) Summarize(ctx context.Context, title, description string) (Result, error) {
	prompt := fmt.Sprintf(summarizePrompt, title, description)
	text, err := o.call(ctx, prompt)
	if err != nil {
		return Result{}, err
	}
	return parseSummaryResponse(text), nil
}

func (o *openaiProvider) Brief(ctx context.Context, titles []string) (string, error) {
	prompt := fmt.Sprintf(briefPrompt, len(titles), strings.Join(titles, "\n"))
	return o.call(ctx, prompt)
}

func (o *openaiProvider) WhyItMatters(ctx context.Context, title, description string) (string, error) {
	prompt := fmt.Sprintf(whyItMattersPrompt, title, description)
	text, err := o.call(ctx, prompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (o *openaiProvider) Themes(ctx context.Context, articles []ArticleSummary) ([]string, error) {
	prompt := fmt.Sprintf(themesPrompt, formatArticlesForThemes(articles))
	text, err := o.call(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return parseThemes(text), nil
}

func (o *openaiProvider) call(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(openaiRequest{
		Model:    o.model,
		Messages: []openaiMessage{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("openai API %d: %s", resp.StatusCode, string(b))
	}

	var or openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
		return "", err
	}
	if len(or.Choices) == 0 {
		return "", fmt.Errorf("empty openai response")
	}
	return or.Choices[0].Message.Content, nil
}
