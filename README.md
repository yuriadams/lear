# Project King Lear Explorer API

Welcome to the **Project King Lear Explorer API**! This API powers the exploration and analysis of books from Project Gutenberg, providing metadata extraction, text analysis, and streaming functionality for real-time insights. It's built with **Golang**, utilizes **Gorilla Mux** for routing, and integrates with the **SambaNova Cloud LLM** for advanced text analysis.

---

## Features

- **Book Metadata Scraping**
  - Extracts details such as Title, Author, Language, Credits, Summary, and Categories directly from Project Gutenberg.

- **Text Analysis**
  - Supports sentiment analysis, key character identification, language detection, and plot summarization using an LLM.

- **Streamed Responses**
  - Real-time streaming of analysis results for enhanced user interaction.

- **Caching System**
  - Reduces redundant book fetches by caching results for future use.

- **Structured Logging**
  - All logs are formatted consistently for easier debugging and monitoring.

---

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Endpoints](#endpoints)
- [Environment Variables](#environment-variables)
- [Error Handling](#error-handling)
- [Running Tests](#running-tests)
- [Future Improvements](#future-improvements)

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yuriadams/lear.git
   cd lear
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up your PostgreSQL database:
   ```bash
   createdb lear
   ```

4. Run migrations (if applicable):
   ```bash
    migrate -path=internal/database/migrations -database "YOUR_DATABASE_URL" -verbose up
   ```

5. Start the server:
   ```bash
   go run main.go
   ```

The API will be available at `http://localhost:3000` by default.

---

## Getting Started

This API relies on specific environment variables and integrates with external services. Ensure that the following are set up:

1. **Environment Variables:**
   - `SAMBA_NOVA_API_KEY`: API key for the SambaNova Cloud.
   - `DATABASE_URL`: Connection string for the PostgreSQL database.

2. Use a tool like [Postman](https://www.postman.com/) or `curl` to test the API endpoints.

---

## Endpoints

### 1. Fetch Book Metadata
- **GET** `/books/{gutenberg_id}`

  Fetches metadata for a given book from Project Gutenberg.

  **Parameters:**
  - `gutenberg_id` (required): The unique ID of the book in Project Gutenberg.

---

### 2. Stream Text Analysis
- **GET** `/analyze`

  Streams real-time analysis of the book content.

  **Query Parameters:**
  - `gutenberg_id` (required): The unique ID of the book in Project Gutenberg.

  **Example:**
  ```bash
  curl -N "http://localhost:3000/books/{gutenberg_id}/analyze"
  ```

  **Response (streamed):**
  ```
  data: {"sentiment":"positive","confidence":0.95}
  data: {"character":"Elizabeth Bennet","role":"protagonist"}
  data: [DONE]
  ```

---

## Environment Variables

| Variable             | Description                                    |
|----------------------|------------------------------------------------|
| `SAMBA_NOVA_API_KEY` | API key for SambaNova Cloud.                  |
| `DATABASE_URL`       | PostgreSQL database connection string.         |

---

## Error Handling

All errors are logged in the following format for consistency:
```
[book-{gutenbergID}] error message here
```

### Common Errors:
- **404 Not Found:** Book metadata is unavailable for the provided Gutenberg ID.
- **500 Internal Server Error:** An unexpected error occurred during text analysis.
- **400 Bad Request:** Missing or invalid parameters in the request.

---

## Running Tests

1. **Unit Tests:**
   ```bash
   go test ./...
   ```

2. **Specific Test File:**
   ```bash
   go test -v ./internal/service/analysis_service_test.go
   ```

3. **Debugging Failed Tests:**
   Run failed tests with verbose output for debugging:
   ```bash
   go test -v -run TestStreamTextAnalysis_Success
   ```

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Contributors

- [Yuri Adams](https://github.com/yuriadams) - Creator & Maintainer

---

Happy exploring! If you have questions, feel free to reach out or create an issue in the [repository](https://github.com/yuriadams/lear).

