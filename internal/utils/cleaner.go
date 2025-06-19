package utils

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

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

// removeUnwantedElements removes common unwanted elements from the document
func removeUnwantedElements(doc *goquery.Document, log *zap.SugaredLogger) {
	unwantedSelectors := []string{
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

	removedCount := 0
	for _, selector := range unwantedSelectors {
		elements := doc.Find(selector)
		count := elements.Length()
		if count > 0 {
			elements.Remove()
			removedCount += count
			log.Debugw("Removed unwanted elements", "selector", selector, "count", count)
		}
	}

	if removedCount > 0 {
		log.Infow("Total unwanted elements removed", "count", removedCount)
	}
}

// cleanTextContent cleans up text content by removing extra whitespace and unwanted characters
func cleanTextContent(content string) string {
	// Remove multiple consecutive newlines
	re := regexp.MustCompile(`\n\s*\n\s*\n+`)
	content = re.ReplaceAllString(content, "\n\n")

	// Remove excessive whitespace
	re = regexp.MustCompile(`[ \t]+`)
	content = re.ReplaceAllString(content, " ")

	// Clean up common unwanted patterns
	unwantedPatterns := []string{
		`(?i)subscribe\s+to\s+our\s+newsletter`,
		`(?i)follow\s+us\s+on`,
		`(?i)share\s+this\s+article`,
		`(?i)related\s+articles?`,
		`(?i)you\s+might\s+also\s+like`,
		`(?i)recommended\s+for\s+you`,
		`(?i)advertisement`,
		`(?i)sponsored\s+content`,
	}

	for _, pattern := range unwantedPatterns {
		re = regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, "")
	}

	// Trim whitespace
	content = strings.TrimSpace(content)

	return content
}

// generateExcerpt creates a brief excerpt from the content
func generateExcerpt(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// Find the last space before maxLength to avoid cutting words
	excerpt := content[:maxLength]
	lastSpace := strings.LastIndex(excerpt, " ")
	if lastSpace > 0 && lastSpace > maxLength-50 { // Don't make it too short
		excerpt = excerpt[:lastSpace]
	}

	return excerpt + "..."
}

// extractOpenGraphData extracts Open Graph and Twitter Card metadata from HTML document
func extractOpenGraphData(doc *goquery.Document, url string, baseURL *http.Request, log *zap.SugaredLogger) *OpenGraphData {
	og := &OpenGraphData{}

	// Extract Open Graph meta tags
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

	// Extract Twitter Card meta tags
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

	// Fallback to standard meta tags if Open Graph is not available
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

	// Set the URL if not provided
	if og.URL == "" {
		og.URL = url
	}

	// Convert relative URLs to absolute URLs
	if og.Image != "" && !strings.HasPrefix(og.Image, "http") {
		if strings.HasPrefix(og.Image, "//") {
			og.Image = "https:" + og.Image
		} else if strings.HasPrefix(og.Image, "/") {
			og.Image = baseURL.URL.Scheme + "://" + baseURL.URL.Host + og.Image
		}
	}

	if og.TwitterImage != "" && !strings.HasPrefix(og.TwitterImage, "http") {
		if strings.HasPrefix(og.TwitterImage, "//") {
			og.TwitterImage = "https:" + og.TwitterImage
		} else if strings.HasPrefix(og.TwitterImage, "/") {
			og.TwitterImage = baseURL.URL.Scheme + "://" + baseURL.URL.Host + og.TwitterImage
		}
	}

	log.Infow("Extracted Open Graph data",
		"title", og.Title,
		"description_length", len(og.Description),
		"image", og.Image,
		"type", og.Type,
		"site_name", og.SiteName,
	)

	return og
}

