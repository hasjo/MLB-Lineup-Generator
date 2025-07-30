package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hasjo/MLBLG/pkg"
)

const BucketName = "scorecards.jjhsk.com"

type TemplateData struct {
	Title string
	Contents string
}


func pushFiletoS3(bucketname string, keyname string, buf bytes.Buffer, contentType string){
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	svc := s3.NewFromConfig(sdkConfig)
	if err != nil {
		log.Fatal(err)
	}
	bufReader := bytes.NewReader(buf.Bytes())
	putObjInput := &s3.PutObjectInput{
			Bucket: aws.String(bucketname),
			Key: aws.String(keyname),
			Body: io.ReadSeeker(bufReader),
	}
	if contentType != "" {
		putObjInput.ContentType = aws.String(contentType)
	}
	svc.PutObject(
		ctx,
		putObjInput,
	)
}

func checkObject(bucketname string, keyname string) bool{
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	svc := s3.NewFromConfig(sdkConfig)
	_, err = svc.HeadObject(
		ctx, 
		&s3.HeadObjectInput{
			Bucket: aws.String(bucketname),
			Key: aws.String(keyname),
		},
	)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func generateListPage(dirname string, title string) bytes.Buffer{
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	svc := s3.NewFromConfig(sdkConfig)
	var dataSlice []types.Object
	listdata, err := svc.ListObjectsV2(
		ctx,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(BucketName),
			Prefix: aws.String(dirname),
		})
	if err != nil {
		log.Fatal(err)
	}
	dataSlice = append(dataSlice, listdata.Contents...)
	for *listdata.IsTruncated {
		listdata, err = svc.ListObjectsV2(
			ctx,
			&s3.ListObjectsV2Input{
				Bucket: aws.String(BucketName),
				ContinuationToken: aws.String(*listdata.NextContinuationToken),
				Prefix: aws.String(dirname),
			})
		dataSlice = append(dataSlice, listdata.Contents...)
	}
	var contentstring string
	for _, thing := range(dataSlice){
		keyname := strings.ReplaceAll(*thing.Key, dirname, "")
		contentstring = fmt.Sprintf("<a href=\"%s%s\">%s</a></br>\n",dirname, keyname, keyname) + contentstring
	}
	file, err := os.ReadFile("list.html")
	if err != nil {
		log.Fatal(err)
	}
	filestring := string(file)
	tmpl, err := template.New("test").Parse(filestring)
	if err != nil {
		log.Fatal(err)
	}
	data := TemplateData{title, contentstring}
	var retBuff bytes.Buffer
	err = tmpl.Execute(&retBuff, data)
	return retBuff
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
		log.Printf("Found %s - Live: %t,  Monitoring...", report.Filename, report.Live)
		if report.Live == true {
			receiptpath := filepath.Join("receipt", report.Filename)
			pagepath := filepath.Join("page", report.Filename)
			receiptData := pkg.GenerateReceiptPDF(report.ReceiptData, config)
			pageData := pkg.GeneratePagePDF(report.PageData, config)

			if checkObject(BucketName, receiptpath) == false {
				pushFiletoS3(BucketName, receiptpath, receiptData, "")
				log.Printf("Pushing %s to s3", receiptpath)
			} else {
				log.Printf("Report exists in s3: %s", receiptpath)
			}
			if checkObject(BucketName, pagepath) == false {
				pushFiletoS3(BucketName, pagepath, pageData, "")
				log.Printf("Pushing %s to s3", pagepath)
			} else {
				log.Printf("Report exists in s3: %s", pagepath)
			}
		}
	}
	pagePage := generateListPage("page/", "PAGES")
	pushFiletoS3(BucketName, "page.html", pagePage, "text/html")
	receiptPage := generateListPage("receipt/", "RECEIPTS")
	pushFiletoS3(BucketName, "receipt.html", receiptPage, "text/html")
}

func main() {
	lambda.Start(runLambda)
}
