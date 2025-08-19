Of course. Here is the complete specification for your AI News Agent, formatted in Markdown.

-----

# **Specification: AI News Agent CLI**

**Version:** 1.0
**Date:** August 19, 2025

## **1. Project Overview**

The project is a Command-Line Interface (CLI) application that acts as an intelligent AI agent. Its purpose is to help the user stay up-to-date with the latest news in the field of Artificial Intelligence by fetching, processing, summarizing, and displaying content from a variety of sources.

The tool is designed for a technical user (a software engineer), intended for personal use and as a portfolio project. It should be powerful, efficient, and have a "hacker-friendly" feel, allowing the user to perform the entire workflow within their terminal.

## **2. System Architecture**

  * **Frontend:** A CLI application.
  * **Backend Logic:** The core application logic responsible for fetching and processing.
  * **AI Engine:** Relies on a **Cloud AI API** (e.g., OpenAI's GPT series, Google's Gemini) for all natural language processing tasks.
  * **Data Storage:** A local **SQLite** database file for persistent storage of all processed news items.
  * **Dependencies:**
      * Content Fetching: A tool like **Jina Reader** to get clean, markdown-formatted article content.
      * Content Rendering: A tool like **Glow** to render markdown beautifully in the terminal.

## **3. Data Ingestion & Sources**

### **3.1. Source Types**

The agent will monitor a pre-defined list of the following source types:

  * Specific news websites
  * YouTube channels
  * Blogs
  * Social media accounts
  * Academic preprint servers (e.g., arXiv)

### **3.2. Source Management**

The list of sources will be **pre-defined in a configuration file or directly in the code**. The user will not manage the source list through the CLI in this version.

### **3.3. Source Priority System**

To ensure the most relevant information is prioritized, sources will be assigned a tier in the configuration. This tiering affects sorting and the selection of primary articles in story groups.

  * **Tier 1: Primary Sources** (e.g., Official company blogs like Google AI, OpenAI)
  * **Tier 2: Top-Tier Technical Journalism & Research** (e.g., Ars Technica, arXiv)
  * **Tier 3: General Tech Media & Reputable Channels** (e.g., The Verge, various blogs)

## **4. Data Processing & Schema**

When a new piece of content is found, the agent will use the Cloud AI API to process it and store it in the SQLite database with the following schema:

### **4.1. Data Schema**

  * **Core Metadata:**
      * `Title`: The original title of the content.
      * `Source`: The name of the source (e.g., "TechCrunch").
      * `URL`: The direct link to the original content.
      * `PublishedDate`: The publication date of the content.
  * **AI-Generated Analysis:**
      * `Summary`: A concise, bullet-point summary of the content.
      * `KeyEntities`: A list of extracted entities.
          * `Organizations`: (e.g., `NVIDIA`, `Hugging Face`)
          * `Products & Models`: (e.g., `GPT-4o`, `Llama 3`)
          * `People`: (e.g., `Geoffrey Hinton`)
  * **AI-Driven Categorization:**
      * `ContentType`: The agent's classification of the content (e.g., `Research Paper`, `Product Launch`, `Tutorial`).
      * `MainTopics`: A list of 2-3 key topics (e.g., `Large Language Models`, `AI Ethics`).
  * **Application Metadata:**
      * `Status`: `unread` or `read`.
      * `StoryGroupID`: An identifier to link duplicate articles together.

### **4.2. Core AI Features**

  * **Summarization:** Generate bullet-point summaries.
  * **Deduplication (Story Clustering):** The agent must identify articles from different sources that cover the same core story. These will be grouped together in the view. The article from the highest-priority source will be designated the "primary" article for the group.

## **5. CLI Functionality & Commands**

The application will be operated via a series of commands.

### **`ai-news fetch`**

  * **Purpose:** Fetches and processes new content.
  * **Behavior:**
    1.  Scans all pre-defined sources.
    2.  Checks the database to see which articles are new.
    3.  For each new article, it sends the content to the AI API for processing (summarization, extraction, etc.).
    4.  It performs deduplication analysis against recent articles.
    5.  Saves the structured data for new items into the SQLite database with an `unread` status.
    6.  This command should provide simple output indicating its progress and how many new items were added.

### **`ai-news view`**

  * **Purpose:** Displays the processed news from the local database.

  * **Behavior:**

    1.  By default, queries the database for all items marked as `unread`.
    2.  Groups items by `StoryGroupID`.
    3.  Sorts the final list of story groups primarily by **Source Priority Tier** (Tier 1 first), and secondarily by `PublishedDate` (newest first).
    4.  Displays each story group in a numbered, detailed "card" format.
        ```
        [1] Google AI Announces Project Astra
            - Source: Google AI Blog (Tier 1)
            - Published: 2025-08-19
            - Summary:
              - A new multimodal AI assistant was announced...
              - It can reason about video and audio in real-time...
            - Also covered by: The Verge, Ars Technica
        ```
    5.  After the command successfully runs, it marks the displayed items as `read` in the database.

  * **Arguments & Flags:**

      * `--all`: Displays all items from the database, both `read` and `unread`.
      * `--source "[Source Name]"`: Filters results for a specific source.
      * `--topic "[Topic Name]"`: Filters results for a specific topic.
      * `--type "[Content Type]"`: Filters results for a specific content type.

### **`ai-news read <number>`**

  * **Purpose:** Displays the full, clean content of an article directly in the terminal.
  * **Behavior:**
    1.  Takes a number corresponding to an item from the last `view` output.
    2.  Retrieves the `URL` of the primary article for that story group.
    3.  Passes the `URL` to a service like **Jina Reader** (`r.jina.ai/[URL]`) to get the full article content as markdown.
    4.  Pipes the resulting markdown to a renderer like **Glow** for formatted display in the terminal.

### **`ai-news open <number>`**

  * **Purpose:** Opens the original article in a web browser.
  * **Behavior:**
    1.  Takes a number corresponding to an item from the last `view` output.
    2.  Retrieves the `URL` of the primary article for that story group.
    3.  Opens the `URL` in the system's default web browser.

## **6. Example User Workflow**

1.  User runs `ai-news fetch` to update the local database with the latest news.
2.  User runs `ai-news view` to see a prioritized, summarized list of only what's new.
3.  User sees an interesting story at `[3]`. They want to read it without leaving the terminal.
4.  User runs `ai-news read 3` to see the full article, beautifully rendered.
5.  User runs `ai-news view --topic "Robotics"` to see if there's anything new specifically on that topic.
6.  User runs `ai-news open 1` on an item from that list to watch a video or see complex diagrams in their browser.

## **7. Future Considerations**

  * **Automation:** The separation of `fetch` and `view` allows the `fetch` command to be run on a `cron` schedule for fully automated data collection.
  * **Integrations:** The structured data in the SQLite database can be easily read by other services, such as a bot that posts daily digests to Discord or Slack.