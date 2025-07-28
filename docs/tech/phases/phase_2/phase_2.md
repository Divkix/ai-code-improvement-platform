# Phase 2: Semantic Understanding (Months 4-6)

## Overview
**Goal**: Understand code behavior and data flow, not just structure

Building on Phase 1's AST foundation, Phase 2 adds deep semantic understanding of how code actually behaves. We move beyond structural analysis to understand execution flow, data dependencies, and hierarchical code organization.

## Why Semantic Understanding Matters

### Phase 1 Limitations
While Phase 1 provides structural understanding, it lacks:
- **Behavioral Analysis**: How code actually executes
- **Data Flow Understanding**: How data moves through variables
- **Control Dependencies**: What conditions affect code execution
- **Semantic Clustering**: Grouping related functionality

### Research Foundation
- **Program Dependence Graphs**: 35-50% improvement in change impact prediction
- **HCGS Study (2024)**: Hierarchical Code Graph Summarization shows 82% improvement in retrieval accuracy
- **Control Flow Analysis**: Essential for understanding code behavior and debugging

## Phase 2 Components

### 2.1 Program Dependence Graph Implementation

#### Why Program Dependence Graphs?
Program Dependence Graphs (PDGs) combine control flow and data flow analysis to create a comprehensive view of program behavior:

- **Control Dependencies**: Understand execution flow and conditionals
- **Data Dependencies**: Track how data flows through variables
- **Def-Use Chains**: Connect variable definitions to their usage
- **What-if Analysis**: Predict impact of code changes

#### Technical Architecture

**Core Data Structures**:
```go
// internal/models/program_dependence.go
type ProgramDependenceGraph struct {
    RepositoryID  primitive.ObjectID `bson:"repositoryId" json:"repositoryId"`
    FilePath      string             `bson:"filePath" json:"filePath"`
    FunctionName  string             `bson:"functionName" json:"functionName"`
    Nodes         []PDGNode          `bson:"nodes" json:"nodes"`
    ControlEdges  []DependenceEdge   `bson:"controlEdges" json:"controlEdges"`
    DataEdges     []DependenceEdge   `bson:"dataEdges" json:"dataEdges"`
    CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type PDGNode struct {
    ID        string    `bson:"id" json:"id"`
    Type      string    `bson:"type" json:"type"` // statement, expression, declaration
    Content   string    `bson:"content" json:"content"`
    FilePath  string    `bson:"filePath" json:"filePath"`
    Line      int       `bson:"line" json:"line"`
    Column    int       `bson:"column" json:"column"`
    Variables []string  `bson:"variables" json:"variables"` // variables used/defined
    ASTNodeID string    `bson:"astNodeId" json:"astNodeId"` // link to AST
}

type DependenceEdge struct {
    FromNodeID   string            `bson:"fromNodeId" json:"fromNodeId"`
    ToNodeID     string            `bson:"toNodeId" json:"toNodeId"`
    EdgeType     string            `bson:"edgeType" json:"edgeType"` // control, data
    DependenceType string          `bson:"dependenceType" json:"dependenceType"` // flow, anti, output
    Variable     string            `bson:"variable,omitempty" json:"variable,omitempty"` // for data edges
    Condition    string            `bson:"condition,omitempty" json:"condition,omitempty"` // for control edges
    Confidence   float64           `bson:"confidence" json:"confidence"` // analysis confidence
}
```

