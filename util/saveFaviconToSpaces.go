package util

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func SaveFaviconToSpaces() string {
	// Step 0: Load config
	config, err := LoadConfig(".")
	if err != nil {
		log.Println(err)
		return ""
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
		log.Fatalf("could not create new session: %s", err)
	}

	s3Client := s3.New(newSession)

	imgFileChan := make(chan *os.File, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := LoadImage(imgFileChan, "favicon.ico"); err != nil {
			return
		}
	}()

	// Step 4: Define the parameters of the object you want to upload.
	object := s3.PutObjectInput{
		Bucket: aws.String("/favicons"),      // The path to the directory you want to upload the object to, starting with your Space name.
		Key:    aws.String(uuid.NewString()), // Object key, referenced whenever you want to access this file later.
		Body:   <-imgFileChan,                // The object's contents.
		ACL:    aws.String("public-read"),    // Defines Access-control List (ACL) permissions, such as private or public.
		Metadata: map[string]*string{ // Required. Defines metadata tags.
			"x-amz-meta-my-key": aws.String("your-value"),
		},
	}

	// Step 5: Run the PutObject function with your parameters, catching for errors.
	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Fatalf("could not put object to digital ocean: %v", err)
	}

	objectLink := fmt.Sprintf("https://sfo3.digitaloceanspaces.com/nested/favicons/%s", *object.Key)

	wg.Wait()

	return objectLink
}
