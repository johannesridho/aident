package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/sns"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

type Configuration struct {
	Region                 string
	TargetSnsTopicArn      string
	FbMessengerAccessToken string
}

type SnsMessage struct {
	JobId string `json:"JobId"`
}

type GetMessageCreativeIdRequest struct {
	Messages []interface{} `json:"messages"`
}

type SimpleFbMessengerText struct {
	Text string `json:"text"`
}

type GetMessageCreativeIdResponse struct {
	MessageCreativeId string `json:"message_creative_id"`
}

type BroadcastMessageRequest struct {
	MessageCreativeId string `json:"message_creative_id"`
	MessagingType     string `json:"messaging_type"`
	Tag               string `json:"tag"`
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
		Region:                 os.Getenv("REGION"),
		TargetSnsTopicArn:      os.Getenv("TARGET_SNS_TOPIC_ARN"),
		FbMessengerAccessToken: os.Getenv("FB_MESSENGER_ACCESS_TOKEN"),
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
	strBuilder.WriteString(fmt.Sprintf("Job Id: %s\n", jobId))

	if len(criminalFacesMap) == 0 {
		strBuilder.WriteString("There is no criminal suspect in this video")
	} else {
		strBuilder.WriteString("Detected criminal suspect:\n")
		for key, val := range criminalFacesMap {
			strBuilder.WriteString(fmt.Sprintf("name: %s - similarity: %f\n", key, val))
		}
	}

	message := strBuilder.String()
	log.Println(message)
	publishToSns(config, message)
	broadcastToFbMessenger(config, message)

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
	log.Print("start publish message to SNS")

	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region)})
	if err != nil {
		return err
	}

	input := sns.PublishInput{Message: &message, TopicArn: &config.TargetSnsTopicArn}

	snsClient := sns.New(session)
	snsClient.Publish(&input)

	log.Print("finished publish message to SNS")

	return nil
}

func broadcastToFbMessenger(config Configuration, message string) {
	log.Print("start broadcast message to FB Messenger")

	messageCreativeId, err := getMessageCreativeId(config, message)
	if err != nil {
		log.Fatal(err)
	}

	broadcastMessageUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/broadcast_messages?access_token=%s",
		config.FbMessengerAccessToken,
	)

	req := BroadcastMessageRequest{
		MessageCreativeId: messageCreativeId,
		MessagingType:     "MESSAGE_TAG",
		Tag:               "NON_PROMOTIONAL_SUBSCRIPTION",
	}

	log.Println("cret", messageCreativeId)

	payload, err := json.Marshal(req)
	res, err := http.Post(broadcastMessageUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	responseBytes, err := ioutil.ReadAll(res.Body)
	log.Printf("received broadcast response: %s, status code: %d", string(responseBytes), res.StatusCode)

	log.Print("finished broadcast message to FB Messenger")
}

func getMessageCreativeId(config Configuration, message string) (string, error) {
	getMessageCreativeIdUrl := fmt.Sprintf(
		"https://graph.facebook.com/v2.11/me/message_creatives?access_token=%s",
		config.FbMessengerAccessToken,
	)

	simpleText := SimpleFbMessengerText{Text: message}

	messages := []interface{}{simpleText}

	req := GetMessageCreativeIdRequest{Messages: messages}

	payload, err := json.Marshal(req)

	res, err := http.Post(getMessageCreativeIdUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	getMessageCreativeIdResponse := GetMessageCreativeIdResponse{}
	err = json.NewDecoder(res.Body).Decode(&getMessageCreativeIdResponse)
	if err != nil {
		return "", err
	}

	log.Printf(
		"received messageCreativeId: %s, status code: %d",
		getMessageCreativeIdResponse.MessageCreativeId, res.StatusCode,
	)

	return getMessageCreativeIdResponse.MessageCreativeId, nil
}
