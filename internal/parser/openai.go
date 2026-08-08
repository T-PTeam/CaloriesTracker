package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/root1/calories-tracker/internal/domain"
)

const systemPrompt = `Ти — нутриціолог-аналітик. Користувач відправляє текст із переліком з'їденого.
Твоя задача — розпізнати продукти, оцінити їхню вагу (якщо не вказана — взяти стандартну порцію), розрахувати калорії та БЖВ і віднести кожен продукт до ОДНІЄЇ категорії:
- high_quality_protein — якісний білок (м'ясо, риба, яйця, молочні, протеїн, бобові як білкове джерело)
- long_acting_carbs — повільні вуглеводи (крупи, цільнозернові, овочі крохмалисті, бобові як вуглеводне джерело)
- lipids — жири (олії, горіхи, авокадо, жирні соуси, вершкове масло)
- fast_acting_carbs — швидкі вуглеводи (цукор, солодощі, білий хліб, соки, фрукти з високим ГІ, десерти)

Також визначи дату прийому їжі з тексту (вчора, позавчора, сьогодні, 3 серпня, 2026-08-05 тощо).
Поле eaten_at обов'язкове у форматі YYYY-MM-DD (дата в часовому поясі Europe/Kyiv).
Якщо дату не вказано — став сьогоднішню дату з повідомлення користувача.
Не плутай дату приготування/покупки з датою споживання: важлива дата, коли з'їв.

Відповідь має бути СУВОРО у форматі JSON без жодного додаткового тексту:
{
  "total_calories": 2050,
  "total_protein": 138,
  "total_fat": 71.5,
  "total_carbs": 193,
  "eaten_at": "2026-08-05",
  "items": [
    {"name": "Вівсянка з молоком", "weight_g": 350, "calories": 325, "protein": 12, "fat": 8, "carbs": 50, "category": "long_acting_carbs"}
  ]
}`

type MealParser interface {
	Parse(ctx context.Context, rawText string) (domain.ParsedMeal, error)
}

type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func NewOpenAIClient(apiKey, model string, timeout time.Duration) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://api.openai.com/v1",
	}
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat responseFmt   `json:"response_format"`
	Temperature    float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFmt struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) Parse(ctx context.Context, rawText string) (domain.ParsedMeal, error) {
	rawText = strings.TrimSpace(rawText)
	if rawText == "" {
		return domain.ParsedMeal{}, fmt.Errorf("empty meal text")
	}

	nowLocal := time.Now().In(MealLocation())
	userContent := fmt.Sprintf(
		"Сьогоднішня дата (Europe/Kyiv): %s\nПоточний час: %s\n\nТекст користувача:\n%s",
		nowLocal.Format("2006-01-02"),
		nowLocal.Format("15:04"),
		rawText,
	)

	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		ResponseFormat: responseFmt{Type: "json_object"},
		Temperature:    0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return domain.ParsedMeal{}, fmt.Errorf("openai status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("decode openai response: %w", err)
	}
	if parsed.Error != nil {
		return domain.ParsedMeal{}, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return domain.ParsedMeal{}, fmt.Errorf("openai returned no choices")
	}

	return DecodeAndValidate(parsed.Choices[0].Message.Content)
}

func DecodeAndValidate(content string) (domain.ParsedMeal, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var meal domain.ParsedMeal
	if err := json.Unmarshal([]byte(content), &meal); err != nil {
		return domain.ParsedMeal{}, fmt.Errorf("decode meal json: %w", err)
	}

	return ValidateParsedMeal(meal)
}

func ValidateParsedMeal(meal domain.ParsedMeal) (domain.ParsedMeal, error) {
	if len(meal.Items) == 0 {
		return domain.ParsedMeal{}, fmt.Errorf("parsed meal has no items")
	}
	if meal.TotalCalories < 0 || meal.TotalProtein < 0 || meal.TotalFat < 0 || meal.TotalCarbs < 0 {
		return domain.ParsedMeal{}, fmt.Errorf("parsed meal has negative totals")
	}

	meal.EatenAt = strings.TrimSpace(meal.EatenAt)
	if meal.EatenAt != "" {
		if _, err := ResolveEatenAt(meal.EatenAt, time.Now().UTC()); err != nil {
			return domain.ParsedMeal{}, err
		}
	}

	var sumCal, sumP, sumF, sumC float64
	for i, item := range meal.Items {
		if strings.TrimSpace(item.Name) == "" {
			return domain.ParsedMeal{}, fmt.Errorf("item %d has empty name", i)
		}
		if item.WeightG < 0 || item.Calories < 0 || item.Protein < 0 || item.Fat < 0 || item.Carbs < 0 {
			return domain.ParsedMeal{}, fmt.Errorf("item %q has negative values", item.Name)
		}
		category := strings.TrimSpace(item.Category)
		if !domain.IsValidCategory(category) {
			return domain.ParsedMeal{}, fmt.Errorf("item %q has invalid category %q", item.Name, item.Category)
		}
		meal.Items[i].Category = category
		sumCal += item.Calories
		sumP += item.Protein
		sumF += item.Fat
		sumC += item.Carbs
	}

	if abs(meal.TotalCalories-sumCal) > max(50, meal.TotalCalories*0.25) {
		meal.TotalCalories = sumCal
	}
	if abs(meal.TotalProtein-sumP) > max(10, meal.TotalProtein*0.25) {
		meal.TotalProtein = sumP
	}
	if abs(meal.TotalFat-sumF) > max(10, meal.TotalFat*0.25) {
		meal.TotalFat = sumF
	}
	if abs(meal.TotalCarbs-sumC) > max(10, meal.TotalCarbs*0.25) {
		meal.TotalCarbs = sumC
	}

	return meal, nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
