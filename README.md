# Aident
Aident detects crime behaviours and recognises criminal suspects in videos using AWS technology. 
This project is built for [AWS Hackdays 2019](https://aws.agorize.com/en/challenges/indonesia-2019)

### Demo
- https://youtu.be/zU3n32rUHEw
- http://aident.wedyzec4mz.ap-northeast-2.elasticbeanstalk.com
- https://www.facebook.com/aidentbot

### Architecture Diagram
Our plan is to build both stream and batch video analysis. With batch video analysis finished, we will start to build the stream video analysis in the next phase.

#### a. Target Implementation
![architecture](https://github.com/johannesridho/aident/blob/master/README_files/Aident%20-%20Target%20Architecture.png)

#### b. Current Implementation
![architecture](https://github.com/johannesridho/aident/blob/master/README_files/Aident%20-%20Current%20Architecture.png)

### How to Use This Code
To be able to use this code, the AWSs need to be configured as seen in the architecture diagram. This code only contains client app which need to be deployed to Elastic Beanstalk and Lambda functions. 
