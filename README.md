# yossid OIDC Provider

yossid is a serverless OpenID Connect (OIDC) provider built with Go (Echo + Ory Fosite) and deployed to AWS Lambda and API Gateway, backed by serverless Neon PostgreSQL.

## Directory Structure

- **[`backend/`](file:///Users/ytakahashi/app/yossid/backend)**: Go Echo application code, OIDC core endpoint logic, database repositories, migrations, and local Docker compose file.
- **[`infra/`](file:///Users/ytakahashi/app/yossid/infra)**:
  - [`terraform/`](file:///Users/ytakahashi/app/yossid/infra/terraform): Neon Postgres project provisioning.
  - [`aws/`](file:///Users/ytakahashi/app/yossid/infra/aws): AWS CDK TypeScript stack (Lambda, API Gateway, IAM, SSM).
  - [`runbooks/`](file:///Users/ytakahashi/app/yossid/infra/runbooks): Automation scripts for configuration setup.

## Quick Start (Local Development)

### 1. Start Local Database
Run the PostgreSQL database container:
```bash
cd backend
make db-up
```

### 2. Run Database Migrations
Create the database tables locally:
```bash
make migrate
```

### 3. Start Local Server
Start the HTTP development server on port `8080`:
```bash
make run-local
```

---

For cloud provisioning and AWS deployment steps, refer to:
- [Infrastructure Setup & Deploy Guide](file:///Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/infra_setup_guide.md)
- [infra/terraform/README.md](file:///Users/ytakahashi/app/yossid/infra/terraform/README.md)
- [infra/aws/README.md](file:///Users/ytakahashi/app/yossid/infra/aws/README.md)
