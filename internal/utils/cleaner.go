package utils

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"page-zen/internal/logger"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"go.uber.org/zap"
)

func mutateFetchReadable(url string, log *zap.SugaredLogger) (readability.Article, error) {
	log.Infow("Starting to fetch and parse article", "url", url)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Errorw("Failed to fetch URL", "url", url, "error", err)
		return readability.Article{}, err
	}
	defer resp.Body.Close()

	log.Infow("Successfully fetched URL", "url", url, "status_code", resp.StatusCode)

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Errorw("Failed to parse HTML document", "url", url, "error", err)
		return readability.Article{}, err
	}

	// Set the base URL for resolving relative URLs
	baseURL := resp.Request.URL
	log.Debugw("Using base URL for relative links", "base_url", baseURL.String())

	// Remove script and style tags
	scriptCount := doc.Find("script").Length()
	doc.Find("script").Remove()
	log.Debugw("Removed script tags", "count", scriptCount)

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
		return readability.Article{}, err
	}

	log.Infow("Successfully processed article",
		"title", article.Title,
		"content_length", len(article.TextContent),
		"url", url,
	)

	// save the doc to a file
	html, err := doc.Html()
	if err != nil {
		log.Errorw("Failed to get HTML", "error", err)
		return readability.Article{}, err
	}

	os.WriteFile("tmp/article.html", []byte(html), 0644)

	return article, nil
}

func GetReadableArticle(url string) string {
	// Initialize logger for this function
	log, err := logger.NewSugaredLogger()
	if err != nil {
		// Fallback to empty string if logger fails
		return ""
	}
	defer log.Sync()

	article, err := mutateFetchReadable(url, log)
	if err != nil {
		log.Errorw("Failed to get readable article", "url", url, "error", err)
		return ""
	}

	return article.TextContent
}
