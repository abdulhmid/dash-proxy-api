package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"api-source-proxy/app/model"
	"api-source-proxy/app/repository"
	apperrors "api-source-proxy/pkg/errors"
)

type ProxyService struct {
	apiSourceRepo repository.ApiSourceRepository
	logRepo       repository.LogRepository
	httpClient    *http.Client
	proxyURL      string
}

func NewProxyService(apiSourceRepo repository.ApiSourceRepository, logRepo repository.LogRepository, proxyURL string) *ProxyService {
	return &ProxyService{
		apiSourceRepo: apiSourceRepo,
		logRepo:       logRepo,
		proxyURL:      proxyURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    100,
				IdleConnTimeout: 90 * time.Second,
			},
		},
	}
}

func (s *ProxyService) ProxyRequest(ctx context.Context, sourceName string, msisdn string, user *model.User, apiKey *model.ApiKey, clientIP, clientName string) (map[string]interface{}, error) {
	source, err := s.apiSourceRepo.GetByName(ctx, sourceName)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, fmt.Sprintf("API source '%s' not found", sourceName))
	}

	if !source.IsActive {
		return nil, apperrors.Wrap(apperrors.ErrForbidden, fmt.Sprintf("API source '%s' is not active", sourceName))
	}

	requestID := uuid.New().String()

	formData := url.Values{
		"msisdn":    {msisdn},
		"username":  {source.Username},
		"requestId": {requestID},
	}

	if source.ExtraParams != "" {
		var extra map[string]string
		if err := json.Unmarshal([]byte(source.ExtraParams), &extra); err == nil {
			for k, v := range extra {
				formData.Set(k, v)
			}
		}
	}

	method := strings.ToUpper(source.Method)
	if method == "" {
		method = http.MethodPost
	}

	upstreamURL := resolvePathParams(source.BaseURL, formData)

	var req *http.Request

	reqBody := formData.Encode()
	if method == http.MethodGet || method == http.MethodDelete {
		reqURL := upstreamURL + "?" + reqBody
		req, err = http.NewRequestWithContext(ctx, method, reqURL, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, upstreamURL, bytes.NewBufferString(reqBody))
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to create request")
	}

	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	s.applyAuth(req, source)

	transport := &http.Transport{}
	if s.proxyURL != "" {
		proxyUrl, err := url.Parse(s.proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(source.TimeoutMs) * time.Millisecond,
	}
	if source.TimeoutMs <= 0 {
		client.Timeout = 30 * time.Second
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		logEntry := s.buildLog(user, apiKey, source, msisdn, requestID, reqBody, 0, err.Error(), clientIP, clientName, duration)
		_ = s.logRepo.Insert(ctx, logEntry)
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Sprintf("Upstream request failed: %v", err))
	}
	defer resp.Body.Close()

	respBodyBytes, _ := io.ReadAll(resp.Body)
	respBodyStr := string(respBodyBytes)

	logEntry := s.buildLog(user, apiKey, source, msisdn, requestID, reqBody, resp.StatusCode, respBodyStr, clientIP, clientName, duration)
	_ = s.logRepo.Insert(ctx, logEntry)

	if resp.StatusCode >= 400 {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Sprintf("Upstream returned error: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBodyBytes, &result); err != nil {
		result = map[string]interface{}{
			"raw_response": respBodyStr,
		}
	}

	return result, nil
}

