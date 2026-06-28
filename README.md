# yossid OIDC Provider

yossid is a serverless OpenID Connect (OIDC) provider built with Go (Echo + Ory Fosite) and deployed to AWS Lambda and API Gateway, backed by serverless Neon PostgreSQL.

## Directory Structure

- **[`backend/`](file:///Users/ytakahashi/app/yossid/backend)**: Go Echo application code, OIDC core endpoint logic, database repositories, migrations, and local Docker compose file.
- **[`infra/`](file:///Users/ytakahashi/app/yossid/infra)**:
  - [`terraform/`](file:///Users/ytakahashi/app/yossid/infra/terraform): Neon Postgres project provisioning.
  - [`aws/`](file:///Users/ytakahashi/app/yossid/infra/aws): AWS CDK TypeScript stack (Lambda, API Gateway, IAM, SSM).
  - [`runbooks/`](file:///Users/ytakahashi/app/yossid/infra/runbooks): Automation scripts for configuration setup.

## Quick Start (Local Development via Makefile & Docker Compose)

We use a root-level `Makefile` to run shorthand commands that orchestrate the Docker Compose environment.

### 1. Build and Start the Stack
From the project root directory, run:
```bash
make up
```
This builds and starts:
- **Next.js Frontend**: Accessible at [http://localhost:3000](http://localhost:3000) (hot-reloaded)
- **Go Backend Server**: Accessible at [http://localhost:8080](http://localhost:8080)
- **PostgreSQL Database**: Port mapping `5433:5432`

### 2. Run Database Migrations
Initialize the local database tables:
```bash
make migrate
```

### 3. Restart Go Backend
If you modify backend Go source files, run this to compile and restart the backend server container:
```bash
make restart
```

### 4. View Service Logs
Tail logs from all running containers:
```bash
make logs
```

### 5. Run Tests & Linting
Run backend unit tests and linter inside the Docker environment:
```bash
make test
make vet
```

### 6. Stop the Stack
Tear down the running containers:
```bash
make down
```

---

For cloud provisioning and AWS deployment steps, refer to:
- [Infrastructure Setup & Deploy Guide](file:///Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/infra_setup_guide.md)
- [infra/terraform/README.md](file:///Users/ytakahashi/app/yossid/infra/terraform/README.md)
- [infra/aws/README.md](file:///Users/ytakahashi/app/yossid/infra/aws/README.md)
