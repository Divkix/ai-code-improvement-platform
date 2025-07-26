# Test Coverage Improvement Plan

This document outlines a comprehensive, step-by-step plan to achieve at least 80% test coverage for both the backend and frontend of the application. The strategy emphasizes the use of mock data and dependencies to ensure that tests are fast, reliable, and focused on the logic of the code under test.

## 1. Backend Test Coverage Strategy

The backend has several critical areas with low or no test coverage. The following is a plan to address these areas, starting with the most critical packages.

### 1.1. `internal/services` (Current Coverage: 1.2%)

This package contains the core business logic and is the highest priority for testing.

**Strategy:**

*   **Mock Dependencies:** All external dependencies, such as the database (`mongodb` and `qdrant`), external APIs (`GitHub`, `Voyage AI`, `OpenAI`), and other services, will be mocked. This will allow us to test the business logic of each service in isolation.
*   **Table-Driven Tests:** We will use table-driven tests to cover a wide range of inputs, outputs, and error conditions for each service method.
*   **GoMock:** We will use the `gomock` library to create mock implementations of the service dependencies.

**Step-by-Step Plan:**

1.  **`user.go`:**
    *   Write tests for `NewUserService`.
    *   Write tests for `CreateUser`, mocking the database dependency to simulate user creation success and failure (e.g., user already exists).
    *   Write tests for `GetUserByEmail`, mocking the database to simulate finding a user and not finding a user.
2.  **`repository.go`:**
    *   Write tests for `NewRepositoryService`.
    *   Write tests for all the methods, mocking the database to test repository creation, retrieval, and updates.
3.  **`github.go`:**
    *   Write tests for `NewGitHubService`.
    *   Mock the GitHub API client to test the logic for fetching repositories and user data from GitHub.
4.  **`embedding.go` and `embedding_pipeline.go`:**
    *   Write tests for the embedding services, mocking the embedding provider (`Voyage AI` or local) and the database.
5.  **`code_processor.go`:**
    *   Write tests for the code processing logic, mocking the file system and the database.
6.  **`search.go` and `chat_rag.go`:**
    *   Write tests for the search and RAG services, mocking the database and the LLM.
7.  **`dashboard.go`:**
    *   Write tests for the dashboard service, mocking the database.

### 1.2. `internal/handlers` (Current Coverage: 0.0%)

This package is responsible for handling API requests and is another high-priority area.

**Strategy:**

*   **Mock Services:** The services that the handlers depend on will be mocked.
*   **`httptest`:** We will use the `net/http/httptest` package to create mock HTTP requests and record the responses.
*   **Test All Endpoints:** We will write tests for all the API endpoints, covering different HTTP methods, request parameters, and request bodies.

**Step-by-Step Plan:**

1.  **`auth.go`:**
    *   Write tests for the `Register` and `Login` handlers, mocking the `UserService` and `AuthService`.
2.  **`repository.go`:**
    *   Write tests for all the repository handlers, mocking the `RepositoryService`.
3.  **`github.go`:**
    *   Write tests for the GitHub handlers, mocking the `GitHubService`.
4.  **`chat.go`:**
    *   Write tests for the chat handlers, mocking the `ChatRAGService`.
5.  **`search.go` and `vector_search.go`:**
    *   Write tests for the search handlers, mocking the `SearchService`.
6.  **`dashboard.go`:**
    *   Write tests for the dashboard handlers, mocking the `DashboardService`.
7.  **`health.go`:**
    *   Write tests for the health check handler.

### 1.3. `internal/database` (Current Coverage: 0.0%)

This package is responsible for database interactions.

**Strategy:**

*   **Integration Tests:** We will write integration tests that interact with a real database (e.g., a test instance of MongoDB and Qdrant).
*   **Test Database:** We will use a separate test database to avoid interfering with the development database.

**Step-by-Step Plan:**

1.  **`mongodb.go`:**
    *   Write integration tests for all the MongoDB-related functions.
2.  **`qdrant.go`:**
    *   Write integration tests for all the Qdrant-related functions.

### 1.4. `internal/models` (Current Coverage: 28.4%)

This package contains the data models.

**Strategy:**

*   **Unit Tests:** We will write unit tests for the model validation logic and any methods on the models.

**Step-by-Step Plan:**

1.  **`user.go`:**
    *   Add more tests for user model validation.
2.  **`repository.go`:**
    *   Add more tests for repository model validation.
3.  **`chat_session.go`:**
    *   Add more tests for chat session model validation.
4.  **`codechunk.go`:**
    *   Add tests for code chunk model validation.

## 2. Frontend Test Coverage Strategy

The frontend tests are currently failing. The first step is to fix the failing test and then proceed with a plan to increase coverage.

### 2.1. Fix Failing Tests

**Strategy:**

*   **Isolate the Issue:** The failing test in `src/routes/page.svelte.test.ts` is due to an error in fetching dashboard data. We will mock the API call to the backend to resolve this.

**Step-by-Step Plan:**

1.  **Mock API:** Use a library like `msw` (Mock Service Worker) to intercept the API request to the backend and return mock dashboard data.
2.  **Fix Test:** Update the test to use the mocked API and assert that the `h1` heading is rendered correctly.

### 2.2. Component Testing

**Strategy:**

*   **`@testing-library/svelte`:** We will use this library to write tests for the Svelte components.
*   **Test Component Logic:** We will focus on testing the logic of each component, such as event handlers, conditional rendering, and data display.

**Step-by-Step Plan:**

1.  **`routes/`:**
    *   Write tests for all the page components in the `routes` directory.
2.  **`lib/components/`:**
    *   Write tests for all the reusable components in the `lib/components` directory.

### 2.3. Store and API Testing

**Strategy:**

*   **Unit Tests:** We will write unit tests for the Svelte stores and the API client.

**Step-by-Step Plan:**

1.  **`lib/stores/`:**
    *   Write tests for the auth store and any other stores.
2.  **`lib/api.ts`:**
    *   Write tests for the API client, mocking the `fetch` function to test the API calls in isolation.

## 3. Summary of the Plan

By following this step-by-step plan, we will systematically increase the test coverage of both the backend and frontend to at least 80%. The use of mocking will ensure that the tests are fast, reliable, and maintainable.
