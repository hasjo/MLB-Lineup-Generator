package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hasjo/MLBLG/pkg"
)

const BucketName = "hasjo-mlb-reports"

func buildReplaceMap() map[string]string{
	returnMap := make(map[string]string)
	for _, team := range(pkg.AllTeams){
		workstring := team
		workstring = strings.ReplaceAll(workstring, " ", "-")
		matchstring := workstring + "at"
		replacestring := workstring + "-at-"
		returnMap[matchstring] = replacestring
	}
	return returnMap
}


func getObjects(ctx context.Context, config aws.Config, bucketname string) {
	replaceMap := buildReplaceMap()
	svc := s3.NewFromConfig(config)
	objects, err := svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketname),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range(objects.Contents){
		if strings.HasPrefix(*item.Key, "receipt/") || strings.HasPrefix(*item.Key, "page/") {
			replacestring := *item.Key
			fmt.Printf("Checking %s...\n", replacestring)
			for key, value := range(replaceMap){
				replacestring = strings.ReplaceAll(replacestring, key, value)
			}
			if replacestring != *item.Key {
				fmt.Printf("STRING REPLACED: FROM %s to %s\n", *item.Key, replacestring)
				_, err := svc.CopyObject(ctx, &s3.CopyObjectInput{
					Bucket: aws.String(bucketname),
					Key: aws.String(replacestring),
					CopySource: aws.String(bucketname+"/"+*item.Key),
				})
				if err != nil {
					log.Fatal(err)
				}
				svc.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(bucketname),
					Key: aws.String(*item.Key),
				})
			}
		}
	}
}

func main() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil { log.Fatal(err) }
	getObjects(ctx, sdkConfig, BucketName)
}
