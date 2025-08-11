---
title: API Reference
type: api
order: 5
---

# API Reference Generation Prompt (English)

Generate comprehensive API reference documentation for the following project:

**Project Information:**
- Project: {{.ProjectName}}
- Primary Language: {{.PrimaryLanguage}}
- Description: {{.Description}}

**Module Information:**
{{range .Modules}}
- **{{.Name}}**: {{.Description}}
  {{range .Functions}}
  - {{.Name}}: {{.Description}}
  {{end}}
{{end}}

**Requirements:**
Create detailed API reference documentation that includes:

## 1. API Overview
- **Introduction**
  - Purpose and scope of the API
  - Target audience and use cases
  - API design principles
- **Base Information**
  - Base URLs or import statements
  - Versioning strategy
  - Supported formats (JSON, XML, etc.)
- **Quick Reference**
  - Most commonly used endpoints/functions
  - Basic usage patterns
  - Key concepts

## 2. Authentication & Authorization
- **Authentication Methods**
  - API keys, tokens, or credentials
  - Authentication flow
  - Security considerations
- **Authorization Levels**
  - Permission scopes
  - Rate limiting
  - Usage quotas

## 3. Request/Response Format
- **Request Structure**
  - Headers and parameters
  - Data formats and encoding
  - Content types
- **Response Structure**
  - Standard response format
  - Success and error responses
  - Status codes and meanings
- **Data Types**
  - Primitive types
  - Complex objects
  - Enumerations

## 4. API Endpoints/Functions
For each endpoint or function, provide:

### Function/Endpoint Name
- **Description**: Clear explanation of purpose
- **Syntax**: Function signature or HTTP method + URL
- **Parameters**:
  - Name, type, required/optional
  - Description and constraints
  - Default values
  - Examples
- **Returns**: Return type and description
- **Examples**: 
  - Request examples
  - Response examples
  - Error scenarios
- **Notes**: Important considerations, limitations

## 5. Error Handling
- **Error Response Format**
  - Standard error structure
  - Error codes and messages
  - Debugging information
- **Common Errors**
  - Typical error scenarios
  - Troubleshooting steps
  - Prevention strategies

## 6. Code Examples
- **Language-Specific Examples**
  - Multiple programming languages
  - Complete working examples
  - Best practices demonstration
- **Use Case Examples**
  - Common integration patterns
  - Real-world scenarios
  - Performance considerations

## 7. SDK and Libraries
- **Official SDKs**
  - Supported languages
  - Installation instructions
  - Basic usage
- **Community Libraries**
  - Third-party implementations
  - Compatibility notes
  - Contribution guidelines

**Format Requirements:**
- Use consistent formatting throughout
- Include syntax highlighting for code
- Provide interactive examples where possible
- Use tables for parameter documentation
- Include navigation/table of contents
- Cross-reference related functions

**Style Guidelines:**
- Be precise and unambiguous
- Use technical language appropriately
- Provide context for complex concepts
- Include practical examples
- Maintain consistency in terminology
