package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/sns"
	"log"
	"math"
	"os"
	"strings"
)

type Configuration struct {
	Region            string
	TargetSnsTopicArn string
}

type SnsMessage struct {
	JobId string `json:"JobId"`
}

func main() {
	lambda.Start(FaceSearchProcessor)
}

func FaceSearchProcessor(ctx context.Context, event events.SNSEvent) (string, error) {
	log.Printf("start FaceSearchProcessor, event: %v", event)

	jsonMessage := event.Records[0].SNS.Message
	log.Printf("SNS message: %v", jsonMessage)

	snsMessage := SnsMessage{}
	json.Unmarshal([]byte(jsonMessage), &snsMessage)

	jobId := snsMessage.JobId
	log.Printf("Rekognition jobId: %s", jobId)

	config := Configuration{
		Region:            os.Getenv("REGION"),
		TargetSnsTopicArn: os.Getenv("TARGET_SNS_TOPIC_ARN"),
	}

	log.Printf("start GetFaceSearch")
	result, err := getFaceSearchResult(config, jobId)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("GetFaceSearch result acquired, status: %s", *result.JobStatus)

	criminalFacesMap := make(map[string]float64)

	for _, person := range result.Persons {
		if len(person.FaceMatches) != 0 {
			for _, faceMatch := range person.FaceMatches {
				val := criminalFacesMap[*faceMatch.Face.ExternalImageId]
				criminalFacesMap[*faceMatch.Face.ExternalImageId] = math.Max(*faceMatch.Similarity, val)
			}
		}
	}

	var strBuilder strings.Builder
	strBuilder.WriteString("Detected criminal suspect: \n")

	for key, val := range criminalFacesMap {
		strBuilder.WriteString(fmt.Sprintf("name: %s - similarity: %f\n", key, val))
	}

	message := strBuilder.String()
	log.Println(message)
	publishToSns(config, message)

	return "success", nil
}

func getFaceSearchResult(config Configuration, jobId string) (*rekognition.GetFaceSearchOutput, error) {
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return nil, err
	}

	rek := rekognition.New(session)

	input := rekognition.GetFaceSearchInput{JobId: &jobId}

	result, err := rek.GetFaceSearch(&input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func publishToSns(config Configuration, message string) error {
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return err
	}

	input := sns.PublishInput{Message: &message, TopicArn: &config.TargetSnsTopicArn}

	snsClient := sns.New(session)
	snsClient.Publish(&input)

	return nil
}
