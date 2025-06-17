# Page Zen API - Enhanced Article Cleaning and Markdown Support

## Overview

The Page Zen API now includes advanced article cleaning capabilities and markdown conversion, powered by Uber Zap logging for comprehensive monitoring.

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

### 📊 Comprehensive Logging
- **Structured Logging**: Uses Uber Zap for structured, contextual logging
- **Environment-Based Configuration**: Different log levels for development and production
- **Request Tracking**: Detailed logging of all HTTP requests and processing steps
- **Performance Monitoring**: Tracks processing times and content metrics

## API Endpoints

### 1. POST /extract
Extract and clean article content with full control over the response format.

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
  "success": true
}
```

### 2. GET /extract
Simple URL-based extraction with query parameters.

**Examples:**

Basic extraction:
```bash
GET /extract?url=https://example.com/article
```

With markdown:
```bash
GET /extract?url=https://example.com/article&markdown=true
```

## Response Fields

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
| `success`      | boolean | Whether extraction succeeded    |
| `message`      | string  | Error message (if applicable)   |

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

**Basic POST request:**
```bash
curl -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article"}'
```

**POST with markdown:**
```bash
curl -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article", "include_markdown": true}'
```

**Simple GET request:**
```bash
curl "http://localhost:8080/extract?url=https://example.com/article"
```

**GET with markdown:**
```bash
curl "http://localhost:8080/extract?url=https://example.com/article&markdown=true"
```

### JavaScript/Fetch Examples

```javascript
// POST request with fetch
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
2024-01-15T10:30:01.500Z	INFO	Total unwanted elements removed	{"count": 15}
2024-01-15T10:30:02.200Z	INFO	Successfully processed article	{"title": "Article Title", "content_length": 1250, "url": "https://example.com/article"}
```

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