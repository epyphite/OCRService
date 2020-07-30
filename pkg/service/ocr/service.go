package ocr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/epyphite/OCRService/pkg/lib/storage"
	"github.com/epyphite/OCRService/pkg/models"
)

//Service main structure for service
type Service struct {
	TextractSession *textract.Textract
	Session         *session.Session
	Config          models.Config
	Jobs            []models.OCRFileProcess
}

// Init creates a instance of the textract session
func (O *Service) Init() {
	if O.Config.Debug == "yes" {
		log.Println(O.Config.CloudSourceRegion)
	}
	O.TextractSession = textract.New(session.Must(session.NewSession(&aws.Config{
		Region: &O.Config.CloudSourceRegion, // OHIO
	})))

	sess := session.Must(session.NewSession(&aws.Config{
		Region: &O.Config.CloudSourceRegion, // OHIO
	}))
	O.Session = sess
}

func (O *Service) readFolder(folder string) ([]string, error) {
	var ret []string
	var err error

	svc := s3.New(O.Session)

	i := 0
	err = svc.ListObjectsPages(&s3.ListObjectsInput{

		Bucket: &O.Config.CloudSourceStorage,
		Prefix: &O.Config.CloudSourceKey,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		i++

		for _, obj := range p.Contents {
			extension := filepath.Ext(*obj.Key)
			if extension == ".pdf" {
				ret = append(ret, *obj.Key)
			}

		}
		return true
	})
	if err != nil {
		log.Println("failed to list objects", err)
		return ret, err
	}

	return ret, err
}

func (O *Service) sendFile(file string) {
	var job models.OCRFileProcess
	hash, err := storage.GetHashFromS3(O.Session, O.Config.CloudSourceStorage, file)

	ret, err := O.checkProcessFile(hash)
	log.Println(ret)
	log.Println(err)
	if ret != "" {
		log.Println(err)
		return
	}
	log.Printf("Processing %v -  %v \n", file, hash)

	resp, err := O.TextractSession.StartDocumentTextDetection(&textract.StartDocumentTextDetectionInput{
		DocumentLocation: &textract.DocumentLocation{
			S3Object: &textract.S3Object{
				Bucket: &O.Config.CloudSourceStorage,
				Name:   &file,
			},
		},
		NotificationChannel: &textract.NotificationChannel{
			RoleArn:     &O.Config.CloudNotificationARN,
			SNSTopicArn: &O.Config.CloudNotificationTopic,
		},
	})
	if err != nil {
		panic(err)
	}
	log.Println(*resp.JobId)

	now := time.Now() // current local time
	job.FileHash = hash
	job.FileName = file
	job.Status = "PENDING"
	job.TimeStamp = now.String()
	O.saveProcessFile(job)

	//O.Jobs = append(O.Jobs, *resp.JobId)
}

//ReadResponse will be a go routine to read the job queue
func (O *Service) ReadResponse(finished chan bool) error {
	log.Println("Launching Reader ")
	var err error
	var timeout int64

	for {

		if timeout < 0 {
			timeout = 0
		}

		if timeout > 12*60*60 {
			timeout = 12 * 60 * 60
		}
		queue := "ocr_sqs"

		svc := sqs.New(O.Session, aws.NewConfig().WithRegion("us-east-2"))

		urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: &queue,
		})

		if err != nil {
			log.Println(err)
			finished <- true
			return err
		}

		queueURL := urlResult.QueueUrl

		var Attributes []*string
		attr := "ApproximateNumberOfMessages"
		Attributes = append(Attributes, &attr)
		resp, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: Attributes,
			QueueUrl:       queueURL,
		})

		if err != nil { // resp is now filled
			log.Println(err)
		}

		total := resp.Attributes["ApproximateNumberOfMessages"]
		log.Println("Total number of messages ", *total)

		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            queueURL,
			MaxNumberOfMessages: aws.Int64(10),
			VisibilityTimeout:   &timeout,
		})

		if msgResult != nil {
			for _, message := range msgResult.Messages {
				var body models.TextractBody
				json.Unmarshal([]byte(*message.Body), &body)

				//			for _, job := range O.Jobs {
				//				if job == body.JobID {

				if body.Status == "SUCCEEDED" {
					var jobID string
					jobID = body.JobID

					var documentAnalysis textract.GetDocumentTextDetectionInput
					documentAnalysis.SetJobId(jobID)
					documentAnalysis.SetMaxResults(1000)

					textractResult, err := O.TextractSession.GetDocumentTextDetection(&documentAnalysis)

					if err != nil {
						log.Println("Error in Textract Result ", err)
						return err
					}

					var ocrRet models.OCRResult
					ocrRet.Filename = body.DocumentLocation.S3ObjectName
					ocrRet.Timestamp = body.Timestamp
					documentAnalysis.NextToken = textractResult.NextToken

					for i := 1; i < len(textractResult.Blocks); i++ {
						if *textractResult.Blocks[i].BlockType == "LINE" {
							ocrRet.Data = append(ocrRet.Data, *textractResult.Blocks[i].Text)
						}
					}
					for documentAnalysis.NextToken != nil {
						textractResult, _ := O.TextractSession.GetDocumentTextDetection(&documentAnalysis)
						documentAnalysis.NextToken = textractResult.NextToken
						for i := 1; i < len(textractResult.Blocks); i++ {
							if *textractResult.Blocks[i].BlockType == "LINE" {
								ocrRet.Data = append(ocrRet.Data, *textractResult.Blocks[i].Text)
							}
						}
					}
					O.saveFile(O.Session, ocrRet)
					O.deleteMessage(O.Session, queueURL, message.ReceiptHandle)
				}

				//				}
				//			}
			}

		}

	}
	finished <- true

	return err
}

