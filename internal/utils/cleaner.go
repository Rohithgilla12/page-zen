package utils

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"page-zen/internal/logger"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"go.uber.org/zap"
)

// OpenGraphData represents Open Graph metadata
type OpenGraphData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	URL         string `json:"url,omitempty"`
	Type        string `json:"type,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	Locale      string `json:"locale,omitempty"`
	// Twitter Card data
	TwitterCard        string `json:"twitter_card,omitempty"`
	TwitterSite        string `json:"twitter_site,omitempty"`
	TwitterCreator     string `json:"twitter_creator,omitempty"`
	TwitterTitle       string `json:"twitter_title,omitempty"`
	TwitterDescription string `json:"twitter_description,omitempty"`
	TwitterImage       string `json:"twitter_image,omitempty"`
	// Additional metadata
	Author      string   `json:"author,omitempty"`
	PublishedAt string   `json:"published_at,omitempty"`
	ModifiedAt  string   `json:"modified_at,omitempty"`
	Section     string   `json:"section,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CleanedArticle represents a cleaned article with both text and markdown content
type CleanedArticle struct {
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	Markdown    string         `json:"markdown"`
	URL         string         `json:"url"`
	Author      string         `json:"author,omitempty"`
	Excerpt     string         `json:"excerpt,omitempty"`
	Length      int            `json:"length"`
	PublishedAt string         `json:"published_at,omitempty"`
	OpenGraph   *OpenGraphData `json:"open_graph,omitempty"`
}

// ArticleCleaner handles the cleaning and processing of web articles
type ArticleCleaner struct {
	logger *zap.SugaredLogger
}

// NewArticleCleaner creates a new ArticleCleaner instance
func NewArticleCleaner() (*ArticleCleaner, error) {
	log, err := logger.NewSugaredLogger()
	if err != nil {
		return nil, err
	}
	return &ArticleCleaner{logger: log}, nil
}

// Close properly closes the logger
func (ac *ArticleCleaner) Close() {
	if ac.logger != nil {
		ac.logger.Sync()
	}
}

// Constants for content cleaning
var (
	unwantedSelectors = []string{
		// Scripts and styles
		"script", "style", "noscript",
		// Navigation and UI elements
		"nav", "header", "footer", ".navigation", ".nav", ".menu",
		// Advertisements and social media
		".ad", ".ads", ".advertisement", ".social", ".share", ".sharing",
		".social-share", ".social-media", ".twitter", ".facebook", ".instagram",
		// Comments and related content
		".comments", ".comment", ".related", ".recommended", ".suggestions",
		// Tracking and analytics
		".analytics", ".tracking", ".pixel",
		// Cookie notices and popups
		".cookie", ".gdpr", ".popup", ".modal", ".overlay",
		// Subscription and newsletter boxes
		".subscribe", ".newsletter", ".signup", ".email-signup",
		// Breadcrumbs and metadata that's not useful
		".breadcrumb", ".breadcrumbs", ".tags", ".categories",
		// Video players that might not work
		".video-player", ".embed", "iframe[src*='youtube']", "iframe[src*='vimeo']",
		// Common class names for unwanted content
		"[class*='sidebar']", "[class*='widget']", "[id*='sidebar']", "[id*='widget']",
		// Forms that are usually subscription/contact forms
		"form:not(.search-form)",
	}

	unwantedTextPatterns = []string{
		`(?i)subscribe\s+to\s+our\s+newsletter`,
		`(?i)follow\s+us\s+on`,
		`(?i)share\s+this\s+article`,
		`(?i)related\s+articles?`,
		`(?i)you\s+might\s+also\s+like`,
		`(?i)recommended\s+for\s+you`,
		`(?i)advertisement`,
		`(?i)sponsored\s+content`,
	}
)

// removeUnwantedElements removes common unwanted elements from the document
func (ac *ArticleCleaner) removeUnwantedElements(doc *goquery.Document) {
	removedCount := 0
	for _, selector := range unwantedSelectors {
		elements := doc.Find(selector)
		count := elements.Length()
		if count > 0 {
			elements.Remove()
			removedCount += count
			ac.logger.Debugw("Removed unwanted elements", "selector", selector, "count", count)
		}
	}

	if removedCount > 0 {
		ac.logger.Infow("Total unwanted elements removed", "count", removedCount)
	}
}

