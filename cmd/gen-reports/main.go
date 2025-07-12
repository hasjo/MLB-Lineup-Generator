package main

import (
	"bytes"
	"io"
	"log"
	"path/filepath"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hasjo/MLBLG/pkg"
)

const BucketName = "hasjo-mlb-reports"

func pushFiletoS3(bucketname string, keyname string, buf bytes.Buffer){
	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	bufReader := bytes.NewReader(buf.Bytes())
	svc := s3.New(session)
	svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketname),
		Key: aws.String(keyname),
		Body: io.ReadSeeker(bufReader),
	})
}

func checkObject(bucketname string, keyname string) bool{
	session, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	svc := s3.New(session)
	svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketname),
		Key: aws.String(keyname),
	})
	if err != nil {
		return false
	}
	return true
}

func runLambda(){
	config := pkg.ConfigData{
		WatchTeams: pkg.AllTeams,
		ReportPath: "",
		ReceiptPath: filepath.Join("", "receipts"),
		PagePath: filepath.Join("", "page"),
	}
	data := pkg.GenerateFullReport(config, false)
	for _, report := range data{
		if report.Live == true {
			receiptpath := filepath.Join("receipt", report.Filename)
			pagepath := filepath.Join("page", report.Filename)
			receiptData := pkg.GenerateReceiptPDF(report.ReceiptData, config)
			pageData := pkg.GeneratePagePDF(report.PageData, config)

			if checkObject(BucketName, receiptpath) == false {
				pushFiletoS3(BucketName, receiptpath, receiptData)
				log.Printf("Pushing %s to s3", receiptpath)
			} else {
				log.Printf("Report exists in s3: %s", receiptpath)
			}
			if checkObject(BucketName, pagepath) == false {
				pushFiletoS3(BucketName, pagepath, pageData)
				log.Printf("Pushing %s to s3", pagepath)
			} else {
				log.Printf("Report exists in s3: %s", pagepath)
			}
		}
	}
}

func main() {
	lambda.Start(runLambda)
}
