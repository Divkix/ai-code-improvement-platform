# Phase 4: Real-Time Updates & Incremental Analysis (Months 10-12)

## Overview
**Goal**: Maintain understanding as code evolves without full reprocessing

Phase 4 transforms our platform from batch processing to real-time understanding. Instead of reanalyzing entire repositories on every change, we implement incremental analysis that tracks changes, propagates impacts, and maintains synchronized understanding across all analysis layers (AST, semantic, multi-modal).

## Why Incremental Analysis is Critical

### Current Batch Processing Limitations
- **Full Repository Reprocessing**: Any code change triggers complete re-analysis
- **High Latency**: Large codebases take hours to reprocess
- **Resource Waste**: 99% of code remains unchanged but gets reanalyzed
- **Expensive Operations**: Re-embedding, re-clustering, and re-graphing unchanged code
- **Poor Developer Experience**: Long waits for updated understanding

### Research Foundation
- **GitHub CodeQL**: 20% performance improvement with incremental analysis
- **Incremental Program Analysis**: 80-90% reduction in analysis time for typical changes
- **Change Impact Analysis**: Critical for maintaining code understanding at scale

## Phase 4 Components

### 4.1 Incremental Analysis Engine

#### Change Detection and Classification

```go
// internal/services/incremental_analyzer.go
type IncrementalAnalyzer struct {
    changeDetector   *ChangeDetector
    impactAnalyzer   *ImpactAnalyzer
    graphUpdater     *GraphUpdater
    embeddingCache   *EmbeddingCache
    astAnalyzer      *ASTAnalyzer
    multiModalAnalyzer *MultiModalAnalyzer
    pdgAnalyzer      *PDGAnalyzer
}

type CodeChange struct {
    ID            string              `bson:"id" json:"id"`
    RepositoryID  primitive.ObjectID  `bson:"repositoryId" json:"repositoryId"`
    CommitHash    string              `bson:"commitHash" json:"commitHash"`
    Type          ChangeType          `bson:"type" json:"type"`
    FilePath      string              `bson:"filePath" json:"filePath"`
    StartLine     int                 `bson:"startLine" json:"startLine"`
    EndLine       int                 `bson:"endLine" json:"endLine"`
    OldContent    string              `bson:"oldContent" json:"oldContent"`
    NewContent    string              `bson:"newContent" json:"newContent"`
    AffectedAST   []string            `bson:"affectedAst" json:"affectedAst"` // AST node IDs
    ChangeScope   ChangeScope         `bson:"changeScope" json:"changeScope"`
    Timestamp     time.Time           `bson:"timestamp" json:"timestamp"`
    Author        string              `bson:"author" json:"author"`
    Message       string              `bson:"message" json:"message"`
    Metadata      ChangeMetadata      `bson:"metadata" json:"metadata"`
}

type ChangeType string
const (
    ChangeTypeAdd    ChangeType = "add"
    ChangeTypeModify ChangeType = "modify"
    ChangeTypeDelete ChangeType = "delete"
    ChangeTypeMove   ChangeType = "move"
    ChangeTypeRename ChangeType = "rename"
)

type ChangeScope string
const (
    ScopeLocal    ChangeScope = "local"      // affects single function/method
    ScopeFile     ChangeScope = "file"       // affects multiple functions in file
    ScopeModule   ChangeScope = "module"     // affects multiple files in module
    ScopeGlobal   ChangeScope = "global"     // affects multiple modules
)

type ChangeMetadata struct {
    FunctionsAffected []string `bson:"functionsAffected" json:"functionsAffected"`
    ClassesAffected   []string `bson:"classesAffected" json:"classesAffected"`
    ImportsChanged    []string `bson:"importsChanged" json:"importsChanged"`
    APIChanges        []APIChange `bson:"apiChanges" json:"apiChanges"`
    TestsAffected     []string `bson:"testsAffected" json:"testsAffected"`
    DocsAffected      []string `bson:"docsAffected" json:"docsAffected"`
}

type APIChange struct {
    Type        string `bson:"type" json:"type"` // signature, visibility, behavior
    Entity      string `bson:"entity" json:"entity"` // function/method/class name
    OldSignature string `bson:"oldSignature" json:"oldSignature"`
    NewSignature string `bson:"newSignature" json:"newSignature"`
    Breaking    bool   `bson:"breaking" json:"breaking"`
}
```

#### Change Detection Implementation

