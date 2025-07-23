// ABOUTME: Search-specific TypeScript types derived from the OpenAPI specification
// ABOUTME: These types will be merged into the main types file once OpenAPI is regenerated

export interface CodeChunk {
    id: string;
    repositoryId: string;
    filePath: string;
    fileName: string;
    language: string;
    startLine: number;
    endLine: number;
    content: string;
    contentHash?: string;
    imports?: string[];
    metadata: ChunkMetadata;
    vectorId?: string;
    createdAt: string;
    updatedAt?: string;
}

export interface ChunkMetadata {
    functions?: string[];
    classes?: string[];
    variables?: string[];
    types?: string[];
    complexity?: number;
}

export interface SearchRequest {
    query: string;
    repositoryId?: string;
    language?: string;
    fileType?: string;
    limit?: number;
    offset?: number;
}

export interface SearchResult extends CodeChunk {
    score: number;
    highlight?: string;
}

export interface SearchResponse {
    results: SearchResult[];
    total: number;
    hasMore: boolean;
    query: string;
}

export interface SearchStats {
    totalChunks: number;
    totalLines: number;
    avgComplexity: number;
    languages: string[];
}

// Quick search response (simplified)
export interface QuickSearchResponse {
    results: Array<{
        id: string;
        filePath: string;
        fileName: string;
        language: string;
        highlight: string;
        score: number;
    }>;
    total: number;
    query: string;
}

// Search suggestions response
export interface SearchSuggestionsResponse {
    suggestions: string[];
    query: string;
}

// Languages response
export interface LanguagesResponse {
    languages: string[];
}

// Recent chunks response
export interface RecentChunksResponse {
    chunks: CodeChunk[];
    total: number;
}