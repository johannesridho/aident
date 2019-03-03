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

type Configuration struct {
	Region       string
	BucketName   string
	SnsTopic     string
	RekRole      string
	CollectionId string
}

func main() {
	lambda.Start(StartCrimeDetectionHandler)
}

func StartCrimeDetectionHandler(ctx context.Context, event events.S3Event) (string, error) {
	log.Printf("start handling event: %v", event)

	videoFileName := event.Records[0].S3.Object.Key
	log.Printf("S3 video file name: %v", videoFileName)

	config := Configuration{
		Region:       os.Getenv("REGION"),
		BucketName:   os.Getenv("S3_BUCKET_NAME"),
		SnsTopic:     os.Getenv("SNS_TOPIC_ARN"),
		RekRole:      os.Getenv("REKOGNITION_ROLE_ARN"),
		CollectionId: os.Getenv("COLLECTION_ID"),
	}

	video := rekognition.Video{
		S3Object: &rekognition.S3Object{
			Bucket: &config.BucketName,
			Name:   &videoFileName,
		},
	}

	notificationChannel := rekognition.NotificationChannel{SNSTopicArn: &config.SnsTopic, RoleArn: &config.RekRole}

	labelDetectionJobId, err := startLabelDetection(config, video, notificationChannel)
	if err != nil {
		log.Fatal(err)
	}

	faceSearchJobId, err := startFaceSearch(config, video, notificationChannel)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("label detection job id = %s", labelDetectionJobId)
	log.Printf("start face search job id = %s", faceSearchJobId)

	return "success", nil
}

func startLabelDetection(config Configuration, video rekognition.Video, notificationChannel rekognition.NotificationChannel) (string, error) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return "", err
	}

	rek := rekognition.New(session)

	minConfidence := float64(50)
	jobTag := "AidentStartLabelDetection"

	input := rekognition.StartLabelDetectionInput{
		Video:               &video,
		MinConfidence:       &minConfidence,
		JobTag:              &jobTag,
		NotificationChannel: &notificationChannel,
	}

	output, err := rek.StartLabelDetection(&input)
	if err != nil {
		return "", err
	}

	jobId := *output.JobId

	return jobId, nil
}

func startFaceSearch(config Configuration, video rekognition.Video, notificationChannel rekognition.NotificationChannel) (string, error) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return "", err
	}

	rek := rekognition.New(session)

	jobTag := "AidentStartFaceSearch"

	input := rekognition.StartFaceSearchInput{
		Video:               &video,
		CollectionId:        &config.CollectionId,
		JobTag:              &jobTag,
		NotificationChannel: &notificationChannel,
	}

	output, err := rek.StartFaceSearch(&input)
	if err != nil {
		return "", err
	}

	jobId := *output.JobId

	return jobId, nil
}