func mutateFetchReadable(url string, log *zap.SugaredLogger) (CleanedArticle, error) {
	log.Infow("Starting to fetch and parse article", "url", url)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Errorw("Failed to fetch URL", "url", url, "error", err)
		return CleanedArticle{}, err
	}
	defer resp.Body.Close()

	log.Infow("Successfully fetched URL", "url", url, "status_code", resp.StatusCode)

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Errorw("Failed to parse HTML document", "url", url, "error", err)
		return CleanedArticle{}, err
	}

	// Set the base URL for resolving relative URLs
	baseURL := resp.Request.URL
	log.Debugw("Using base URL for relative links", "base_url", baseURL.String())

	// Extract Open Graph data before removing elements
	openGraphData := extractOpenGraphData(doc, url, resp.Request, log)

	// Remove unwanted elements first
	removeUnwantedElements(doc, log)

	// Remove script and style tags (additional cleanup)
	scriptCount := doc.Find("script").Length()
	doc.Find("script").Remove()
	styleCount := doc.Find("style").Length()
	doc.Find("style").Remove()
	log.Debugw("Removed script and style tags", "scripts", scriptCount, "styles", styleCount)

	// Handle picture elements
	pictureCount := 0
	doc.Find("picture").Each(func(i int, picture *goquery.Selection) {
		pictureCount++
		// Find the img element within picture
		img := picture.Find("img")
		if img.Length() == 0 {
			return
		}

		// Try to get the highest quality source
		highestQualitySource := ""
		maxWidth := 0

		picture.Find("source").Each(func(i int, source *goquery.Selection) {
			if srcset, exists := source.Attr("srcset"); exists {
				sources := strings.Split(srcset, ",")
				for _, src := range sources {
					parts := strings.Fields(strings.TrimSpace(src))
					if len(parts) != 2 {
						continue
					}

					// Parse the width value (remove 'w' suffix)
					width := strings.TrimSuffix(parts[1], "w")
					w, err := strconv.Atoi(width)
					if err != nil {
						log.Debugw("Failed to parse width from srcset", "width_string", width, "error", err)
						continue
					}

					// Keep track of highest quality source
					if w > maxWidth {
						maxWidth = w
						highestQualitySource = parts[0]
					}
				}
			}
		})

		if highestQualitySource != "" {
			// replace webp with png
			highestQualitySource = strings.Replace(highestQualitySource, "webp", "png", 1)
			// Handle relative URLs
			if !strings.HasPrefix(highestQualitySource, "http://") && !strings.HasPrefix(highestQualitySource, "https://") {
				if strings.HasPrefix(highestQualitySource, "//") {
					highestQualitySource = "https:" + highestQualitySource
				} else if strings.HasPrefix(highestQualitySource, "/") {
					highestQualitySource = baseURL.Scheme + "://" + baseURL.Host + highestQualitySource
				} else {
					highestQualitySource = baseURL.Scheme + "://" + baseURL.Host + "/" + highestQualitySource
				}
			}
			img.SetAttr("src", highestQualitySource)
			log.Debugw("Updated picture element source", "new_src", highestQualitySource, "width", maxWidth)
		}

		// Remove loading and decoding attributes
		img.RemoveAttr("loading")
		img.RemoveAttr("decoding")

		// Replace picture with the img element
		picture.ReplaceWithSelection(img)
	})

	if pictureCount > 0 {
		log.Debugw("Processed picture elements", "count", pictureCount)
	}

	// Handle standalone img elements
	imgCount := 0
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		imgCount++
		// Remove loading and decoding attributes
		img.RemoveAttr("loading")
		img.RemoveAttr("decoding")
		img.RemoveAttr("srcset")

		// Handle relative URLs for src attribute
		if src, exists := img.Attr("src"); exists {
			originalSrc := src
			if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
				if strings.HasPrefix(src, "//") {
					src = "https:" + src
				} else if strings.HasPrefix(src, "/") {
					src = baseURL.Scheme + "://" + baseURL.Host + src
				} else {
					src = baseURL.Scheme + "://" + baseURL.Host + "/" + src
				}
				img.SetAttr("src", src)
				log.Debugw("Updated img src", "original", originalSrc, "new", src)
			}
		}
	})

	if imgCount > 0 {
		log.Debugw("Processed img elements", "count", imgCount)
	}

	// Convert the document back to a Node for readability
	htmlNode := doc.Get(0)

	log.Info("Converting document to readability format")
	article, err := readability.FromDocument(htmlNode, nil)
	if err != nil {
		log.Errorw("Failed to convert document to readability format", "error", err)
		return CleanedArticle{}, err
	}

	// Clean the text content
	cleanedTextContent := cleanTextContent(article.TextContent)

	// Convert HTML content to markdown
	log.Info("Converting HTML content to markdown")
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(article.Content)
	if err != nil {
		log.Warnw("Failed to convert to markdown, using HTML content", "error", err)
		markdown = article.Content // Fallback to HTML if markdown conversion fails
	}

	// Generate excerpt
	excerpt := generateExcerpt(cleanedTextContent, 200)

	// Handle published time safely
	var publishedAt string
	if article.PublishedTime != nil {
		publishedAt = article.PublishedTime.Format("2006-01-02T15:04:05Z07:00")
	}

	cleanedArticle := CleanedArticle{
		Title:       strings.TrimSpace(article.Title),
		Content:     cleanedTextContent,
		Markdown:    markdown,
		URL:         url,
		Author:      strings.TrimSpace(article.Byline),
		Excerpt:     excerpt,
		Length:      len(cleanedTextContent),
		PublishedAt: publishedAt,
		OpenGraph:   openGraphData,
	}

	log.Infow("Successfully processed article",
		"title_length", len(cleanedArticle.Title),
		"content_length", cleanedArticle.Length,
		"markdown_length", len(cleanedArticle.Markdown),
		"url", url,
	)

	// Save the cleaned HTML to a file for debugging
	html, err := doc.Html()
	if err != nil {
		log.Errorw("Failed to get HTML", "error", err)
	} else {
		os.WriteFile("tmp/article.html", []byte(html), 0644)
		log.Debugw("Saved cleaned HTML to file", "file", "tmp/article.html")
	}

	return cleanedArticle, nil
}