//deleteMessage from the queue
func (O *Service) deleteMessage(sess *session.Session, queueURL *string, messageHandle *string) error {
	svc := sqs.New(sess, aws.NewConfig().WithRegion(O.Config.CloudSourceRegion))

	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: messageHandle,
	})
	if err != nil {
		log.Println("Error in Delete message ", err)
		return err
	}

	return nil
}

func (O *Service) saveFile(sess *session.Session, results models.OCRResult) error {
	var err error
	fileContent, _ := json.MarshalIndent(results, "", " ")
	filename := O.Config.TempDir + filepath.Base(results.Filename) + ".json"
	_ = ioutil.WriteFile(filename, fileContent, 0644)
	f, err := os.Open(filename)

	if O.Config.EnableCloud == "yes" {
		err = storage.AddFileToS3(sess, f, O.Config.CloudDestinationStorage, O.Config.CloudDestinationKey)
		if err != nil {
			log.Println("Save File - ", err)
		} else {
			log.Println("File", filename, "Saved ")
		}

	}
	return err
}

func (O *Service) saveProcessFile(job models.OCRFileProcess) error {
	var err error

	svc := dynamodb.New(O.Session)
	av, err := dynamodbattribute.MarshalMap(job)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(O.Config.ProcessTable),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Got error calling PutItem:")
		log.Println(err.Error())
		os.Exit(1)
	}
	return err
}

func (O *Service) checkProcessFile(hash string) (string, error) {
	svc := dynamodb.New(O.Session)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(O.Config.ProcessTable),
		Key: map[string]*dynamodb.AttributeValue{
			"FileHash": {
				S: aws.String(hash),
			},
		},
	})

	var item models.OCRFileProcess
	if result.Item == nil {
		return "", err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	log.Printf("Found file %v - HASH  %v  STATUS -> %v \n", item.FileName, item.FileHash, item.Status)
	return item.FileHash, fmt.Errorf("File already being processed ")
}

func (O *Service) getProcessFiles() ([]models.OCRFileProcess, error) {

	var err error
	var jobs []models.OCRFileProcess
	svc := dynamodb.New(O.Session)

	params := &dynamodb.ScanInput{
		TableName: aws.String("Teams"),
	}
	result, _ := svc.Scan(params)

	if len(result.Items) > 0 {
		for _, item := range result.Items {
			var job models.OCRFileProcess

			err := dynamodbattribute.UnmarshalMap(item, &job)
			if err != nil {
				log.Println(err)
				break
			}
			jobs = append(jobs, job)
		}
	}
	return nil, err
}

//Process will launch the entire process of the file or the folder
func (O *Service) Process(finished chan bool) error {
	var err error

	if O.Config.CloudSourceProvider == "aws" {
		if O.Config.CloudSourceStorage != "" {

			fileList, err := O.readFolder(O.Config.CloudSourceStorage)
			if err != nil {
				return err
			}
			for i, file := range fileList {
				c := math.Mod(float64(i), 10)
				if c == 9 {
					fmt.Println("Sleeping for 10 seconds")
					time.Sleep(10 * time.Second)
				}
				O.sendFile(file)
			}
		}
	}
	finished <- true

	return err
}