**PDG Construction Algorithm**:
```go
// internal/services/pdg_analyzer.go
type PDGAnalyzer struct {
    astService      *ASTAnalyzer
    cfgBuilder      *ControlFlowGraphBuilder
    dfgBuilder      *DataFlowGraphBuilder
    dominanceAnalyzer *DominanceAnalyzer
}

func (p *PDGAnalyzer) BuildPDG(function *ASTNode) (*ProgramDependenceGraph, error) {
    // 1. Build Control Flow Graph from AST
    cfg, err := p.cfgBuilder.BuildCFG(function)
    if err != nil {
        return nil, fmt.Errorf("failed to build CFG: %w", err)
    }
    
    // 2. Compute dominance relationships
    dominance := p.dominanceAnalyzer.ComputeDominance(cfg)
    postDominance := p.dominanceAnalyzer.ComputePostDominance(cfg)
    
    // 3. Build Data Flow Graph
    dfg, err := p.dfgBuilder.BuildDFG(function)
    if err != nil {
        return nil, fmt.Errorf("failed to build DFG: %w", err)
    }
    
    // 4. Combine into Program Dependence Graph
    return p.combineToPDG(cfg, dfg, dominance, postDominance)
}

func (p *PDGAnalyzer) combineToPDG(cfg *ControlFlowGraph, dfg *DataFlowGraph, 
                                  dominance, postDominance *DominanceTree) *ProgramDependenceGraph {
    
    pdg := &ProgramDependenceGraph{
        Nodes: make([]PDGNode, 0),
        ControlEdges: make([]DependenceEdge, 0),
        DataEdges: make([]DependenceEdge, 0),
    }
    
    // Convert CFG nodes to PDG nodes
    for _, cfgNode := range cfg.Nodes {
        pdgNode := PDGNode{
            ID: cfgNode.ID,
            Type: cfgNode.Type,
            Content: cfgNode.Content,
            FilePath: cfgNode.FilePath,
            Line: cfgNode.Line,
            Variables: cfgNode.Variables,
        }
        pdg.Nodes = append(pdg.Nodes, pdgNode)
    }
    
    // Add control dependence edges using dominance
    pdg.ControlEdges = p.buildControlDependences(cfg, postDominance)
    
    // Add data dependence edges from DFG
    pdg.DataEdges = p.buildDataDependences(dfg)
    
    return pdg
}
```

#### Expected Benefits
- **Change impact analysis**: Understand what breaks when code changes
- **Bug prediction**: Identify potential data flow issues
- **Refactoring safety**: Ensure changes don't break dependencies
- **Testing guidance**: Focus tests on critical dependency paths

### 2.2 Control and Data Flow Analysis

#### Control Flow Analysis

**Why Control Flow?**
- **Execution Understanding**: How code actually runs, not just structure
- **Conditional Logic**: Understand branch conditions and their effects
- **Loop Analysis**: Detect infinite loops and performance issues
- **Exception Handling**: Track error propagation paths

**Control Flow Graph Construction**:
```go
// internal/services/control_flow.go
type ControlFlowGraphBuilder struct {
    astService *ASTAnalyzer
}

type ControlFlowGraph struct {
    Nodes []CFGNode     `json:"nodes"`
    Edges []CFGEdge     `json:"edges"`
    Entry string        `json:"entry"` // entry node ID
    Exit  string        `json:"exit"`  // exit node ID
}

type CFGNode struct {
    ID         string   `json:"id"`
    Type       string   `json:"type"`       // basic_block, condition, loop_header
    Statements []string `json:"statements"` // AST node IDs
    Content    string   `json:"content"`    // human-readable content
    FilePath   string   `json:"filePath"`
    StartLine  int      `json:"startLine"`
    EndLine    int      `json:"endLine"`
}

type CFGEdge struct {
    From      string `json:"from"`
    To        string `json:"to"`
    Type      string `json:"type"`      // unconditional, true_branch, false_branch
    Condition string `json:"condition"` // condition for conditional edges
}

func (c *ControlFlowGraphBuilder) BuildCFG(function *ASTNode) (*ControlFlowGraph, error) {
    cfg := &ControlFlowGraph{
        Nodes: make([]CFGNode, 0),
        Edges: make([]CFGEdge, 0),
    }
    
    // Parse function body and build basic blocks
    basicBlocks := c.identifyBasicBlocks(function)
    
    // Connect basic blocks with edges
    c.connectBasicBlocks(basicBlocks, cfg)
    
    // Add special entry and exit nodes
    c.addEntryExitNodes(cfg)
    
    return cfg, nil
}
```