```go
// internal/services/change_detector.go
type ChangeDetector struct {
    gitService    *GitService
    astService    *ASTAnalyzer
    diffAnalyzer  *DiffAnalyzer
}

func (c *ChangeDetector) DetectChanges(repositoryID primitive.ObjectID, fromCommit, toCommit string) ([]CodeChange, error) {
    // 1. Get git diff between commits
    gitDiff, err := c.gitService.GetDiff(fromCommit, toCommit)
    if err != nil {
        return nil, fmt.Errorf("failed to get git diff: %w", err)
    }
    
    changes := make([]CodeChange, 0)
    
    // 2. Process each changed file
    for _, fileDiff := range gitDiff.Files {
        fileChanges, err := c.analyzeFileChanges(repositoryID, fileDiff, toCommit)
        if err != nil {
            log.Printf("Error analyzing file changes for %s: %v", fileDiff.Path, err)
            continue
        }
        changes = append(changes, fileChanges...)
    }
    
    // 3. Classify change scope and impact
    for i := range changes {
        c.classifyChangeScope(&changes[i])
        c.extractChangeMetadata(&changes[i])
    }
    
    return changes, nil
}

func (c *ChangeDetector) analyzeFileChanges(repositoryID primitive.ObjectID, fileDiff *GitFileDiff, commitHash string) ([]CodeChange, error) {
    changes := make([]CodeChange, 0)
    
    // Handle different change types
    switch fileDiff.Status {
    case "added":
        change := CodeChange{
            ID:           generateChangeID(),
            RepositoryID: repositoryID,
            CommitHash:   commitHash,
            Type:         ChangeTypeAdd,
            FilePath:     fileDiff.Path,
            NewContent:   fileDiff.NewContent,
            ChangeScope:  c.determineScope(fileDiff.NewContent),
            Timestamp:    time.Now(),
        }
        changes = append(changes, change)
        
    case "deleted":
        change := CodeChange{
            ID:           generateChangeID(),
            RepositoryID: repositoryID,
            CommitHash:   commitHash,
            Type:         ChangeTypeDelete,
            FilePath:     fileDiff.Path,
            OldContent:   fileDiff.OldContent,
            ChangeScope:  c.determineScope(fileDiff.OldContent),
            Timestamp:    time.Now(),
        }
        changes = append(changes, change)
        
    case "modified":
        // Parse line-by-line changes for modified files
        lineChanges := c.parseLineChanges(fileDiff)
        for _, lineChange := range lineChanges {
            change := CodeChange{
                ID:           generateChangeID(),
                RepositoryID: repositoryID,
                CommitHash:   commitHash,
                Type:         ChangeTypeModify,
                FilePath:     fileDiff.Path,
                StartLine:    lineChange.StartLine,
                EndLine:      lineChange.EndLine,
                OldContent:   lineChange.OldContent,
                NewContent:   lineChange.NewContent,
                Timestamp:    time.Now(),
            }
            changes = append(changes, change)
        }
    }
    
    return changes, nil
}

func (c *ChangeDetector) classifyChangeScope(change *CodeChange) {
    // Use AST analysis to determine scope
    if change.Type == ChangeTypeAdd || change.Type == ChangeTypeModify {
        content := change.NewContent
        if content == "" {
            content = change.OldContent
        }
        
        // Parse content to understand structural impact
        astNodes, err := c.astService.ParseContent(content, change.FilePath)
        if err != nil {
            change.ChangeScope = ScopeLocal // default fallback
            return
        }
        
        // Classify based on AST structure
        if c.affectsMultipleModules(astNodes) {
            change.ChangeScope = ScopeGlobal
        } else if c.affectsMultipleFiles(astNodes) {
            change.ChangeScope = ScopeModule
        } else if c.affectsMultipleFunctions(astNodes) {
            change.ChangeScope = ScopeFile
        } else {
            change.ChangeScope = ScopeLocal
        }
        
        // Store affected AST nodes
        change.AffectedAST = c.extractAffectedNodes(astNodes)
    }
}
```

### 4.2 Change Propagation System

#### Impact Analysis Engine

