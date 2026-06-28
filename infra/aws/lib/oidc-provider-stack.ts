import * as path from "path";
import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigwv2 from "aws-cdk-lib/aws-apigatewayv2";
import * as integrations from "aws-cdk-lib/aws-apigatewayv2-integrations";
import * as iam from "aws-cdk-lib/aws-iam";
import * as logs from "aws-cdk-lib/aws-logs";

// ── SSM Parameter names (values are set manually via runbook/setup-ssm.sh) ──
const SSM_PREFIX = "/yossid";
const SSM_PARAMS = {
  databaseUrl: `${SSM_PREFIX}/database/url`,
  jwtPrivateKey: `${SSM_PREFIX}/jwt/private-key`,
  cookieSigningKey: `${SSM_PREFIX}/cookie/signing-key`,
  tokenPepper: `${SSM_PREFIX}/token/pepper`,
  otpPepper: `${SSM_PREFIX}/otp/pepper`,
  sesFromEmail: `${SSM_PREFIX}/ses/from-email`,
};

export class OidcProviderStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // ── CloudWatch Log Group ─────────────────────────────────────────────────
    const logGroup = new logs.LogGroup(this, "OidcProviderLogGroup", {
      logGroupName: "/aws/lambda/yossid-oidc-provider",
      retention: logs.RetentionDays.ONE_MONTH,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });

    // ── IAM Role ────────────────────────────────────────────────────────────
    const lambdaRole = new iam.Role(this, "OidcProviderLambdaRole", {
      roleName: "yossid-oidc-provider-lambda-role",
      assumedBy: new iam.ServicePrincipal("lambda.amazonaws.com"),
      description: "Execution role for yossid OIDC Provider Lambda",
    });

    // CloudWatch Logs
    lambdaRole.addManagedPolicy(
      iam.ManagedPolicy.fromAwsManagedPolicyName(
        "service-role/AWSLambdaBasicExecutionRole",
      ),
    );

    // SSM: GetParameter for each secret (principle of least privilege)
    lambdaRole.addToPolicy(
      new iam.PolicyStatement({
        sid: "SsmGetParameters",
        effect: iam.Effect.ALLOW,
        actions: ["ssm:GetParameter"],
        resources: Object.values(SSM_PARAMS).map(
          (name) =>
            `arn:aws:ssm:${this.region}:${this.account}:parameter${name}`,
        ),
      }),
    );

    // SSM: Decrypt SecureString values with default AWS managed key
    lambdaRole.addToPolicy(
      new iam.PolicyStatement({
        sid: "KmsDecryptSsm",
        effect: iam.Effect.ALLOW,
        actions: ["kms:Decrypt"],
        resources: [`arn:aws:kms:${this.region}:${this.account}:alias/aws/ssm`],
      }),
    );

    // SES: Send emails (OTP / email verification)
    lambdaRole.addToPolicy(
      new iam.PolicyStatement({
        sid: "SesSendEmail",
        effect: iam.Effect.ALLOW,
        actions: ["ses:SendEmail", "ses:SendRawEmail"],
        resources: ["*"], // Restrict to SES identity ARN once email is verified
      }),
    );

    // ── API Gateway HTTP API ─────────────────────────────────────────────────
    const httpApi = new apigwv2.HttpApi(this, "OidcProviderApi", {
      apiName: "yossid-oidc-provider-api",
      description: "yossid OIDC Provider HTTP API",
      // Default stage ($default) → shortest URL
      createDefaultStage: true,
    });

    // ── Lambda Function ──────────────────────────────────────────────────────
    // Docker bundling: compiles Go inside a container for reproducible Linux binary.
    // The `bootstrap` binary name is required for provided.al2023 runtime.
    const fn = new lambda.DockerImageFunction(this, "OidcProviderFunction", {
      functionName: "yossid-oidc-provider",
      description: "yossid OIDC Provider — Echo on Go Lambda (provided.al2023)",
      role: lambdaRole,
      logGroup,
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      code: lambda.DockerImageCode.fromImageAsset(
        path.join(__dirname, "../../../backend"),
        {
          file: "Dockerfile",
          buildArgs: {},
        },
      ),
      environment: {
        // Issuer is set dynamically to the API Gateway endpoint
        ISSUER: httpApi.apiEndpoint,
        DATABASE_URL_PARAMETER: SSM_PARAMS.databaseUrl,
        JWT_PRIVATE_KEY_PARAMETER: SSM_PARAMS.jwtPrivateKey,
        COOKIE_SIGNING_KEY_PARAMETER: SSM_PARAMS.cookieSigningKey,
        TOKEN_PEPPER_PARAMETER: SSM_PARAMS.tokenPepper,
        OTP_PEPPER_PARAMETER: SSM_PARAMS.otpPepper,
        SES_FROM_EMAIL_PARAMETER: SSM_PARAMS.sesFromEmail,
      },
    });

    const lambdaIntegration = new integrations.HttpLambdaIntegration(
      "LambdaIntegration",
      fn,
    );

    // Route all requests to Lambda (proxy integration)
    httpApi.addRoutes({
      path: "/{proxy+}",
      methods: [apigwv2.HttpMethod.ANY],
      integration: lambdaIntegration,
    });

    // Root path (e.g. /healthz without proxy segment)
    httpApi.addRoutes({
      path: "/",
      methods: [apigwv2.HttpMethod.ANY],
      integration: lambdaIntegration,
    });

    // ── Outputs ──────────────────────────────────────────────────────────────
    new cdk.CfnOutput(this, "ApiEndpoint", {
      value: httpApi.apiEndpoint,
      description: "OIDC Provider API Gateway endpoint URL",
      exportName: "YossidOidcApiEndpoint",
    });

    new cdk.CfnOutput(this, "LambdaFunctionName", {
      value: fn.functionName,
      description: "Lambda function name",
    });

    new cdk.CfnOutput(this, "SsmParameterPrefix", {
      value: SSM_PREFIX,
      description:
        "SSM Parameter Store prefix — see infra/runbooks/setup-ssm.sh",
    });
  }
}
