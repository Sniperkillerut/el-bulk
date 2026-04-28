package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

type ProxyHandler struct {
	client *http.Client
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GET /api/proxy/image?url=...
func (h *ProxyHandler) Image(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		render.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		render.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	// Whitelist check
	allowedDomains := []string{
		"lorcana-api.com",
		"cardboardcrack.com",
		"dragonshield.com",
		"unsplash.com",
		"shopify.com",
	}

	allowed := false
	for _, domain := range allowedDomains {
		if strings.HasSuffix(parsedURL.Host, domain) {
			allowed = true
			break
		}
	}

	if !allowed {
		logger.WarnCtx(r.Context(), "[proxy] Rejected request to non-whitelisted domain: %s", parsedURL.Host)
		render.Error(w, "domain not allowed", http.StatusForbidden)
		return
	}

	// Fetch the image
	req, err := http.NewRequestWithContext(r.Context(), "GET", targetURL, nil)
	if err != nil {
		render.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	// Some providers might block requests without a proper User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := h.client.Do(req)
	if err != nil {
		logger.ErrorCtx(r.Context(), "[proxy] Failed to fetch image: %v", err)
		render.Error(w, "failed to fetch image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WarnCtx(r.Context(), "[proxy] Target returned status %d for %s", resp.StatusCode, targetURL)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	// Copy headers
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	} else {
		w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day by default
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, resp.Body)
}
