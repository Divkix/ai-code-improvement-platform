# GitHub Repository Analyzer - Technical Specification

## Project Overview

### Vision
A cost-effective, AI-powered code analysis platform that helps development teams onboard faster, maintain code quality, and modernize legacy codebases through intelligent conversation with their repositories.

### Target Audience
Enterprise development teams looking to reduce costs while integrating AI capabilities into their development workflow.

### Core Value Propositions
1. **Cost Efficiency**: Self-hosted solution with API-based pricing vs GitHub Copilot's per-seat model
2. **Onboarding Acceleration**: New developers can understand codebases through natural conversation
3. **Code Quality Insights**: Proactive identification of optimization opportunities and code issues
4. **Easy Integration**: Simple GitHub App installation with minimal configuration

### MVP Scope (3 Core Features)
1. **Strategic Dashboard**: Immediate visual impact showing analyzed code metrics and cost savings
2. **Repository Import**: Seamless GitHub integration for repository analysis
3. **AI Chat Interface**: Natural language queries about code with context-aware responses

## Technical Architecture

### Technology Stack
- **Frontend**: SvelteKit (using Bun as runtime)
- **Backend**: Go 1.21+
- **Database**: MongoDB 7.0
- **Vector Database**: Qdrant 1.7+
- **Embedding Model**: Voyage AI (voyage-code-3)
- **LLM**: Claude 3.5 Sonnet via Anthropic API
- **Containerization**: Docker Compose
- **Version Control Integration**: GitHub API v3/GraphQL

### System Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  SvelteKit UI   │────▶│   Go Backend    │────▶│    MongoDB      │
│   (Port 3000)   │     │   (Port 8080)   │     │   (Port 27017)  │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    ▼            ▼            ▼
            ┌──────────┐ ┌──────────┐ ┌──────────┐
            │ GitHub   │ │ Qdrant   │ │ Voyage/  │
            │   API    │ │(Port     │ │ Claude   │
            │          │ │ 6333)    │ │   APIs   │
            └──────────┘ └──────────┘ └──────────┘
```

### Docker Compose Structure

```yaml
version: '3.8'

