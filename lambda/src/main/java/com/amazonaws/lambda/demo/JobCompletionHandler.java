//Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//PDX-License-Identifier: MIT-0 (For details, see https://github.com/awsdocs/amazon-rekognition-developer-guide/blob/master/LICENSE-SAMPLECODE.)

package com.amazonaws.lambda.demo;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.LambdaLogger;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.SNSEvent;

import java.util.ArrayList;
import java.util.List;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.rekognition.AmazonRekognition;
import com.amazonaws.services.rekognition.AmazonRekognitionClientBuilder;
import com.amazonaws.services.rekognition.model.GetLabelDetectionRequest;
import com.amazonaws.services.rekognition.model.GetLabelDetectionResult;
import com.amazonaws.services.rekognition.model.LabelDetection;
import com.amazonaws.services.rekognition.model.LabelDetectionSortBy;
import com.amazonaws.services.rekognition.model.VideoMetadata;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.amazonaws.services.sns.AmazonSNS;
import com.amazonaws.services.sns.AmazonSNSClientBuilder;
import com.amazonaws.services.sns.model.PublishRequest;
import com.amazonaws.services.sns.model.PublishResult;


public class JobCompletionHandler implements RequestHandler<SNSEvent, String> {

   @Override
   public String handleRequest(SNSEvent event, Context context) {

      String message = event.getRecords().get(0).getSNS().getMessage();
      LambdaLogger logger = context.getLogger();

      // Parse SNS event for analysis results. Log results
      try {
         ObjectMapper operationResultMapper = new ObjectMapper();
         JsonNode jsonResultTree = operationResultMapper.readTree(message);
         logger.log("Rekognition Video Operation:=========================");
         logger.log("Job id: " + jsonResultTree.get("JobId"));
         logger.log("Status : " + jsonResultTree.get("Status"));
         logger.log("Job tag : " + jsonResultTree.get("JobTag"));
         logger.log("Operation : " + jsonResultTree.get("API"));

         if (jsonResultTree.get("API").asText().equals("StartLabelDetection")) {

            if (jsonResultTree.get("Status").asText().equals("SUCCEEDED")){
               GetResultsLabels(jsonResultTree.get("JobId").asText(), context);
            }
            else{
               String errorMessage = "Video analysis failed for job " 
                     + jsonResultTree.get("JobId") 
                     + "State " + jsonResultTree.get("Status");
               throw new Exception(errorMessage); 
            }

         } else
            logger.log("Operation not StartLabelDetection");

      } catch (Exception e) {
         logger.log("Error: " + e.getMessage());
         throw new RuntimeException (e);


      }

      return message;
   }

   void GetResultsLabels(String startJobId, Context context) throws Exception {

      LambdaLogger logger = context.getLogger();

      AmazonRekognition rek = AmazonRekognitionClientBuilder.standard().withRegion(Regions.AP_NORTHEAST_2).build();

      int maxResults = 1000;
      String paginationToken = null;
      GetLabelDetectionResult labelDetectionResult = null;
      String labels = "";
      Integer labelsCount = 0;
      String label = "";
      String currentLabel = "";
      Integer suspiciousLabelsCount = 0;
      String suspiciousLabel = "";
      List<String> allSuspiciousLabels = getSuspiciousLabels();
     
      //Get label detection results and log them. 
      do {

         GetLabelDetectionRequest labelDetectionRequest = new GetLabelDetectionRequest().withJobId(startJobId)
               .withSortBy(LabelDetectionSortBy.NAME).withMaxResults(maxResults).withNextToken(paginationToken);

         labelDetectionResult = rek.getLabelDetection(labelDetectionRequest);
         
         paginationToken = labelDetectionResult.getNextToken();
         VideoMetadata videoMetaData = labelDetectionResult.getVideoMetadata();

         // Add labels to log
         List<LabelDetection> detectedLabels = labelDetectionResult.getLabels();
         
         for (LabelDetection detectedLabel : detectedLabels) {
            label = detectedLabel.getLabel().getName();
            if (label.equals(currentLabel)) {
               continue;
            }
            labels = labels + label + " / ";
            currentLabel = label;
            labelsCount++;

            if (allSuspiciousLabels.contains(label) ) {
            	suspiciousLabel = suspiciousLabel + label + " / ";
            	suspiciousLabelsCount++;
            }
         }
      } while (labelDetectionResult != null && labelDetectionResult.getNextToken() != null);
      
      if (suspiciousLabelsCount > 0) {
    	  publishSNS(label, context);
	  }
      
      logger.log("Total number of labels : " + labelsCount);
      logger.log("labels : " + labels);

      logger.log("Total number of suspicious labels : " + suspiciousLabelsCount);
      logger.log("suspicious labels : " + suspiciousLabel);
   }


   List<String> getSuspiciousLabels() {
	      List<String> suspiciousLabels = new ArrayList<>();
	      suspiciousLabels.add("kicking");
	      suspiciousLabels.add("punching");
	      suspiciousLabels.add("fighting");
	      suspiciousLabels.add("martial art");
	      suspiciousLabels.add("wrestling");
	      
	      return suspiciousLabels;
   }
   
   void publishSNS(String label, Context context) {
	   LambdaLogger logger = context.getLogger();
	   AmazonSNS snsClient = AmazonSNSClientBuilder.standard().withRegion(Regions.AP_NORTHEAST_2).build();
	   
	   String topicArn = "arn:aws:sns:ap-northeast-2:964962553544:Email";
	   
	   //publish to an SNS topic
	   String msg = "Suspicious activity detected with label: " + label;
	   PublishRequest publishRequest = new PublishRequest(topicArn, msg);
	   PublishResult publishResult = snsClient.publish(publishRequest);
	   //print MessageId of message published to SNS topic
	   logger.log("MessageId - " + publishResult.getMessageId());

   }
}


