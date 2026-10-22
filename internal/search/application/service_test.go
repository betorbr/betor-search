package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewService_UsesDownloadItemsOverrideFromEnv(t *testing.T) {
	t.Setenv("BETOR_SEARCH_DOWNLOAD_ITEMS_URL", "https://example.com/custom-download/download-items/")

	service := NewService("", time.Minute)
	if got, want := service.downloadURL, "https://example.com/custom-download/download-items/"; got != want {
		t.Fatalf("downloadURL = %q, want %q", got, want)
	}
}

func TestService_Sync_UsesPublicDownloadItemsEndpointWithoutAuth(t *testing.T) {
	var metadataRequested bool
	var itemsRequested bool

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/download-items/":
			metadataRequested = true
			if got := r.Header.Get("Authorization"); got != "" {
				http.Error(w, "forbidden", http.StatusForbidden)
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

	service := NewService(server.URL+"/download-items/", time.Minute)
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

func TestDecodeItems_MapsSeasonAndEpisodeFieldsFromSourcePayload(t *testing.T) {
	body := []byte(`{
		"id": "tv-1",
		"provider_slug": "hdrtorrents",
		"provider_url": "https://example.com/tv",
		"item_type": "tv",
		"magnet_uri": "magnet:?xt=urn:btih:1",
		"magnet_xt": "urn:btih:1",
		"torrent_name": "Series S02",
		"episodes": [
			{"season": 2, "episode": 10},
			{"season": 2, "episode": 1}
		],
		"seasons": [2],
		"inserted_at": "2024-01-01T00:00:00Z"
	}`)

	items, err := decodeItems(body)
	if err != nil {
		t.Fatalf("decodeItems() returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if got, want := len(items[0].Episodes), 2; got != want {
		t.Fatalf("len(episodes) = %d, want %d", got, want)
	}
	if got, want := items[0].Episodes[0].Season, 2; got != want {
		t.Fatalf("episodes[0].Season = %d, want %d", got, want)
	}
	if got, want := items[0].Episodes[0].Episode, 10; got != want {
		t.Fatalf("episodes[0].Episode = %d, want %d", got, want)
	}
	if got, want := items[0].Seasons[0], 2; got != want {
		t.Fatalf("seasons[0] = %d, want %d", got, want)
	}
}

func TestParseHTTPQuery_ParsesSeasonAndEpisodeFilters(t *testing.T) {
	filter := Service{}.ParseHTTPQuery("season=2&episode=10,11")
	if got, want := len(filter.Season), 1; got != want {
		t.Fatalf("len(filter.Season) = %d, want %d", got, want)
	}
	if got, want := filter.Season[0], 2; got != want {
		t.Fatalf("filter.Season[0] = %d, want %d", got, want)
	}
	if got, want := len(filter.Episode), 2; got != want {
		t.Fatalf("len(filter.Episode) = %d, want %d", got, want)
	}
	if got, want := filter.Episode[0], 10; got != want {
		t.Fatalf("filter.Episode[0] = %d, want %d", got, want)
	}
	if got, want := filter.Episode[1], 11; got != want {
		t.Fatalf("filter.Episode[1] = %d, want %d", got, want)
	}
}

func TestService_RespectsConfiguredMinuteInterval(t *testing.T) {
	t.Setenv("BETOR_SEARCH_UPDATE_INTERVAL_MINUTES", "1")
	service := NewService("https://example.com", 0)
	if got, want := service.updateInterval, time.Minute; got != want {
		t.Fatalf("updateInterval = %v, want %v", got, want)
	}
}

func TestService_Sync_PreservesPriorCatalogAndMarksDegradedOnEmptyResponse(t *testing.T) {
	items := []Item{{
		ID:           "existing-item",
		ProviderSlug: "test",
		ProviderURL:  "https://example.com/1",
		ItemType:     "movie",
		MagnetURI:    "magnet:?xt=urn:btih:1",
		MagnetXT:     "urn:btih:1",
		TorrentName:  "Movie A",
		InsertedAt:   time.Now().UTC(),
	}}
	service := NewServiceWithItems(items, "https://example.com", time.Minute)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	service.downloadURL = server.URL
	if err := service.Sync(); err == nil {
		t.Fatal("Sync() expected error when upstream responds with an empty catalog")
	}
	if got, want := len(service.Items()), 1; got != want {
		t.Fatalf("len(items) = %d, want %d", got, want)
	}
	if got, want := service.Status(), "DEGRADED"; got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
}

func TestService_Sync_IgnoresStaleCatalogPayload(t *testing.T) {
	currentTime := time.Now().UTC()
	service := NewServiceWithItems([]Item{{
		ID:           "newer",
		ProviderSlug: "test",
		ProviderURL:  "https://example.com/newer",
		ItemType:     "movie",
		MagnetURI:    "magnet:?xt=urn:btih:newer",
		MagnetXT:     "urn:btih:newer",
		TorrentName:  "Newest Movie",
		InsertedAt:   currentTime.Add(-1 * time.Hour),
		UpdatedAt:    currentTime,
	}}, "https://example.com", time.Minute)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"older","provider_slug":"test","provider_url":"https://example.com/older","item_type":"movie","magnet_uri":"magnet:?xt=urn:btih:older","magnet_xt":"urn:btih:older","torrent_name":"Older Movie","inserted_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}]}`))
	}))
	defer server.Close()

	service.downloadURL = server.URL
	if err := service.Sync(); err == nil {
		t.Fatal("Sync() expected stale payload to be rejected")
	}
	if got, want := len(service.Items()), 1; got != want {
		t.Fatalf("len(items) = %d, want %d", got, want)
	}
	if got := service.Items()[0].ID; got != "newer" {
		t.Fatalf("catalog id = %q, want newer", got)
	}
}

