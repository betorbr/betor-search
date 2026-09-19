package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Episode struct {
	Number int `json:"number"`
	Title  string `json:"title,omitempty"`
}

type Item struct {
	ID              string    `json:"id"`
	ProviderSlug    string    `json:"provider_slug"`
	ProviderURL     string    `json:"provider_url"`
	IMDBID          string    `json:"imdb_id,omitempty"`
	TMDBID          string    `json:"tmdb_id,omitempty"`
	ItemType        string    `json:"item_type"`
	MagnetURI       string    `json:"magnet_uri"`
	MagnetXT        string    `json:"magnet_xt"`
	MagnetDN        string    `json:"magnet_dn,omitempty"`
	TorrentName     string    `json:"torrent_name,omitempty"`
	TorrentNumPeers int      `json:"torrent_num_peers,omitempty"`
	TorrentNumSeeds int      `json:"torrent_num_seeds,omitempty"`
	TorrentFiles    []string  `json:"torrent_files,omitempty"`
	TorrentSize     int       `json:"torrent_size,omitempty"`
	Languages       []string  `json:"languages,omitempty"`
	Episodes        []Episode `json:"episodes,omitempty"`
	Seasons         []int     `json:"seasons,omitempty"`
	InsertedAt      time.Time `json:"inserted_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
	DownloadURL     string    `json:"download_url,omitempty"`
}

func (i *Item) UnmarshalJSON(data []byte) error {
	type itemAlias struct {
		ID              string    `json:"id"`
		ProviderSlug    string    `json:"provider_slug"`
		ProviderURL     string    `json:"provider_url"`
		IMDBID          string    `json:"imdb_id"`
		TMDBID          string    `json:"tmdb_id"`
		ItemType        string    `json:"item_type"`
		MagnetURI       string    `json:"magnet_uri"`
		MagnetXT        string    `json:"magnet_xt"`
		MagnetDN        string    `json:"magnet_dn"`
		TorrentName     string    `json:"torrent_name"`
		TorrentNumPeers int       `json:"torrent_num_peers"`
		TorrentNumSeeds int       `json:"torrent_num_seeds"`
		TorrentFiles    []string  `json:"torrent_files"`
		TorrentSize     int       `json:"torrent_size"`
		Languages       []string  `json:"languages"`
		Episodes        []Episode `json:"episodes"`
		Seasons         []int     `json:"seasons"`
		InsertedAt      json.RawMessage `json:"inserted_at"`
		UpdatedAt       json.RawMessage `json:"updated_at"`
		DownloadURL     string    `json:"download_url"`
	}

	var aux itemAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	i.ID = aux.ID
	i.ProviderSlug = aux.ProviderSlug
	i.ProviderURL = aux.ProviderURL
	i.IMDBID = aux.IMDBID
	i.TMDBID = aux.TMDBID
	i.ItemType = aux.ItemType
	i.MagnetURI = aux.MagnetURI
	i.MagnetXT = aux.MagnetXT
	i.MagnetDN = aux.MagnetDN
	i.TorrentName = aux.TorrentName
	i.TorrentNumPeers = aux.TorrentNumPeers
	i.TorrentNumSeeds = aux.TorrentNumSeeds
	i.TorrentFiles = aux.TorrentFiles
	i.TorrentSize = aux.TorrentSize
	i.Languages = aux.Languages
	i.Episodes = aux.Episodes
	i.Seasons = aux.Seasons
	i.DownloadURL = aux.DownloadURL

	if len(aux.InsertedAt) > 0 && string(aux.InsertedAt) != "null" {
		parsed, err := parseFlexibleTimeString(strings.Trim(string(aux.InsertedAt), "\""))
		if err != nil {
			return err
		}
		i.InsertedAt = parsed
	}
	if len(aux.UpdatedAt) > 0 && string(aux.UpdatedAt) != "null" {
		parsed, err := parseFlexibleTimeString(strings.Trim(string(aux.UpdatedAt), "\""))
		if err != nil {
			return err
		}
		i.UpdatedAt = parsed
	}
	return nil
}

func parseFlexibleTimeString(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported datetime format: %q", value)
}

type SearchFilter struct {
	Q        string
	IMDBID   string
	TMDBID   string
	ItemType string
	Season   []int
	Episode  []int
	Page     int
	Size     int
}

type SearchResult struct {
	Items []Item `json:"items"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

type Service struct {
	mu             sync.RWMutex
	items          []Item
	status         string
	lastExecution  time.Time
	lastSuccess    *time.Time
	lastError      string
	baseURL        string
	updateInterval time.Duration
}

func NewService(baseURL string, interval time.Duration) Service {
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	if baseURL == "" {
		baseURL = resolveBaseURLFromEnv()
	}
	return Service{
		status:         "DOWN",
		baseURL:        normalizeBaseURL(baseURL),
		updateInterval: interval,
	}
}

func resolveBaseURLFromEnv() string {
	if override := strings.TrimSpace(os.Getenv("BETOR_SEARCH_DOWNLOAD_ITEMS_URL")); override != "" {
		return normalizeBaseURL(stripCredentialsFromURL(override))
	}
	base := strings.TrimSpace(os.Getenv("BETOR_SEARCH_API_BASE_URL"))
	if base == "" {
		base = "https://api.betor.top/"
	}
	return strings.TrimRight(base, "/") + "/v1/admin/download-items/"
}

func normalizeBaseURL(baseURL string) string {
	if baseURL == "" {
		return "https://api.betor.top/v1/admin/download-items/"
	}
	baseURL = strings.TrimSpace(baseURL)
	baseURL = stripCredentialsFromURL(baseURL)
	if strings.Contains(baseURL, "/v1/admin/download-items/") {
		return baseURL
	}
	return strings.TrimRight(baseURL, "/") + "/v1/admin/download-items/"
}

func stripCredentialsFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err == nil {
		if parsed.User != nil {
			parsed.User = nil
			return parsed.String()
		}
		return rawURL
	}

	if idx := strings.Index(rawURL, "//"); idx >= 0 {
		prefix := rawURL[:idx+2]
		remainder := rawURL[idx+2:]
		if at := strings.Index(remainder, "@"); at >= 0 {
			return prefix + remainder[at+1:]
		}
	}
	return rawURL
}

