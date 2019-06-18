package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	Region      = "ap-southeast-2"
	CredPath    = "/home/jeff/.aws/credentials"
	CredProfile = "aws-cred-profile"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("Usage: go run s3.go <the bucket name> <the file name>\n" +
			"Example: go run s3.go my-test-bucket my-upload-file\n")
		os.Exit(1)
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := s3.New(sess, &aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
		MaxRetries:  aws.Int(5),
	})

	listMyBuckets(svc)
	createMyBucket(svc, os.Args[1], Region)
	uploadObj(svc, os.Args[1], os.Args[2])
	listObj(svc, os.Args[1])
	listMyBuckets(svc)
	deleteMyBucket(svc, os.Args[1])
	listMyBuckets(svc)
}

func listMyBuckets(svc *s3.S3) {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}
	fmt.Println("My buckets now are:")
	for _, b := range result.Buckets {
		fmt.Printf(aws.StringValue(b.Name) + "\n")
	}
	fmt.Printf("\n")
}

func createMyBucket(svc *s3.S3, bucketName string, region string) {
	fmt.Printf("\nCreating a new bucket named '" + bucketName + "'...\n\n")
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	})
	if err != nil {
		exitErrorf("Unable to create bucket, %v", err)
	}
	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
}

func deleteMyBucket(svc *s3.S3, bucketName string) {
	fmt.Printf("\nDeleting the bucket named '" + bucketName + "'...\n\n")
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		exitErrorf("Unable to delete bucket, %v", err)
	}
	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
}

func copyObj(svc *s3.S3, bucket, item, other string) {

	source := bucket + "/" + item
	// Copy the item
	_, err := svc.CopyObject(&s3.CopyObjectInput{Bucket: aws.String(other), CopySource: aws.String(source), Key: aws.String(item)})
	if err != nil {
		exitErrorf("Unable to copy item from bucket %q to bucket %q, %v", bucket, other, err)
	}
	err = svc.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(other), Key: aws.String(item)})
	if err != nil {
		exitErrorf("Error occurred while waiting for item %q to be copied to bucket %q, %v", bucket, item, other, err)
	}
	fmt.Printf("Item %q successfully copied from bucket %q to bucket %q\n", item, bucket, other)
}

func uploadObj(svc *s3.S3, bucket, filename string) {

	file, err := os.Open(filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer file.Close()
	// Setup the S3 Upload Manager
	uploader := s3manager.NewUploaderWithClient(svc)
	// Upload the file's body to S3 bucket as an object with the key being the same as the filename.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}
	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func listObj(svc *s3.S3, bucket string) {

	// Get the list of items
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}
	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
	fmt.Println("Found", len(resp.Contents), "items in bucket", bucket)
	fmt.Println("")
}

// If there's an error, display it.
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