// cleanTextContent cleans up text content by removing extra whitespace and unwanted characters
func (ac *ArticleCleaner) cleanTextContent(content string) string {
	// Remove multiple consecutive newlines
	re := regexp.MustCompile(`\n\s*\n\s*\n+`)
	content = re.ReplaceAllString(content, "\n\n")

	// Remove excessive whitespace
	re = regexp.MustCompile(`[ \t]+`)
	content = re.ReplaceAllString(content, " ")

	// Clean up common unwanted patterns
	for _, pattern := range unwantedTextPatterns {
		re = regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, "")
	}

	return strings.TrimSpace(content)
}

// generateExcerpt creates a brief excerpt from the content
func (ac *ArticleCleaner) generateExcerpt(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// Find the last space before maxLength to avoid cutting words
	excerpt := content[:maxLength]
	if lastSpace := strings.LastIndex(excerpt, " "); lastSpace > 0 && lastSpace > maxLength-50 {
		excerpt = excerpt[:lastSpace]
	}

	return excerpt + "..."
}

// resolveURL converts relative URLs to absolute URLs
func (ac *ArticleCleaner) resolveURL(rawURL string, baseURL *url.URL) string {
	if rawURL == "" {
		return ""
	}

	// Already absolute URL
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}

	// Protocol-relative URL
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}

	// Absolute path
	if strings.HasPrefix(rawURL, "/") {
		return baseURL.Scheme + "://" + baseURL.Host + rawURL
	}

	// Relative path
	return baseURL.Scheme + "://" + baseURL.Host + "/" + rawURL
}

// processImages handles both picture elements and standalone img elements
func (ac *ArticleCleaner) processImages(doc *goquery.Document, baseURL *url.URL) {
	ac.processPictureElements(doc, baseURL)
	ac.processImgElements(doc, baseURL)
}

// processPictureElements handles picture elements and converts them to img elements
func (ac *ArticleCleaner) processPictureElements(doc *goquery.Document, baseURL *url.URL) {
	pictureCount := 0
	doc.Find("picture").Each(func(i int, picture *goquery.Selection) {
		pictureCount++
		img := picture.Find("img")
		if img.Length() == 0 {
			return
		}

		// Find the highest quality source
		if highestQualitySource := ac.findHighestQualitySource(picture); highestQualitySource != "" {
			// Replace webp with png for better compatibility
			highestQualitySource = strings.Replace(highestQualitySource, "webp", "png", 1)
			resolvedURL := ac.resolveURL(highestQualitySource, baseURL)
			img.SetAttr("src", resolvedURL)
			ac.logger.Debugw("Updated picture element source", "new_src", resolvedURL)
		}

		// Clean up attributes
		img.RemoveAttr("loading")
		img.RemoveAttr("decoding")

		// Replace picture with the img element
		picture.ReplaceWithSelection(img)
	})

	if pictureCount > 0 {
		ac.logger.Debugw("Processed picture elements", "count", pictureCount)
	}
}

// findHighestQualitySource finds the highest quality source from picture element
func (ac *ArticleCleaner) findHighestQualitySource(picture *goquery.Selection) string {
	highestQualitySource := ""
	maxWidth := 0

	picture.Find("source").Each(func(i int, source *goquery.Selection) {
		srcset, exists := source.Attr("srcset")
		if !exists {
			return
		}

		sources := strings.Split(srcset, ",")
		for _, src := range sources {
			parts := strings.Fields(strings.TrimSpace(src))
			if len(parts) != 2 {
				continue
			}

			// Parse the width value (remove 'w' suffix)
			widthStr := strings.TrimSuffix(parts[1], "w")
			width, err := strconv.Atoi(widthStr)
			if err != nil {
				ac.logger.Debugw("Failed to parse width from srcset", "width_string", widthStr, "error", err)
				continue
			}

			// Keep track of highest quality source
			if width > maxWidth {
				maxWidth = width
				highestQualitySource = parts[0]
			}
		}
	})

	return highestQualitySource
}

// processImgElements handles standalone img elements
func (ac *ArticleCleaner) processImgElements(doc *goquery.Document, baseURL *url.URL) {
	imgCount := 0
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		imgCount++

		// Remove attributes that might cause issues
		img.RemoveAttr("loading")
		img.RemoveAttr("decoding")
		img.RemoveAttr("srcset")

		// Handle relative URLs for src attribute
		if src, exists := img.Attr("src"); exists {
			resolvedURL := ac.resolveURL(src, baseURL)
			if resolvedURL != src {
				img.SetAttr("src", resolvedURL)
				ac.logger.Debugw("Updated img src", "original", src, "new", resolvedURL)
			}
		}
	})

	if imgCount > 0 {
		ac.logger.Debugw("Processed img elements", "count", imgCount)
	}
}

