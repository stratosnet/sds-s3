package main

import (
	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/stratosnet/sds/pp/file"
	mys3 "github.com/stratosnet/sds/s3/cmd/sdss3/s3"
	"github.com/stratosnet/sds/s3/cmd/sdss3/sds"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp/setting"
	"github.com/stratosnet/sds/rpc"
	"log"
)

const (
	HttpRpcUrl               = "httpRpcUrl"
	RpcModeFlag              = "rpcMode"
	RpcModeHttpRpc           = "httpRpc"
	RpcModeIpc               = "ipc"
	IpcEndpoint              = "ipcEndpoint"
	Home                     = "home"
	PasswordFlag             = "password"
	WalletAddressFlag        = "walletAddress"
	LocalFolderPath          = "s3"
	DefaultConfigPath string = "./config/config.toml"
)

var (
	SdsUploader *sds.Uploader
	S3Bucket    *mys3.BucketBasics
)

func PreRunE(cmd *cobra.Command, args []string) error {
	homePath, configPath, err := getPaths(cmd)
	if err != nil {
		log.Fatal("failed to get 'home' path for the client")
	}
	setting.SetIPCEndpoint(homePath)

	config, err := loadConfig(cmd, configPath)
	if err != nil {
		log.Fatalf("failed to load config and parameters: %v", err)
	}

	requester := getRequester(config)
	SdsUploader, err = sds.CreateSdsUploader(&requester, config.Keys.WalletAddress, config.Keys.WalletPassword)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	S3Bucket = &mys3.BucketBasics{
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

	exists, err := S3Bucket.BucketExists(myBucket)
	if err != nil || !exists {
		utils.ErrorLogf("failed to find the bucket %v: %v", myBucket, err)
	}

	files, err := S3Bucket.ListObjects(myBucket)
	if err != nil {
		utils.ErrorLogf("failed to read file list in the bucket %v: %v", myBucket, err)
	}

	folder := filepath.Join(file.GetTmpDownloadPath(), LocalFolderPath, myBucket)
	defer os.RemoveAll(folder)

	for _, file := range files {
		if myString != "" && myString != *file.Key {
			continue
		}
		fileKey := *file.Key
		utils.Logf("start downloading file %v from bucket %v", fileKey, myBucket)
		downloadPath, err := S3Bucket.DownloadFile(myBucket, fileKey, folder, *file.Size)
		if err != nil {
			utils.ErrorLogf("failed to download file %v from bucket %v: %v", fileKey, myBucket, err)
			continue
		}
		utils.Logf("downloaded file %v from buckt %v", fileKey, myBucket)
		utils.Logf("start uploading file %v to sds", downloadPath)
		err = SdsUploader.Upload(downloadPath)
		if err != nil {
			utils.ErrorLogf("failed to uplaod file %v to sds: %v", fileKey, err)
			continue
		}
	}
}

func getRequester(config *Config) sds.Requester {
	if config.Connectivity.RpcMode == RpcModeIpc {
		ipcEndpoint := setting.IpcEndpoint
		if config.Connectivity.IpcEndpoint != "" {
			ipcEndpoint = config.Connectivity.IpcEndpoint
		}
		c, err := rpc.Dial(ipcEndpoint)
		if err != nil {
			panic("failed to dial ipc endpoint, please make sure sds is launched.")
		}
		return sds.GetIpcRequester(c)
	} else if config.Connectivity.RpcMode == RpcModeHttpRpc {
		return sds.GetHttpRequester(config.Connectivity.HttpRpcUrl)
	} else {
		panic("unsupported rpc mode")
	}
}

func loadConfig(cmd *cobra.Command, configPath string) (*Config, error) {
	config := Config{}
	err := utils.LoadTomlConfig(&config, configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	walletAddress, err := cmd.Flags().GetString(WalletAddressFlag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wallet address from the parameters")
	}
	if walletAddress != "" {
		config.Keys.WalletAddress = walletAddress
	}

	password, err := cmd.Flags().GetString(PasswordFlag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get password from the parameters")
	}
	if password != "" {
		config.Keys.WalletPassword = password
	}

	rpcModeParam, err := cmd.Flags().GetString(RpcModeFlag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rpc mode from the parameters")
	}
	if rpcModeParam != "" {
		config.Connectivity.RpcMode = rpcModeParam
	}

	ipcEndpointParam, err := cmd.Flags().GetString(IpcEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ipc endpoint from the parameters")
	}
	if ipcEndpointParam != "" {
		config.Connectivity.IpcEndpoint = ipcEndpointParam
	}

	httpRpcUrl, err := cmd.Flags().GetString(HttpRpcUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get http rpc url from the parameters")
	}
	if httpRpcUrl != "" {
		config.Connectivity.HttpRpcUrl = httpRpcUrl
	}
	return &config, nil
}

func getPaths(cmd *cobra.Command) (homePath, configPath string, err error) {
	homePath, err = cmd.Flags().GetString(Home)
	if err != nil {
		utils.ErrorLog("failed to get 'home' path for the node")
		return
	}
	homePath, err = utils.Absolute(homePath)
	if err != nil {
		return
	}
	configPath = filepath.Join(homePath, DefaultConfigPath)
	return
}
