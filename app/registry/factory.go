package registry

import (
	"clean-serverless-book-sample/adapter"
	"clean-serverless-book-sample/domain"
	"clean-serverless-book-sample/interactor"
	"clean-serverless-book-sample/usecase"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Factory 様々なインスタンスを生成する構造体
type Factory struct {
	Envs *Envs
}

// GetFactory Factoryのインスタンスを取得する
func GetFactory() *Factory {
	return &Factory{
		Envs: NewEnvs(),
	}
}

// BuildDynamoClient DynamoDBに接続するためのインスタンスを生成
func (f *Factory) BuildDynamoClient() *adapter.DynamoClient {
	config := &aws.Config{
		Region: aws.String("ap-northeast-1"),
	}

	if f.Envs.DynamoLocalEndpoint() != "" {
		config.Credentials = credentials.NewStaticCredentials("dummy", "dummy", "dummy")
		config.Endpoint = aws.String(f.Envs.DynamoLocalEndpoint())
	}
	return adapter.NewClient(config)
}

// BuildResourceTableOperator DynamoDBのテーブルに接続するためのインスタンスを生成
func (f *Factory) BuildResourceTableOperator() *adapter.ResourceTableOperator {
	return adapter.NewResourceTableOperator(
		f.BuildDynamoClient(),
		f.Envs.DynamoTableName())
}

// BuildDynamoModelMapper ModelからDynamoDBに保存する形式に変換するためのインスタンスを生成
func (f *Factory) BuildDynamoModelMapper() *adapter.DynamoModelMapper {
	return &adapter.DynamoModelMapper{
		Client:    f.BuildResourceTableOperator(),
		TableName: f.Envs.DynamoTableName(),
		PKName:    f.Envs.DynamoPKName(),
		SKName:    f.Envs.DynamoSKName(),
	}
}

// BuildUserEmailUniqGenerator ユーザーのメールアドレス重複チェック用のレコード生成機のインスタンスを生成
func (f *Factory) BuildUserEmailUniqGenerator() *adapter.UserEmailUniqGenerator {
	return adapter.NewUserEmailUniqGenerator(
		f.BuildDynamoModelMapper(),
		f.BuildResourceTableOperator(),
		f.Envs.DynamoPKName(),
		f.Envs.DynamoSKName())
}

// BuildUserOperator ユーザー情報関連の操作を行うインスタンスを生成
func (f *Factory) BuildUserOperator() domain.UserRepository {
	return &adapter.UserOperator{
		Client:                 f.BuildResourceTableOperator(),
		Mapper:                 f.BuildDynamoModelMapper(),
		UserEmailUniqGenerator: f.BuildUserEmailUniqGenerator(),
	}
}

// BuildUserEmailUniqChecker ユーザーのメールアドレス重複チェックインスタンスを生成
func (f *Factory) BuildUserEmailUniqChecker() *domain.UserEmailUniqChecker {
	return domain.NewUserEmailUniqChecker(f.BuildUserOperator())
}

// BuildMicropostOperator マイクロポスト情報関連の操作を行うインスタンスを生成
func (f *Factory) BuildMicropostOperator() *adapter.MicropostOperator {
	return &adapter.MicropostOperator{
		Client: f.BuildResourceTableOperator(),
		Mapper: f.BuildDynamoModelMapper(),
	}
}

// BuildCreateUser ユーザー作成UseCaseインスタンスを生成
func (f *Factory) BuildCreateUser() usecase.ICreateUser {
	return interactor.NewCreateUser(
		f.BuildUserOperator(),
		f.BuildUserEmailUniqChecker())
}

// BuildUpdateUser ユーザー更新UseCaseインスタンスを生成
func (f *Factory) BuildUpdateUser() usecase.IUpdateUser {
	return interactor.NewUpdateUser(
		f.BuildUserOperator(),
		f.BuildUserEmailUniqChecker())
}

// BuildGetUserList ユーザー取得UseCaseインスタンスを生成
func (f *Factory) BuildGetUserList() usecase.IGetUserList {
	return interactor.NewGetUserList(f.BuildUserOperator())
}

// BuildGetUserByID ユーザー取得UseCaseインスタンスを生成
func (f *Factory) BuildGetUserByID() usecase.IGetUserByID {
	return interactor.NewGetUserByID(f.BuildUserOperator())
}

// BuildUserDeleter ユーザー削除Usecaseインスタンスを生成
func (f *Factory) BuildUserDeleter() usecase.IDeleteUser {
	return interactor.NewUserDeleter(
		f.BuildUserOperator(),
		f.BuildGetUserByID())
}

// BuildCreateMicropost マイクロポスト作成UseCaseインスタンスを生成
func (f *Factory) BuildCreateMicropost() usecase.ICreateMicropost {
	return interactor.NewCreateMicropost(
		f.BuildMicropostOperator())
}

// BuildGetMicropostList マイクロポスト取得UseCaseインスタンスを生成
func (f *Factory) BuildGetMicropostList() usecase.IGetMicropostList {
	return interactor.NewGetMicropostList(
		f.BuildMicropostOperator())
}

// BuildGetMicropostByID マイクロポスト取得UseCaseインスタンスを生成
func (f *Factory) BuildGetMicropostByID() usecase.IGetMicropostByID {
	return interactor.NewGetMicropostByID(
		f.BuildMicropostOperator())
}

// BuildUpdateMicropost マイクロポスト更新UseCaseインスタンスを生成
func (f *Factory) BuildUpdateMicropost() usecase.IUpdateMicropost {
	return interactor.NewUpdateMicropost(
		f.BuildMicropostOperator())
}

// BuildDeleteMicropost マイクロポスト削除UseCaseインスタンスを生成
func (f *Factory) BuildDeleteMicropost() usecase.IDeleteMicropost {
	return interactor.NewDeleteMicropost(
		f.BuildGetMicropostByID(),
		f.BuildMicropostOperator())
}
