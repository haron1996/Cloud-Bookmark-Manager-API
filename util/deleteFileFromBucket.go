package util

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func DeleteFileFromBucket(bucket string, key string) error {
	// Step 0: Load config
	config, err := LoadConfig(".")
	if err != nil {
		log.Printf("failed to load config file: %v", err)
		return err
	}

	s3Config := &aws.Config{
		// Credentials:      credentials.NewStaticCredentials(key, secret, ""),        // Specifies your credentials.
		Credentials:      credentials.NewStaticCredentials(config.DOSpacesKey, config.DOSecretKey, ""),
		Endpoint:         aws.String("https://nested.sfo3.digitaloceanspaces.com"), // Find your endpoint in the control panel, under Settings. Prepend "https://".
		S3ForcePathStyle: aws.Bool(false),                                          // // Configures to use subdomain/virtual calling format. Depending on your version, alternatively use o.UsePathStyle = false
		Region:           aws.String("sfo3"),                                       // Must be "us-east-1" when creating new Spaces. Otherwise, use the region in your endpoint, such as "nyc3".
	}

	// Step 3: The new session validates your request and directs it to your Space's specified endpoint using the AWS SDK.
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.Printf("failed to create new session: %v", err)
		return err
	}

	s3Client := s3.New(newSession)

	if _, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