```go
// internal/services/impact_analyzer.go
type ImpactAnalyzer struct {
    knowledgeGraph    *KnowledgeGraphService
    pdgService        *PDGAnalyzer
    dependencyService *DependencyService
    embeddingService  *EmbeddingService
}

type ImpactAnalysis struct {
    ChangeID             string                `bson:"changeId" json:"changeId"`
    DirectlyAffected     []AffectedEntity      `bson:"directlyAffected" json:"directlyAffected"`
    IndirectlyAffected   []AffectedEntity      `bson:"indirectlyAffected" json:"indirectlyAffected"`
    CallGraphChanges     []CallGraphChange     `bson:"callGraphChanges" json:"callGraphChanges"`
    DataFlowChanges      []DataFlowChange      `bson:"dataFlowChanges" json:"dataFlowChanges"`
    TestsToUpdate        []TestEntity          `bson:"testsToUpdate" json:"testsToUpdate"`
    DocsToUpdate         []DocumentEntity      `bson:"docsToUpdate" json:"docsToUpdate"`
    EmbeddingsToUpdate   []EmbeddingEntity     `bson:"embeddingsToUpdate" json:"embeddingsToUpdate"`
    KnowledgeGraphUpdates []GraphUpdate       `bson:"knowledgeGraphUpdates" json:"knowledgeGraphUpdates"`
    MultiModalUpdates    []MultiModalUpdate   `bson:"multiModalUpdates" json:"multiModalUpdates"`
    PropagationDepth     int                   `bson:"propagationDepth" json:"propagationDepth"`
    EstimatedUpdateTime  time.Duration         `bson:"estimatedUpdateTime" json:"estimatedUpdateTime"`
    ImpactScore          float64               `bson:"impactScore" json:"impactScore"`
}

type AffectedEntity struct {
    Type         string  `bson:"type" json:"type"` // function, class, file, module
    ID           string  `bson:"id" json:"id"`
    Name         string  `bson:"name" json:"name"`
    FilePath     string  `bson:"filePath" json:"filePath"`
    ImpactType   string  `bson:"impactType" json:"impactType"` // signature, behavior, dependencies
    ImpactLevel  int     `bson:"impactLevel" json:"impactLevel"` // 1=direct, 2=indirect, etc.
    Confidence   float64 `bson:"confidence" json:"confidence"`
    RequiresReanalysis bool `bson:"requiresReanalysis" json:"requiresReanalysis"`
}

func (i *ImpactAnalyzer) AnalyzeImpact(change CodeChange) (*ImpactAnalysis, error) {
    analysis := &ImpactAnalysis{
        ChangeID:            change.ID,
        DirectlyAffected:    make([]AffectedEntity, 0),
        IndirectlyAffected:  make([]AffectedEntity, 0),
        CallGraphChanges:    make([]CallGraphChange, 0),
        DataFlowChanges:     make([]DataFlowChange, 0),
        TestsToUpdate:       make([]TestEntity, 0),
        DocsToUpdate:        make([]DocumentEntity, 0),
        EmbeddingsToUpdate:  make([]EmbeddingEntity, 0),
        KnowledgeGraphUpdates: make([]GraphUpdate, 0),
        MultiModalUpdates:   make([]MultiModalUpdate, 0),
    }
    
    // 1. Find directly affected entities
    directlyAffected := i.findDirectlyAffected(change)
    analysis.DirectlyAffected = directlyAffected
    
    // 2. Traverse dependency graph for indirect impact
    indirectlyAffected := i.findIndirectlyAffected(change, directlyAffected)
    analysis.IndirectlyAffected = indirectlyAffected
    
    // 3. Analyze call graph changes
    callGraphChanges := i.analyzeCallGraphImpact(change)
    analysis.CallGraphChanges = callGraphChanges
    
    // 4. Analyze data flow changes
    dataFlowChanges := i.analyzeDataFlowImpact(change)
    analysis.DataFlowChanges = dataFlowChanges
    
    // 5. Find affected tests
    testsToUpdate := i.findAffectedTests(change, directlyAffected)
    analysis.TestsToUpdate = testsToUpdate
    
    // 6. Find affected documentation
    docsToUpdate := i.findAffectedDocs(change, directlyAffected)
    analysis.DocsToUpdate = docsToUpdate
    
    // 7. Determine embeddings that need updates
    embeddingsToUpdate := i.findEmbeddingsToUpdate(directlyAffected, indirectlyAffected)
    analysis.EmbeddingsToUpdate = embeddingsToUpdate
    
    // 8. Plan knowledge graph updates
    graphUpdates := i.planKnowledgeGraphUpdates(change, directlyAffected)
    analysis.KnowledgeGraphUpdates = graphUpdates
    
    // 9. Plan multi-modal context updates
    multiModalUpdates := i.planMultiModalUpdates(change, directlyAffected)
    analysis.MultiModalUpdates = multiModalUpdates
    
    // 10. Calculate metrics
    analysis.PropagationDepth = i.calculatePropagationDepth(analysis)
    analysis.EstimatedUpdateTime = i.estimateUpdateTime(analysis)
    analysis.ImpactScore = i.calculateImpactScore(analysis)
    
    return analysis, nil
}

func (i *ImpactAnalyzer) findDirectlyAffected(change CodeChange) []AffectedEntity {
    affected := make([]AffectedEntity, 0)
    
    // Find entities directly modified by the change
    for _, astNodeID := range change.AffectedAST {
        node, err := i.knowledgeGraph.GetNode(astNodeID)
        if err != nil {
            continue
        }
        
        entity := AffectedEntity{
            Type:        node.Type,
            ID:          node.ID,
            Name:        node.Name,
            FilePath:    change.FilePath,
            ImpactType:  i.determineImpactType(change, node),
            ImpactLevel: 1, // direct impact
            Confidence:  0.9,
            RequiresReanalysis: i.requiresReanalysis(change, node),
        }
        affected = append(affected, entity)
    }
    
    return affected
}

func (i *ImpactAnalyzer) findIndirectlyAffected(change CodeChange, directlyAffected []AffectedEntity) []AffectedEntity {
    indirectlyAffected := make([]AffectedEntity, 0)
    visited := make(map[string]bool)
    
    // BFS traversal of dependency graph
    queue := make([]AffectedEntity, len(directlyAffected))
    copy(queue, directlyAffected)
    
    for _, entity := range directlyAffected {
        visited[entity.ID] = true
    }
    
    currentLevel := 2
    maxDepth := 5 // prevent infinite traversal
    
    for len(queue) > 0 && currentLevel <= maxDepth {
        levelSize := len(queue)
        nextLevel := make([]AffectedEntity, 0)
        
        for i := 0; i < levelSize; i++ {
            current := queue[i]
            
            // Find dependencies of current entity
            dependencies := i.knowledgeGraph.GetDependencies(current.ID)
            for _, dep := range dependencies {
                if visited[dep.ID] {
                    continue
                }
                
                affected := AffectedEntity{
                    Type:        dep.Type,
                    ID:          dep.ID,
                    Name:        dep.Name,
                    FilePath:    dep.FilePath,
                    ImpactType:  "dependency",
                    ImpactLevel: currentLevel,
                    Confidence:  i.calculatePropagationConfidence(current, dep, currentLevel),
                    RequiresReanalysis: i.requiresPropagatedReanalysis(dep, currentLevel),
                }
                
                if affected.Confidence > 0.3 { // threshold for inclusion
                    indirectlyAffected = append(indirectlyAffected, affected)
                    nextLevel = append(nextLevel, affected)
                    visited[dep.ID] = true
                }
            }
        }
        
        queue = nextLevel
        currentLevel++
    }
    
    return indirectlyAffected
}
```