// extractOpenGraphData extracts Open Graph and Twitter Card metadata from HTML document
func (ac *ArticleCleaner) extractOpenGraphData(doc *goquery.Document, pageURL string, baseURL *url.URL) *OpenGraphData {
	og := &OpenGraphData{URL: pageURL}

	// Extract Open Graph meta tags
	ac.extractOGMetaTags(doc, og)

	// Extract Twitter Card meta tags
	ac.extractTwitterMetaTags(doc, og)

	// Fallback to standard meta tags if Open Graph is not available
	ac.extractFallbackMetaTags(doc, og)

	// Resolve relative URLs
	og.Image = ac.resolveURL(og.Image, baseURL)
	og.TwitterImage = ac.resolveURL(og.TwitterImage, baseURL)

	ac.logger.Infow("Extracted Open Graph data",
		"title", og.Title,
		"description_length", len(og.Description),
		"image", og.Image,
		"type", og.Type,
		"site_name", og.SiteName,
	)

	return og
}

// extractOGMetaTags extracts Open Graph meta tags
func (ac *ArticleCleaner) extractOGMetaTags(doc *goquery.Document, og *OpenGraphData) {
	doc.Find("meta[property^='og:']").Each(func(i int, s *goquery.Selection) {
		property, exists := s.Attr("property")
		if !exists {
			return
		}

		content, exists := s.Attr("content")
		if !exists || content == "" {
			return
		}

		switch property {
		case "og:title":
			og.Title = content
		case "og:description":
			og.Description = content
		case "og:image":
			og.Image = content
		case "og:url":
			og.URL = content
		case "og:type":
			og.Type = content
		case "og:site_name":
			og.SiteName = content
		case "og:locale":
			og.Locale = content
		case "article:author":
			og.Author = content
		case "article:published_time":
			og.PublishedAt = content
		case "article:modified_time":
			og.ModifiedAt = content
		case "article:section":
			og.Section = content
		case "article:tag":
			og.Tags = append(og.Tags, content)
		}
	})
}

// extractTwitterMetaTags extracts Twitter Card meta tags
func (ac *ArticleCleaner) extractTwitterMetaTags(doc *goquery.Document, og *OpenGraphData) {
	doc.Find("meta[name^='twitter:']").Each(func(i int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if !exists {
			return
		}

		content, exists := s.Attr("content")
		if !exists || content == "" {
			return
		}

		switch name {
		case "twitter:card":
			og.TwitterCard = content
		case "twitter:site":
			og.TwitterSite = content
		case "twitter:creator":
			og.TwitterCreator = content
		case "twitter:title":
			og.TwitterTitle = content
		case "twitter:description":
			og.TwitterDescription = content
		case "twitter:image":
			og.TwitterImage = content
		}
	})
}

// extractFallbackMetaTags extracts fallback meta tags when Open Graph is not available
func (ac *ArticleCleaner) extractFallbackMetaTags(doc *goquery.Document, og *OpenGraphData) {
	if og.Title == "" {
		if title := doc.Find("title").First().Text(); title != "" {
			og.Title = strings.TrimSpace(title)
		}
	}

	if og.Description == "" {
		if desc, exists := doc.Find("meta[name='description']").Attr("content"); exists && desc != "" {
			og.Description = strings.TrimSpace(desc)
		}
	}

	if og.Author == "" {
		if author, exists := doc.Find("meta[name='author']").Attr("content"); exists && author != "" {
			og.Author = strings.TrimSpace(author)
		}
	}
}

// fetchAndParseDocument fetches a URL and returns a parsed goquery document
func (ac *ArticleCleaner) fetchAndParseDocument(pageURL string) (*goquery.Document, *url.URL, error) {
	ac.logger.Infow("Starting to fetch and parse article", "url", pageURL)

	resp, err := http.Get(pageURL)
	if err != nil {
		ac.logger.Errorw("Failed to fetch URL", "url", pageURL, "error", err)
		return nil, nil, err
	}
	defer resp.Body.Close()

	ac.logger.Infow("Successfully fetched URL", "url", pageURL, "status_code", resp.StatusCode)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		ac.logger.Errorw("Failed to parse HTML document", "url", pageURL, "error", err)
		return nil, nil, err
	}

	return doc, resp.Request.URL, nil
}

// convertToMarkdown converts HTML content to markdown
func (ac *ArticleCleaner) convertToMarkdown(htmlContent string) string {
	ac.logger.Info("Converting HTML content to markdown")
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		ac.logger.Warnw("Failed to convert to markdown, using HTML content", "error", err)
		return htmlContent // Fallback to HTML if markdown conversion fails
	}
	return markdown
}

