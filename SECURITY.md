# Security Policy

## Supported Versions

Security updates are provided for the latest released version only. Please make
sure you are running the most recent release before reporting an issue.

| Version | Supported          |
|---------|--------------------|
| latest  | :white_check_mark: |
| older   | :x:                |

## Reporting a Vulnerability

Please **do not** report security vulnerabilities through public GitHub issues,
pull requests, or discussions.

Instead, report them privately using GitHub's
[private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability):

1. Go to the repository's **Security** tab.
2. Click **Report a vulnerability**.
3. Fill in the details of the issue.

Please include as much of the following as possible to help us triage quickly:

- A description of the vulnerability and its potential impact
- Steps to reproduce, including a minimal configuration if relevant
- The affected version, Go version, and deployment method (binary or Docker)
- Any known workarounds

## Response Process

- We aim to acknowledge new reports within **7 days**.
- We will keep you informed of the progress toward a fix and full announcement.
- Once a fix is available, we will publish a new release and, where
  appropriate, a GitHub Security Advisory crediting the reporter (unless you
  prefer to remain anonymous).

## Scope and Hardening Notes

This project is a Prometheus exporter that runs MongoDB queries and exposes the
results as metrics. When operating it, keep the following security
considerations in mind — these are deployment concerns rather than
vulnerabilities in the exporter itself:

- **Transport security:** The exporter's HTTP endpoints (`/prometheus`,
  `/health`, `/live`) do **not** support HTTPS. Run it behind a reverse proxy
  or within a trusted network segment, and do not expose the endpoints to
  untrusted networks.
- **Credentials:** The MongoDB connection string (via `mongodb.uri` or the
  `MONGODB_URI` environment variable) may contain credentials. Provide it
  through environment variables or a secret manager rather than committing it
  to source control, and never include real credentials in issues or pull
  requests.
- **Least privilege:** Configure the MongoDB user used by the exporter with
  read-only access to only the databases and collections it needs to query.
- **Query configuration:** The `find` and `aggregate` queries are defined in
  configuration. Treat the configuration file as trusted input and restrict who
  can modify it.

## Dependencies

Dependencies are monitored and updated via
[Dependabot](https://docs.github.com/en/code-security/dependabot). Security-
related dependency updates are prioritized.
