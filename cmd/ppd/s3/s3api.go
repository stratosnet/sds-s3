package s3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/pkg/errors"
	"github.com/stratosnet/sds/pp/file"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	fwcryptotypes "github.com/stratosnet/sds/framework/crypto/types"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp/setting"
	"github.com/stratosnet/sds/rpc"
	"log"
)

const (
	IpcNamespace             = "remoterpc"
	HttpRpcNamespace         = "user"
	HttpRpcUrl               = "httpRpcUrl"
	RpcModeFlag              = "rpcMode"
	RpcModeHttpRpc           = "httpRpc"
	RpcModeIpc               = "ipc"
	IpcEndpoint              = "ipcEndpoint"
	IpfsPortFlag             = "port"
	HttpRpcDefaultUrl        = "http://127.0.0.1:9301"
	HOME              string = "home"
	PasswordFlag             = "password"
	LocalFolderPath          = "s3"
	BigFileSizeInMB          = 10
)

var (
	WalletPrivateKey fwcryptotypes.PrivKey
	WalletPublicKey  fwcryptotypes.PubKey
	WalletAddress    string
	WalletPassword   string
	S3Bucket         *BucketBasics
)

func PreRunE(cmd *cobra.Command, args []string) error {
	homePath, err := cmd.Flags().GetString(HOME)
	if err != nil {
		utils.ErrorLog("failed to get 'home' path for the client")
		return err
	}
	setting.SetIPCEndpoint(homePath)

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		panic(errors.New("failed to get password from the parameters"))
	}
	WalletPassword = password

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	S3Bucket = &BucketBasics{
		S3Client: s3.NewFromConfig(cfg),
	}

	return nil
}

func S3Migrate(cmd *cobra.Command, args []string) {
	if len(args) < 1 || args[0] == "" {
		panic("missing bucket")
	}
	myBucket := ""
	myString := ""

	if len(args) > 0 && args[0] != "" {
		myBucket = args[0]
	}

	if len(args) > 1 && args[1] != "" {
		myString = args[1]
	}
	requester := getRequester(cmd)

	exists, err := S3Bucket.BucketExists(myBucket)
	if err != nil || !exists {
		utils.ErrorLog("failed to find the bucket %v: %v", myBucket, err)
	}

	files, err := S3Bucket.ListObjects(myBucket)
	if err != nil {
		utils.ErrorLog("failed to read file list in the bucket %v: %v", myBucket, err)
	}

	folder := filepath.Join(file.GetTmpDownloadPath(), LocalFolderPath, myBucket)
	defer os.RemoveAll(folder)

	for _, file := range files {
		if myString != "" && myString != *file.Key {
			continue
		}
		fileKey := *file.Key
		utils.Log("start downloading file %v from bucket %v", fileKey, myBucket)
		downloadPath, err := S3Bucket.DownloadFile(myBucket, fileKey, folder, *file.Size)
		if err != nil {
			utils.ErrorLog("failed to download file %v from bucket %v: %v", fileKey, myBucket, err)
			continue
		}
		utils.Log("downloaded file %v from buckt %v", fileKey, myBucket)
		utils.Log("start uploading file %v to sds", downloadPath)
		err = UploadToSds(requester, downloadPath)
		if err != nil {
			utils.ErrorLog("failed to uplaod file %v to sds: %v", fileKey, err)
			continue
		}
	}
}

func getRequester(cmd *cobra.Command) Requester {
	rpcModeParam, _ := cmd.Flags().GetString(RpcModeFlag)
	ipcEndpointParam, _ := cmd.Flags().GetString(IpcEndpoint)
	httpRpcUrl, _ := cmd.Flags().GetString(HttpRpcUrl)
	if rpcModeParam == RpcModeIpc {
		ipcEndpoint := setting.IpcEndpoint
		if ipcEndpointParam != "" {
			ipcEndpoint = ipcEndpointParam
		}
		c, err := rpc.Dial(ipcEndpoint)
		if err != nil {
			panic("failed to dial ipc endpoint, please make sure sds is launched.")
		}
		return getIpcRequester(c)
	} else if rpcModeParam == RpcModeHttpRpc {
		return GetHttpRequester(httpRpcUrl)
	} else {
		panic("unsupported rpc mode")
	}
}
