# Phase 1: Foundation Enhancement (Months 1-3)

## Overview
**Goal**: Replace text chunking with structural code analysis

This phase transforms our platform from text-based analysis to structural code understanding using AST (Abstract Syntax Tree) analysis and knowledge graphs. Instead of treating code like natural language text, we'll understand its logical structure and relationships.

## Why This Phase is Critical

### Current Limitations
- **Text-based chunking**: 30-line chunks with overlap lose structural context
- **Embedding similarity**: Voyage AI embeddings treat code like natural language
- **Basic metadata**: Functions, classes, variables extracted via regex patterns
- **Missing relationships**: No understanding of how code components connect

### Research Evidence
- **SimAST-GCN (2024)**: AST-based analysis outperforms text-based by 60% for code understanding
- **HCGS Study (2024)**: Hierarchical Code Graph Summarization shows 82% improvement in retrieval accuracy
- **CodePlan Framework (2025)**: Repository-level reasoning achieves 21.4% vs 20.6% issue resolution

## Phase 1 Components

### 1.1 AST-Based Code Analysis Engine

#### Why AST Analysis?
- **Structural Understanding**: Parse code into logical components (functions, classes, variables)
- **Language Agnostic**: Tree-sitter parsers support 40+ languages
- **Precise Boundaries**: Exact function/class boundaries, not arbitrary line chunks
- **Dependency Extraction**: Understand imports, calls, inheritance relationships

#### Implementation Architecture

**Core Service Structure**:
```go
// New service: internal/services/ast_analyzer.go
type ASTAnalyzer struct {
    parsers map[string]ASTParser // language-specific parsers
    graphDB *neo4j.Driver       // knowledge graph storage
}

type ASTNode struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`     // function, class, variable
    Name         string                 `json:"name"`
    FilePath     string                 `json:"filePath"`
    StartLine    int                    `json:"startLine"`
    EndLine      int                    `json:"endLine"`
    Children     []string               `json:"children"`
    Dependencies []string               `json:"dependencies"`
    Metadata     map[string]interface{} `json:"metadata"`
}
```

**Parser Integration**:
- **Tree-sitter Integration**: Add parsers for JavaScript, Python, Go, Java, TypeScript
- **Language Support**: Start with top 5 languages, expand incrementally
- **Parser Management**: Lazy loading and caching of language parsers
- **Error Handling**: Graceful fallback to text-based chunking for unsupported files

#### Expected Benefits
- **40-60% accuracy improvement** in code understanding
- Enable relationship queries: "What functions call this method?"
- Precise code boundaries eliminate context bleeding
- Language-aware parsing respects syntax rules

### 1.2 Knowledge Graph Infrastructure

#### Why Knowledge Graphs?
- **Relationship Modeling**: Explicitly model code relationships
- **Graph Traversal**: Find related code through dependency paths
- **Multi-hop Reasoning**: "What would break if I change this function?"
- **Scalable Queries**: Neo4j optimized for relationship queries

#### Graph Schema Design

**Node Types**:
```cypher
// Core entities
CREATE (f:Function {
    name: "calculateTotal", 
    filePath: "src/billing.js", 
    startLine: 45,
    endLine: 67,
    parameters: ["items", "taxRate"],
    returnType: "number",
    complexity: 5
})

CREATE (c:Class {
    name: "BillingService", 
    filePath: "src/billing.js",
    startLine: 10,
    endLine: 120,
    methods: ["calculateTotal", "applyDiscount"],
    extends: "BaseService"
})

CREATE (v:Variable {
    name: "taxRate", 
    type: "const", 
    scope: "global",
    dataType: "number",
    filePath: "src/config.js",
    line: 15
})

CREATE (file:File {
    path: "src/billing.js",
    language: "javascript",
    size: 2500,
    lastModified: "2025-01-15"
})

CREATE (lib:Library {
    name: "lodash",
    version: "4.17.21",
    type: "npm"
})
```

**Relationship Types**:
```cypher
// Function relationships
CREATE (f1:Function)-[:CALLS {callCount: 5}]->(f2:Function)
CREATE (f:Function)-[:USES {lineNumber: 45}]->(v:Variable)
CREATE (f:Function)-[:DEFINED_IN]->(file:File)

// Class relationships
CREATE (c1:Class)-[:EXTENDS]->(c2:Class)
CREATE (c:Class)-[:IMPLEMENTS]->(i:Interface)
CREATE (c:Class)-[:HAS_METHOD]->(f:Function)

