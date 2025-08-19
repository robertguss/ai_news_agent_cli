Of course. Switching to Go and the Charm ecosystem is a fantastic choice for building a polished, high-performance CLI. We'll adopt idiomatic Go practices, including a strong emphasis on interfaces, robust error handling, and a clear project structure.

Here is the updated, step-by-step blueprint and the corresponding prompts for a code-generation LLM, tailored specifically for Go.

### **Development Blueprint & Strategy (Go Edition)**

Our approach remains **Test-Driven Development (TDD)**, starting with a "walking skeleton." We will build the application from the database and core logic outward to the CLI commands.

**Go-Specific Toolchain:**

  * **CLI Framework:** `Cobra` for powerful command, subcommand, and flag management.
  * **Configuration:** `Viper` for handling `config.yaml`.
  * **Database:** Standard `database/sql` package with the `modernc.org/sqlite` driver (to avoid CGo). We will use `sqlc` to generate type-safe Go code from our SQL queries, ensuring a clean separation between SQL and Go.
  * **HTTP & Parsing:** Standard `net/http` for API calls and `mmcdole/gofeed` for parsing RSS/Atom feeds.
  * **Styling & Rendering:** `Lip Gloss` for styling the "card" output in the terminal and `Glow` (called via `os/exec`) for rendering markdown.
  * **Testing:** Go's built-in `testing` package with `stretchr/testify` for expressive assertions and `mockery` for auto-generating mocks from our interfaces.

**Phases of Development:**

1.  **Phase 1: Project Skeleton & Database Core:** Establish the Go module, Cobra CLI structure, and a fully tested, type-safe database layer using `sqlc`.
2.  **Phase 2: Basic Content Pipeline:** Implement configuration loading with Viper and a robust fetcher for RSS feeds. Wire this into a basic `fetch` command that populates the database without AI.
3.  **Phase 3: AI Processing Layer:** Define a clear `Processor` interface. Build a mock implementation for testing, then a real implementation that calls a cloud AI API.
4.  **Phase 4: Advanced Features & User Experience:** Enhance the `view` command with styled output using `Lip Gloss`. Implement filtering and the interactive `read`/`open` commands.
5.  **Phase 5: Finalization & Refinement:** Implement robust, idiomatic error handling throughout the application and finalize the configuration.

-----

### **Prompts for Code-Generation LLM (Go Edition)**

Here are the sequential prompts to build the application in Go. Each prompt is designed to be a self-contained, testable step.

-----

### **Phase 1: Project Skeleton & Database Core**

-----

#### **Prompt 1: Initial Project & Cobra CLI Setup**

This prompt establishes the Go module, project structure, and a "hello world" root command to verify the setup.

````text
Create a new Go project for an application named "AI News Agent". Use `Cobra` for the CLI framework and `stretchr/testify` for testing.

1.  **Initialize the Go module:**
    `go mod init github.com/user/ai-news-agent`

2.  **Create the project directory structure:**
    ```
    ai-news-agent/
    ├── cmd/           // Cobra commands
    │   └── root.go
    ├── internal/      // Internal application logic
    ├── main.go
    └── go.mod
    ```

3.  **Add dependencies:**
    `go get github.com/spf13/cobra@latest`
    `go get github.com/stretchr/testify@latest`

4.  **`main.go`:**
    Create the main function that simply calls `cmd.Execute()`.

5.  **`cmd/root.go`:**
    - Create the `rootCmd` using Cobra. It should have a `Short` description.
    - The `Execute` function will be called by `main.go`.
    - For now, the `Run` function for the root command can just print "AI News Agent".

6.  **Create `cmd/root_test.go`:**
    Write a test function `TestRootCmd`. It should execute the root command and capture the standard output. Use `testify/assert` to check that the output contains "AI News Agent".
````

-----

#### **Prompt 2: Database Layer with `sqlc`**

This prompt creates the database schema and generates type-safe Go code for all database interactions. This is a crucial step in idiomatic Go development.

````text
Based on the previous step, let's create the database layer using the standard `database/sql` package, the `modernc.org/sqlite` driver, and `sqlc` for code generation.

