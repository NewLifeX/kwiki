---
title: Architecture
type: architecture
order: 4
---

# Architecture Documentation Generation Prompt (English)

Generate comprehensive architecture documentation for the following project:

**Project Information:**
- Project: {{.ProjectName}}
- Primary Language: {{.PrimaryLanguage}}
- Description: {{.Description}}

**Module Structure:**
{{range .Modules}}
- **{{.Name}}**: {{.Description}}
  {{range .Functions}}
  - {{.Name}}: {{.Description}}
  {{end}}
{{end}}

**Requirements:**
Create detailed architecture documentation that includes:

## 1. System Overview
- **High-Level Architecture**
  - System purpose and scope
  - Key architectural decisions
  - Design principles and patterns
  - Technology stack overview
- **System Context**
  - External dependencies
  - Integration points
  - User interactions
  - Environmental considerations

## 2. Architectural Patterns
- **Design Patterns Used**
  - Primary architectural patterns
  - Pattern implementation details
  - Benefits and trade-offs
  - Pattern interactions
- **Architectural Styles**
  - Layered architecture
  - Microservices/Monolithic approach
  - Event-driven components
  - Data flow patterns

## 3. Core Components
{{range .Modules}}
### {{.Name}} Module
- **Purpose**: {{.Description}}
- **Responsibilities**
  - Primary functions and capabilities
  - Data processing responsibilities
  - Integration responsibilities
- **Key Functions**
  {{range .Functions}}
  - **{{.Name}}**: {{.Description}}
  {{end}}
- **Dependencies**
  - Internal module dependencies
  - External library dependencies
  - Service dependencies
- **Interfaces**
  - Public APIs
  - Internal interfaces
  - Data contracts
{{end}}

## 4. Data Architecture
- **Data Models**
  - Core data entities
  - Relationships and constraints
  - Data validation rules
- **Data Flow**
  - Input data sources
  - Processing pipelines
  - Output destinations
  - Data transformation points
- **Storage Strategy**
  - Database design
  - Caching mechanisms
  - Data persistence patterns
  - Backup and recovery

## 5. System Interactions
- **Internal Communication**
  - Module-to-module communication
  - Message passing mechanisms
  - Event handling
  - Synchronous vs asynchronous patterns
- **External Integrations**
  - API integrations
  - Third-party services
  - Database connections
  - File system interactions

## 6. Security Architecture
- **Security Principles**
  - Authentication mechanisms
  - Authorization strategies
  - Data protection measures
  - Security boundaries
- **Threat Model**
  - Identified security risks
  - Mitigation strategies
  - Security controls
  - Compliance considerations

## 7. Performance and Scalability
- **Performance Characteristics**
  - Expected load patterns
  - Performance bottlenecks
  - Optimization strategies
  - Monitoring approaches
- **Scalability Design**
  - Horizontal scaling capabilities
  - Vertical scaling considerations
  - Load distribution
  - Resource management

## 8. Deployment Architecture
- **Deployment Model**
  - Deployment environments
  - Infrastructure requirements
  - Container strategies
  - Service orchestration
- **Configuration Management**
  - Environment-specific configurations
  - Configuration sources
  - Runtime configuration
  - Feature flags

## 9. Quality Attributes
- **Reliability**
  - Fault tolerance mechanisms
  - Error handling strategies
  - Recovery procedures
  - Availability targets
- **Maintainability**
  - Code organization
  - Testing strategies
  - Documentation standards
  - Development workflows

## 10. Future Considerations
- **Extensibility**
  - Plugin architectures
  - Extension points
  - API evolution
  - Backward compatibility
- **Evolution Path**
  - Planned improvements
  - Technology migration paths
  - Architectural debt
  - Refactoring opportunities

**Format Requirements:**
- Use clear architectural diagrams (describe in text)
- Include component interaction diagrams
- Provide code examples for key patterns
- Use consistent terminology throughout
- Include decision rationale for major choices
- Add cross-references between sections

**Style Guidelines:**
- Write for both technical and non-technical stakeholders
- Balance high-level overview with technical details
- Include practical examples and use cases
- Explain the "why" behind architectural decisions
- Use visual descriptions where diagrams would help