#### Smart Caching and Update Strategy

```go
// internal/services/embedding_cache.go
type EmbeddingCache struct {
    redisClient    *redis.Client
    vectorDB       *qdrant.Client
    cachePolicy    *CachePolicy
    updateQueue    *UpdateQueue
}

type CachePolicy struct {
    TTL              time.Duration `json:"ttl"`
    MaxSize          int64         `json:"maxSize"`
    EvictionPolicy   string        `json:"evictionPolicy"` // lru, lfu, ttl
    UpdateThreshold  float64       `json:"updateThreshold"` // similarity threshold for updates
    BatchSize        int           `json:"batchSize"`
    MaxRetries       int           `json:"maxRetries"`
}

type CachedEmbedding struct {
    ChunkID      string    `json:"chunkId"`
    ContentHash  string    `json:"contentHash"` // SHA256 of content
    Embedding    []float32 `json:"embedding"`
    LastUpdated  time.Time `json:"lastUpdated"`
    AccessCount  int       `json:"accessCount"`
    Version      int       `json:"version"`
}

func (e *EmbeddingCache) GetOrComputeEmbedding(chunkID, content string) ([]float32, error) {
    contentHash := e.calculateContentHash(content)
    
    // 1. Check cache first
    cached, err := e.getCachedEmbedding(chunkID)
    if err == nil && cached.ContentHash == contentHash {
        // Cache hit - update access count
        e.updateAccessCount(chunkID)
        return cached.Embedding, nil
    }
    
    // 2. Cache miss or content changed - compute new embedding
    embedding, err := e.computeEmbedding(content)
    if err != nil {
        return nil, fmt.Errorf("failed to compute embedding: %w", err)
    }
    
    // 3. Store in cache
    cachedEmbedding := CachedEmbedding{
        ChunkID:     chunkID,
        ContentHash: contentHash,
        Embedding:   embedding,
        LastUpdated: time.Now(),
        AccessCount: 1,
        Version:     cached.Version + 1,
    }
    
    err = e.storeCachedEmbedding(cachedEmbedding)
    if err != nil {
        log.Printf("Failed to cache embedding for %s: %v", chunkID, err)
    }
    
    // 4. Update vector database asynchronously
    e.updateQueue.Enqueue(VectorUpdate{
        ChunkID:   chunkID,
        Embedding: embedding,
        Operation: "upsert",
    })
    
    return embedding, nil
}

func (e *EmbeddingCache) InvalidateByImpact(impact *ImpactAnalysis) error {
    // 1. Invalidate directly affected embeddings
    for _, entity := range impact.EmbeddingsToUpdate {
        err := e.invalidateEmbedding(entity.ChunkID)
        if err != nil {
            log.Printf("Failed to invalidate embedding %s: %v", entity.ChunkID, err)
        }
    }
    
    // 2. Schedule batch recomputation
    updateBatch := make([]string, 0)
    for _, entity := range impact.EmbeddingsToUpdate {
        if entity.Priority == "high" {
            updateBatch = append(updateBatch, entity.ChunkID)
        }
    }
    
    if len(updateBatch) > 0 {
        e.updateQueue.EnqueueBatch(updateBatch, "high")
    }
    
    return nil
}
```