1.  **Add dependencies:**
    `go get modernc.org/sqlite@latest`
    `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

2.  **Create a new directory `internal/database`**.

3.  **Inside `internal/database`, create `sqlc.yaml`:**
    Configure `sqlc` to read SQL files and generate Go code in the same directory.
    ```yaml
    version: "2"
    sql:
      - engine: "sqlite"
        queries: "queries.sql"
        schema: "schema.sql"
        gen:
          go:
            package: "database"
            out: "."
            sql_package: "database/sql"
    ```

4.  **Create `internal/database/schema.sql`:**
    Define the `articles` table exactly as specified, using SQLite syntax.

5.  **Create `internal/database/queries.sql`:**
    Write the SQL queries with `sqlc` annotations that we will need.
    - `-- name: CreateArticle :one` (An INSERT statement that returns the inserted article)
    - `-- name: GetArticleByUrl :one` (A SELECT statement to check if a URL exists)
    - `-- name: ListArticles :many` (A simple SELECT * FROM articles)

6.  **Run `sqlc generate`** from the `internal/database` directory. This will create `db.go`, `models.go`, and `queries.sql.go`.

7.  **Create a test file `internal/database/db_test.go`:**
    - Write a test function that opens an in-memory SQLite database for testing.
    - Use the generated `New(db)` function to get a `Queries` object.
    - Test the `CreateArticle` function by inserting a sample article and asserting that the returned object's fields match the input, using `testify/assert`.
````

-----

#### **Prompt 3: Wire Database into a Basic `view` Command**

This prompt connects the generated database code to a new Cobra command, creating the first piece of real functionality.

```text
Now, let's integrate our `sqlc`-generated database module with the CLI by creating a `view` command.

1.  **Create `cmd/view.go`:**
    - Define a new `viewCmd` using Cobra and add it to the `rootCmd` in `root.go`'s `init()` function.
    - The `RunE` function for this command should handle errors idiomatically (it returns an `error`).

2.  **In the `viewCmd`'s `RunE` function:**
    - Open a connection to the SQLite database file (e.g., `news.db`). Handle errors.
    - Create a new `database.Queries` object using the connection.
    - Call the `ListArticles` method.
    - If there are no articles, print "No articles found."
    - If there are articles, iterate through them and print the `Title` and `SourceName` for each.
    - Return `nil` on success.

3.  **Create `cmd/view_test.go`:**
    - Write a test for `TestViewCmdEmpty`. It should execute the `view` command with a temporary, empty database and assert that the output contains "No articles found."
    - Write a test for `TestViewCmdWithArticles`. It should:
      a. Create a temporary database and use the generated `sqlc` code to insert a sample article.
      b. Execute the `view` command.
      c. Assert that the command's output contains the title of the sample article.
```

-----

### **Phase 2: Basic Content Pipeline**

-----

#### **Prompt 4: Configuration and RSS Fetcher**

This prompt introduces configuration management with Viper and creates a dedicated, testable module for fetching and parsing RSS feeds.

````text
Let's build the foundation for the `fetch` command. We will use `Viper` for configuration and `gofeed` for parsing.

1.  **Add dependencies:**
    `go get github.com/spf13/viper@latest`
    `go get github.com/mmcdole/gofeed@latest`

2.  **Create `config.yaml` in the project root:**
    ```yaml
    sources:
      - name: "Ars Technica"
        url: "[http://feeds.arstechnica.com/arstechnica/index](http://feeds.arstechnica.com/arstechnica/index)"
        type: "rss"
        priority: 2
    ```

3.  **Create a new package `internal/fetcher`:**
    - Create `fetcher.go`. Define a `Source` struct to match the config and an `Article` struct for the parsed data (`Title`, `Link`, `PublishedDate`).
    - Create a `Fetch(source Source)` function. It should use `gofeed.NewParser()` to parse the source URL and return a slice of `Article` structs and an error.

4.  **Create `internal/fetcher/fetcher_test.go`:**
    - Write a test for `Fetch`.
    - Use Go's `net/http/httptest` to create a mock HTTP server that returns a sample, hardcoded RSS XML string.
    - Pass the URL of the test server to your `Fetch` function.
    - Assert that the function returns the correctly parsed articles and no error.

5.  **Create a new package `internal/config`:**
    - Create `config.go`. Define a function `Load()` that uses Viper to read `config.yaml` and unmarshal the sources into a slice of `fetcher.Source` structs.
````

-----

#### **Prompt 5: Implement the Basic `fetch` Command**

This prompt connects the fetcher to the database, completing the non-AI data pipeline in a new `fetch` command.

```text
Now, let's create the `fetch` command, wiring together the config, fetcher, and database modules. This version will not perform any AI processing.

