package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAndParseDocumentSetsUserAgent(t *testing.T) {
	var receivedUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test</title></head><body><p>Hello</p></body></html>`))
	}))
	defer ts.Close()

	ac, err := NewArticleCleaner()
	if err != nil {
		t.Fatalf("Failed to create ArticleCleaner: %v", err)
	}
	defer ac.Close()

	doc, baseURL, err := ac.fetchAndParseDocument(ts.URL)
	if err != nil {
		t.Fatalf("fetchAndParseDocument failed: %v", err)
	}

	if receivedUA == "" {
		t.Fatal("User-Agent header was not set")
	}
	if receivedUA == "Go-http-client/1.1" || receivedUA == "Go-http-client/2.0" {
		t.Errorf("User-Agent is Go's default (%q), should be custom", receivedUA)
	}
	if doc == nil {
		t.Fatal("Returned document is nil")
	}
	if baseURL == nil {
		t.Fatal("Returned baseURL is nil")
	}

	title := doc.Find("title").Text()
	if title != "Test" {
		t.Errorf("Expected title 'Test', got %q", title)
	}
}

func TestFetchAndParseDocumentFollowsRedirectWithUserAgent(t *testing.T) {
	var finalUA string

	// Destination server
	dest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Redirected</title></head><body><p>Content</p></body></html>`))
	}))
	defer dest.Close()

	// Origin server that issues a 307 redirect (mimics Medium's behavior)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, dest.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	ac, err := NewArticleCleaner()
	if err != nil {
		t.Fatalf("Failed to create ArticleCleaner: %v", err)
	}
	defer ac.Close()

	doc, _, err := ac.fetchAndParseDocument(origin.URL)
	if err != nil {
		t.Fatalf("fetchAndParseDocument failed: %v", err)
	}

	if finalUA == "" || finalUA == "Go-http-client/1.1" || finalUA == "Go-http-client/2.0" {
		t.Errorf("User-Agent not preserved through redirect, got %q", finalUA)
	}

	title := doc.Find("title").Text()
	if title != "Redirected" {
		t.Errorf("Expected title 'Redirected', got %q", title)
	}
}

