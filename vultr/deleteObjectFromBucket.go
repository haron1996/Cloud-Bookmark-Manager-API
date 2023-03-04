package vultr

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func DeleteObjectFromBucket(bucket, key string) {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panicf("could not load conig file: %v", err)
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.VultrAccessKey, config.VultrSecretKey, ""),
		Endpoint:         aws.String("https://ewr1.vultrobjects.com/"),
		S3ForcePathStyle: aws.Bool(false),
		Region:           aws.String("ewr"),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.Panicf("could not create new vultr s3 session: %v", err)
	}

	s3Client := s3.New(newSession)

	object := &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	_, err = s3Client.DeleteObject(object)
	if err != nil {
		log.Panicf("could not delete object from vultr: %v", err)
	}
}