### 4.3 Real-Time Update Pipeline

#### Incremental Processing Pipeline

```go
// internal/services/incremental_pipeline.go
type IncrementalPipeline struct {
    webhookHandler    *WebhookHandler
    changeDetector    *ChangeDetector
    impactAnalyzer    *ImpactAnalyzer
    updateOrchestrator *UpdateOrchestrator
    notificationService *NotificationService
}

type UpdateOrchestrator struct {
    astUpdater        *ASTUpdater
    graphUpdater      *GraphUpdater
    embeddingUpdater  *EmbeddingUpdater
    multiModalUpdater *MultiModalUpdater
    hierarchyUpdater  *HierarchyUpdater
    taskQueue        *TaskQueue
}

func (p *IncrementalPipeline) ProcessRepositoryUpdate(repositoryID primitive.ObjectID, commitHash string) error {
    log.Printf("Processing incremental update for repository %s, commit %s", repositoryID.Hex(), commitHash)
    
    // 1. Get last processed commit
    lastCommit, err := p.getLastProcessedCommit(repositoryID)
    if err != nil {
        return fmt.Errorf("failed to get last processed commit: %w", err)
    }
    
    // 2. Detect changes since last commit
    changes, err := p.changeDetector.DetectChanges(repositoryID, lastCommit, commitHash)
    if err != nil {
        return fmt.Errorf("failed to detect changes: %w", err)
    }
    
    if len(changes) == 0 {
        log.Printf("No changes detected for repository %s", repositoryID.Hex())
        return nil
    }
    
    log.Printf("Detected %d changes for repository %s", len(changes), repositoryID.Hex())
    
    // 3. Analyze impact for each change
    updatePlan := &IncrementalUpdatePlan{
        RepositoryID: repositoryID,
        CommitHash:   commitHash,
        Changes:      changes,
        Tasks:        make([]UpdateTask, 0),
        StartTime:    time.Now(),
    }
    
    for _, change := range changes {
        impact, err := p.impactAnalyzer.AnalyzeImpact(change)
        if err != nil {
            log.Printf("Failed to analyze impact for change %s: %v", change.ID, err)
            continue
        }
        
        // Create update tasks based on impact analysis
        tasks := p.createUpdateTasks(change, impact)
        updatePlan.Tasks = append(updatePlan.Tasks, tasks...)
    }
    
    // 4. Execute update plan
    err = p.updateOrchestrator.ExecuteUpdatePlan(updatePlan)
    if err != nil {
        return fmt.Errorf("failed to execute update plan: %w", err)
    }
    
    // 5. Update last processed commit
    err = p.updateLastProcessedCommit(repositoryID, commitHash)
    if err != nil {
        log.Printf("Failed to update last processed commit: %v", err)
    }
    
    // 6. Send completion notification
    p.notificationService.NotifyUpdateComplete(updatePlan)
    
    log.Printf("Completed incremental update for repository %s in %v", 
        repositoryID.Hex(), time.Since(updatePlan.StartTime))
    
    return nil
}

type IncrementalUpdatePlan struct {
    RepositoryID  primitive.ObjectID `json:"repositoryId"`
    CommitHash    string            `json:"commitHash"`
    Changes       []CodeChange      `json:"changes"`
    Tasks         []UpdateTask      `json:"tasks"`
    StartTime     time.Time         `json:"startTime"`
    Status        string            `json:"status"`
    Progress      float64           `json:"progress"`
    EstimatedTime time.Duration     `json:"estimatedTime"`
}

type UpdateTask struct {
    ID            string        `json:"id"`
    Type          string        `json:"type"` // ast, graph, embedding, multimodal, hierarchy
    Priority      string        `json:"priority"` // high, medium, low
    Dependencies  []string      `json:"dependencies"` // other task IDs
    EntityID      string        `json:"entityId"`
    EntityType    string        `json:"entityType"`
    Operation     string        `json:"operation"` // create, update, delete
    Payload       interface{}   `json:"payload"`
    Status        string        `json:"status"` // pending, running, completed, failed
    StartTime     time.Time     `json:"startTime"`
    EndTime       time.Time     `json:"endTime"`
    Error         string        `json:"error,omitempty"`
    RetryCount    int           `json:"retryCount"`
}

func (u *UpdateOrchestrator) ExecuteUpdatePlan(plan *IncrementalUpdatePlan) error {
    // 1. Sort tasks by dependencies and priority
    sortedTasks := u.sortTasksByDependencies(plan.Tasks)
    
    // 2. Execute tasks in parallel where possible
    return u.executeTasksWithConcurrency(sortedTasks, 5) // max 5 concurrent tasks
}

func (u *UpdateOrchestrator) executeTasksWithConcurrency(tasks []UpdateTask, maxConcurrency int) error {
    taskQueue := make(chan UpdateTask, len(tasks))
    results := make(chan TaskResult, len(tasks))
    
    // Add tasks to queue
    for _, task := range tasks {
        taskQueue <- task
    }
    close(taskQueue)
    
    // Start workers
    for i := 0; i < maxConcurrency; i++ {
        go u.taskWorker(taskQueue, results)
    }
    
    // Collect results
    completedTasks := 0
    failedTasks := 0
    
    for i := 0; i < len(tasks); i++ {
        result := <-results
        if result.Error != nil {
            log.Printf("Task %s failed: %v", result.TaskID, result.Error)
            failedTasks++
        } else {
            completedTasks++
        }
    }
    
    log.Printf("Completed %d tasks, failed %d tasks", completedTasks, failedTasks)
    
    if failedTasks > 0 {
        return fmt.Errorf("failed to complete %d tasks", failedTasks)
    }
    
    return nil
}
```