func TestCleanArticleWithMediumHTML(t *testing.T) {
	// Simulate a Medium-style page with OG tags and article content
	mediumHTML := `<html>
<head>
	<title>Test Article - Netflix Tech Blog</title>
	<meta property="og:title" content="Test Article" />
	<meta property="og:description" content="A test article description" />
	<meta property="og:image" content="/img/hero.png" />
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content="Netflix TechBlog" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="author" content="Test Author" />
</head>
<body>
	<nav><a href="/">Home</a></nav>
	<article>
		<h1>Test Article</h1>
		<p>This is a substantial test article with enough content to be extracted by readability.
		It needs multiple paragraphs to pass the content length threshold that readability uses
		to determine if something is actual article content or just noise.</p>
		<p>Here is a second paragraph with more meaningful content about distributed systems
		and how they handle failure modes in production environments.</p>
		<p>And a third paragraph discussing the architecture decisions that were made during
		the design phase of this particular system component.</p>
	</article>
	<footer><p>Footer content</p></footer>
	<div class="newsletter"><p>Subscribe to our newsletter</p></div>
</body>
</html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(mediumHTML))
	}))
	defer ts.Close()

	ac, err := NewArticleCleaner()
	if err != nil {
		t.Fatalf("Failed to create ArticleCleaner: %v", err)
	}
	defer ac.Close()

	article, err := ac.CleanArticle(ts.URL)
	if err != nil {
		t.Fatalf("CleanArticle failed: %v", err)
	}

	if article.Title == "" {
		t.Error("Expected non-empty title")
	}
	if article.Content == "" {
		t.Error("Expected non-empty content")
	}
	if article.Length == 0 {
		t.Error("Expected non-zero content length")
	}
	if article.OpenGraph == nil {
		t.Fatal("Expected OpenGraph data to be present")
	}
	if article.OpenGraph.Title != "Test Article" {
		t.Errorf("Expected OG title 'Test Article', got %q", article.OpenGraph.Title)
	}
	if article.OpenGraph.Description != "A test article description" {
		t.Errorf("Expected OG description 'A test article description', got %q", article.OpenGraph.Description)
	}
	if article.OpenGraph.SiteName != "Netflix TechBlog" {
		t.Errorf("Expected OG site_name 'Netflix TechBlog', got %q", article.OpenGraph.SiteName)
	}
}

func TestExtractOpenGraphData(t *testing.T) {
	html := `<html>
<head>
	<meta property="og:title" content="OG Title" />
	<meta property="og:description" content="OG Description" />
	<meta property="og:image" content="https://example.com/image.png" />
	<meta property="og:url" content="https://example.com/article" />
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content="Example Blog" />
	<meta property="article:author" content="Jane Doe" />
	<meta property="article:published_time" content="2025-01-15T10:00:00Z" />
	<meta property="article:tag" content="go" />
	<meta property="article:tag" content="testing" />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:site" content="@example" />
	<meta name="twitter:creator" content="@janedoe" />
</head>
<body><p>Content</p></body>
</html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer ts.Close()

	ac, err := NewArticleCleaner()
	if err != nil {
		t.Fatalf("Failed to create ArticleCleaner: %v", err)
	}
	defer ac.Close()

	og, err := ac.ExtractOpenGraphData(ts.URL)
	if err != nil {
		t.Fatalf("ExtractOpenGraphData failed: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Title", og.Title, "OG Title"},
		{"Description", og.Description, "OG Description"},
		{"Image", og.Image, "https://example.com/image.png"},
		{"Type", og.Type, "article"},
		{"SiteName", og.SiteName, "Example Blog"},
		{"Author", og.Author, "Jane Doe"},
		{"PublishedAt", og.PublishedAt, "2025-01-15T10:00:00Z"},
		{"TwitterCard", og.TwitterCard, "summary_large_image"},
		{"TwitterSite", og.TwitterSite, "@example"},
		{"TwitterCreator", og.TwitterCreator, "@janedoe"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	if len(og.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(og.Tags))
	}
	if og.Tags[0] != "go" || og.Tags[1] != "testing" {
		t.Errorf("Expected tags [go, testing], got %v", og.Tags)
	}
}

// Integration tests that hit real Medium/Netflix URLs.
// Skipped by default — run with: go test -run TestIntegration -tags integration ./...
// or simply: go test -run TestIntegration -count=1 ./internal/utils/

func TestIntegrationNetflixTechBlog(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []struct {
		name string
		url  string
	}{
		{"RSS tracked URL", "https://netflixtechblog.com/netflix-live-origin-41f1b0ad5371?source=rss----2615bd06b42e---4"},
		{"Canonical URL", "https://netflixtechblog.com/netflix-live-origin-41f1b0ad5371"},
		{"AV1 article", "https://netflixtechblog.com/av1-now-powering-30-of-netflix-streaming-02f592242d80"},
		{"Temporal article", "https://netflixtechblog.com/how-temporal-powers-reliable-cloud-operations-at-netflix-73c69ccb5953"},
	}

	ac, err := NewArticleCleaner()
	if err != nil {
		t.Fatalf("Failed to create ArticleCleaner: %v", err)
	}
	defer ac.Close()

	for _, tc := range urls {
		t.Run(tc.name, func(t *testing.T) {
			article, err := ac.CleanArticle(tc.url)
			if err != nil {
				t.Fatalf("CleanArticle failed for %s: %v", tc.url, err)
			}

			if article.Title == "" {
				t.Errorf("Expected non-empty title for %s", tc.url)
			}
			if article.Content == "" {
				t.Errorf("Expected non-empty content for %s", tc.url)
			}
			if article.Length < 100 {
				t.Errorf("Content too short (%d chars) for %s — likely extraction failed", article.Length, tc.url)
			}
			if article.OpenGraph == nil || article.OpenGraph.Title == "" {
				t.Errorf("Expected OpenGraph title for %s", tc.url)
			}

			t.Logf("OK: %q (%d chars)", article.Title, article.Length)
		})
	}
}