func basicAuthHeaderValue() (string, bool) {
	value := strings.TrimSpace(os.Getenv("BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE"))
	if value == "" {
		return "", false
	}
	return value, true
}

func NewServiceWithItems(items []Item, baseURL string, interval time.Duration) Service {
	service := NewService(baseURL, interval)
	service.items = append([]Item(nil), items...)
	if len(items) > 0 {
		service.status = "UP"
		now := time.Now().UTC()
		service.lastSuccess = &now
		service.lastExecution = now
	}
	return service
}

func (s Service) Search(query string, filter SearchFilter) (SearchResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 || filter.Size > 100 {
		filter.Size = 50
	}

	matches := make([]Item, 0)
	for _, item := range s.items {
		if !matchQuery(item, query, filter) {
			continue
		}
		matches = append(matches, item)
	}

	start := (filter.Page - 1) * filter.Size
	if start >= len(matches) {
		return SearchResult{Items: []Item{}, Total: len(matches), Page: filter.Page, Size: filter.Size}, nil
	}

	end := start + filter.Size
	if end > len(matches) {
		end = len(matches)
	}

	return SearchResult{Items: matches[start:end], Total: len(matches), Page: filter.Page, Size: filter.Size}, nil
}

func (s *Service) Sync() error {
	url := s.baseURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://localhost:8080" + url
	}

	now := time.Now().UTC()
	s.mu.Lock()
	s.lastExecution = now
	s.status = "DEGRADED"
	s.lastError = ""
	s.mu.Unlock()

	items, err := s.fetchItemsFromURL(url)
	if err != nil {
		s.mu.Lock()
		s.status = "DOWN"
		s.lastError = err.Error()
		s.mu.Unlock()
		return err
	}

	if len(items) == 0 {
		return errors.New("empty catalog returned by BeTor")
	}

	s.mu.Lock()
	s.items = items
	s.status = "UP"
	s.lastError = ""
	nowSuccess := time.Now().UTC()
	s.lastSuccess = &nowSuccess
	s.mu.Unlock()
	return nil
}

func (s Service) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.status == "" {
		return "DOWN"
	}
	return s.status
}

func (s Service) LastExecution() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastExecution
}

func (s Service) LastSuccess() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastSuccess == nil {
		return nil
	}
	v := *s.lastSuccess
	return &v
}

func (s Service) LastError() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

func (s Service) Items() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Item, len(s.items))
	copy(items, s.items)
	return items
}

func (s Service) HasHealthyData() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items) > 0
}

