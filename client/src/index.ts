import * as AWS from "aws-sdk";
import bodyParser from "body-parser";
import dotenv from "dotenv";
import express from "express";
import multer from "multer";
import s3Storage from "multer-s3";
import path from "path";

// initialize configuration
dotenv.config();

// port is now available to the Node.js runtime
// as if it were an environment variable
const port = process.env.SERVER_PORT;
const accessKeyId =  process.env.AWS_ACCESS_KEY;
const bucket = process.env.AWS_BUCKET;
const region = process.env.AWS_REGION;
const secretAccessKey = process.env.AWS_SECRET_KEY;
AWS.config.update({ accessKeyId, region, secretAccessKey });

const app = express();
const s3 = new AWS.S3();
app.use(bodyParser.json());

// Configure Express to use EJS
app.set( "views", path.join( __dirname, "views" ) );
app.set( "view engine", "ejs" );

const upload = multer({
    storage: s3Storage({
        bucket,
        key(req, file, cb) {
            // tslint:disable-next-line:no-console
            console.log(file);
            cb(null, file.originalname);
        },
        s3
    })
});

// define a route handler for the default home page
app.get( "/", (req, res) => {
    // render the index template
    res.render( "index" );
} );

// use by upload form
app.post("/upload", upload.array("upl", 1), (req, res, next) => {
    res.send("Uploaded!");
});

// start the express server
app.listen( port, () => {
    // tslint:disable-next-line:no-console
    console.log( `server started at http://localhost:${ port }` );
} );
