package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp/file"
	"io"
)

type BucketBasics struct {
	S3Client *s3.Client
}

func (basics BucketBasics) ListBuckets() ([]types.Bucket, error) {
	result, err := basics.S3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	var buckets []types.Bucket
	if err != nil {
		utils.ErrorLogf("Couldn't list buckets for your account. Here's why: %v\n", err)
	} else {
		buckets = result.Buckets
	}
	return buckets, err
}

func (basics BucketBasics) DownloadLargeObject(bucketName string, objectKey string) ([]byte, error) {
	var partMiBs int64 = 10
	downloader := manager.NewDownloader(basics.S3Client, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	buffer := manager.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(context.TODO(), buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		utils.ErrorLogf("Couldn't download large object from %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return buffer.Bytes(), err
}

func (basics BucketBasics) DownloadObject(bucketName string, objectKey string) ([]byte, error) {
	result, err := basics.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		utils.ErrorLogf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return nil, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

func (basics BucketBasics) ListObjects(bucketName string) ([]types.Object, error) {
	result, err := basics.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	var contents []types.Object
	if err != nil {
		utils.ErrorLogf("Couldn't list objects in bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		contents = result.Contents
	}
	return contents, err
}

func (basics BucketBasics) DownloadFile(bucketName, objectKey, downloadFolder string, fileSize int64) (string, error) {
	var body []byte
	f, err := file.CreateFolderAndReopenFile(downloadFolder, objectKey)
	if err != nil {
		utils.ErrorLogf("Couldn't create file %v. Here's why: %v\n", objectKey, err)
		return "", err
	}
	defer f.Close()

	if fileSize < BigFileSizeInMB*1024*1024 {
		body, err = basics.DownloadObject(bucketName, objectKey)
	} else {
		body, err = basics.DownloadLargeObject(bucketName, objectKey)
	}
	if err != nil {
		utils.ErrorLogf("Couldn't read object body from %v: %v\n", objectKey, err)
	}
	_, err = f.Write(body)
	return f.Name(), err
}

// BucketExists checks whether a bucket exists in the current account.
func (basics BucketBasics) BucketExists(bucketName string) (bool, error) {
	_, err := basics.S3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				utils.ErrorLogf("Bucket %v is available.\n", bucketName)
				exists = false
				err = nil
			default:
				utils.ErrorLogf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err)
			}
		}
	} else {
		utils.Logf("Bucket %v exists and you already own it.", bucketName)
	}

	return exists, err
}