1.  **Create `cmd/fetch.go`:**
    - Define a new `fetchCmd` using Cobra and add it to the `rootCmd`.

2.  **In the `fetchCmd`'s `RunE` function:**
    a. Load the configuration using your `config.Load()` function.
    b. Open the database connection and create a `database.Queries` object.
    c. Iterate through the configured sources. For each source, call `fetcher.Fetch()`.
    d. For each article returned by the fetcher:
       i. Use the `GetArticleByUrl` query to check if it already exists.
       ii. If it doesn't exist, call the `CreateArticle` query. Populate the `database.CreateArticleParams` struct. For now, AI-related fields can be empty strings. The `Status` should be 'unread'.
    e. Print a summary message, e.g., "Fetch complete. Added X new articles."

3.  **Create `cmd/fetch_test.go`:**
    - This is an integration test.
    - Use `httptest` to mock the RSS feed endpoint.
    - Use a temporary database file.
    - Execute the `fetch` command.
    - Assert that the output contains the correct summary message ("Added 1 new articles.").
    - Query the test database directly to confirm the article was inserted.
    - Execute the `fetch` command again and assert the summary message says "Added 0 new articles."
```

-----

### **Phase 3: AI Processing Layer**

-----

#### **Prompt 6: AI Processor Interface and Mock**

This prompt defines a clean boundary for the AI logic using an interface and creates a mock implementation for TDD using `mockery`.

````text
Let's create the AI processing module. We will start by defining an interface and using `mockery` to generate a mock for testing.

1.  **Add mockery:**
    `go install github.com/vektra/mockery/v2@latest`

2.  **Create a new package `internal/processor`**.

3.  **In `internal/processor/processor.go`, define the interface:**
    ```go
    package processor

    //go:generate mockery --name AIProcessor
    type AIProcessor interface {
        AnalyzeContent(content string) (*AnalysisResult, error)
    }

    type AnalysisResult struct {
        Summary      string
        Entities     []byte // JSON
        Topics       []byte // JSON
        ContentType  string
        StoryGroupID string
    }
    ```
    The `//go:generate` comment is crucial.

4.  **Run `go generate ./...`** from the root directory. This will create a `mocks` directory inside `internal/processor` with a mock implementation of `AIProcessor`.

5.  **Create a new package `internal/scraper`:**
    - Create `scraper.go` with a function `Scrape(url string) (string, error)` that fetches a URL and extracts the main text content. Use a library like `go-trafilatura`. Add it as a dependency.

6.  **Update the `fetch` command in `cmd/fetch.go`:**
    - This is a temporary change for testing the wiring. In the `fetchCmd`'s `RunE`, instantiate the *mock* processor: `mockProcessor := new(mocks.AIProcessor)`.
    - Define the mock's expected behavior: `mockProcessor.On("AnalyzeContent", mock.Anything).Return(&processor.AnalysisResult{Summary: "mock summary"}, nil)`.
    - When a new article is found, call `scraper.Scrape()` on its URL.
    - Then call `mockProcessor.AnalyzeContent()` with the scraped text.
    - Use the mock's return value to populate the database record.

7.  **Update `cmd/fetch_test.go`:**
    - Modify the `fetch` command test. After running `fetch`, query the database and assert that the `summary` field of the new article is "mock summary". This proves the interface is wired correctly.
````

-----

#### **Prompt 7: Implement Real AI API Client**

Now we build the real implementation of the `AIProcessor` interface.

````text
Let's implement the real `AIProcessor` that calls a cloud AI API.

1.  **Add the OpenAI Go client library:**
    `go get github.com/sashabaranov/go-openai@latest`

2.  **In `internal/processor/processor.go`, create the real implementation:**
    ```go
    type OpenAIProcessor struct {
        client *openai.Client
    }

    func NewOpenAIProcessor(apiKey string) *OpenAIProcessor {
        // ... create and return a new processor
    }

    func (p *OpenAIProcessor) AnalyzeContent(content string) (*AnalysisResult, error) {
        // ... implementation details
    }
    ```
    - The `AnalyzeContent` method should:
      a. Get the API key from an environment variable (e.g., `OPENAI_API_KEY`).
      b. Create a chat completion request. The prompt should instruct the model to return a JSON object matching the `AnalysisResult` struct. Use the model's JSON mode.
      c. Make the API call.
      d. Unmarshal the JSON response into an `AnalysisResult` struct.
      e. Return the result. Handle all potential errors.