// saveDebugHTML saves the cleaned HTML to a file for debugging
func (ac *ArticleCleaner) saveDebugHTML(doc *goquery.Document) {
	html, err := doc.Html()
	if err != nil {
		ac.logger.Errorw("Failed to get HTML", "error", err)
		return
	}

	if err := os.WriteFile("tmp/article.html", []byte(html), 0644); err != nil {
		ac.logger.Errorw("Failed to save debug HTML", "error", err)
		return
	}

	ac.logger.Debugw("Saved cleaned HTML to file", "file", "tmp/article.html")
}

// CleanArticle processes a URL and returns a comprehensive cleaned article
func (ac *ArticleCleaner) CleanArticle(pageURL string) (CleanedArticle, error) {
	// Fetch and parse the document
	doc, baseURL, err := ac.fetchAndParseDocument(pageURL)
	if err != nil {
		return CleanedArticle{}, err
	}

	// Extract Open Graph data before removing elements
	openGraphData := ac.extractOpenGraphData(doc, pageURL, baseURL)

	// Remove unwanted elements
	ac.removeUnwantedElements(doc)

	// Process images
	ac.processImages(doc, baseURL)

	// Convert to readability format
	ac.logger.Info("Converting document to readability format")
	article, err := readability.FromDocument(doc.Get(0), nil)
	if err != nil {
		ac.logger.Errorw("Failed to convert document to readability format", "error", err)
		return CleanedArticle{}, err
	}

	// Clean the text content
	cleanedTextContent := ac.cleanTextContent(article.TextContent)

	// Convert to markdown
	markdown := ac.convertToMarkdown(article.Content)

	// Generate excerpt
	excerpt := ac.generateExcerpt(cleanedTextContent, 200)

	// Handle published time safely
	var publishedAt string
	if article.PublishedTime != nil {
		publishedAt = article.PublishedTime.Format(time.RFC3339)
	}

	cleanedArticle := CleanedArticle{
		Title:       strings.TrimSpace(article.Title),
		Content:     cleanedTextContent,
		Markdown:    markdown,
		URL:         pageURL,
		Author:      strings.TrimSpace(article.Byline),
		Excerpt:     excerpt,
		Length:      len(cleanedTextContent),
		PublishedAt: publishedAt,
		OpenGraph:   openGraphData,
	}

	ac.logger.Infow("Successfully processed article",
		"title_length", len(cleanedArticle.Title),
		"content_length", cleanedArticle.Length,
		"markdown_length", len(cleanedArticle.Markdown),
		"url", pageURL,
	)

	// Save debug HTML
	ac.saveDebugHTML(doc)

	return cleanedArticle, nil
}

// ExtractOpenGraphData extracts only Open Graph metadata from a URL
func (ac *ArticleCleaner) ExtractOpenGraphData(pageURL string) (*OpenGraphData, error) {
	ac.logger.Infow("Starting to fetch Open Graph data", "url", pageURL)

	doc, baseURL, err := ac.fetchAndParseDocument(pageURL)
	if err != nil {
		return &OpenGraphData{}, err
	}

	openGraphData := ac.extractOpenGraphData(doc, pageURL, baseURL)
	ac.logger.Infow("Successfully extracted Open Graph data", "url", pageURL, "title", openGraphData.Title)

	return openGraphData, nil
}

// Public API functions for backward compatibility

// GetReadableArticle returns just the text content (for backward compatibility)
func GetReadableArticle(url string) string {
	article := GetCleanedArticle(url)
	return article.Content
}

// GetCleanedArticle returns a comprehensive cleaned article with markdown
func GetCleanedArticle(url string) CleanedArticle {
	cleaner, err := NewArticleCleaner()
	if err != nil {
		return CleanedArticle{}
	}
	defer cleaner.Close()

	article, err := cleaner.CleanArticle(url)
	if err != nil {
		cleaner.logger.Errorw("Failed to get readable article", "url", url, "error", err)
		return CleanedArticle{}
	}

	return article
}

// GetOpenGraphData extracts only Open Graph metadata from a URL
func GetOpenGraphData(url string) *OpenGraphData {
	cleaner, err := NewArticleCleaner()
	if err != nil {
		return &OpenGraphData{}
	}
	defer cleaner.Close()

	data, err := cleaner.ExtractOpenGraphData(url)
	if err != nil {
		cleaner.logger.Errorw("Failed to extract Open Graph data", "url", url, "error", err)
		return &OpenGraphData{}
	}

	return data
}