// GetReadableArticle returns just the text content (for backward compatibility)
func GetReadableArticle(url string) string {
	article := GetCleanedArticle(url)
	return article.Content
}

// GetCleanedArticle returns a comprehensive cleaned article with markdown
func GetCleanedArticle(url string) CleanedArticle {
	// Initialize logger for this function
	log, err := logger.NewSugaredLogger()
	if err != nil {
		// Fallback to empty article if logger fails
		return CleanedArticle{}
	}
	defer log.Sync()

	article, err := mutateFetchReadable(url, log)
	if err != nil {
		log.Errorw("Failed to get readable article", "url", url, "error", err)
		return CleanedArticle{}
	}

	return article
}

// GetOpenGraphData extracts only Open Graph metadata from a URL
func GetOpenGraphData(url string) *OpenGraphData {
	// Initialize logger for this function
	log, err := logger.NewSugaredLogger()
	if err != nil {
		// Fallback to empty Open Graph data if logger fails
		return &OpenGraphData{}
	}
	defer log.Sync()

	log.Infow("Starting to fetch Open Graph data", "url", url)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Errorw("Failed to fetch URL for Open Graph data", "url", url, "error", err)
		return &OpenGraphData{}
	}
	defer resp.Body.Close()

	log.Infow("Successfully fetched URL for Open Graph data", "url", url, "status_code", resp.StatusCode)

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Errorw("Failed to parse HTML document for Open Graph data", "url", url, "error", err)
		return &OpenGraphData{}
	}

	// Extract Open Graph data
	openGraphData := extractOpenGraphData(doc, url, resp.Request, log)

	log.Infow("Successfully extracted Open Graph data", "url", url, "title", openGraphData.Title)

	return openGraphData
}