#### Data Flow Analysis

**Why Data Flow?**
- **Variable Lifecycle**: Track variable creation, modification, usage
- **Def-Use Chains**: Connect variable definitions to usage points
- **Reaching Definitions**: What definitions reach each use
- **Live Variable Analysis**: What variables are live at each point

**Data Flow Graph Construction**:
```go
// internal/services/data_flow.go
type DataFlowGraphBuilder struct {
    astService *ASTAnalyzer
}

type DataFlowGraph struct {
    Variables    []VariableInfo  `json:"variables"`
    Definitions  []Definition    `json:"definitions"`
    Uses         []Use           `json:"uses"`
    DefUseChains []DefUseChain   `json:"defUseChains"`
}

type VariableInfo struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Scope    string `json:"scope"`
    FilePath string `json:"filePath"`
}

type Definition struct {
    ID        string `json:"id"`
    Variable  string `json:"variable"`
    Line      int    `json:"line"`
    Statement string `json:"statement"`
    Type      string `json:"type"` // assignment, declaration, parameter
}

type Use struct {
    ID        string `json:"id"`
    Variable  string `json:"variable"`
    Line      int    `json:"line"`
    Statement string `json:"statement"`
    Type      string `json:"type"` // read, write, read_write
}

type DefUseChain struct {
    DefinitionID string `json:"definitionId"`
    UseID        string `json:"useId"`
    Variable     string `json:"variable"`
    Path         []int  `json:"path"` // line numbers in execution path
}

func (d *DataFlowGraphBuilder) BuildDFG(function *ASTNode) (*DataFlowGraph, error) {
    dfg := &DataFlowGraph{
        Variables: make([]VariableInfo, 0),
        Definitions: make([]Definition, 0),
        Uses: make([]Use, 0),
        DefUseChains: make([]DefUseChain, 0),
    }
    
    // 1. Identify all variables in the function
    variables := d.extractVariables(function)
    dfg.Variables = variables
    
    // 2. Find all variable definitions
    definitions := d.findDefinitions(function, variables)
    dfg.Definitions = definitions
    
    // 3. Find all variable uses
    uses := d.findUses(function, variables)
    dfg.Uses = uses
    
    // 4. Build def-use chains
    defUseChains := d.buildDefUseChains(definitions, uses)
    dfg.DefUseChains = defUseChains
    
    return dfg, nil
}
```

### 2.3 Hierarchical Code Summarization

#### Why Hierarchical Summarization?

**Research Evidence**: HCGS (Hierarchical Code Graph Summarization) shows 82% improvement in retrieval accuracy for large codebases by creating multi-level semantic clusters.

**Problem with Current Approach**:
- Flat chunking loses hierarchical relationships
- No semantic clustering of related functionality
- Large codebases become overwhelming without structure

#### Multi-Level Clustering Architecture