## Implementation Plan

### Month 10: Change Detection and Impact Analysis
**Week 1-2**: Change Detection Engine
- Implement Git-based change detection
- Add AST-aware change classification
- Create change scope determination

**Week 3-4**: Impact Analysis Foundation
- Implement dependency graph traversal
- Add impact confidence scoring
- Create update task generation

### Month 11: Incremental Update Systems
**Week 1-2**: Smart Caching Implementation
- Implement embedding cache with Redis
- Add cache invalidation strategies
- Create batch update queuing

**Week 3-4**: Update Orchestration
- Implement task dependency resolution
- Add parallel update execution
- Create rollback mechanisms

### Month 12: Real-Time Pipeline and Optimization
**Week 1-2**: Real-Time Processing Pipeline
- Implement webhook handling
- Add real-time update orchestration
- Create progress tracking and notifications

**Week 3-4**: Performance Optimization
- Optimize update algorithms
- Add performance monitoring
- Comprehensive testing and validation

## Technical Requirements

### Infrastructure Additions

```go
// Add to config
type IncrementalConfig struct {
    // Change Detection
    EnableIncremental    bool          `env:"ENABLE_INCREMENTAL" envDefault:"true"`
    MaxCommitHistory     int           `env:"MAX_COMMIT_HISTORY" envDefault:"100"`
    ChangeDetectionTimeout time.Duration `env:"CHANGE_DETECTION_TIMEOUT" envDefault:"30s"`
    
    // Impact Analysis
    MaxPropagationDepth  int           `env:"MAX_PROPAGATION_DEPTH" envDefault:"5"`
    ImpactConfidenceThreshold float64  `env:"IMPACT_CONFIDENCE_THRESHOLD" envDefault:"0.3"`
    
    // Caching
    EnableSmartCaching   bool          `env:"ENABLE_SMART_CACHING" envDefault:"true"`
    CacheTTL            time.Duration `env:"CACHE_TTL" envDefault:"24h"`
    MaxCacheSize        int64         `env:"MAX_CACHE_SIZE" envDefault:"10737418240"` // 10GB
    
    // Update Processing
    MaxConcurrentUpdates int           `env:"MAX_CONCURRENT_UPDATES" envDefault:"5"`
    UpdateBatchSize     int           `env:"UPDATE_BATCH_SIZE" envDefault:"100"`
    UpdateTimeout       time.Duration `env:"UPDATE_TIMEOUT" envDefault:"10m"`
    
    // Real-time Processing
    WebhookSecret       string        `env:"WEBHOOK_SECRET"`
    ProcessingDelay     time.Duration `env:"PROCESSING_DELAY" envDefault:"30s"`
}
```