func matchQuery(item Item, query string, filter SearchFilter) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query != "" {
		searchText := strings.ToLower(item.TorrentName + " " + item.MagnetDN + " " + item.ProviderSlug)
		if !strings.Contains(searchText, query) {
			return false
		}
	}
	if filter.IMDBID != "" && strings.TrimSpace(item.IMDBID) != strings.TrimSpace(filter.IMDBID) {
		return false
	}
	if filter.TMDBID != "" && strings.TrimSpace(item.TMDBID) != strings.TrimSpace(filter.TMDBID) {
		return false
	}
	if filter.ItemType != "" && !strings.EqualFold(item.ItemType, filter.ItemType) {
		return false
	}
	if len(filter.Season) > 0 && !containsInt(item.Seasons, filter.Season) {
		return false
	}
	if len(filter.Episode) > 0 {
		matched := false
		for _, episode := range item.Episodes {
			for _, target := range filter.Episode {
				if episode.Number == target {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func containsInt(values []int, targets []int) bool {
	for _, target := range targets {
		for _, value := range values {
			if value == target {
				return true
			}
		}
	}
	return false
}

func decodeItems(body []byte) ([]Item, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, errors.New("empty payload")
	}

	if strings.HasPrefix(trimmed, "[") {
		items, err := normalizeItems(json.RawMessage(body))
		if err == nil && len(items) > 0 {
			return items, nil
		}
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	for _, key := range []string{"items", "data", "results", "records"} {
		if raw, ok := payload[key]; ok {
			items, err := normalizeItems(raw)
			if err == nil && len(items) > 0 {
				return items, nil
			}
		}
	}

	if raw, ok := payload["item"]; ok {
		items, err := normalizeItems(raw)
		if err == nil && len(items) > 0 {
			return items, nil
		}
	}

	if looksLikeSingleItem(payload) {
		items, err := normalizeItems(payload)
		if err == nil && len(items) > 0 {
			return items, nil
		}
	}

	if len(payload) == 0 {
		return nil, errors.New("empty payload")
	}

	return nil, errors.New("no item list found in response")
}

func looksLikeSingleItem(payload map[string]any) bool {
	for _, key := range []string{"id", "provider_slug", "provider_url", "item_type", "magnet_uri", "torrent_name"} {
		if _, ok := payload[key]; ok {
			return true
		}
	}
	return false
}

func normalizeItems(raw any) ([]Item, error) {
	bytesValue, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var item Item
	if err := json.Unmarshal(bytesValue, &item); err == nil && item.ProviderSlug != "" {
		return []Item{item}, nil
	}

	var list []Item
	if err := json.Unmarshal(bytesValue, &list); err == nil {
		return list, nil
	}

	var mapped []map[string]any
	if err := json.Unmarshal(bytesValue, &mapped); err == nil {
		items := make([]Item, 0, len(mapped))
		for _, entry := range mapped {
			itemBytes, err := json.Marshal(entry)
			if err != nil {
				return nil, err
			}
			var decoded Item
			if err := json.Unmarshal(itemBytes, &decoded); err != nil {
				return nil, err
			}
			if decoded.ProviderSlug != "" || decoded.TorrentName != "" || decoded.MagnetURI != "" {
				items = append(items, decoded)
			}
		}
		return items, nil
	}

	return nil, errors.New("cannot decode payload into item list")
}

func (s *Service) fetchItemsFromURL(requestURL string) ([]Item, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	if value, ok := basicAuthHeaderValue(); ok {
		req.Header.Set("Authorization", "Basic "+value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("metadata request to %s failed with status %d: %s", requestURL, resp.StatusCode, summarizeResponseBody(body))
	}

	items, err := decodeItems(body)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	var metadata struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.Unmarshal(body, &metadata); err == nil && metadata.DownloadURL != "" {
		downloadReq, buildErr := http.NewRequest(http.MethodGet, metadata.DownloadURL, nil)
		if buildErr != nil {
			return nil, buildErr
		}
		downloadResp, downloadErr := client.Do(downloadReq)
		if downloadErr != nil {
			return nil, downloadErr
		}
		defer downloadResp.Body.Close()
		downloadBody, readErr := io.ReadAll(downloadResp.Body)
		if readErr != nil {
			return nil, readErr
		}
		if downloadResp.StatusCode < http.StatusOK || downloadResp.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("download_url %s failed with status %d: %s", metadata.DownloadURL, downloadResp.StatusCode, summarizeResponseBody(downloadBody))
		}
		items, err = decodeItems(downloadBody)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		return nil, fmt.Errorf("download_url %s responded with status %d but no usable item list was found: %s", metadata.DownloadURL, downloadResp.StatusCode, summarizeResponseBody(downloadBody))
	}

	return nil, fmt.Errorf("metadata request to %s succeeded but did not contain an item list or download_url: %s", requestURL, summarizeResponseBody(body))
}

func summarizeResponseBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "empty body"
	}
	if len(trimmed) > 220 {
		trimmed = trimmed[:220] + "..."
	}
	return trimmed
}

func (s Service) ParseHTTPQuery(query string) SearchFilter {
	values := map[string]string{}
	for _, part := range strings.Split(query, "&") {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		values[kv[0]] = kv[1]
	}
	filter := SearchFilter{Page: 1, Size: 50}
	if values["page"] != "" {
		if parsed, err := strconv.Atoi(values["page"]); err == nil && parsed > 0 {
			filter.Page = parsed
		}
	}
	if values["size"] != "" {
		if parsed, err := strconv.Atoi(values["size"]); err == nil && parsed > 0 {
			filter.Size = parsed
		}
	}
	if values["item_type"] != "" {
		filter.ItemType = values["item_type"]
	}
	if values["imdb_id"] != "" {
		filter.IMDBID = values["imdb_id"]
	}
	if values["tmdb_id"] != "" {
		filter.TMDBID = values["tmdb_id"]
	}
	if values["q"] != "" {
		filter.Q = values["q"]
	}
	return filter
}
