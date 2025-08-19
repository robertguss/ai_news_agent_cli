Of course. Here is a comprehensive `todo.md` checklist based on the Go development blueprint we created.

---

# AI News Agent CLI - Development Checklist

This checklist breaks down the construction of the Go-based AI News Agent into small, manageable tasks. Follow the steps in order to build the project incrementally with testing at each stage.

## ✅ **Phase 1: Project Skeleton & Database Core**

This phase establishes the project's foundation, including the CLI structure and the type-safe database layer.

* [ ] **Project Setup**
    * [ ] Initialize the Go module (`go mod init`).
    * [ ] Create the initial directory structure (`cmd/`, `internal/`, `main.go`).
    * [ ] Add initial dependencies: `Cobra` and `testify`.

* [ ] **CLI Entrypoint (`rootCmd`)**
    * [ ] Implement the basic `rootCmd` in `cmd/root.go`.
    * [ ] Implement `main.go` to execute the root command.
    * [ ] Write a test (`cmd/root_test.go`) to verify the root command's output.

* [ ] **Database Layer (`sqlc`)**
    * [ ] Add `modernc.org/sqlite` and install the `sqlc` tool.
    * [ ] Create the `internal/database` directory.
    * [ ] Define the database schema in `internal/database/schema.sql`.
    * [ ] Define SQL queries for `CreateArticle`, `GetArticleByUrl`, and `ListArticles` in `internal/database/queries.sql`.
    * [ ] Create and configure `internal/database/sqlc.yaml`.
    * [ ] Run `sqlc generate` to create the type-safe Go database code.
    * [ ] Write unit tests (`internal/database/db_test.go`) for the generated `CreateArticle` function using an in-memory database.

* [ ] **`view` Command (Initial)**
    * [ ] Create the `view` command in `cmd/view.go` and register it with `rootCmd`.
    * [ ] Implement the command to connect to the database and list article titles using the generated `ListArticles` function.
    * [ ] Write a test (`cmd/view_test.go`) for the `view` command against an empty database.
    * [ ] Write a second test for the `view` command against a database with pre-populated data.

## ✅ **Phase 2: Basic Content Pipeline**

This phase implements the ability to fetch and store articles from an external source, completing the data flow without AI processing.

* [ ] **Configuration**
    * [ ] Add `Viper` and `gofeed` as dependencies.
    * [ ] Create the `config.yaml` file with an initial RSS source.
    * [ ] Implement the `internal/config` package to load and parse the YAML file.

* [ ] **RSS Fetcher**
    * [ ] Implement the `internal/fetcher` package with a `Fetch` function to parse an RSS feed URL.
    * [ ] Write a unit test (`internal/fetcher/fetcher_test.go`) for the `Fetch` function using a mock HTTP server (`net/http/httptest`).

* [ ] **`fetch` Command (Initial)**
    * [ ] Create the `fetch` command in `cmd/fetch.go` and register it.
    * [ ] Implement the command logic: load config, iterate sources, call the fetcher, and check for existing articles using `GetArticleByUrl`.
    * [ ] Add new, non-duplicate articles to the database using `CreateArticle`.
    * [ ] Write an integration test (`cmd/fetch_test.go`) that uses a mock RSS feed and a temporary database to verify the command's behavior, including its duplicate-checking logic.

## ✅ **Phase 3: AI Processing Layer**

This phase introduces the AI capabilities, starting with a clean interface and mock implementation before wiring up the real API.

* [ ] **Processor Interface & Mock**
    * [ ] Add the `mockery` tool.
    * [ ] Define the `AIProcessor` interface and `AnalysisResult` struct in `internal/processor/processor.go`.
    * [ ] Add the `//go:generate` directive and run `go generate ./...` to create the mock.

* [ ] **Content Scraping**
    * [ ] Add a content scraping library dependency (e.g., `go-trafilatura`).
    * [ ] Implement the `internal/scraper` package with a `Scrape` function.
    * [ ] Write a unit test for the scraper using a mock HTTP server.

* [ ] **Integrate Mock Processor**
    * [ ] Update the `fetch` command to call the scraper and then the *mock* `AIProcessor`.
    * [ ] Update the database insertion logic to use the data returned from the mock processor.
    * [ ] Update the `fetch` command's test to assert that the mock's data (e.g., "mock summary") is correctly saved to the database.

* [ ] **Implement Real AI Client**
    * [ ] Add the Go client for your chosen Cloud AI API (e.g., `go-openai`).
    * [ ] Implement the real `OpenAIProcessor` that fulfills the `AIProcessor` interface. It should read the API key from an environment variable.
    * [ ] Write a unit test (`internal/processor/processor_test.go`) for the real processor that mocks the API endpoint using `net/http/httptest`.
    * [ ] Update the `fetch` command to instantiate and use the real processor.

## ✅ **Phase 4: Advanced Features & User Experience**

This phase builds the user-facing features that make the tool powerful and enjoyable to use.

* [ ] **Enhance `view` Command**
    * [ ] Add `Lip Gloss` as a dependency.
    * [ ] Update `queries.sql` with queries for filtering and marking articles as read, then run `sqlc generate`.
    * [ ] Implement the `read`/`unread` status logic in the `view` command.
    * [ ] Add the `--all`, `--source`, and `--topic` flags and wire them to the new database queries.
    * [ ] Use `Lip Gloss` to style the output into a visually appealing "card" format.
    * [ ] Update tests for `view` to cover the new status and filtering behaviors.

* [ ] **Interactive Commands State**
    * [ ] Implement the logic in the `view` command to save a mapping of displayed numbers to article IDs/URLs in a temporary file.

* [ ] **`open` Command**
    * [ ] Add the `pkg/browser` dependency.
    * [ ] Create the `open` command in `cmd/open.go`.
    * [ ] Implement the logic to read the mapping file and open the correct URL in a browser.
    * [ ] Write a test (`cmd/open_test.go`) that verifies the command reads the mapping and attempts to open the correct URL.

* [ ] **`read` Command**
    * [ ] Create the `read` command in `cmd/read.go`.
    * [ ] Implement the logic to fetch markdown from Jina Reader.
    * [ ] Use `os/exec` to pipe the markdown content to `glow`.
    * [ ] Write a test (`cmd/read_test.go`) that mocks the Jina Reader request and the `glow` subprocess execution.

## ✅ **Phase 5: Finalization & Refinement**

This final phase focuses on making the application robust and user-friendly.

* [ ] **Error Handling**
    * [ ] Review all network-facing code (`fetch`, `read`, processor) and ensure robust, user-friendly error handling for timeouts, DNS issues, etc.
    * [ ] Implement graceful error handling for API failures (e.g., invalid API key, rate limits).
    * [ ] Ensure database connection and query errors are handled properly.

* [ ] **Code Quality & Documentation**
    * [ ] Add comments to all public functions and complex logic blocks.
    * [ ] Run `go fmt` and `go vet` to ensure idiomatic formatting and catch potential issues.
    * [ ] Create a `README.md` file explaining what the project is, how to install it (including dependencies like `glow`), how to configure it (API key, `config.yaml`), and how to use each command.