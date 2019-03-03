package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"log"
	"os"
)

func main() {
	lambda.Start(CrimeDetectionHandler)
}

func CrimeDetectionHandler(ctx context.Context, event events.S3Event) (string, error) {
	log.Printf("start handling event: %v", event)

	name := event.Records[0].S3.Object.Key
	log.Printf("S3 filename: %v", name)

	region := os.Getenv("REGION")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	snsTopic := os.Getenv("SNS_TOPIC_ARN")
	rekRole := os.Getenv("REKOGNITION_ROLE_ARN")

	session, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}

	rek := rekognition.New(session)

	video := rekognition.Video{
		S3Object: &rekognition.S3Object{
			Bucket: &bucketName,
			Name:   &name,
		},
	}

	minConfidence := float64(0)
	jobTag := "AidentCrimeDetection"

	notificationChannel := rekognition.NotificationChannel{SNSTopicArn: &snsTopic, RoleArn: &rekRole}

	input := rekognition.StartLabelDetectionInput{
		Video:               &video,
		MinConfidence:       &minConfidence,
		JobTag:              &jobTag,
		NotificationChannel: &notificationChannel,
	}

	output, err := rek.StartLabelDetection(&input)
	if err != nil {
		log.Fatal(err)
	}

	jobId := *output.JobId
	log.Printf("job id = %s", jobId)

	return "success", nil
}
