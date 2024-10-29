FROM golang:1.23 AS build
WORKDIR /go/src/clean-serverless-book-sample
COPY . /go/src/clean-serverless-book-sample
RUN scripts/build-handlers.sh


FROM public.ecr.aws/lambda/provided:al2023 AS api
COPY --from=build /go/src/clean-serverless-book-sample/adapter/handlers/api/main ./main
ENTRYPOINT [ "./main" ]
