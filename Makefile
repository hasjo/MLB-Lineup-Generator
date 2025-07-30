build:
	CGO_ENABLED=0 go build -o bootstrap cmd/gen-reports/main.go
	cp /usr/share/fonts/liberation/LiberationMono-Regular.ttf LiberationMono-Regular.ttf
	zip lambda-handler.zip bootstrap LiberationMono-Regular.ttf
	rm LiberationMono-Regular.ttf
	rm bootstrap

deploy: build
	aws s3 cp lambda-handler.zip s3://hasjo-lambda-zip-bucket/go-handler.zip
	rm lambda-handler.zip

deploy-index:
	aws s3 cp site/index.html s3://scorecards.jjhsk.com/index.html
	aws s3 cp site/baseball.gif s3://scorecards.jjhsk.com/baseball.gif

update-lambda: deploy
	aws lambda update-function-code --function-name hasjo-scorecard-report-generator --s3-bucket hasjo-lambda-zip-bucket --s3-key go-handler.zip

deploy-cf:
	aws cloudformation deploy --template-file cf.yml --stack-name scorecard-site-and-report-stack --capabilities CAPABILITY_IAM

deploy-zipbucket:
	aws cloudformation deploy --template-file zipbucket.yml --stack-name lambda-zip-bucket --capabilities CAPABILITY_IAM

deploy-static-site:
	aws cloudformation deploy --template-file site.yml --stack-name scorecard-report-site
