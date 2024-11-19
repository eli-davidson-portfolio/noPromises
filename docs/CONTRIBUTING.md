# Contributing to noPromises

## Development Requirements
- Go 1.21+
- golangci-lint
- staticcheck
- go-mod-outdated
- Make (optional)

## Development Workflow
1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Code Standards
- All code must be thoroughly tested
- Follow Go best practices
- Maintain FBP principles
- Document all exported items

## Documentation Contribution Guidelines

### Documentation Structure
```
docs/
├── api/              # API documentation
│   ├── endpoints.md  # API endpoint details
│   └── swagger.json  # OpenAPI specification
├── architecture/     # Architecture documentation
│   ├── patterns/    # Design patterns
│   ├── subsystems/  # Subsystem details
│   └── README.md    # Architecture overview
├── guides/          # User guides
└── README.md        # Documentation home
```

### Documentation Standards
1. File Format
   - Use Markdown for all documentation
   - Include language tags in code blocks
   - Follow consistent heading hierarchy
   - Use relative links for navigation

2. Code Examples
   - Include language identifier
   - Provide complete, working examples
   - Add explanatory comments
   - Show both success and error cases

3. API Documentation
   - Follow OpenAPI 3.0 specification
   - Include request/response examples
   - Document all status codes
   - Specify content types

4. Diagrams
   - Use Mermaid for diagrams
   - Include diagram source
   - Provide diagram description
   - Follow consistent styling

### Testing Requirements
1. Documentation Tests
   - Verify all links work
   - Ensure code examples compile
   - Check Markdown formatting
   - Validate OpenAPI spec

2. Integration Tests
   - Test documentation server
   - Verify HTML rendering
   - Check diagram generation
   - Test API documentation

### Visualization Guidelines
1. Network Diagrams
   - Use consistent node shapes
   - Follow left-to-right flow
   - Label all connections
   - Include state indicators

2. State Diagrams
   - Show all valid states
   - Include transition conditions
   - Mark initial/final states
   - Use consistent styling

### Pull Request Process
1. Documentation Updates
   - Update relevant docs
   - Add new docs as needed
   - Update table of contents
   - Check formatting

2. Review Process
   - Technical accuracy
   - Documentation clarity
   - Code example quality
   - Link validation

## Best Practices
1. Keep documentation close to code
2. Update docs with code changes
3. Use consistent terminology
4. Include practical examples
5. Consider documentation impact in reviews

## Getting Help
- Join our Discord server
- Check existing issues
- Review documentation
- Ask in discussions