func TestSearchService_SortsByInsertedAtByDefault(t *testing.T) {
	base := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	items := []Item{
		{ID: "older-inserted", ItemType: "movie", TorrentName: "Older Inserted", ProviderSlug: "test", ProviderURL: "https://example.com/older-inserted", MagnetURI: "magnet:?xt=urn:btih:older-inserted", MagnetXT: "urn:btih:older-inserted", Languages: []string{"pt-BR"}, InsertedAt: base.Add(1 * time.Hour), UpdatedAt: base.Add(5 * time.Hour)},
		{ID: "newer-inserted", ItemType: "movie", TorrentName: "Newer Inserted", ProviderSlug: "test", ProviderURL: "https://example.com/newer-inserted", MagnetURI: "magnet:?xt=urn:btih:newer-inserted", MagnetXT: "urn:btih:newer-inserted", Languages: []string{"pt-BR"}, InsertedAt: base.Add(3 * time.Hour), UpdatedAt: base.Add(1 * time.Hour)},
	}

	service := Service{items: items}
	results, err := service.Search("", SearchFilter{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(results.Items))
	}
	if got := results.Items[0].ID; got != "newer-inserted" {
		t.Fatalf("first result id = %q, want newer-inserted", got)
	}
	if got := results.Items[1].ID; got != "older-inserted" {
		t.Fatalf("second result id = %q, want older-inserted", got)
	}
}

func TestService_StartBackgroundSync_RefreshesOnInterval(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"` + string(rune('A'+calls-1)) + `","provider_slug":"test","provider_url":"https://example.com/1","item_type":"movie","magnet_uri":"magnet:?xt=urn:btih:1","magnet_xt":"urn:btih:1","torrent_name":"Movie ` + string(rune('A'+calls-1)) + `","inserted_at":"2024-01-01T00:00:00Z"}]}`))
	}))
	defer server.Close()

	service := NewService(server.URL, 10*time.Millisecond)
	if err := service.Sync(); err != nil {
		t.Fatalf("initial Sync() returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls after initial sync = %d, want 1", calls)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.StartBackgroundSync(ctx)

	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if calls >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 sync calls within 200ms, got %d", calls)
	}
	if !service.HasHealthyData() {
		t.Fatal("catalog was not populated after background syncs")
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
	base := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	items := []Item{
		{ID: "1", ItemType: "movie", TorrentName: "Movie 1", ProviderSlug: "test", ProviderURL: "https://example.com/1", MagnetURI: "magnet:?xt=urn:btih:1", MagnetXT: "urn:btih:1", Languages: []string{"pt-BR"}, InsertedAt: base},
		{ID: "2", ItemType: "movie", TorrentName: "Movie 2", ProviderSlug: "test", ProviderURL: "https://example.com/2", MagnetURI: "magnet:?xt=urn:btih:2", MagnetXT: "urn:btih:2", Languages: []string{"pt-BR"}, InsertedAt: base.Add(1 * time.Hour)},
		{ID: "3", ItemType: "movie", TorrentName: "Movie 3", ProviderSlug: "test", ProviderURL: "https://example.com/3", MagnetURI: "magnet:?xt=urn:btih:3", MagnetXT: "urn:btih:3", Languages: []string{"pt-BR"}, InsertedAt: base.Add(2 * time.Hour)},
	}

	service := Service{items: items}
	results, err := service.Search("Movie", SearchFilter{Page: 2, Size: 2})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(results.Items))
	}
	if got := results.Items[0].ID; got != "1" {
		t.Fatalf("result id = %q, want 1", got)
	}
}

func TestSearchService_SortsNewestFirst(t *testing.T) {
	base := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	items := []Item{
		{ID: "older", ItemType: "movie", TorrentName: "Older Movie", ProviderSlug: "test", ProviderURL: "https://example.com/older", MagnetURI: "magnet:?xt=urn:btih:older", MagnetXT: "urn:btih:older", Languages: []string{"pt-BR"}, InsertedAt: base.Add(-2 * time.Hour)},
		{ID: "newest", ItemType: "movie", TorrentName: "Newest Movie", ProviderSlug: "test", ProviderURL: "https://example.com/newest", MagnetURI: "magnet:?xt=urn:btih:newest", MagnetXT: "urn:btih:newest", Languages: []string{"pt-BR"}, InsertedAt: base.Add(2 * time.Hour)},
		{ID: "mid", ItemType: "movie", TorrentName: "Mid Movie", ProviderSlug: "test", ProviderURL: "https://example.com/mid", MagnetURI: "magnet:?xt=urn:btih:mid", MagnetXT: "urn:btih:mid", Languages: []string{"pt-BR"}, InsertedAt: base},
	}

	service := Service{items: items}
	results, err := service.Search("Movie", SearchFilter{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results.Items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(results.Items))
	}
	if got := results.Items[0].ID; got != "newest" {
		t.Fatalf("first result id = %q, want newest", got)
	}
	if got := results.Items[1].ID; got != "mid" {
		t.Fatalf("second result id = %q, want mid", got)
	}
	if got := results.Items[2].ID; got != "older" {
		t.Fatalf("third result id = %q, want older", got)
	}
}
