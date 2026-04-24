package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/el-bulk/backend/utils/logger"
)

type EDHRECPage struct {
	Container struct {
		JSONDict struct {
			CardLists []struct {
				Header    string `json:"header"`
				Tag       string `json:"tag"`
				CardViews []struct {
					Name string `json:"name"`
				} `json:"cardviews"`
			} `json:"cardlists"`
		} `json:"json_dict"`
	} `json:"container"`
}

var edhrecClient = &http.Client{
	Timeout: 10 * time.Second,
}

func FetchEDHRECRecommendations(ctx context.Context, commanderName string) ([]string, error) {
	sanitized := strings.ToLower(commanderName)
	// Remove symbols and replace spaces with dashes
	sanitized = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, sanitized)
	// Remove double dashes
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	url := fmt.Sprintf("https://json.edhrec.com/pages/commanders/%s.json", sanitized)
	logger.DebugCtx(ctx, "[EDHREC] Fetching recommendations from %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := edhrecClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return []string{}, nil
		}
		return nil, fmt.Errorf("edhrec returned status %d", resp.StatusCode)
	}

	var page EDHRECPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}

	var recommendations []string
	for _, list := range page.Container.JSONDict.CardLists {
		if list.Tag == "highsynergycards" || list.Tag == "topcards" {
			for _, card := range list.CardViews {
				recommendations = append(recommendations, card.Name)
			}
		}
	}

	return recommendations, nil
}