### New Database Collections

```go
// MongoDB collections for incremental analysis
type IncrementalCollections struct {
    CodeChanges           string // "code_changes"
    ImpactAnalyses        string // "impact_analyses"
    UpdateTasks           string // "update_tasks"
    UpdatePlans           string // "update_plans"
    ProcessingStatus      string // "processing_status"
    CacheMetadata         string // "cache_metadata"
    ChangeNotifications   string // "change_notifications"
}
```

### API Endpoints

```yaml
# Add to OpenAPI spec
paths:
  /repositories/{id}/incremental-status:
    get:
      summary: Get incremental processing status
      responses:
        200:
          description: Processing status and metrics
          
  /repositories/{id}/trigger-update:
    post:
      summary: Manually trigger incremental update
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                commitHash:
                  type: string
                force:
                  type: boolean
      responses:
        202:
          description: Update triggered successfully
          
  /repositories/{id}/update-progress:
    get:
      summary: Get real-time update progress
      responses:
        200:
          description: Update progress information
          
  /webhooks/github:
    post:
      summary: GitHub webhook endpoint for push events
      responses:
        200:
          description: Webhook processed successfully
```

## Success Metrics

### Performance Metrics
- **Update Processing Time**: 90% reduction compared to full reprocessing
- **Change Detection Latency**: <30 seconds from commit to detection
- **Impact Analysis Speed**: <60 seconds for typical changes
- **Cache Hit Rate**: >80% for unchanged content

### Quality Metrics
- **Impact Accuracy**: >85% accuracy in predicting affected entities
- **Update Consistency**: 100% consistency across all analysis layers
- **Change Classification**: >90% accuracy in scope determination
- **Incremental Correctness**: Results identical to full reprocessing

### Business Metrics
- **Developer Experience**: <2 minutes from commit to updated understanding
- **Resource Efficiency**: 80% reduction in computational resources
- **System Availability**: 99.9% uptime during incremental updates
- **Cost Savings**: 70% reduction in processing costs

## Risk Mitigation

### Technical Risks
1. **Propagation Complexity**: Implement depth limits and confidence thresholds
2. **Cache Consistency**: Use content hashing and versioning
3. **Update Failures**: Implement rollback and retry mechanisms

### Data Integrity Risks
1. **Incremental Correctness**: Regular full reprocessing validation
2. **Dependency Tracking**: Comprehensive testing of impact analysis
3. **Race Conditions**: Proper synchronization and queuing

## Integration Points

Phase 4 enhances all previous phases:
- **Phase 1 (AST)**: Incremental AST updates and graph maintenance
- **Phase 2 (Semantic)**: PDG updates and hierarchical re-clustering
- **Phase 3 (Multi-Modal)**: Context updates and consistency re-checking

## Preparation for Phase 5

Real-time updates enable Phase 5's automated fix generation:
- **Fresh Understanding**: Always up-to-date code analysis
- **Change Tracking**: History of modifications for learning
- **Impact Prediction**: Understanding consequences before fixes