services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - VITE_API_URL=http://localhost:8080
    depends_on:
      - backend

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=mongodb://mongodb:27017/github-analyzer
      - QDRANT_URL=http://qdrant:6333
      - JWT_SECRET=${JWT_SECRET}
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - VOYAGE_API_KEY=${VOYAGE_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    depends_on:
      - mongodb
      - qdrant

  mongodb:
    image: mongo:7.0
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

  qdrant:
    image: qdrant/qdrant:v1.7.4
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage

volumes:
  mongodb_data:
  qdrant_data:
```

## Database Schemas

### MongoDB Collections

#### users
```javascript
{
  _id: ObjectId,
  email: string,
  password: string, // bcrypt hashed
  name: string,
  githubAccessToken: string, // encrypted
  createdAt: Date,
  updatedAt: Date
}
```

#### repositories
```javascript
{
  _id: ObjectId,
  userId: ObjectId,
  githubRepoId: number,
  owner: string,
  name: string,
  fullName: string, // "owner/name"
  description: string,
  primaryLanguage: string,
  isPrivate: boolean,
  indexedAt: Date,
  lastSyncedAt: Date,
  stats: {
    totalFiles: number,
    totalLines: number,
    languages: Map<string, number>, // {"Go": 15000, "JavaScript": 3000}
    lastCommitDate: Date
  },
  status: string, // "importing", "ready", "error"
  importProgress: number, // 0-100
  createdAt: Date,
  updatedAt: Date
}
```

#### code_chunks
```javascript
{
  _id: ObjectId,
  repositoryId: ObjectId,
  filePath: string,
  fileName: string,
  language: string,
  startLine: number,
  endLine: number,
  content: string,
  contentHash: string, // SHA256 of content for deduplication
  imports: [string], // extracted import statements
  metadata: {
    functions: [string], // function names in chunk
    classes: [string], // class names in chunk
    complexity: number // cyclomatic complexity estimate
  },
  vectorId: string, // reference to Qdrant point ID
  createdAt: Date
}
```

#### chat_sessions
```javascript
{
  _id: ObjectId,
  userId: ObjectId,
  repositoryId: ObjectId,
  title: string, // auto-generated from first question
  messages: [{
    role: string, // "user" or "assistant"
    content: string,
    timestamp: Date,
    retrievedChunks: [ObjectId], // code_chunks references for context
    tokensUsed: number
  }],
  createdAt: Date,
  updatedAt: Date
}
```

#### analytics_events
```javascript
{
  _id: ObjectId,
  userId: ObjectId,
  repositoryId: ObjectId,
  eventType: string, // "optimization_found", "query_asked", "repository_imported"
  eventData: Object, // flexible schema for different event types
  severity: string, // "info", "warning", "critical"
  createdAt: Date
}
```

### Qdrant Collections

#### code_embeddings
```javascript
{
  collection_name: "code_embeddings",
  vectors: {
    size: 1024, // voyage-code-3 embedding dimension
    distance: "Cosine"
  },
  payload_schema: {
    chunk_id: string, // MongoDB ObjectId as string
    repository_id: string,
    file_path: string,
    language: string,
    start_line: number,
    end_line: number
  }
}
```

## API Endpoints

### Authentication
```
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
GET    /api/auth/me
```

### Repositories
```
GET    /api/repositories              # List user's repositories
POST   /api/repositories/import       # Import a new repository
GET    /api/repositories/:id          # Get repository details
DELETE /api/repositories/:id          # Remove repository
GET    /api/repositories/:id/stats    # Get repository statistics
```

### Chat
```
GET    /api/chat/sessions             # List chat sessions
POST   /api/chat/sessions             # Create new session
GET    /api/chat/sessions/:id         # Get session with messages
POST   /api/chat/sessions/:id/message # Send message to assistant
DELETE /api/chat/sessions/:id         # Delete session
```

### Dashboard
```
GET    /api/dashboard/stats           # Get aggregated statistics
GET    /api/dashboard/activity        # Get recent activity
GET    /api/dashboard/trends          # Get trend data (hardcoded for MVP)
```

## Feature Specifications

### 1. Dashboard with Strategic Metrics

#### Visual Layout
The dashboard presents four key components arranged in a grid layout:

```
┌─────────────────────────────────────────────────────────┐
│                    Hero Metrics Bar                      │
│  Total Lines: 1.2M  |  Active Repos: 5  |  Savings: $8k │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────┬───────────────────────────┐
│                             │                           │
│    Code Quality Trend       │     Recent Activity       │
│    (Line Chart)             │     (List View)           │
│                             │                           │
└─────────────────────────────┴───────────────────────────┘
```

#### Implementation Details

**Hero Metrics Calculation**:
```go
type DashboardStats struct {
    TotalLinesAnalyzed int64  `json:"totalLinesAnalyzed"`
    ActiveRepositories int    `json:"activeRepositories"`
    MonthlySavings     string `json:"monthlySavings"`
    IssuesFoundWeek    int    `json:"issuesFoundWeek"`
}

// Calculate savings based on:
// - Number of repositories analyzed
// - Estimated developer count (5 per repository)
// - GitHub Copilot pricing ($19/user/month)
// - Our estimated cost (API usage ~$50/month total)
monthlySavings = (repoCount * 5 * 19) - 50
```

**Trend Data Structure** (Hardcoded for MVP):
```javascript
const trendData = {
  labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
  datasets: [{
    label: 'Code Quality Score',
    data: [72, 75, 78, 83], // Shows improvement
    borderColor: '#10b981',
    tension: 0.4
  }]
}
```

**Recent Activity Items**:
```javascript
const recentActivity = [
  {
    id: 1,
    type: 'optimization',
    severity: 'medium',
    message: 'Found 3 redundant database queries in UserService',
    repository: 'backend-api',
    timestamp: '2 hours ago',
    isRead: false
  },
  // ... more items
]
```

### 2. Repository Import Flow

#### User Journey
1. User clicks "Import Repository" button
2. Modal appears with GitHub URL input
3. System validates access and repository existence
4. Progress indicator shows import stages
5. Repository appears in list with "Ready" status

#### Technical Implementation

**Import Process**:
```go
func ImportRepository(userId, repoURL string) error {
    // 1. Parse repository URL
    owner, name := parseGitHubURL(repoURL)
    
    // 2. Verify user has access
    hasAccess := checkGitHubAccess(userId, owner, name)
    if !hasAccess {
        return ErrNoAccess
    }
    
    // 3. Create repository record
    repo := createRepositoryRecord(userId, owner, name)
    
    // 4. Start import job
    go importRepositoryAsync(repo.ID)
    
    return nil
}

func importRepositoryAsync(repoId string) {
    // Update progress: 10%
    updateImportProgress(repoId, 10, "Fetching repository structure...")
    
    // Clone repository content via API
    files := fetchRepositoryFiles(repoId)
    updateImportProgress(repoId, 30, "Analyzing code files...")
    
    // Process files in batches
    for i, batch := range chunkFiles(files, 10) {
        processBatch(batch, repoId)
        progress := 30 + (i * 50 / len(batches))
        updateImportProgress(repoId, progress, fmt.Sprintf("Processing files... %d%%", progress))
    }
    
    // Generate embeddings
    updateImportProgress(repoId, 80, "Building semantic index...")
    generateEmbeddings(repoId)
    
    // Finalize
    updateImportProgress(repoId, 100, "Ready for analysis!")
    updateRepositoryStatus(repoId, "ready")
}
```

**Progress Updates via WebSocket** (Simplified for MVP - use polling):
```javascript
// Frontend polls every 2 seconds during import
async function checkImportProgress(repoId) {
  const response = await fetch(`/api/repositories/${repoId}`);
  const data = await response.json();
  
  if (data.status === 'importing') {
    updateProgressBar(data.importProgress);
    updateStatusMessage(data.importMessage);
    setTimeout(() => checkImportProgress(repoId), 2000);
  } else if (data.status === 'ready') {
    showSuccessMessage();
    redirectToRepository(repoId);
  }
}
```

### 3. Chat Interface with RAG Pipeline

#### RAG Pipeline Architecture

The RAG (Retrieval-Augmented Generation) pipeline is the heart of your AI capabilities. Here's how it works in simple terms:

**Step 1: Code Chunking and Indexing**
When a repository is imported, we break down each file into overlapping chunks:

```go
func chunkFile(filePath string, content string) []CodeChunk {
    lines := strings.Split(content, "\n")
    chunks := []CodeChunk{}
    chunkSize := 150
    overlap := 50
    
    for i := 0; i < len(lines); i += (chunkSize - overlap) {
        end := min(i + chunkSize, len(lines))
        chunk := CodeChunk{
            FilePath: filePath,
            StartLine: i + 1,
            EndLine: end,
            Content: strings.Join(lines[i:end], "\n"),
        }
        chunks = append(chunks, chunk)
    }
    return chunks
}
```

**Step 2: Embedding Generation**
Each chunk is converted to a vector representation using Voyage AI:

```go
func generateEmbedding(text string) ([]float32, error) {
    payload := map[string]interface{}{
        "input": text,
        "model": "voyage-code-3",
    }
    
    resp, err := http.Post("https://api.voyageai.com/v1/embeddings", 
                           "application/json", 
                           jsonPayload(payload))
    // Parse response and extract embedding vector
    return embedding, nil
}
```

**Step 3: Vector Storage in Qdrant**
```go
func storeEmbedding(chunkId string, embedding []float32, metadata map[string]interface{}) error {
    point := qdrant.Point{
        ID: chunkId,
        Vector: embedding,
        Payload: metadata,
    }
    return qdrantClient.Upsert("code_embeddings", []qdrant.Point{point})
}
```

**Step 4: Query Processing**
When a user asks a question:

```go
func processQuery(query string, repositoryId string) (string, error) {
    // 1. Generate embedding for the query
    queryEmbedding := generateEmbedding(query)
    
    // 2. Search for similar chunks
    searchResults := qdrantClient.Search(qdrant.SearchRequest{
        Collection: "code_embeddings",
        Vector: queryEmbedding,
        Filter: qdrant.Filter{
            Must: []qdrant.Condition{{
                Key: "repository_id",
                Match: qdrant.Match{Value: repositoryId},
            }},
        },
        Limit: 8,
        WithPayload: true,
    })
    
    // 3. Retrieve full chunk content from MongoDB
    chunks := retrieveChunks(searchResults)
    
    // 4. Construct prompt for Claude
    prompt := constructPrompt(query, chunks)
    
    // 5. Get response from Claude
    response := callClaude(prompt)
    
    return response, nil
}
```

**Step 5: Prompt Construction**
This is crucial for getting good responses:

```go
func constructPrompt(query string, chunks []CodeChunk) string {
    prompt := `You are a helpful code analysis assistant. You have access to code from a repository.
    
Based on the following code segments, please answer the user's question. Be specific and reference actual code when explaining.

Code segments:
`
    for i, chunk := range chunks {
        prompt += fmt.Sprintf("\n--- File: %s (lines %d-%d) ---\n%s\n",
                              chunk.FilePath, chunk.StartLine, chunk.EndLine, chunk.Content)
    }
    
    prompt += fmt.Sprintf("\nUser question: %s\n\nPlease provide a clear, helpful answer:", query)
    return prompt
}
```

#### Chat UI Implementation

**Frontend Component Structure**:
```svelte
<!-- ChatInterface.svelte -->
<script>
  export let repositoryId;
  
  let messages = [];
  let inputText = '';
  let isLoading = false;
  let analyzingFiles = [];
  
  const suggestedQuestions = [
    "Explain the authentication flow",
    "What does the main API handler do?",
    "Find potential improvements in the database queries"
  ];
  
  async function sendMessage() {
    if (!inputText.trim()) return;
    
    // Add user message
    messages = [...messages, { role: 'user', content: inputText }];
    const query = inputText;
    inputText = '';
    isLoading = true;
    
    try {
      // Send to backend
      const response = await fetch(`/api/chat/sessions/${sessionId}/message`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: query })
      });
      
      const data = await response.json();
      
      // Show analyzing files
      analyzingFiles = data.analyzingFiles || [];
      
      // Simulate progressive file analysis display
      for (let file of analyzingFiles) {
        await sleep(500); // Show each file for 500ms
      }
      
      // Add assistant response
      messages = [...messages, { 
        role: 'assistant', 
        content: data.response,
        chunks: data.retrievedChunks 
      }];
      
    } finally {
      isLoading = false;
      analyzingFiles = [];
    }
  }
