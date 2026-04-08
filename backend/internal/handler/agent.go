package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/bedrock"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

// ChatRequest represents a chat request to the AI agent.
type ChatRequest struct {
	Message      string `json:"message"`
	Symbol       string `json:"symbol"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	ApiKey       string `json:"apiKey"`
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
}

// ModelsRequest represents a request to list available models for a provider.
type ModelsRequest struct {
	Provider     string `json:"provider"`
	ApiKey       string `json:"apiKey"`
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
}

// InitLLM initializes default required ENV vars checks or general startup logic.
func (h *Handlers) InitLLM() {
	if os.Getenv("AWS_REGION") == "" {
		log.Println("WARNING: AWS_REGION not set. Make sure your AWS environment variables are configured for Bedrock.")
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Println("WARNING: OPENAI_API_KEY not set. OpenAI won't work.")
	}
}

// GetLLM dynamically instantiates the appropriate LLM client.
func (h *Handlers) GetLLM(ctx context.Context, provider, model, apiKey, awsAccessKey, awsSecretKey, awsRegion string) (llms.Model, error) {
	switch provider {
	case "openai":
		if model == "" {
			model = "gpt-3.5-turbo"
		}
		opts := []openai.Option{openai.WithModel(model)}
		if apiKey != "" {
			opts = append(opts, openai.WithToken(apiKey))
		}
		return openai.New(opts...)
	case "anthropic":
		if model == "" {
			model = "claude-3-sonnet-20240229"
		}
		opts := []anthropic.Option{anthropic.WithModel(model)}
		if apiKey != "" {
			opts = append(opts, anthropic.WithToken(apiKey))
		}
		return anthropic.New(opts...)
	case "google":
		if model == "" {
			model = "gemini-1.5-pro"
		}
		opts := []googleai.Option{googleai.WithDefaultModel(model)}
		if apiKey != "" {
			opts = append(opts, googleai.WithAPIKey(apiKey))
		}
		return googleai.New(ctx, opts...)
	case "qwen":
		if model == "" {
			model = "qwen-turbo"
		}
		opts := []openai.Option{
			openai.WithModel(model),
			openai.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		}
		if apiKey != "" {
			opts = append(opts, openai.WithToken(apiKey))
		}
		return openai.New(opts...)
	case "bedrock":
		if model == "" {
			model = "anthropic.claude-3-sonnet-20240229-v1:0"
		}
		// If the user-supplied AWS credentials, build the bedrock runtime client using them.
		if awsAccessKey != "" && awsSecretKey != "" {
			region := awsRegion
			if region == "" {
				region = "us-east-1"
			}
			cfg, err := config.LoadDefaultConfig(ctx,
				config.WithRegion(region),
				config.WithCredentialsProvider(
					credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, ""),
				),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to build AWS config: %w", err)
			}
			brc := bedrockruntime.NewFromConfig(cfg)
			return bedrock.New(
				bedrock.WithModel(model),
				bedrock.WithClient(brc),
			)
		}
		// Fall back to default AWS CLI profile / IAM role.
		return bedrock.New(bedrock.WithModel(model))
	default:
		return bedrock.New(bedrock.WithModel("anthropic.claude-3-sonnet-20240229-v1:0"))
	}
}

// AnalyzeMarketData creates a synthesis of market data for a given symbol.
func (h *Handlers) AnalyzeMarketData(ctx context.Context, symbol string) (string, error) {
	if symbol == "" {
		return "No symbol provided for analysis.", nil
	}

	// Fetch real-time data
	quotes, err := h.VnstockClient.RealTimeQuotes(ctx, []string{symbol})
	if err != nil {
		return "", fmt.Errorf("failed to fetch quotes: %v", err)
	}

	if len(quotes) == 0 {
		return "No data found for symbol " + symbol, nil
	}

	q := quotes[0]
	dataStr := fmt.Sprintf("Symbol: %s, Real-time Close: %.2f, Volume: %d", q.Symbol, q.Close, q.Volume)
	return dataStr, nil
}

// HandleChat serves POST /api/chat — the main endpoint orchestrating the "multi-agent" pipeline.
func (h *Handlers) HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format."})
		return
	}

	ctx := c.Request.Context()

	llm, err := h.GetLLM(ctx, req.Provider, req.Model, req.ApiKey, req.AwsAccessKey, req.AwsSecretKey, req.AwsRegion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to init LLM: %v", err)})
		return
	}

	// 1. Data Agent Step: Fetch market data
	marketContext := ""
	if req.Symbol != "" {
		marketData, err := h.AnalyzeMarketData(ctx, req.Symbol)
		if err != nil {
			log.Printf("Data Agent Error: %v", err)
			marketContext = "Error fetching market data."
		} else {
			marketContext = "Market Data Context: " + marketData
		}
	} else {
		marketContext = "General Finance Query (no specific symbol provided)."
	}

	// 2. Analyzer & Advisor Step: Produce response
	systemPrompt := `You are an AI financial advisor agent. 
You are composed of a multi-agent framework where the data agent has provided you with the following context based on real-time fetching.
Review the market context below (if any), analyze it quickly, and formulate your advice and recommendation directly answering the user's question. Be professional but succinct.

CONTEXT:
%s
`
	promptContext := fmt.Sprintf(systemPrompt, marketContext)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, promptContext),
		llms.TextParts(llms.ChatMessageTypeHuman, req.Message),
	}

	completion, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(completion.Choices) > 0 {
		c.JSON(http.StatusOK, gin.H{"reply": completion.Choices[0].Content})
	} else {
		c.JSON(http.StatusOK, gin.H{"reply": "No response generated."})
	}
}

// HandleModels serves POST /api/models — fetches available models for a given provider.
func (h *Handlers) HandleModels(c *gin.Context) {
	var req ModelsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	var models []string

	switch req.Provider {
	case "openai":
		if req.ApiKey == "" {
			c.JSON(http.StatusOK, gin.H{"models": []string{"gpt-3.5-turbo", "gpt-4-turbo", "gpt-4o"}})
			return
		}
		client := &http.Client{}
		httpReq, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
		httpReq.Header.Set("Authorization", "Bearer "+req.ApiKey)
		resp, err := client.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var res struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &res); err == nil && len(res.Data) > 0 {
				for _, m := range res.Data {
					models = append(models, m.ID)
				}
			}
		}

	case "google":
		if req.ApiKey == "" {
			c.JSON(http.StatusOK, gin.H{"models": []string{"gemini-1.5-pro", "gemini-1.5-flash"}})
			return
		}
		client := &http.Client{}
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", req.ApiKey)
		httpReq, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var res struct {
				Models []struct {
					Name string `json:"name"`
				} `json:"models"`
			}
			if err := json.Unmarshal(body, &res); err == nil && len(res.Models) > 0 {
				for _, m := range res.Models {
					models = append(models, m.Name)
				}
			}
		}

	case "anthropic":
		if req.ApiKey == "" {
			c.JSON(http.StatusOK, gin.H{"models": []string{"claude-3-5-sonnet-20240620", "claude-3-opus-20240229", "claude-3-haiku-20240307"}})
			return
		}
		client := &http.Client{}
		httpReq, _ := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
		httpReq.Header.Set("x-api-key", req.ApiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
		resp, err := client.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var res struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &res); err == nil && len(res.Data) > 0 {
				for _, m := range res.Data {
					models = append(models, m.ID)
				}
			}
		}

	case "qwen":
		if req.ApiKey == "" {
			c.JSON(http.StatusOK, gin.H{"models": []string{"qwen-turbo", "qwen-plus", "qwen-max"}})
			return
		}
		client := &http.Client{}
		httpReq, _ := http.NewRequest("GET", "https://dashscope.aliyuncs.com/compatible-mode/v1/models", nil)
		httpReq.Header.Set("Authorization", "Bearer "+req.ApiKey)
		resp, err := client.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var res struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &res); err == nil && len(res.Data) > 0 {
				for _, m := range res.Data {
					models = append(models, m.ID)
				}
			}
		}

	case "bedrock":
		models = []string{
			// === Anthropic Claude 4.x — VERIFIED IDs only ===
			"anthropic.claude-opus-4-6-v1",
			"anthropic.claude-sonnet-4-6",
			"anthropic.claude-haiku-4-5-20251001-v1:0",
			// === Anthropic Claude 3.7 ===
			"anthropic.claude-3-7-sonnet-20250219-v1:0",
			// === Anthropic Claude 3.5 ===
			"anthropic.claude-3-5-sonnet-20241022-v2:0",
			"anthropic.claude-3-5-sonnet-20240620-v1:0",
			"anthropic.claude-3-5-haiku-20241022-v1:0",
			// === Anthropic Claude 3 ===
			"anthropic.claude-3-opus-20240229-v1:0",
			"anthropic.claude-3-sonnet-20240229-v1:0",
			"anthropic.claude-3-haiku-20240307-v1:0",
			// === Anthropic Claude 2 (Legacy) ===
			"anthropic.claude-v2:1",
			"anthropic.claude-instant-v1",
			// === Amazon Nova (2025) ===
			"amazon.nova-lite-v2:0",
			"amazon.nova-premier-v1:0",
			"amazon.nova-pro-v1:0",
			"amazon.nova-lite-v1:0",
			"amazon.nova-micro-v1:0",
			// === Amazon Titan ===
			"amazon.titan-text-premier-v1:0",
			"amazon.titan-text-express-v1",
			"amazon.titan-text-lite-v1",
			// === DeepSeek (2025) ===
			"deepseek.r1-v1:0",
			"deepseek.deepseek-v3-1-v1:0",
			// === Meta Llama 4 ===
			"meta.llama4-maverick-17b-instruct-v1:0",
			"meta.llama4-scout-17b-instruct-v1:0",
			// === Meta Llama 3.3 / 3.2 / 3.1 ===
			"meta.llama3-3-70b-instruct-v1:0",
			"meta.llama3-2-90b-instruct-v2:0",
			"meta.llama3-2-11b-instruct-v2:0",
			"meta.llama3-2-3b-instruct-v2:0",
			"meta.llama3-2-1b-instruct-v2:0",
			"meta.llama3-1-405b-instruct-v1:0",
			"meta.llama3-1-70b-instruct-v1:0",
			"meta.llama3-1-8b-instruct-v1:0",
			// === Mistral ===
			"mistral.mistral-large-2402-v1:0",
			"mistral.mistral-large-2407-v1:0",
			"mistral.mistral-small-2402-v1:0",
			"mistral.mixtral-8x7b-instruct-v0:1",
			// === Cohere ===
			"cohere.command-r-plus-v1:0",
			"cohere.command-r-v1:0",
			// === AI21 Jamba ===
			"ai21.jamba-1-5-large-v1:0",
			"ai21.jamba-1-5-mini-v1:0",
		}
	}

	if len(models) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "Failed to fetch models or invalid API key."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}
