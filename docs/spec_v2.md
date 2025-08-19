Of course. Here is the comprehensive, developer-ready specification for the AI News Agent CLI.

## **AI News Agent CLI: Developer Specification**

This document outlines the complete technical and functional requirements for building the AI News Agent, a command-line tool designed to intelligently fetch, process, and display news about Artificial Intelligence.

---

### **1. System Architecture**

The application is a standalone command-line tool built with a modular architecture.

* **Application Type**: Command-Line Interface (CLI).
* **AI Processing**: All Natural Language Processing (NLP) tasks will be handled by a **Cloud AI API** (e.g., OpenAI's GPT series, Google's Gemini). This choice prioritizes processing power and ease of implementation over local execution.
* **Data Storage**: A local **SQLite database** will be used for persistent storage of all processed news items. This provides a lightweight, serverless, and robust solution for managing application data.
* **External Services**:
    * **Content Extraction**: Use a service like **Jina Reader** (`r.jina.ai`) to fetch clean, reader-friendly markdown from article URLs.
    * **Terminal Rendering**: Use a markdown renderer like **Glow** to display article content within the terminal.



---

### **2. Core Features & Functionality**

#### **Data Ingestion & Sources**

* **Source Types**: The agent will monitor a pre-defined list of sources, including news websites, academic preprint servers (e.g., arXiv), blogs, YouTube channels, and social media accounts.
* **Source Management**: The list of sources and their configurations will be **hardcoded or managed in a local configuration file** (e.g., `config.json` or `config.yaml`). There is no user-facing feature for adding sources in this version.
* **Source Prioritization**: A tiered priority system will be implemented in the configuration to rank sources. This ranking dictates the sort order of news items and determines which article is primary when duplicates are found.
    * **Tier 1**: Primary sources (e.g., official company blogs).
    * **Tier 2**: Top-tier technical journalism and research hubs.
    * **Tier 3**: General tech media and other reputable sources.

#### **AI-Powered Data Processing**

* **Summarization**: Generate concise, bullet-point summaries for each piece of content.
* **Entity Extraction**: Identify and extract key entities, categorized as `Organizations`, `Products & Models`, and `People`.
* **Topic Classification**: Assign 2-3 `MainTopics` (e.g., "Large Language Models," "AI Ethics") to each item.
* **Content-Type Classification**: Determine the `ContentType` (e.g., "Research Paper," "Product Launch," "Opinion Piece").
* **Deduplication (Story Clustering)**: The agent must identify articles from different sources that cover the same core news event. These items will be linked in the database and displayed as a single, grouped story.

---

### **3. Data Schema (SQLite)**

The `articles` table will store all processed items with the following structure:

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | INTEGER | Primary Key |
| `title` | TEXT | The original title of the content. |
| `url` | TEXT | The direct link to the original source. |
| `source_name` | TEXT | The name of the source publication or channel. |
| `published_date` | DATETIME | The date and time the content was published. |
| `summary` | TEXT | The AI-generated bullet-point summary. |
| `entities` | JSON | A JSON object storing lists of extracted entities. |
| `content_type` | TEXT | The classified type of the content. |
| `topics` | JSON | A JSON array of the main topics. |
| `status` | TEXT | The current status of the item (`unread` or `read`). |
| `story_group_id` | TEXT | A unique identifier linking duplicate articles. |

---

### **4. CLI Command Reference**

The tool will expose four main commands:

#### **`ai-news fetch`**

* **Purpose**: Fetches new content from all sources, processes it, and saves it to the database.
* **Behavior**:
    1.  Iterates through the configured sources.
    2.  For each source, it retrieves the latest content, checking the `url` against the database to avoid re-processing.
    3.  For each new item, it calls the Cloud AI API to perform all processing tasks.
    4.  It runs a deduplication check against recent entries to identify story clusters.
    5.  Saves the fully processed data to the SQLite database with a default `status` of `unread`.
    6.  The command should output a status message, e.g., "Fetch complete. Added 12 new items."

#### **`ai-news view`**

* **Purpose**: Displays processed news from the local database.
* **Behavior**:
    1.  By default, queries the database for all items with `status = 'unread'`.
    2.  Groups items by `story_group_id` to present duplicate stories as a single entry.
    3.  Sorts the results first by **Source Priority Tier** (ascending), then by `published_date` (descending).
    4.  Displays each story group in a numbered, detailed "card" format. The primary article is determined by the highest-priority source within the group.
    5.  After displaying, it updates the `status` of the viewed items to `read`.
* **Arguments & Flags**:
    * `--all`: Shows all items, including those marked `read`.
    * `--source "[Source Name]"`: Filters results by source.
    * `--topic "[Topic Name]"`: Filters results by a specific topic.

#### **`ai-news read <number>`**

* **Purpose**: Renders the full article content within the terminal.
* **Behavior**:
    1.  Accepts a number corresponding to an item from the last `view` output.
    2.  Retrieves the `url` for the primary article in the selected story group.
    3.  Passes the `url` to the Jina Reader service to fetch the content as markdown.
    4.  Pipes the markdown content to the Glow renderer for display.

#### **`ai-news open <number>`**

* **Purpose**: Opens the original article in the user's default web browser.
* **Behavior**:
    1.  Accepts a number corresponding to an item from the last `view` output.
    2.  Retrieves the `url` for the primary article.
    3.  Launches the URL using the system's default browser.

---

### **5. Error Handling**

Robust error handling is critical for a good user experience.

* **Network Errors**: Any command that makes a network request (`fetch`, `read`) must handle connection timeouts, DNS failures, and other network issues gracefully, exiting with a clear error message (e.g., "Error: Could not connect to the internet.").
* **API Failures**:
    * Handle API key errors (missing or invalid) with a specific message guiding the user to configure their key.
    * Implement retries with exponential backoff for transient API errors (e.g., rate limits, server-side errors).
    * If an API call fails permanently for a specific item, log the error and continue the fetch process for other items where possible.
* **Parsing Errors**: If a source's content cannot be parsed or the Jina Reader service fails, the tool should skip that item and report the failure in the `fetch` summary (e.g., "Fetch complete. Added 11 new items. 1 item failed to process.").
* **Database Errors**: Handle potential SQLite errors, such as a corrupt database file or permission issues, with clear, actionable error messages.

---

### **6. Testing Plan**

A multi-layered testing strategy should be implemented to ensure reliability.

#### **Unit Tests**

* Test individual functions in isolation.
* **Examples**:
    * Test the database query functions (e.g., `get_unread_articles`, `mark_as_read`).
    * Test data transformation logic.
    * Mock API calls to test the AI processing module's handling of successful and failed API responses.

#### **Integration Tests**

* Test the interaction between different modules of the application.
* **Examples**:
    * Test the full `fetch` command's workflow: can it connect to a mock source, send data to a mock API, and correctly write the results to a test database?
    * Test the interaction between the `view` and `read` commands, ensuring the correct data is passed and status updates (`unread` -> `read`) occur as expected.

#### **End-to-End (E2E) Tests**

* Test the application as a whole from the user's perspective by running the actual CLI commands.
* **Examples**:
    * A test script that runs `ai-news fetch`, then `ai-news view`, captures the output, and verifies that it has the expected format and content.
    * A test that runs `ai-news view`, then `ai-news open 1`, and confirms that the correct URL is being opened (this may require mocking the browser-opening function).