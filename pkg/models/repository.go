package models

import (
	"time"
)

// Repository represents a code repository
type Repository struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Provider    string    `json:"provider"` // github, gitlab, bitbucket
	Branch      string    `json:"branch"`
	LocalPath   string    `json:"local_path"`
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`
	Languages   []string  `json:"languages"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	Topics      []string  `json:"topics"`
	License     string    `json:"license"`
	Stars       int       `json:"stars"`
	Forks       int       `json:"forks"`
}

// FileInfo represents information about a file in the repository
type FileInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	Language     string    `json:"language"`
	Content      string    `json:"content,omitempty"`
	Hash         string    `json:"hash"`
	ModifiedAt   time.Time `json:"modified_at"`
	IsDirectory  bool      `json:"is_directory"`
	LineCount    int       `json:"line_count"`
	Complexity   int       `json:"complexity,omitempty"`
}

// CodeStructure represents the analyzed structure of the codebase
type CodeStructure struct {
	RepositoryID string                 `json:"repository_id"`
	Files        []FileInfo             `json:"files"`
	Directories  []DirectoryInfo        `json:"directories"`
	Dependencies []Dependency           `json:"dependencies"`
	Modules      []Module               `json:"modules"`
	Functions    []Function             `json:"functions"`
	Classes      []Class                `json:"classes"`
	Interfaces   []Interface            `json:"interfaces"`
	Constants    []Constant             `json:"constants"`
	Variables    []Variable             `json:"variables"`
	Imports      []Import               `json:"imports"`
	Exports      []Export               `json:"exports"`
	Tests        []TestInfo             `json:"tests"`
	Metrics      CodeMetrics            `json:"metrics"`
	Relationships []Relationship        `json:"relationships"`
}

// DirectoryInfo represents information about a directory
type DirectoryInfo struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	FileCount   int      `json:"file_count"`
	SubDirCount int      `json:"subdir_count"`
	Size        int64    `json:"size"`
	Purpose     string   `json:"purpose,omitempty"`
	Languages   []string `json:"languages"`
}

// Dependency represents a project dependency
type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"` // direct, dev, peer, optional
	Source      string `json:"source"` // npm, pip, cargo, go mod, etc.
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
}

// Module represents a code module or package
type Module struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Language    string   `json:"language"`
	Description string   `json:"description,omitempty"`
	Exports     []string `json:"exports"`
	Imports     []string `json:"imports"`
	Functions   []string `json:"functions"`
	Classes     []string `json:"classes"`
	LineCount   int      `json:"line_count"`
}

// Function represents a function or method
type Function struct {
	Name        string      `json:"name"`
	Module      string      `json:"module"`
	File        string      `json:"file"`
	StartLine   int         `json:"start_line"`
	EndLine     int         `json:"end_line"`
	Language    string      `json:"language"`
	Signature   string      `json:"signature"`
	Parameters  []Parameter `json:"parameters"`
	ReturnType  string      `json:"return_type,omitempty"`
	Description string      `json:"description,omitempty"`
	Complexity  int         `json:"complexity"`
	IsPublic    bool        `json:"is_public"`
	IsAsync     bool        `json:"is_async"`
	IsStatic    bool        `json:"is_static"`
	Decorators  []string    `json:"decorators,omitempty"`
	Calls       []string    `json:"calls,omitempty"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	IsOptional   bool   `json:"is_optional"`
	Description  string `json:"description,omitempty"`
}

// Class represents a class or struct
type Class struct {
	Name        string     `json:"name"`
	Module      string     `json:"module"`
	File        string     `json:"file"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Language    string     `json:"language"`
	Description string     `json:"description,omitempty"`
	Methods     []Function `json:"methods"`
	Properties  []Property `json:"properties"`
	Inherits    []string   `json:"inherits,omitempty"`
	Implements  []string   `json:"implements,omitempty"`
	IsAbstract  bool       `json:"is_abstract"`
	IsPublic    bool       `json:"is_public"`
	Decorators  []string   `json:"decorators,omitempty"`
}

// Property represents a class property or field
type Property struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPublic     bool   `json:"is_public"`
	IsStatic     bool   `json:"is_static"`
	IsReadonly   bool   `json:"is_readonly"`
	Description  string `json:"description,omitempty"`
}

// Interface represents an interface definition
type Interface struct {
	Name        string     `json:"name"`
	Module      string     `json:"module"`
	File        string     `json:"file"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Language    string     `json:"language"`
	Description string     `json:"description,omitempty"`
	Methods     []Function `json:"methods"`
	Properties  []Property `json:"properties"`
	Extends     []string   `json:"extends,omitempty"`
}

// Constant represents a constant definition
type Constant struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type,omitempty"`
	Module      string `json:"module"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

// Variable represents a variable definition
type Variable struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Module      string `json:"module"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
	IsGlobal    bool   `json:"is_global"`
}

// Import represents an import statement
type Import struct {
	Module      string   `json:"module"`
	Alias       string   `json:"alias,omitempty"`
	Items       []string `json:"items,omitempty"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	IsWildcard  bool     `json:"is_wildcard"`
}

// Export represents an export statement
type Export struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // function, class, variable, etc.
	Module      string `json:"module"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	IsDefault   bool   `json:"is_default"`
}

// TestInfo represents test information
type TestInfo struct {
	Name        string   `json:"name"`
	File        string   `json:"file"`
	StartLine   int      `json:"start_line"`
	EndLine     int      `json:"end_line"`
	Type        string   `json:"type"` // unit, integration, e2e
	Framework   string   `json:"framework"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CodeMetrics represents code quality metrics
type CodeMetrics struct {
	TotalLines          int     `json:"total_lines"`
	CodeLines           int     `json:"code_lines"`
	CommentLines        int     `json:"comment_lines"`
	BlankLines          int     `json:"blank_lines"`
	TotalFiles          int     `json:"total_files"`
	TotalFunctions      int     `json:"total_functions"`
	TotalClasses        int     `json:"total_classes"`
	TotalInterfaces     int     `json:"total_interfaces"`
	AverageComplexity   float64 `json:"average_complexity"`
	MaxComplexity       int     `json:"max_complexity"`
	TestCoverage        float64 `json:"test_coverage,omitempty"`
	DuplicationRatio    float64 `json:"duplication_ratio,omitempty"`
	TechnicalDebt       string  `json:"technical_debt,omitempty"`
}

// Relationship represents relationships between code elements
type Relationship struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Type        string `json:"type"` // calls, inherits, implements, imports, etc.
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description,omitempty"`
}
