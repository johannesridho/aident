// Configure the AWS. In this case for the simplicity I'm using access key and secret.
AWS.config.update({
    credentials: {
        accessKeyId: "???",
        secretAccessKey: "???",
        region: "???"
    }
});


const BUCKET_NAME = "aident-videos";

// console.log(AWS.config);