```go
// internal/services/hierarchical_summarizer.go
type HierarchicalSummarizer struct {
    embeddingService *EmbeddingService
    clusteringService *ClusteringService
    summaryService   *SummaryService
    knowledgeGraph   *KnowledgeGraphService
}

type HierarchicalSummary struct {
    RepositoryID primitive.ObjectID `bson:"repositoryId" json:"repositoryId"`
    Levels       []SummaryLevel     `bson:"levels" json:"levels"`
    CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type SummaryLevel struct {
    Level      int           `bson:"level" json:"level"`    // 0=chunks, 1=functions, 2=modules, 3=system
    Clusters   []CodeCluster `bson:"clusters" json:"clusters"`
    Summary    string        `bson:"summary" json:"summary"`
    Metadata   LevelMetadata `bson:"metadata" json:"metadata"`
}

type CodeCluster struct {
    ID           string    `bson:"id" json:"id"`
    Name         string    `bson:"name" json:"name"`          // descriptive cluster name
    Chunks       []string  `bson:"chunkIds" json:"chunkIds"`  // chunk IDs in cluster
    Functions    []string  `bson:"functionIds" json:"functionIds"` // function IDs in cluster
    Centroid     []float32 `bson:"centroid" json:"centroid"`  // embedding centroid
    Summary      string    `bson:"summary" json:"summary"`    // cluster purpose
    SemanticType string    `bson:"semanticType" json:"semanticType"` // auth, api, database, ui
    Purpose      string    `bson:"purpose" json:"purpose"`    // what this cluster does
    Keywords     []string  `bson:"keywords" json:"keywords"`  // key terms
    Complexity   float64   `bson:"complexity" json:"complexity"` // average complexity
    Size         int       `bson:"size" json:"size"`          // number of elements
}

type LevelMetadata struct {
    TotalClusters    int     `bson:"totalClusters" json:"totalClusters"`
    AvgClusterSize   float64 `bson:"avgClusterSize" json:"avgClusterSize"`
    SilhouetteScore  float64 `bson:"silhouetteScore" json:"silhouetteScore"` // clustering quality
    CoverageScore    float64 `bson:"coverageScore" json:"coverageScore"`     // how much code is covered
}
```

#### Clustering Strategy Implementation

**Level 0: Code Chunks (Base Level)**
```go
func (h *HierarchicalSummarizer) ClusterChunks(chunks []*models.CodeChunk) ([]CodeCluster, error) {
    // 1. Extract embeddings for all chunks
    embeddings := make([][]float32, len(chunks))
    for i, chunk := range chunks {
        embedding, err := h.embeddingService.GetEmbedding(chunk.Content)
        if err != nil {
            return nil, err
        }
        embeddings[i] = embedding
    }
    
    // 2. Apply K-means clustering
    numClusters := h.calculateOptimalClusters(len(chunks)) // use elbow method
    clusters := h.clusteringService.KMeans(embeddings, numClusters)
    
    // 3. Generate cluster summaries
    codeClusters := make([]CodeCluster, len(clusters))
    for i, cluster := range clusters {
        summary := h.generateClusterSummary(cluster.ChunkIDs, chunks)
        semanticType := h.classifySemanticType(summary)
        
        codeClusters[i] = CodeCluster{
            ID:           fmt.Sprintf("chunk_cluster_%d", i),
            Name:         h.generateClusterName(summary, semanticType),
            Chunks:       cluster.ChunkIDs,
            Centroid:     cluster.Centroid,
            Summary:      summary,
            SemanticType: semanticType,
            Purpose:      h.extractPurpose(summary),
            Keywords:     h.extractKeywords(summary),
        }
    }
    
    return codeClusters, nil
}
```

**Level 1: Function-Level Clusters**
```go
func (h *HierarchicalSummarizer) ClusterFunctions(functions []*ASTNode) ([]CodeCluster, error) {
    // 1. Group chunks by function
    functionGroups := h.groupChunksByFunction(functions)
    
    // 2. Create function-level embeddings (average of chunk embeddings)
    functionEmbeddings := make([][]float32, len(functionGroups))
    for i, group := range functionGroups {
        functionEmbeddings[i] = h.averageEmbeddings(group.ChunkEmbeddings)
    }
    
    // 3. Cluster functions by semantic similarity
    numClusters := h.calculateOptimalClusters(len(functions))
    clusters := h.clusteringService.KMeans(functionEmbeddings, numClusters)
    
    // 4. Generate function cluster summaries
    return h.generateFunctionClusters(clusters, functionGroups)
}
```

**Level 2: Module-Level Clusters**
```go
func (h *HierarchicalSummarizer) ClusterModules(files []*models.FileInfo) ([]CodeCluster, error) {
    // 1. Create module representations from file clusters
    moduleEmbeddings := h.createModuleEmbeddings(files)
    
    // 2. Apply hierarchical clustering for modules
    clusters := h.clusteringService.HierarchicalClustering(moduleEmbeddings)
    
    // 3. Generate module cluster summaries
    return h.generateModuleClusters(clusters, files)
}
```

