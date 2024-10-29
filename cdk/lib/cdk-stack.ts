import {Architecture, DockerImageCode, DockerImageFunction} from "aws-cdk-lib/aws-lambda";
import { Duration, RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { RestApi, LambdaIntegration } from 'aws-cdk-lib/aws-apigateway';
import { AttributeType, Table, BillingMode } from 'aws-cdk-lib/aws-dynamodb';
import { Effect, PolicyStatement } from 'aws-cdk-lib/aws-iam';
import {Bucket, EventType} from 'aws-cdk-lib/aws-s3';
import { LambdaDestination } from 'aws-cdk-lib/aws-s3-notifications';
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

    // S3 Bucket
    const bucket = new Bucket(this, 'CleanServerlessTestBucket', {
      bucketName: 'clean-serverless-test',
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

    // Create S3 Event Handler Lambda
    const s3HandlerFunction = createLambdaFunction('s3event', 's3Handler');

    // Grant S3 permissions to the handler
    bucket.grantReadWrite(s3HandlerFunction);

    // Add DynamoDB permissions
    dynamoTable.grantFullAccess(s3HandlerFunction);

    // Add additional permissions if needed
    s3HandlerFunction.addToRolePolicy(new PolicyStatement({
      actions: ['dynamodb:*', 'logs:*'],
      effect: Effect.ALLOW,
      resources: ['*'],
    }));

    // Add S3 event notification
    bucket.addEventNotification(
        EventType.OBJECT_CREATED,
        new LambdaDestination(s3HandlerFunction)
    );
    bucket.addEventNotification(
        EventType.OBJECT_REMOVED,
        new LambdaDestination(s3HandlerFunction)
    );
  }
}
