package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

var ErrLLMUnavailable = errors.New("llm service unavailable")

type OfferInfo struct {
	OfferID     uuid.UUID
	Name        string
	Description string
	Tags        []string
}

type RankedOffer struct {
	OfferID uuid.UUID
	Comment string
}

type Ranker interface {
	Rank(ctx context.Context, target OfferInfo, candidates []OfferInfo) ([]RankedOffer, error)
}

type DeepSeekRanker struct {
	client openai.Client
	model  string
}

func NewDeepSeekRanker(apiKey, baseURL, model string) *DeepSeekRanker {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return &DeepSeekRanker{client: client, model: model}
}

const systemPrompt = `
		You are a barter matching assistant. Given a target offer and candidate offers, rank the candidates by relevance to the target, prioritizing items with approximately equal market value, category similarity, condition, and exchange fairness.
		
		Important:
		- Prefer candidates that are close in estimated value to the target offer.
		- Penalize offers that are significantly more expensive or significantly cheaper.
		- If possible, infer value from brand, condition, rarity, specifications, and category.
		- Relevance should balance BOTH semantic similarity and comparable value.
		- Do not recommend unfair exchanges unless no better alternatives exist.
		- Briefly explain why the offer matches, mentioning value equivalence when relevant.
		
		Respond with a JSON array only, no extra text:
		[{"offerId":"<uuid>","comment":"<why it matches>"}]
		
		Ordered from best to worst match.
		Language: Russian.`

func (r *DeepSeekRanker) Rank(ctx context.Context, target OfferInfo, candidates []OfferInfo) ([]RankedOffer, error) {
	userPrompt, err := buildUserPrompt(target, candidates)
	if err != nil {
		return nil, fmt.Errorf("%w: build prompt: %w", ErrLLMUnavailable, err)
	}

	completion, err := r.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: r.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{
				Type: "json_object",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMUnavailable, err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("%w: empty choices", ErrLLMUnavailable)
	}

	raw := completion.Choices[0].Message.Content

	ranked, err := parseRankedOffers(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: parse response: %w", ErrLLMUnavailable, err)
	}

	return ranked, nil
}

type candidateJSON struct {
	OfferID     string   `json:"offerId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type rankedOfferJSON struct {
	OfferID string `json:"offerId"`
	Comment string `json:"comment"`
}

func buildUserPrompt(target OfferInfo, candidates []OfferInfo) (string, error) {
	cs := make([]candidateJSON, 0, len(candidates))
	for _, c := range candidates {
		tags := c.Tags
		if tags == nil {
			tags = []string{}
		}
		cs = append(cs, candidateJSON{
			OfferID:     c.OfferID.String(),
			Name:        c.Name,
			Description: c.Description,
			Tags:        tags,
		})
	}

	candidatesJSON, err := json.Marshal(cs)
	if err != nil {
		return "", err
	}

	targetTags := target.Tags
	if targetTags == nil {
		targetTags = []string{}
	}
	tagsJSON, err := json.Marshal(targetTags)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Target offer:\n")
	sb.WriteString("Name: " + target.Name + "\n")
	sb.WriteString("Description: " + target.Description + "\n")
	sb.WriteString("Tags: " + string(tagsJSON) + "\n\n")
	sb.WriteString("Candidates:\n")
	sb.Write(candidatesJSON)

	return sb.String(), nil
}

// parseRankedOffers handles two response shapes:
//  1. A bare JSON array: [{"offerId":"...","comment":"..."},...]
//  2. A JSON object wrapping an array (some models do this despite json_object mode):
//     {"offers":[...]} or {"result":[...]}
func parseRankedOffers(raw string) ([]RankedOffer, error) {
	raw = strings.TrimSpace(raw)

	var items []rankedOfferJSON
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		return toRankedOffers(items)
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil, fmt.Errorf("unexpected response format: %s", raw)
	}
	for _, v := range obj {
		if err := json.Unmarshal(v, &items); err == nil {
			return toRankedOffers(items)
		}
	}

	return nil, fmt.Errorf("no ranked array found in response: %s", raw)
}

func toRankedOffers(items []rankedOfferJSON) ([]RankedOffer, error) {
	result := make([]RankedOffer, 0, len(items))
	for _, item := range items {
		id, err := uuid.Parse(item.OfferID)
		if err != nil {
			return nil, fmt.Errorf("invalid offerId %q: %w", item.OfferID, err)
		}
		result = append(result, RankedOffer{OfferID: id, Comment: item.Comment})
	}
	return result, nil
}
