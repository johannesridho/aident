'use strict';
var AWS = require('aws-sdk'),
    transcoder = new AWS.ElasticTranscoder({
        apiVersion: '2012-09-25',
        region: 'ap-northeast-1'
    });
exports.handler = (event, context, callback) => {
    let fileName = event.Records[0].s3.object.key;
    var srcKey =  decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, " "));
    var newKey = fileName.split('.')[0];
    console.log('New video has been uploaded:', fileName);
transcoder.createJob({
     PipelineId: process.env.PIPELINE_ID,
     Input: {
      Key: srcKey,
      FrameRate: 'auto',
      Resolution: 'auto',
      AspectRatio: 'auto',
      Interlaced: 'auto',
      Container: 'auto'
     },
     Output: {
      Key: getOutputName(fileName),
      ThumbnailPattern: '',
    //   PresetId: '1351620000001-300040',
      PresetId: '1351620000001-000010',
      Rotate: 'auto'
     }
    }, function(err, data){
        if(err){
            console.log('Something went wrong:',err)
        }else{
            console.log('Converting is done');
        }
     callback(err, data);
    });
};
function getOutputName(srcKey){
//  let baseName = srcKey.replace('videos/','');
//  let withOutExtension = removeExtension(baseName);
//  return 'audios/' + withOutExtension + '.mp3';
    return srcKey.replace('webn','mp4');
}
function removeExtension(srcKey){
    let lastDotPosition = srcKey.lastIndexOf(".");
    if (lastDotPosition === -1) return srcKey;
    else return srcKey.substr(0, lastDotPosition);
}
