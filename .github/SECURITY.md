# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version  | Supported          |
| -------- | ------------------ |
| latest   | :white_check_mark: |
| < latest | :x:                |

We recommend always running the latest version for security patches.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### Private Reporting (Preferred)

Report security vulnerabilities using [GitHub Security Advisories](https://github.com/Recipe-Web-App/user-management-service/security/advisories/new).

This allows us to:

- Discuss the vulnerability privately
- Develop and test a fix
- Coordinate disclosure timing
- Issue a CVE if necessary

### What to Include

When reporting a vulnerability, please include:

1. **Description** - Clear description of the vulnerability
2. **Impact** - What can an attacker achieve?
3. **Reproduction Steps** - Step-by-step instructions to reproduce
4. **Affected Components** - Which parts of the service are affected
5. **Suggested Fix** - If you have ideas for remediation
6. **Environment** - Version, configuration, deployment details
7. **Proof of Concept** - Code or requests demonstrating the issue (if safe to share)

### Example Report

```text
Title: JWT Token Signature Bypass

Description: The JWT validation does not properly verify signatures...

Impact: An attacker can forge tokens and gain unauthorized access...

Steps to Reproduce:
1. Create a JWT with algorithm "none"
2. Send to /api/v1/auth/userinfo
3. Token is accepted without signature verification

Affected: app/services/token_service.py line 45

Suggested Fix: Enforce algorithm whitelist and reject "none"

Environment: v1.0.0, Docker deployment
```

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Varies by severity (critical: days, high: weeks, medium: months)

## Severity Levels

### Critical

- Remote code execution
- Authentication bypass
- Privilege escalation to admin
- Mass data exposure

### High

- Token forgery/manipulation
- SQL injection
- Unauthorized access to user data
- Denial of service affecting all users

### Medium

- Information disclosure (limited)
- CSRF vulnerabilities
- Rate limiting bypass
- Session fixation

### Low

- Verbose error messages
- Security header issues
- Best practice violations

## Security Features

This service implements multiple security layers:

### Authentication & Authorization

- **JWT-based Authentication** - Secure token-based authentication
  - Access tokens: Short-lived (30 minutes)
  - Refresh tokens: Long-lived (7 days)
  - Password reset tokens: Time-limited (15 minutes)
- **bcrypt Password Hashing** - Industry-standard password storage
- **Role-Based Access Control** - USER/ADMIN roles
- **Session Management** - Redis-backed JWT session tracking

### Input Validation

- **Pydantic Models** - Strict request validation
- **SQL Injection Prevention** - Parameterized queries via SQLAlchemy
- **XSS Protection** - Input sanitization

### API Security

- **CORS Configuration** - Configurable allowed origins
- **Rate Limiting** - Protection against abuse (future enhancement)
- **Request ID Tracking** - Audit trail for requests
- **Sensitive Data Masking** - Automatic redaction in logs

### Database Security

- **Connection Encryption** - TLS for PostgreSQL connections
- **Principle of Least Privilege** - Minimal database permissions
- **Prepared Statements** - SQL injection prevention

### Transport Security

- **HTTPS Enforced** - TLS 1.2+ required in production
- **Secure Headers** - Security headers configured
- **HSTS Support** - HTTP Strict Transport Security

## Security Best Practices

### For Operators

1. **Environment Variables** - Never commit secrets to git

   ```bash
   # Use environment variables
   JWT_SECRET_KEY=<generate-secure-random-key>
   POSTGRES_PASSWORD=<strong-password>
   ```

2. **Database Security**
   - Use strong passwords for PostgreSQL
   - Enable SSL/TLS connections
   - Restrict network access to database
   - Regular backups with encryption

3. **Redis Security**
   - Set Redis password (`REDIS_PASSWORD`)
   - Bind to localhost or private network
   - Enable persistence with encryption

4. **JWT Configuration**
   - Use long, random secret keys (256+ bits)
   - Rotate JWT secrets periodically
   - Monitor for suspicious token usage

5. **Network Security**
   - Deploy behind reverse proxy (nginx, Traefik)
   - Use HTTPS in production
   - Configure firewall rules
   - Enable DDoS protection

### For Developers

1. **Code Security**
   - Run `bandit -r app/` before committing
   - Run `safety check` for dependency vulnerabilities
   - Use type hints everywhere
   - Validate all inputs

2. **Secret Management**
   - Use `.env` files (never commit!)
   - Use secret managers in production (AWS Secrets Manager, Vault)
   - Rotate secrets regularly
   - Never log secrets

3. **Dependency Management**
   - Keep dependencies updated
   - Review security advisories
   - Use `poetry update` cautiously
   - Monitor Dependabot PRs

4. **Testing**
   - Include security test cases
   - Test authentication flows
   - Test authorization boundaries
   - Fuzz testing for inputs

## Security Checklist

Before deploying to production:

- [ ] JWT secret key is strong and random (256+ bits)
- [ ] All secrets are in environment variables (not code)
- [ ] Database passwords are strong
- [ ] PostgreSQL SSL/TLS is enabled
- [ ] Redis password is set
- [ ] HTTPS is enforced
- [ ] CORS origins are properly configured
- [ ] Security headers are enabled
- [ ] Logging is configured (without sensitive data)
- [ ] Dependencies are up to date
- [ ] Security scans pass (bandit, safety)
- [ ] Rate limiting is configured (future)
- [ ] Backups are configured and encrypted
- [ ] Monitoring and alerting is set up

## Known Security Considerations

### JWT Token Storage

- **Access tokens** are short-lived (30 minutes) to limit exposure
- **Refresh tokens** should be stored securely by clients
- Tokens are tracked in Redis for revocation capability

### Password Reset Flow

- Reset tokens expire after 15 minutes
- Tokens are single-use only
- Email verification required

### Session Management

- Redis-backed session tracking
- Automatic cleanup of expired sessions
- Logout invalidates tokens immediately

### Database Queries

- All queries use SQLAlchemy ORM
- No raw SQL execution
- Parameterized queries prevent SQL injection

### Privacy Features

- User privacy preferences are enforced
- Sensitive data is masked in logs
- Data export capabilities for GDPR compliance

## Disclosure Policy

We follow coordinated disclosure:

1. **Report received** - Acknowledge within 48 hours
2. **Investigation** - Assess severity and impact
3. **Fix development** - Create and test patch
4. **Disclosure coordination** - Agree on timeline with reporter
5. **Release** - Deploy fix to production
6. **Public disclosure** - Publish advisory after fix is deployed

**Standard timeline**: 90 days from report to public disclosure

## Security Updates

Stay informed about security updates:

- Watch this repository for security advisories
- Subscribe to GitHub Security Advisories
- Check release notes for security fixes
- Follow conventional commit messages (type: `security`)

## Acknowledgments

We appreciate security researchers who report vulnerabilities responsibly.

Contributors who report valid security issues will be:

- Credited in security advisories (unless they prefer to remain anonymous)
- Acknowledged in release notes
- Added to this SECURITY.md file

### Security Hall of Fame

<!-- Security researchers will be listed here -->

## Contact

- **Security Issues**: [GitHub Security Advisories](https://github.com/Recipe-Web-App/user-management-service/security/advisories/new)
- **General Questions**: [GitHub Discussions](https://github.com/Recipe-Web-App/user-management-service/discussions)
- **Other Issues**: [GitHub Issues](https://github.com/Recipe-Web-App/user-management-service/issues)

---

Thank you for helping keep User Management Service and its users safe!