3.  **Update `cmd/fetch.go`:**
    - In `fetchCmd`, replace the mock instantiation with the real one: `processor := processor.NewOpenAIProcessor(os.Getenv("OPENAI_API_KEY"))`.

4.  **Create `internal/processor/processor_test.go`:**
    - Write a unit test for `OpenAIProcessor`.
    - Do *not* make a real API call. Instead, use `net/http/httptest` to mock the OpenAI API endpoint.
    - Your test should create an `OpenAIProcessor` that points to the test server's URL.
    - Call `AnalyzeContent` and assert that it correctly parses the mock JSON response from your test server.
````

-----

### **Phase 4: Advanced Features & UX**

-----

#### **Prompt 8: Implement `read`/`unread` and Filtering in `view`**

This prompt makes the `view` command powerful by adding state management, filtering, and styled output.

```text
Let's enhance the `view` command with `read`/`unread` logic, filtering, and styled output using `Lip Gloss`.

1.  **Add `Lip Gloss` dependency:**
    `go get github.com/charmbracelet/lipgloss@latest`

2.  **Update `internal/database/queries.sql`:**
    - Create a new query `MarkArticlesAsRead :exec` that takes a slice of IDs and updates their status.
    - Modify `ListArticles` to be `ListUnreadArticles :many` which adds `WHERE status = 'unread'`.
    - Add new queries for filtering: `ListArticlesBySource :many` and `ListArticlesByTopic :many`.
    - Run `sqlc generate` to update the Go code.

3.  **Update the `viewCmd` in `cmd/view.go`:**
    - Add flags for `--all`, `--source`, and `--topic` to the command definition.
    - In the `RunE` function, use logic to decide which `sqlc` query to call based on the flags.
    - After retrieving the articles, collect their IDs.
    - If the command is not in `--all` mode, call `MarkArticlesAsRead` with the collected IDs.
    - Use `Lip Gloss` to style the output. Create styles for titles, sources, and summaries to build a visually appealing "card" for each article.

4.  **Update `cmd/view_test.go`:**
    - Add new tests to verify the filtering logic for `--source` and `--topic`.
    - Add a test to verify the `read`/`unread` workflow:
      a. Insert two unread articles.
      b. Run `view` and confirm both are shown.
      c. Run `view` again and confirm "No articles found" is shown.
      d. Run `view --all` and confirm both articles are shown again.
```

-----

#### **Prompt 9: Implement the `read` and `open` Commands**

This final prompt adds the interactive commands, completing the core application workflow.

```text
Let's add the `read` and `open` commands.

1.  **Add a cross-platform "open" library:**
    `go get github.com/pkg/browser@latest`

2.  **Update `cmd/view.go`:**
    - When the `view` command displays articles, it must store the mapping of the displayed number (1, 2, 3...) to the article's database ID and URL. A simple approach is to write this mapping to a temporary file (e.g., in `os.TempDir()`).

3.  **Create `cmd/open.go` and `cmd/read.go`:**
    - Add `openCmd` and `readCmd` to `rootCmd`.
    - Both commands should take a single argument: the article number.
    - They should read the temporary mapping file created by `view` to find the URL for the given number.
    - **`openCmd`:** should use `browser.OpenURL(url)` to open the article in the default web browser.
    - **`readCmd`:** should first fetch the markdown content from Jina Reader (`https://r.jina.ai/{url}`). Then, it should use `os/exec` to run `glow` as a subprocess, passing the markdown content to its standard input. Handle the error if `glow` is not in the system's PATH.

4.  **Create `cmd/open_test.go` and `cmd/read_test.go`:**
    - For these tests, you will need to mock the external interactions.
    - First, create the temporary mapping file that `view` would have created.
    - **For the `open` test:** You cannot easily test that a browser opened. Instead, test that the command correctly reads the mapping file and attempts to open the correct URL.
    - **For the `read` test:** Mock the `os/exec` call. You can do this by changing the `PATH` environment variable for the test process to point to a directory with a fake `glow` script that just records its arguments. Assert that your command tried to execute `glow` with the correct markdown content.
```