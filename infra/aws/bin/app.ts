import * as cdk from "aws-cdk-lib";
import { OidcProviderStack } from "../lib/oidc-provider-stack";

const app = new cdk.App();

new OidcProviderStack(app, "YossidOidcProviderStack", {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: "ap-northeast-1",
  },
  description: "YossID OIDC Provider — Lambda + API Gateway HTTP API",
});
