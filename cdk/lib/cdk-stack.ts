import {Architecture, DockerImageCode, DockerImageFunction} from "aws-cdk-lib/aws-lambda";
import {Duration, RemovalPolicy, Stack, StackProps} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import {LambdaIntegration, RestApi} from 'aws-cdk-lib/aws-apigateway';
import {AttributeType, BillingMode, Table} from 'aws-cdk-lib/aws-dynamodb';
import {Effect, PolicyStatement} from 'aws-cdk-lib/aws-iam';
import * as dotenv from 'dotenv';

dotenv.config({path: '../.env'});

export class CdkStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // DynamoDB Table
    const dynamoTable = new Table(this, 'ResourceTable', {
      partitionKey: { name: 'PK', type: AttributeType.STRING },
      sortKey: { name: 'SK', type: AttributeType.STRING },
      tableName: process.env.DYNAMO_TABLE_NAME,
      billingMode: BillingMode.PAY_PER_REQUEST,
      removalPolicy: RemovalPolicy.DESTROY,
    });

    // API Gateway
    const api = new RestApi(this, 'CleanServerlessBookSampleApi', {
      restApiName: 'CleanServerlessBookSampleAPI',
      deployOptions: {
        stageName: 'dev',
      },
    });

    // Lambda Functions and API Gateway Integrations
    const imagePath = "../app"
    const createLambdaFunction = (target: string, functionName: string) => {
      return new DockerImageFunction(this, functionName, {
        functionName: `clean-serverless-${functionName}`,
        code: DockerImageCode.fromImageAsset(imagePath, {
          target: target,
        }),
        architecture: Architecture.ARM_64,
        timeout: Duration.seconds(30),
        memorySize: 1280,
        environment: {
          DYNAMO_TABLE_NAME: process.env.DYNAMO_TABLE_NAME || '',
          DYNAMO_PK_NAME: process.env.DYNAMO_PK_NAME || '',
          DYNAMO_SK_NAME: process.env.DYNAMO_SK_NAME || '',
        }
      });
    };

    const addApiIntegration = (path: string, method: string, lambdaFunction: DockerImageFunction) => {
      const integration = new LambdaIntegration(lambdaFunction);
      api.root.resourceForPath(path).addMethod(method, integration);
    };

    // Define all Lambda functions
    const functions = [
      { name: 'deleteMicropost', method: 'DELETE', apiPath: '/v1/users/{user_id}/microposts/{micropost_id}' },
      { name: 'deleteUser', method: 'DELETE', apiPath: '/v1/users/{user_id}' },
      { name: 'getMicropost', method: 'GET', apiPath: '/v1/users/{user_id}/microposts/{micropost_id}' },
      { name: 'getMicroposts', method: 'GET', apiPath: '/v1/users/{user_id}/microposts' },
      { name: 'getUser', method: 'GET', apiPath: '/v1/users/{user_id}' },
      { name: 'getUsers', method: 'GET', apiPath: '/v1/users' },
      { name: 'postMicroposts', method: 'POST', apiPath: '/v1/users/{user_id}/microposts' },
      { name: 'postUsers', method: 'POST', apiPath: '/v1/users' },
      { name: 'putMicropost', method: 'PUT', apiPath: '/v1/users/{user_id}/microposts/{micropost_id}' },
      { name: 'putUser', method: 'PUT', apiPath: '/v1/users/{user_id}' }
    ];

    // Create Lambda functions and integrate them with API Gateway
    functions.forEach(({ name, apiPath, method }) => {
      const lambdaFunction = createLambdaFunction('api', name);
      dynamoTable.grantFullAccess(lambdaFunction);
      lambdaFunction.addToRolePolicy(new PolicyStatement({
        actions: ['dynamodb:*', 'logs:*'],
        effect: Effect.ALLOW,
        resources: ['*'],
      }));
      addApiIntegration(apiPath, method, lambdaFunction);
    });
  }
}