func (s *ProxyService) applyAuth(req *http.Request, source *model.ApiSource) {
	authType := source.AuthType
	if authType == "" {
		authType = "custom"
	}

	switch authType {
	case "none":
		return

	case "basic":
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(source.AuthHeaders), &creds); err == nil && creds.Username != "" {
			req.SetBasicAuth(creds.Username, creds.Password)
		}

	case "bearer":
		var creds struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(source.AuthHeaders), &creds); err == nil && creds.Token != "" {
			req.Header.Set("Authorization", "Bearer "+creds.Token)
		}

	case "api-key":
		var creds struct {
			Key  string `json:"key"`
			In   string `json:"in"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(source.AuthHeaders), &creds); err == nil && creds.Key != "" && creds.Name != "" {
			if creds.In == "query" {
				q := req.URL.Query()
				q.Set(creds.Name, creds.Key)
				req.URL.RawQuery = q.Encode()
			} else {
				req.Header.Set(creds.Name, creds.Key)
			}
		}

	default:
		if source.AuthHeaders != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(source.AuthHeaders), &headers); err == nil {
				for k, v := range headers {
					req.Header.Set(k, v)
				}
			}
		}
	}
}

func (s *ProxyService) buildLog(user *model.User, apiKey *model.ApiKey, source *model.ApiSource, msisdn, requestID, reqBody string, statusCode int, respBody, clientIP, client string, durationMs int64) *model.ActivityLog {
	username := ""
	if user != nil {
		username = user.Username
	}
	if username == "" && apiKey != nil {
		username = apiKey.Name
	}
	if client == "" {
		client = "internal-team"
	}
	if username == "" {
		username = client
	}

	apiKeyID := ""
	if apiKey != nil {
		apiKeyID = apiKey.ID
	}

	userID := ""
	if user != nil {
		userID = user.ID
	}

	return &model.ActivityLog{
		UserID:        userID,
		ApiKeyID:      apiKeyID,
		Username:      username,
		Client:        client,
		ApiSourceName: source.Name,
		Method:        strings.ToUpper(source.Method),
		Path:          source.BaseURL,
		RequestBody:   reqBody,
		ResponseCode:  statusCode,
		ResponseBody:  respBody,
		ClientIP:      clientIP,
		DurationMs:    durationMs,
	}
}

func (s *ProxyService) ProxyRequestWithOverride(ctx context.Context, sourceName string, user *model.User, clientIP, clientName, methodOverride string, params map[string]interface{}) (map[string]interface{}, error) {
	source, err := s.apiSourceRepo.GetByName(ctx, sourceName)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, fmt.Sprintf("API source '%s' not found", sourceName))
	}

	if !source.IsActive {
		return nil, apperrors.Wrap(apperrors.ErrForbidden, fmt.Sprintf("API source '%s' is not active", sourceName))
	}

	requestID := uuid.New().String()

	method := strings.ToUpper(source.Method)
	if methodOverride != "" {
		method = strings.ToUpper(methodOverride)
	}

	formData := url.Values{}
	formData.Set("requestId", requestID)

	// Build log body from user params as JSON
	var logBodyBytes []byte
	if params != nil {
		logBodyBytes, _ = json.Marshal(params)
	}

	// Source extra params (static config)
	if source.ExtraParams != "" {
		var extra map[string]string
		if err := json.Unmarshal([]byte(source.ExtraParams), &extra); err == nil {
			for k, v := range extra {
				formData.Set(k, v)
			}
		}
	}

	// Merge params from request (override everything)
	if params != nil {
		for k, v := range params {
			formData.Set(k, fmt.Sprintf("%v", v))
		}
	}

	msisdn := formData.Get("msisdn")

	// Resolve path params in URL
	upstreamURL := resolvePathParams(source.BaseURL, formData)

	var req *http.Request

	reqBody := formData.Encode()
	if method == http.MethodGet || method == http.MethodDelete {
		reqURL := upstreamURL + "?" + reqBody
		req, err = http.NewRequestWithContext(ctx, method, reqURL, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, upstreamURL, bytes.NewBufferString(reqBody))
	}

	if logBodyBytes != nil {
		reqBody = string(logBodyBytes)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to create request")
	}

	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	s.applyAuth(req, source)

	transport := &http.Transport{}
	if s.proxyURL != "" {
		proxyUrl, err := url.Parse(s.proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(source.TimeoutMs) * time.Millisecond,
	}
	if source.TimeoutMs <= 0 {
		client.Timeout = 30 * time.Second
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		logEntry := s.buildLog(user, nil, source, msisdn, requestID, reqBody, 0, err.Error(), clientIP, clientName, duration)
		_ = s.logRepo.Insert(ctx, logEntry)
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Sprintf("Upstream request failed: %v", err))
	}
	defer resp.Body.Close()

	respBodyBytes, _ := io.ReadAll(resp.Body)
	respBodyStr := string(respBodyBytes)

	logEntry := s.buildLog(user, nil, source, msisdn, requestID, reqBody, resp.StatusCode, respBodyStr, clientIP, clientName, duration)
	_ = s.logRepo.Insert(ctx, logEntry)

	if resp.StatusCode >= 400 {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Sprintf("Upstream returned error: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBodyBytes, &result); err != nil {
		result = map[string]interface{}{
			"raw_response": respBodyStr,
		}
	}

	return result, nil
}

func (s *ProxyService) GetApiSources(ctx context.Context) ([]model.ApiSource, error) {
	return s.apiSourceRepo.List(ctx)
}

func (s *ProxyService) GetApiSourceByID(ctx context.Context, id string) (*model.ApiSource, error) {
	return s.apiSourceRepo.GetByID(ctx, id)
}

func (s *ProxyService) GetLogs(ctx context.Context, filter map[string]interface{}, page, limit int) ([]model.ActivityLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.logRepo.List(ctx, filter, page, limit)
}

func (s *ProxyService) ProcessLogFilter(userID, apiSourceName, startDate, endDate string) map[string]interface{} {
	filter := map[string]interface{}{}

	if userID != "" {
		filter["user_id"] = userID
	}
	if apiSourceName != "" {
		filter["api_source_name"] = apiSourceName
	}
	if startDate != "" {
		filter["start_date"] = startDate
	}
	if endDate != "" {
		filter["end_date"] = endDate
	}

	return filter
}

func (s *ProxyService) InitDefaultSources(ctx context.Context) error {
	sources := []struct {
		Name     string
		BaseURL  string
		Username string
	}{
		{
			Name:     "cp-cobra",
			BaseURL:  "https://xxxxx.xyz/getLocation",
			Username: "api-int-proxy",
		},
		{
			Name:     "cp-snake",
			BaseURL:  "https://xxxxxx.xyz/getLocation",
			Username: "api-int-proxyxxx",
		},
	}

		for _, src := range sources {
			_, err := s.apiSourceRepo.GetByName(ctx, src.Name)
			if err == apperrors.ErrNotFound {
				_ = s.apiSourceRepo.Create(ctx, &model.ApiSource{
					Name:        src.Name,
					BaseURL:     src.BaseURL,
					Username:    src.Username,
					AuthType:    "custom",
					Method:      "POST",
					IsActive:    true,
					TimeoutMs:   30000,
					ExtraParams: "",
				})
			}
		}
	return nil
}

func (s *ProxyService) CreateApiSource(ctx context.Context, source *model.ApiSource) error {
	return s.apiSourceRepo.Create(ctx, source)
}

func (s *ProxyService) UpdateApiSource(ctx context.Context, source *model.ApiSource) error {
	existing, err := s.apiSourceRepo.GetByID(ctx, source.ID)
	if err != nil {
		return err
	}
	source.CreatedAt = existing.CreatedAt
	return s.apiSourceRepo.Update(ctx, source)
}

func (s *ProxyService) DeleteApiSource(ctx context.Context, id string) error {
	return s.apiSourceRepo.Delete(ctx, id)
}

func (s *ProxyService) BuildProxyClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}
	return ip
}

func (s *ProxyService) LogRepo() repository.LogRepository {
	return s.logRepo
}

// resolvePathParams replaces {key} placeholders in the URL with formData values
// and removes those keys from formData so they aren't sent as query/body params.
func resolvePathParams(rawURL string, formData url.Values) string {
	result := rawURL
	for key, vals := range formData {
		placeholder := "{" + key + "}"
		if strings.Contains(result, placeholder) {
			if len(vals) > 0 {
				result = strings.ReplaceAll(result, placeholder, vals[0])
			}
			formData.Del(key)
		}
	}
	return result
}
