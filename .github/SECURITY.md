# Security Policy

## Supported Versions

BrewSource MCP Server is currently in active development. We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| Phase 1 MVP | :white_check_mark: |

## Reporting a Vulnerability

We take the security of BrewSource MCP Server seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please send an email to: **<security@brewsource-mcp.dev>** (or create a private security advisory on GitHub)

Include the following information in your report:

1. **Description** - A clear description of the vulnerability
2. **Impact** - How the vulnerability could be exploited and its potential impact
3. **Reproduction** - Step-by-step instructions to reproduce the issue
4. **Environment** - Version, operating system, configuration details
5. **Suggested Fix** - If you have ideas for how to fix the issue

### What to Expect

- **Acknowledgment** - We will acknowledge receipt of your report within 48 hours
- **Initial Assessment** - We will provide an initial assessment within 5 business days
- **Regular Updates** - We will keep you informed of our progress
- **Resolution Timeline** - We aim to resolve critical vulnerabilities within 30 days
- **Credit** - We will credit you in our security advisory (unless you prefer to remain anonymous)

## Security Considerations

### MCP Protocol Security

As an MCP server, BrewSource handles external connections and data requests. Key security considerations include:

#### Connection Security

- **HTTP Connections** - Validate all incoming connections
- **Input Validation** - Sanitize all MCP tool parameters
- **Rate Limiting** - Prevent abuse through excessive requests
- **Error Handling** - Avoid exposing sensitive information in error messages

#### Data Protection

- **Database Security** - Use parameterized queries to prevent SQL injection
- **User Data** - Minimal data collection, secure storage practices
- **API Keys** - Secure handling of external API credentials
- **Logging** - Avoid logging sensitive information

### Current Security Measures

#### Input Validation

```go
// Example: Validate BJCP style codes
func validateStyleCode(code string) error {
    if code == "" {
        return errors.New("style code cannot be empty")
    }
    if len(code) > 10 {
        return errors.New("style code too long")
    }
    if !regexp.MustCompile(`^[0-9A-Z]+$`).MatchString(code) {
        return errors.New("invalid style code format")
    }
    return nil
}
```

#### Database Security

- Parameterized queries for all database operations
- Database connection pooling with limits
- Regular dependency updates via Dependabot

#### Error Handling

- Generic error messages to external clients
- Detailed logging for debugging (without sensitive data)
- Proper HTTP status codes

### Potential Security Risks

#### MCP-Specific Risks

- **Malicious Tool Calls** - Ensure tool parameters are validated
- **Resource Exhaustion** - Implement rate limiting and timeouts
- **Data Injection** - Validate all user inputs before database queries

#### Application-Specific Risks

- **Database Injection** - Use parameterized queries
- **External API Abuse** - Rate limit external API calls
- **Memory Exhaustion** - Limit response sizes and processing time

### Security Best Practices for Contributors

#### Code Security

- Always validate user inputs
- Use parameterized database queries
- Handle errors without exposing sensitive information
- Implement proper authentication and authorization
- Keep dependencies up to date

#### Example Secure Code Patterns

```go
// Good: Parameterized query
func (r *Repository) GetStyle(ctx context.Context, code string) (*Style, error) {
    var style Style
    err := r.db.GetContext(ctx, &style,
        "SELECT * FROM styles WHERE code = $1", code)
    return &style, err
}

// Bad: String concatenation (SQL injection risk)
func (r *Repository) GetStyleUnsafe(ctx context.Context, code string) (*Style, error) {
    query := "SELECT * FROM styles WHERE code = '" + code + "'"
    // This is vulnerable to SQL injection!
}
```

#### MCP Tool Security

```go
// Validate all tool parameters
func (h *ToolHandlers) BJCPLookup(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
    styleCode, ok := args["style_code"].(string)
    if !ok {
        return nil, &mcp.Error{
            Code:    mcp.InvalidParams,
            Message: "style_code must be a string",
        }
    }

    if err := validateStyleCode(styleCode); err != nil {
        return nil, &mcp.Error{
            Code:    mcp.InvalidParams,
            Message: "invalid style code",
        }
    }

    // Proceed with validated input
}
```

### Dependencies and Supply Chain Security

#### Dependency Management

- Regular dependency updates via Dependabot
- Security scanning of dependencies
- Minimal dependency footprint
- Prefer well-maintained, popular libraries

#### Go Module Security

```bash
# Check for known vulnerabilities
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Update dependencies
go get -u ./...
go mod tidy
```

### Deployment Security

#### Environment Configuration

- Never commit secrets to version control
- Use environment variables for sensitive configuration
- Implement proper secrets management
- Regular security updates for base images and OS

#### Database Security

- Use connection pooling with limits
- Implement database user with minimal required permissions
- Regular database backups with encryption
- Network isolation where possible

### Incident Response

#### In Case of a Security Incident

1. **Immediate Response**
   - Assess the scope and impact
   - Contain the vulnerability if possible
   - Document the incident

2. **Communication**
   - Notify users if data may be compromised
   - Coordinate with security researchers if applicable
   - Prepare public disclosure timeline

3. **Resolution**
   - Develop and test fix
   - Deploy security update
   - Monitor for ongoing issues

4. **Post-Incident**
   - Conduct post-mortem analysis
   - Update security practices
   - Improve monitoring and detection

### Security Contact

For security-related questions or concerns:

- **Email**: <security@brewsource-mcp.dev>
- **GitHub**: Create a private security advisory
- **Response Time**: Within 48 hours for acknowledgment

### Security Disclosure Policy

We believe in responsible disclosure and will work with security researchers to:

- Understand and verify security reports
- Develop appropriate fixes
- Coordinate public disclosure timing
- Provide credit for responsible reporting

### Regular Security Reviews

We conduct regular security reviews including:

- **Code Reviews** - All code changes reviewed for security implications
- **Dependency Audits** - Regular checks for vulnerable dependencies
- **Threat Modeling** - Ongoing assessment of potential attack vectors
- **Security Testing** - Automated and manual security testing

### Security Resources

- [OWASP Go Security Guide](https://owasp.org/www-project-go-secure-coding-practices-guide/)
- [Go Security Best Practices](https://golang.org/doc/security/best-practices)
- [MCP Security Considerations](https://modelcontextprotocol.io/docs/security)

---

**Last Updated**: July 5, 2025
**Version**: Phase 1 MVP

Thank you for helping keep BrewSource MCP Server secure! 🔒
