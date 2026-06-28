# AWS Infrastructure Deployment (AWS CDK)

This directory manages the AWS serverless stack (Lambda Function with Go, API Gateway HTTP API proxy, IAM Execution Roles, and CloudWatch Log Group) using the AWS Cloud Development Kit (CDK) in TypeScript.

## Prerequisites

1. Install Node.js & npm.
2. Authenticate AWS CLI (set `AWS_PROFILE` in your terminal).
3. Ensure SSM Parameters are populated under the `/yossid` namespace. You can automatically generate the keys and upload them by running:
   ```bash
   ../../infra/runbooks/setup-ssm.sh
   ```

## Usage

```bash
# 1. Install Node dependencies
npm install

# 2. Compile TypeScript
npm run build

# 3. Bootstrap CDK (first time only)
npx cdk bootstrap

# 4. Deploy the stack
npx cdk deploy
```