// File relationships
CREATE (file:File)-[:IMPORTS {importType: "default"}]->(lib:Library)
CREATE (file1:File)-[:DEPENDS_ON]->(file2:File)

// Module relationships
CREATE (m1:Module)-[:EXPORTS]->(f:Function)
CREATE (m:Module)-[:CONTAINS]->(file:File)
```

#### Neo4j Integration

**Database Setup**:
```go
// internal/database/neo4j.go
type Neo4jConnection struct {
    driver neo4j.Driver
    session neo4j.Session
}

func (n *Neo4jConnection) CreateASTNode(node *ASTNode) error {
    query := `
        CREATE (n:` + node.Type + ` {
            id: $id,
            name: $name,
            filePath: $filePath,
            startLine: $startLine,
            endLine: $endLine,
            metadata: $metadata
        })
    `
    
    _, err := n.session.Run(query, map[string]interface{}{
        "id": node.ID,
        "name": node.Name,
        "filePath": node.FilePath,
        "startLine": node.StartLine,
        "endLine": node.EndLine,
        "metadata": node.Metadata,
    })
    
    return err
}
```

#### Expected Benefits
- **Graph-based context retrieval**: Follow relationships to find relevant code
- **Impact analysis**: Understand change propagation through dependency chains
- **Architectural insights**: Visualize system structure and patterns
- **Performance**: Neo4j optimized for relationship queries at scale

### 1.3 Enhanced Metadata Extraction

#### Why Enhanced Metadata?
Current metadata extraction is limited to basic pattern matching. Enhanced extraction provides:
- **Semantic understanding** of code purpose
- **Call graph analysis** for function relationships
- **Data flow tracking** for variable usage
- **Complexity metrics** for code quality assessment

#### Enhanced Data Structure

```go
// internal/models/enhanced_chunk.go
type AdvancedChunkMetadata struct {
    Functions     []FunctionInfo    `bson:"functions"`
    Classes       []ClassInfo       `bson:"classes"`
    Variables     []VariableInfo    `bson:"variables"`
    Dependencies  []DependencyInfo  `bson:"dependencies"`
    CallGraph     []CallInfo        `bson:"callGraph"`
    DataFlow      []DataFlowInfo    `bson:"dataFlow"`
    ControlFlow   []ControlFlowInfo `bson:"controlFlow"`
    Complexity    ComplexityInfo    `bson:"complexity"`
    ASTFingerprint string           `bson:"astFingerprint"`
}

type FunctionInfo struct {
    Name        string            `bson:"name"`
    Parameters  []ParameterInfo   `bson:"parameters"`
    ReturnType  string            `bson:"returnType"`
    Visibility  string            `bson:"visibility"`  // public, private, protected
    Complexity  int               `bson:"complexity"`  // cyclomatic complexity
    CallCount   int               `bson:"callCount"`   // how many times called
    CalledBy    []string          `bson:"calledBy"`    // functions that call this
    Calls       []string          `bson:"calls"`       // functions this calls
}

type ClassInfo struct {
    Name         string   `bson:"name"`
    SuperClass   string   `bson:"superClass"`
    Interfaces   []string `bson:"interfaces"`
    Methods      []string `bson:"methods"`
    Properties   []string `bson:"properties"`
    IsAbstract   bool     `bson:"isAbstract"`
    Visibility   string   `bson:"visibility"`
}

type DependencyInfo struct {
    Type         string `bson:"type"`         // import, require, include
    Source       string `bson:"source"`       // what's being imported
    Alias        string `bson:"alias"`        // import alias
    IsExternal   bool   `bson:"isExternal"`   // external library vs local file
    Usage        []int  `bson:"usage"`        // line numbers where used
}

