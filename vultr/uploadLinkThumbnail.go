package vultr

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
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func UploadLinkThumbnail(linkThumbnailChannel chan string) {
	imgFileChan := make(chan *os.File, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := util.LoadImage(imgFileChan, "a.png"); err != nil {
			log.Panicf("could not load link thumbnail: %v", err)
		}
	}()

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
		Bucket: aws.String("/link-thumbnails"),
		Key:    aws.String(uuid.NewString()),
		Body:   <-imgFileChan,
		ACL:    aws.String("public-read"),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		log.Panicf("could not upload link thumbnail to vultr: %v", err)
	}

	log.Printf("link thumbnail url: %s", fmt.Sprintf("https://ewr1.vultrobjects.com/link-thumbnails/%s", *object.Key))

	linkThumbnailChannel <- fmt.Sprintf("https://ewr1.vultrobjects.com/link-thumbnails/%s", *object.Key)

	wg.Wait()
}