**Level 3: System-Level Architecture**
```go
func (h *HierarchicalSummarizer) CreateSystemSummary(moduleClusters []CodeCluster) (*CodeCluster, error) {
    // 1. Analyze architectural patterns
    architecturalPatterns := h.detectArchitecturalPatterns(moduleClusters)
    
    // 2. Create system-level summary
    systemSummary := h.generateSystemSummary(moduleClusters, architecturalPatterns)
    
    // 3. Generate overall system cluster
    return &CodeCluster{
        ID:           "system_architecture",
        Name:         "System Architecture",
        Summary:      systemSummary,
        SemanticType: "architecture",
        Purpose:      h.extractSystemPurpose(systemSummary),
        Keywords:     h.extractArchitecturalTerms(systemSummary),
    }, nil
}
```

#### Semantic Type Classification

```go
func (h *HierarchicalSummarizer) classifySemanticType(summary string) string {
    // Use LLM to classify cluster purpose
    prompt := fmt.Sprintf(`
    Analyze this code cluster summary and classify its primary semantic type:
    
    Summary: %s
    
    Choose the most appropriate category:
    - authentication: Login, user management, security
    - api: REST endpoints, HTTP handlers, routing
    - database: Data access, queries, persistence
    - ui: User interface, frontend components
    - business_logic: Core business rules and processing
    - utilities: Helper functions, common utilities
    - testing: Test code, mocks, fixtures
    - configuration: Settings, environment, constants
    - infrastructure: Deployment, monitoring, logging
    
    Return only the category name.
    `, summary)
    
    response, err := h.summaryService.GenerateSummary(prompt)
    if err != nil {
        return "unknown"
    }
    
    return strings.TrimSpace(strings.ToLower(response))
}
```

#### Expected Benefits
- **82% retrieval improvement** (per HCGS research)
- **Semantic organization**: Group related functionality automatically
- **Scalable understanding**: Handle large codebases efficiently
- **Multi-granularity search**: From specific functions to system architecture

## Implementation Plan

### Month 4: Program Dependence Foundations
**Week 1-2**: Control Flow Analysis
- Implement Control Flow Graph construction
- Add basic block identification algorithms
- Create dominance analysis for control dependencies

**Week 3-4**: Data Flow Analysis
- Implement Data Flow Graph construction
- Add def-use chain analysis
- Create reaching definitions analysis

### Month 5: PDG Integration & Hierarchical Clustering
**Week 1-2**: PDG Construction
- Combine control and data flow into Program Dependence Graphs
- Implement PDG storage and querying
- Add impact analysis using PDGs

**Week 3-4**: Hierarchical Clustering Foundation
- Implement K-means clustering for code chunks
- Create embedding-based similarity metrics
- Add cluster quality evaluation (silhouette score)

### Month 6: Advanced Summarization & Integration
**Week 1-2**: Multi-Level Clustering
- Implement function-level clustering
- Add module-level clustering
- Create system-level architectural analysis

**Week 3-4**: Integration & Optimization
- Integrate hierarchical summaries with search
- Performance optimization and caching
- Comprehensive testing and validation

## Technical Requirements

### New Database Collections
```go
// MongoDB collections to add
type DatabaseCollections struct {
    ProgramDependenceGraphs  string // "program_dependence_graphs"
    ControlFlowGraphs       string // "control_flow_graphs"
    DataFlowGraphs          string // "data_flow_graphs"
    HierarchicalSummaries   string // "hierarchical_summaries"
    CodeClusters            string // "code_clusters"
    DefUseChains           string // "def_use_chains"
}
```

