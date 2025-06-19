# Page Zen API - Enhanced Article Cleaning, Markdown Support & Open Graph Extraction

![Adobe Express - file](https://github.com/user-attachments/assets/282bd1f2-4604-4a2d-ab3b-9996d476748f)



## Overview

The Page Zen API provides comprehensive web content extraction with advanced article cleaning capabilities, markdown conversion, and Open Graph metadata extraction, powered by Uber Zap logging for comprehensive monitoring.

## Features

### 🧹 Enhanced Article Cleaning
- **Unwanted Element Removal**: Automatically removes ads, social media widgets, navigation menus, comments, and other non-content elements
- **Script & Style Cleaning**: Removes all JavaScript and CSS that could interfere with content
- **Image Optimization**: Processes picture elements and selects highest quality sources
- **Text Content Cleaning**: Removes excessive whitespace and unwanted text patterns
- **Relative URL Resolution**: Converts relative URLs to absolute URLs

### 📝 Markdown Conversion
- **HTML to Markdown**: Converts cleaned HTML content to markdown format
- **Optional Inclusion**: Choose whether to include markdown in the response
- **Fallback Support**: If markdown conversion fails, falls back to cleaned HTML

### 🏷️ Open Graph Metadata Extraction
- **Complete Open Graph Support**: Extracts all standard Open Graph meta tags (og:title, og:description, og:image, etc.)
- **Twitter Card Integration**: Captures Twitter Card metadata for enhanced social sharing
- **Article Metadata**: Extracts article-specific data like author, publication date, tags, and sections
- **Fallback to Standard Meta Tags**: Uses standard HTML meta tags when Open Graph data is unavailable
- **URL Normalization**: Converts relative URLs to absolute URLs for images and links
- **Standalone Endpoint**: Dedicated endpoint for Open Graph-only extraction

### 📊 Comprehensive Logging
- **Structured Logging**: Uses Uber Zap for structured, contextual logging
- **Environment-Based Configuration**: Different log levels for development and production
- **Request Tracking**: Detailed logging of all HTTP requests and processing steps
- **Performance Monitoring**: Tracks processing times and content metrics

## API Endpoints

### Article Extraction with Open Graph Data

### 1. POST /extract
Extract and clean article content with full control over the response format, including Open Graph metadata.

**Request:**
```json
{
  "url": "https://example.com/article",
  "include_markdown": true
}
```

**Response:**
```json
{
  "url": "https://example.com/article",
  "title": "Article Title",
  "content": "Cleaned text content...",
  "markdown": "# Article Title\n\nCleaned markdown content...",
  "author": "Author Name",
  "excerpt": "Brief excerpt of the article...",
  "length": 1250,
  "published_at": "2024-01-15T10:30:00Z",
  "open_graph": {
    "title": "Article Title",
    "description": "Article description for social sharing",
    "image": "https://example.com/featured-image.jpg",
    "url": "https://example.com/article",
    "type": "article",
    "site_name": "Example Site",
    "locale": "en_US",
    "twitter_card": "summary_large_image",
    "twitter_site": "@examplesite",
    "twitter_creator": "@author",
    "twitter_title": "Article Title",
    "twitter_description": "Twitter-optimized description",
    "twitter_image": "https://example.com/twitter-image.jpg",
    "author": "Author Name",
    "published_at": "2024-01-15T10:30:00Z",
    "section": "Technology",
    "tags": ["web", "api", "extraction"]
  },
  "success": true
}
```

### 2. GET /extract
Simple URL-based extraction with query parameters, including Open Graph metadata.

**Examples:**

Basic extraction with Open Graph data:
```bash
GET /extract?url=https://example.com/article
```

With markdown and Open Graph data:
```bash
GET /extract?url=https://example.com/article&markdown=true
```

### Open Graph Only Extraction

### 3. POST /opengraph
Extract only Open Graph metadata without article content processing.

**Request:**
```json
{
  "url": "https://example.com/article"
}
```

**Response:**
```json
{
  "url": "https://example.com/article",
  "open_graph": {
    "title": "Article Title",
    "description": "Article description for social sharing",
    "image": "https://example.com/featured-image.jpg",
    "url": "https://example.com/article",
    "type": "article",
    "site_name": "Example Site",
    "locale": "en_US",
    "twitter_card": "summary_large_image",
    "twitter_site": "@examplesite",
    "twitter_creator": "@author",
    "twitter_title": "Article Title",
    "twitter_description": "Twitter-optimized description",
    "twitter_image": "https://example.com/twitter-image.jpg",
    "author": "Author Name",
    "published_at": "2024-01-15T10:30:00Z",
    "section": "Technology",
    "tags": ["web", "api", "extraction"]
  },
  "success": true
}
```

### 4. GET /opengraph
Simple URL-based Open Graph extraction.

**Example:**
```bash
GET /opengraph?url=https://example.com/article
```

## Response Fields

### Article Extraction Response

| Field          | Type    | Description                     |
| -------------- | ------- | ------------------------------- |
| `url`          | string  | Original URL                    |
| `title`        | string  | Extracted article title         |
| `content`      | string  | Cleaned text content            |
| `markdown`     | string  | Markdown version (if requested) |
| `author`       | string  | Article author (if available)   |
| `excerpt`      | string  | Brief excerpt (200 chars max)   |
| `length`       | integer | Length of content in characters |
| `published_at` | string  | Publication date (ISO 8601)     |
| `open_graph`   | object  | Open Graph metadata (see below) |
| `success`      | boolean | Whether extraction succeeded    |
| `message`      | string  | Error message (if applicable)   |

### Open Graph Data Fields

| Field                 | Type   | Description                     |
| --------------------- | ------ | ------------------------------- |
| `title`               | string | Open Graph title                |
| `description`         | string | Open Graph description          |
| `image`               | string | Featured image URL              |
| `url`                 | string | Canonical URL                   |
| `type`                | string | Content type (article, website) |
| `site_name`           | string | Site name                       |
| `locale`              | string | Content locale                  |
| `twitter_card`        | string | Twitter Card type               |
| `twitter_site`        | string | Twitter site handle             |
| `twitter_creator`     | string | Twitter creator handle          |
| `twitter_title`       | string | Twitter-specific title          |
| `twitter_description` | string | Twitter-specific description    |
| `twitter_image`       | string | Twitter-specific image          |
| `author`              | string | Article author                  |
| `published_at`        | string | Publication date                |
| `modified_at`         | string | Last modification date          |
| `section`             | string | Article section/category        |
| `tags`                | array  | Article tags                    |

## Cleaning Process

### Elements Removed
- Scripts, styles, and noscript tags
- Navigation menus, headers, and footers
- Advertisements and social media widgets
- Comments and related content sections
- Tracking and analytics code
- Cookie notices and popups
- Subscription and newsletter boxes
- Video embeds and iframes
- Sidebar and widget content
- Forms (except search forms)

### Text Patterns Cleaned
- "Subscribe to our newsletter"
- "Follow us on [social media]"
- "Share this article"
- "Related articles"
- "You might also like"
- "Recommended for you"
- "Advertisement"
- "Sponsored content"

## Environment Configuration

### Development Mode
```bash
ENV=development LOG_LEVEL=debug ./bin/api
```
- Pretty formatted logs with colors
- Debug level logging enabled
- Detailed processing information

### Production Mode
```bash
ENV=production LOG_LEVEL=info ./bin/api
```
- JSON formatted logs
- Info level and above
- Performance optimized

## Usage Examples

### cURL Examples

**Article extraction with Open Graph data:**
```bash
curl -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article"}'
```

**Article extraction with markdown:**
```bash
curl -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article", "include_markdown": true}'
```

**Open Graph only extraction:**
```bash
curl -X POST http://localhost:8080/opengraph \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article"}'
```

**Simple GET requests:**
```bash
# Article extraction
curl "http://localhost:8080/extract?url=https://example.com/article"

# Article extraction with markdown
curl "http://localhost:8080/extract?url=https://example.com/article&markdown=true"

# Open Graph only
curl "http://localhost:8080/opengraph?url=https://example.com/article"
```

### JavaScript/Fetch Examples

```javascript
// Article extraction with Open Graph data
const response = await fetch('http://localhost:8080/extract', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    url: 'https://example.com/article',
    include_markdown: true
  })
});

const article = await response.json();
console.log(article.title, article.content);
console.log('Open Graph:', article.open_graph);

// Open Graph only extraction
const ogResponse = await fetch('http://localhost:8080/opengraph', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    url: 'https://example.com/article'
  })
});

const ogData = await ogResponse.json();
console.log('Social sharing data:', ogData.open_graph);
```

## Error Handling

The API returns appropriate HTTP status codes:
- `200 OK`: Successful extraction
- `400 Bad Request`: Invalid request (missing URL, invalid JSON)
- `500 Internal Server Error`: Extraction failed

Error responses include a descriptive message:
```json
{
  "success": false,
  "message": "Failed to extract article content",
  "url": "https://example.com/article"
}
```

## Logging

All requests and processing steps are logged with structured data. Example log entries:

```
2024-01-15T10:30:00.123Z	INFO	HTTP Request	{"client_ip": "127.0.0.1", "method": "POST", "path": "/extract", "status_code": 200, "latency": "2.5s"}
2024-01-15T10:30:00.124Z	INFO	Starting to fetch and parse article	{"url": "https://example.com/article"}
2024-01-15T10:30:00.500Z	INFO	Extracted Open Graph data	{"title": "Article Title", "description_length": 150, "image": "https://example.com/image.jpg", "type": "article", "site_name": "Example Site"}
2024-01-15T10:30:01.500Z	INFO	Total unwanted elements removed	{"count": 15}
2024-01-15T10:30:02.200Z	INFO	Successfully processed article	{"title_length": 50, "content_length": 1250, "markdown_length": 1400, "url": "https://example.com/article"}
```

## Use Cases

### Content Management Systems
- Import articles from external sources with clean content and metadata
- Preserve social sharing information during content migration

### Social Media Tools
- Extract Open Graph data for link previews
- Generate social media cards with proper metadata

### SEO and Analytics
- Analyze competitor content and metadata
- Audit Open Graph implementation across websites

### Content Aggregation
- Build news aggregators with rich metadata
- Create reading lists with proper article previews

### Browser Extensions
- Clean and save articles for offline reading
- Extract social sharing data for bookmarking tools

## Development

To run the enhanced version:

1. Build the application:
   ```bash
   go build -o bin/api cmd/api/main.go
   ```

2. Run with development settings:
   ```bash
   PORT=8080 ENV=development ./bin/api
   ```

3. Test with the included script:
   ```bash
   ./test_api.sh
   ```

The application will create a `tmp/` directory for debugging files and log all processing steps with detailed information. 