type ComplexityInfo struct {
    Cyclomatic   int `bson:"cyclomatic"`   // cyclomatic complexity
    Cognitive    int `bson:"cognitive"`    // cognitive complexity
    Lines        int `bson:"lines"`        // lines of code
    Maintainability float64 `bson:"maintainability"` // maintainability index
}
```

## Implementation Plan

### Month 1: AST Foundation
**Week 1-2**: Tree-sitter Integration
- Set up Tree-sitter parsers for JavaScript, Python, Go
- Create parser management system with lazy loading
- Implement AST traversal and node extraction

**Week 3-4**: Basic AST Analysis
- Extract functions, classes, variables from AST
- Generate unique IDs for AST nodes
- Create basic relationship detection (calls, imports)

### Month 2: Knowledge Graph Setup
**Week 1-2**: Neo4j Integration
- Set up Neo4j database connection and schema
- Implement node and relationship creation APIs
- Create graph population pipeline

**Week 3-4**: Graph Population
- Populate knowledge graph from AST analysis
- Implement relationship detection algorithms
- Create graph query interfaces

### Month 3: Enhanced Metadata & Integration
**Week 1-2**: Advanced Metadata Extraction
- Implement complexity calculation algorithms
- Create call graph analysis
- Add data flow tracking basics

**Week 3-4**: Integration & Testing
- Integrate AST analysis with existing pipeline
- Performance optimization and caching
- Comprehensive testing and validation

## Technical Requirements

### Infrastructure
- **Neo4j Database**: Version 5.0+ for knowledge graph storage
- **Tree-sitter**: Language parsers for AST analysis
- **Enhanced MongoDB Schema**: New collections for AST metadata
- **Qdrant Integration**: Vector storage for AST-aware embeddings

### Configuration Updates
```go
// Add to internal/config/config.go
type ASTAnalysisConfig struct {
    EnableAST          bool   `env:"ENABLE_AST_ANALYSIS" envDefault:"true"`
    TreeSitterPath     string `env:"TREE_SITTER_PATH" envDefault:"/usr/local/lib/tree-sitter"`
    
    // Knowledge Graph
    EnableKnowledgeGraph bool   `env:"ENABLE_KNOWLEDGE_GRAPH" envDefault:"true"`
    Neo4jURI            string `env:"NEO4J_URI" envDefault:"bolt://localhost:7687"`
    Neo4jUser           string `env:"NEO4J_USER" envDefault:"neo4j"`
    Neo4jPassword       string `env:"NEO4J_PASSWORD"`
    
    // Analysis Settings
    AnalysisDepth      string `env:"ANALYSIS_DEPTH" envDefault:"basic"` // basic, intermediate, full
    MaxNodesPerFile    int    `env:"MAX_NODES_PER_FILE" envDefault:"1000"`
    EnableCallGraph    bool   `env:"ENABLE_CALL_GRAPH" envDefault:"true"`
}
```

### New Services

1. **AST Analyzer Service** (`internal/services/ast_analyzer.go`)
   - Tree-sitter parser management
   - AST traversal and node extraction
   - Relationship detection

2. **Knowledge Graph Service** (`internal/services/knowledge_graph.go`)
   - Neo4j connection management
   - Graph population and updates
   - Query interfaces

3. **Enhanced Metadata Service** (`internal/services/enhanced_metadata.go`)
   - Advanced metadata extraction
   - Complexity calculations
   - Call graph analysis

## Success Metrics

### Technical Metrics
- **AST Parsing Accuracy**: >95% for supported languages
- **Knowledge Graph Population**: Complete graphs for all analyzed repositories
- **Query Performance**: <100ms for basic relationship queries
- **Memory Usage**: <2GB additional memory per repository

### Quality Metrics
- **Code Understanding Improvement**: 40-60% better accuracy in code queries
- **Relationship Query Success**: 90% success rate for "what calls this function" queries
- **Developer Satisfaction**: Improved relevance scores in user feedback

### Performance Metrics
- **Processing Time**: <5 minutes additional per 10,000 lines of code
- **Storage Growth**: ~20% increase in storage requirements
- **API Response Time**: <200ms for enhanced context retrieval

## Risk Mitigation

### Technical Risks
1. **Parser Complexity**: Start with 3 languages, expand gradually
2. **Memory Usage**: Implement streaming and caching strategies
3. **Neo4j Performance**: Optimize queries and use appropriate indexes

### Integration Risks
1. **Backward Compatibility**: Maintain existing API contracts
2. **Migration**: Gradual rollout with feature flags
3. **Fallback Strategy**: Fall back to text-based analysis for unsupported languages

## Next Phase Preparation

Phase 1 establishes the foundation for Phase 2 (Semantic Understanding):
- **Program Dependence Graphs**: Build on AST structure
- **Control/Data Flow Analysis**: Extend call graph analysis
- **Hierarchical Summarization**: Use knowledge graph for clustering

The AST analysis and knowledge graph infrastructure created in Phase 1 are essential prerequisites for the advanced semantic analysis planned in Phase 2.