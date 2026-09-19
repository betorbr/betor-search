package application

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewService_UsesDownloadItemsOverrideFromEnv(t *testing.T) {
	t.Setenv("BETOR_SEARCH_API_BASE_URL", "https://api.betor.top/")
	t.Setenv("BETOR_SEARCH_DOWNLOAD_ITEMS_URL", "https://example.com/custom-download/")

	service := NewService("", time.Minute)
	if got, want := service.baseURL, "https://example.com/custom-download/v1/admin/download-items/"; got != want {
		t.Fatalf("baseURL = %q, want %q", got, want)
	}
}

func TestService_Sync_UsesBasicAuthForMetadataDownloadURL(t *testing.T) {
	var metadataRequested bool
	var itemsRequested bool

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/admin/download-items/":
			metadataRequested = true
			if got := r.Header.Get("Authorization"); got != "Basic YmV0b3I6c2VjcmV0" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items_total":1,"download_url":"` + serverURL + `/items"}`))
		case "/items":
			itemsRequested = true
			if got := r.Header.Get("Authorization"); got != "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":[{"id":"1","provider_slug":"test","provider_url":"https://example.com/1","item_type":"movie","magnet_uri":"magnet:?xt=urn:btih:1","magnet_xt":"urn:btih:1","torrent_name":"Movie A","inserted_at":"2024-01-01T00:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	t.Setenv("BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE", "YmV0b3I6c2VjcmV0")

	service := NewService(server.URL, time.Minute)
	if err := service.Sync(); err != nil {
		t.Fatalf("Sync() returned error: %v", err)
	}
	if !metadataRequested {
		t.Fatal("metadata endpoint was not requested")
	}
	if !itemsRequested {
		t.Fatal("download_url endpoint was not requested")
	}
	if len(service.Items()) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(service.Items()))
	}
}

func TestDecodeItems_AllowsTopLevelArray(t *testing.T) {
	body := []byte(`[
		{
			"id": "1",
			"provider_slug": "bludv",
			"provider_url": "https://example.com/movie",
			"item_type": "movie",
			"magnet_uri": "magnet:?xt=urn:btih:1",
			"magnet_xt": "urn:btih:1",
			"torrent_name": "Movie A",
			"inserted_at": "2024-01-01T00:00:00Z"
		}
	]`)

	items, err := decodeItems(body)
	if err != nil {
		t.Fatalf("decodeItems() returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ID != "1" {
		t.Fatalf("items[0].ID = %q, want 1", items[0].ID)
	}
}

func TestDecodeItems_AllowsSingleItemObject(t *testing.T) {
	body := []byte(`{
		"id": "1",
		"provider_slug": "bludv",
		"provider_url": "https://example.com/movie",
		"item_type": "movie",
		"magnet_uri": "magnet:?xt=urn:btih:1",
		"magnet_xt": "urn:btih:1",
		"torrent_name": "Movie A",
		"inserted_at": "2024-01-01T00:00:00Z"
	}`)

	items, err := decodeItems(body)
	if err != nil {
		t.Fatalf("decodeItems() returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].TorrentName != "Movie A" {
		t.Fatalf("items[0].TorrentName = %q, want Movie A", items[0].TorrentName)
	}
}

func TestSearchService_QueryByKeyword(t *testing.T) {
	items := []Item{
		{ID: "1", ItemType: "movie", TorrentName: "Dune Part Two", ProviderSlug: "test", ProviderURL: "https://example.com/1", MagnetURI: "magnet:?xt=urn:btih:1", MagnetXT: "urn:btih:1", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
		{ID: "2", ItemType: "tv", TorrentName: "The Office", ProviderSlug: "test", ProviderURL: "https://example.com/2", MagnetURI: "magnet:?xt=urn:btih:2", MagnetXT: "urn:btih:2", Languages: []string{"en"}, InsertedAt: time.Now()},
	}

	service := Service{items: items}
	results, err := service.Search("Dune", SearchFilter{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(results.Items))
	}
	if got := results.Items[0].TorrentName; got != "Dune Part Two" {
		t.Fatalf("result title = %q, want Dune Part Two", got)
	}
}

func TestSearchService_QueryByItemType(t *testing.T) {
	items := []Item{
		{ID: "1", ItemType: "movie", TorrentName: "Movie A", ProviderSlug: "test", ProviderURL: "https://example.com/1", MagnetURI: "magnet:?xt=urn:btih:1", MagnetXT: "urn:btih:1", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
		{ID: "2", ItemType: "tv", TorrentName: "Series B", ProviderSlug: "test", ProviderURL: "https://example.com/2", MagnetURI: "magnet:?xt=urn:btih:2", MagnetXT: "urn:btih:2", Languages: []string{"en"}, InsertedAt: time.Now()},
	}

	service := Service{items: items}
	results, err := service.Search("", SearchFilter{ItemType: "tv", Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(results.Items))
	}
	if results.Items[0].ItemType != "tv" {
		t.Fatalf("item_type = %q, want tv", results.Items[0].ItemType)
	}
}

func TestSearchService_Pagination(t *testing.T) {
	items := []Item{
		{ID: "1", ItemType: "movie", TorrentName: "Movie 1", ProviderSlug: "test", ProviderURL: "https://example.com/1", MagnetURI: "magnet:?xt=urn:btih:1", MagnetXT: "urn:btih:1", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
		{ID: "2", ItemType: "movie", TorrentName: "Movie 2", ProviderSlug: "test", ProviderURL: "https://example.com/2", MagnetURI: "magnet:?xt=urn:btih:2", MagnetXT: "urn:btih:2", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
		{ID: "3", ItemType: "movie", TorrentName: "Movie 3", ProviderSlug: "test", ProviderURL: "https://example.com/3", MagnetURI: "magnet:?xt=urn:btih:3", MagnetXT: "urn:btih:3", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
	}

	service := Service{items: items}
	results, err := service.Search("Movie", SearchFilter{Page: 2, Size: 2})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(results.Items))
	}
	if got := results.Items[0].ID; got != "3" {
		t.Fatalf("result id = %q, want 3", got)
	}
}
