build:
	CGO_ENABLED=0 go build -o bootstrap cmd/gen-reports/main.go
	cp /usr/share/fonts/liberation/LiberationMono-Regular.ttf LiberationMono-Regular.ttf
	zip lambda-handler.zip bootstrap LiberationMono-Regular.ttf
	rm LiberationMono-Regular.ttf
	rm bootstrap

deploy: build
	aws s3 cp lambda-handler.zip s3://hasjo-lambda-zip-bucket/go-handler.zip
	rm lambda-handler.zip

update-lambda: deploy
	aws lambda update-function-code --function-name hasjo-mlb-report-generator --s3-bucket hasjo-lambda-zip-bucket --s3-key go-handler.zip

deploy-cf:
	aws cloudformation deploy --template-file cf.yml --stack-name mlb-report-generator --capabilities CAPABILITY_IAM

deploy-zipbucket:
	aws cloudformation deploy --template-file zipbucket.yml --stack-name lambda-zip-bucket --capabilities CAPABILITY_IAM
