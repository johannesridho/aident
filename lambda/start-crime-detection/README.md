### General Setup

1. Create S3 bucket
2. Create SNS topics
3. Create Rekognition role
4. Create Lambda function
5. Give permission to S3 to be able to invoke Lambda
6. Build package
    - `GOOS=linux go build main.go`
    - `zip main.zip ./main`
7. Upload main.zip with AWS Lambda console
8. Set the env vars

### Setup Face Search

1. Create a collection with aws-cli
    ```
    aws rekognition create-collection \
        --collection-id "criminalFaces"
    ```
2. Add photos of target criminal to S3 and run the command below with aws-cli
    ```
    aws rekognition index-faces \      
          --image '{"S3Object":{"Bucket":"aident-criminal-faces","Name":"avenger-villains.jpg"}}' \
          --collection-id "criminalFaces" \
          --quality-filter "AUTO" \
          --detection-attributes "ALL" \
          --external-image-id "avenger-villains.jpg"
    ```
3. You can check the collection's content with the command below
    ```
    aws rekognition list-faces \
          --collection-id "criminalFaces"  
    ```