</script>

<div class="chat-container">
  <div class="chat-header">
    <RepositorySelector bind:repositoryId />
  </div>
  
  <div class="messages">
    {#each messages as message}
      <Message {message} />
    {/each}
    
    {#if isLoading}
      <div class="analyzing-indicator">
        {#if analyzingFiles.length > 0}
          <p>Analyzing: {analyzingFiles[analyzingFiles.length - 1]}</p>
        {:else}
          <p>Thinking...</p>
        {/if}
      </div>
    {/if}
  </div>
  
  {#if messages.length === 0}
    <div class="suggested-questions">
      <p>Try asking:</p>
      {#each suggestedQuestions as question}
        <button on:click={() => { inputText = question; sendMessage(); }}>
          {question}
        </button>
      {/each}
    </div>
  {/if}
  
  <div class="input-area">
    <input 
      bind:value={inputText}
      on:keypress={e => e.key === 'Enter' && sendMessage()}
      placeholder="Ask about the code..."
    />
    <button on:click={sendMessage} disabled={isLoading}>
      Send
    </button>
  </div>
</div>
```

**Message Formatting with Syntax Highlighting**:
```javascript
// Use Prism.js or similar for syntax highlighting
function formatCodeBlocks(content) {
  return content.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
    const highlighted = Prism.highlight(code, Prism.languages[lang] || Prism.languages.plaintext, lang);
    return `<pre><code class="language-${lang}">${highlighted}</code></pre>`;
  });
}
```

## Implementation Guidelines

### Phase 1: Foundation (Days 1-5)
1. Set up Docker Compose environment
2. Create basic Go project structure with routing
3. Implement JWT authentication
4. Set up MongoDB schemas and connections
5. Create SvelteKit project with basic routing
6. Implement login/register UI

### Phase 2: GitHub Integration (Days 6-10)
1. Implement GitHub OAuth App authentication
2. Build repository import API endpoint
3. Create file fetching logic using GitHub API
4. Implement code chunking algorithm
5. Set up Qdrant and create collections
6. Build repository import UI with progress tracking

### Phase 3: RAG Pipeline (Days 11-15)
1. Integrate Voyage AI API for embeddings
2. Implement chunk storage in MongoDB
3. Build vector storage logic for Qdrant
4. Create search functionality
5. Integrate Claude API
6. Build chat UI with message formatting

### Phase 4: Dashboard & Polish (Days 16-18)
1. Create dashboard API endpoints
2. Build dashboard UI with charts
3. Add loading states and error handling
4. Implement suggested questions
5. Add syntax highlighting for code
6. Create demo data and scenarios

### Phase 5: Testing & Demo Prep (Days 19-21)
1. End-to-end testing of all features
2. Prepare demo repositories
3. Create demo script
4. Test with different code questions
5. Performance optimization
6. Bug fixes and polish

## Key Implementation Tips

### For RAG Pipeline (Since It's New to You)

1. **Start Simple**: Begin with basic text matching before adding semantic search. Get the pipeline working end-to-end first.

2. **Test with Small Data**: Use a small repository initially to test your chunking and embedding logic without waiting long or using много API credits.

3. **Log Everything**: Add detailed logging to understand what chunks are being retrieved and why:
```go
log.Printf("Query: %s, Retrieved %d chunks, Top match: %.3f similarity", 
           query, len(chunks), topScore)
```

4. **Prompt Engineering**: The quality of Claude's responses depends heavily on your prompt. Test different prompt structures:
```go
// Version 1: Simple
prompt := fmt.Sprintf("Code: %s\nQuestion: %s", code, question)

// Version 2: Detailed (Better)
prompt := fmt.Sprintf(`You are analyzing code from the %s repository.
                      
Context: The user is trying to understand how this code works.
                      
Code segments:
%s

User question: %s

Please provide a clear explanation that references specific parts of the code.`, 
repoName, code, question)
```

5. **Handle Edge Cases**:
   - What if no relevant chunks are found?
   - What if the question is too vague?
   - What if the repository has no code files?

### For Performance

1. **Batch Operations**: When importing a repository, process files in batches:
```go
const batchSize = 10
for i := 0; i < len(files); i += batchSize {
    batch := files[i:min(i+batchSize, len(files))]
    processBatch(batch)
}
```

2. **Concurrent Processing**: Use goroutines for parallel operations:
```go
var wg sync.WaitGroup
semaphore := make(chan struct{}, 5) // Limit to 5 concurrent operations

for _, file := range files {
    wg.Add(1)
    semaphore <- struct{}{}
    go func(f File) {
        defer wg.Done()
        defer func() { <-semaphore }()
        processFile(f)
    }(file)
}
wg.Wait()
```

3. **Caching**: Cache embeddings for unchanged files:
```go
contentHash := sha256.Sum256([]byte(fileContent))
if existingChunk := findByHash(contentHash); existingChunk != nil {
    // Reuse existing embedding
    return existingChunk.VectorId
}
```

## Demo Scenarios

### Scenario 1: First Impression
1. Login with pre-created account
2. Show dashboard with impressive metrics
3. Highlight cost savings calculation
4. Point out the trending improvement in code quality

### Scenario 2: Repository Import
1. Click "Import Repository"
2. Paste a prepared GitHub URL
3. Show the progress animation
4. Once complete, navigate to the repository

### Scenario 3: Intelligent Conversation
1. Select the imported repository
2. Ask: "How does authentication work in this codebase?"
3. Show the AI analyzing specific files
4. Highlight how the response references actual code
5. Ask a follow-up: "What security improvements would you suggest?"
6. Show the AI providing specific, actionable suggestions

### Scenario 4: Value Demonstration
1. Return to dashboard
2. Show accumulated metrics
3. Calculate ROI: "With 5 repositories and 25 developers, this saves $475/month"
4. Mention scalability: "This prototype analyzes samples, but the architecture supports full codebase analysis"

## Environment Variables

Create a `.env` file with:
```bash
# JWT
JWT_SECRET=your-secret-key-here

# GitHub OAuth App
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Voyage AI
VOYAGE_API_KEY=your-voyage-api-key

# Anthropic
ANTHROPIC_API_KEY=your-anthropic-api-key

# MongoDB (if not using default)
MONGODB_URI=mongodb://localhost:27017/github-analyzer

# Qdrant (if not using default)
QDRANT_URL=http://localhost:6333
```

## Testing Checklist

- [ ] User can register and login
- [ ] Dashboard shows realistic metrics
- [ ] Repository import completes successfully
- [ ] Progress indicator updates during import
- [ ] Chat interface loads imported repository
- [ ] Questions receive relevant, code-aware responses
- [ ] Code snippets are properly formatted
- [ ] Suggested questions work correctly
- [ ] Error states are handled gracefully
- [ ] The application works in Docker Compose

## Future Enhancements (To Mention in Demo)

1. **Real-time Analysis**: Webhook integration for automatic code review on push
2. **Team Collaboration**: Shared chat sessions and knowledge base
3. **Advanced Analytics**: Dependency analysis, security scanning, complexity metrics
4. **IDE Integration**: VS Code extension for in-editor chat
5. **Custom Rules**: Company-specific coding standards and patterns
6. **SSO Integration**: Enterprise authentication support
7. **Comprehensive Analysis**: Full codebase scanning with job queues
8. **Automated Fixes**: Generate pull requests with improvements

## Success Metrics for Demo

1. **Technical Execution**: All features work smoothly without errors
2. **Performance**: Responses appear within 3-5 seconds
3. **User Experience**: Interface is intuitive and professional
4. **Value Communication**: Cost savings and benefits are clear
5. **Vision**: Future possibilities are articulated well

This specification provides everything you need to build an impressive MVP that demonstrates both immediate value and future potential. Remember, the goal is not perfection but rather showing your ability to identify a problem, design a solution, and execute effectively within constraints.