### Enhanced Configuration
```go
// Add to existing config
type SemanticAnalysisConfig struct {
    // PDG Configuration
    EnablePDG            bool `env:"ENABLE_PDG_ANALYSIS" envDefault:"true"`
    MaxPDGNodes          int  `env:"MAX_PDG_NODES" envDefault:"5000"`
    PDGAnalysisTimeout   time.Duration `env:"PDG_ANALYSIS_TIMEOUT" envDefault:"30s"`
    
    // Clustering Configuration
    EnableHierarchical   bool    `env:"ENABLE_HIERARCHICAL" envDefault:"true"`
    MaxClustersPerLevel  int     `env:"MAX_CLUSTERS_PER_LEVEL" envDefault:"20"`
    MinClusterSize       int     `env:"MIN_CLUSTER_SIZE" envDefault:"3"`
    ClusteringAlgorithm  string  `env:"CLUSTERING_ALGORITHM" envDefault:"kmeans"`
    SilhouetteThreshold  float64 `env:"SILHOUETTE_THRESHOLD" envDefault:"0.5"`
    
    // Performance Tuning
    MaxConcurrentPDG     int `env:"MAX_CONCURRENT_PDG" envDefault:"5"`
    CacheClusterResults  bool `env:"CACHE_CLUSTER_RESULTS" envDefault:"true"`
    ClusterCacheTTL      time.Duration `env:"CLUSTER_CACHE_TTL" envDefault:"1h"`
}
```

### New API Endpoints
```yaml
# Add to OpenAPI spec
paths:
  /repositories/{id}/pdg:
    get:
      summary: Get Program Dependence Graph for repository
      parameters:
        - name: function
          in: query
          description: Specific function to analyze
      responses:
        200:
          description: Program Dependence Graph data
          
  /repositories/{id}/hierarchical-summary:
    get:
      summary: Get hierarchical code summary
      parameters:
        - name: level
          in: query
          description: Summary level (0-3)
      responses:
        200:
          description: Hierarchical summary data
          
  /repositories/{id}/impact-analysis:
    post:
      summary: Analyze impact of potential code changes
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                changes:
                  type: array
                  items:
                    $ref: '#/components/schemas/CodeChange'
      responses:
        200:
          description: Impact analysis results
```

## Success Metrics

### Technical Metrics
- **PDG Construction Time**: <30 seconds per function
- **Clustering Quality**: Silhouette score >0.5 for all levels
- **Memory Usage**: <5GB additional per repository
- **Query Performance**: <500ms for PDG-based queries

### Quality Metrics
- **Code Understanding**: 50% improvement in relationship queries
- **Impact Analysis Accuracy**: >85% accuracy in predicting change effects
- **Clustering Relevance**: >80% user satisfaction with cluster organization
- **Summary Quality**: >4/5 rating for hierarchical summaries

### Business Metrics
- **Query Success Rate**: >90% for complex architectural questions
- **Developer Satisfaction**: Improved relevance in code exploration
- **Time Savings**: 30% reduction in time to understand code relationships

## Risk Mitigation

### Performance Risks
1. **PDG Construction Complexity**: Implement timeouts and progressive analysis
2. **Memory Usage**: Use streaming processing and disk-based caching
3. **Clustering Performance**: Implement approximate algorithms for large datasets

### Quality Risks
1. **Clustering Accuracy**: Multiple validation metrics and manual verification
2. **PDG Completeness**: Fallback to simpler analysis for complex functions
3. **Summary Quality**: Human validation of summary generation

## Integration with Phase 1

Phase 2 builds directly on Phase 1 infrastructure:
- **AST Foundation**: PDG construction uses AST nodes from Phase 1
- **Knowledge Graph**: Enhanced with PDG relationships
- **Enhanced Metadata**: Enriched with flow analysis data

## Preparation for Phase 3

Phase 2 prepares for Phase 3 (Multi-Modal Context Integration):
- **Behavioral Understanding**: PDGs provide execution context for intent analysis
- **Hierarchical Structure**: Multi-level summaries enable better context selection
- **Semantic Classification**: Cluster types inform multi-modal analysis focus