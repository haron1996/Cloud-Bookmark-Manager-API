package vultr

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func UploadHeroImage(heroImage *os.File) string {
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

	object := s3.PutObjectInput{
		Bucket: aws.String("/app-assets"),
		Key:    aws.String(uuid.NewString()),
		Body:   heroImage,
		ACL:    aws.String("public-read"),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload hero image to app-assets bucket: %v", err)
	}

	return fmt.Sprintf("https://ewr1.vultrobjects.com/app-assets/%s", *object.Key)
}
