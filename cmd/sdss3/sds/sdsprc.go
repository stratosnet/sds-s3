package sds

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	fwcryptotypes "github.com/stratosnet/sds/framework/crypto/types"
	"github.com/stratosnet/sds/pp/setting"
	"os"
	"path/filepath"
	"time"

	fwtypes "github.com/stratosnet/sds/framework/types"
	"github.com/stratosnet/sds/framework/utils"
	rpc_api "github.com/stratosnet/sds/pp/api/rpc"
	"github.com/stratosnet/sds/pp/file"
	"github.com/stratosnet/sds/sds-msg/protos"
	msgutils "github.com/stratosnet/sds/sds-msg/utils"
)

const (
	IpcNamespace     = "remoterpc"
	HttpRpcNamespace = "user"
)

type Uploader struct {
	walletPrivateKey fwcryptotypes.PrivKey
	walletPublicKey  fwcryptotypes.PubKey
	walletAddress    string
	requester        Requester
}

func CreateSdsUploader(requester *Requester, walletAddress, walletPassword string) (*Uploader, error) {
	up := Uploader{
		requester: *requester,
	}
	err := up.readWalletKeys(walletAddress, walletPassword)
	if err != nil {
		return nil, err
	}
	return &up, nil
}

func (up *Uploader) Upload(filePah string) error {
	fileName := filePah
	hash := file.GetFileHash(fileName, "")
	utils.Log("- start uploading the file:", fileName)

	// compose reqOzone for the SN
	paramReqGetOzone, err := up.reqOzone()
	if err != nil {
		return errors.Wrap(err, "failed to create request message")
	}

	utils.Log("- request get ozone (method: user_requestGetOzone)")

	var resOzone rpc_api.GetOzoneResult
	err = up.requester(paramReqGetOzone, &resOzone, "requestGetOzone")
	if err != nil {
		return errors.Wrap(err, "failed to send upload file request")
	}

	// compose request file upload params
	paramsFile, err := up.reqUploadMsg(fileName, hash, resOzone.SequenceNumber)
	if err != nil {
		return errors.Wrap(err, "failed to create request message")
	}

	utils.Log("- request uploading file (method: user_requestUpload)")

	var res rpc_api.Result
	err = up.requester(paramsFile, &res, "requestUpload")
	if err != nil {
		return errors.Wrap(err, "failed to send upload file request")
	}

	// Handle result:1 sending the content
	for res.Return == rpc_api.UPLOAD_DATA {
		utils.Log("- received response (return: UPLOAD_DATA)")
		// get the data from the file
		so := &protos.SliceOffset{
			SliceOffsetStart: *res.OffsetStart,
			SliceOffsetEnd:   *res.OffsetEnd,
		}
		rawData, err := file.GetFileData(fileName, so)
		if err != nil {
			return errors.Wrap(err, "failed to get data from file")
		}
		encoded := base64.StdEncoding.EncodeToString(rawData)
		paramsData, err := up.uploadDataMsg(hash, encoded, resOzone.SequenceNumber)
		if err != nil {
			return errors.Wrap(err, "failed to prepare upload data request")
		}
		utils.Log("- request upload date (method: user_uploadData)")

		err = up.requester(paramsData, &res, "uploadData")
		if err != nil {
			return errors.Wrap(err, "failed to send upload data request")
		}
	}
	if res.Return == rpc_api.SUCCESS {
		utils.Log("- uploading is done")
		return nil
	} else {
		utils.Log("- received response (return: ", res.Return, ")")
		return errors.Wrap(err, "failed to upload")
	}
}

func (up *Uploader) reqOzone() (*rpc_api.ParamReqGetOzone, error) {
	return &rpc_api.ParamReqGetOzone{
		WalletAddr: up.walletAddress,
	}, nil
}

func (up *Uploader) reqUploadMsg(filePath, hash, sn string) (*rpc_api.ParamReqUploadFile, error) {
	// file size
	info, err := file.GetFileInfo(filePath)
	if info == nil || err != nil {
		return nil, errors.New("failed to get file information")
	}
	fileName := info.Name()
	nowSec := time.Now().Unix()
	// signature
	sign, err := up.walletPrivateKey.Sign([]byte(msgutils.GetFileUploadWalletSignMessage(hash, up.walletAddress, sn, nowSec)))
	if err != nil {
		return nil, err
	}
	wpk, err := fwtypes.WalletPubKeyToBech32(up.walletPublicKey)
	if err != nil {
		return nil, err
	}

	return &rpc_api.ParamReqUploadFile{
		FileName: fileName,
		FileSize: int(info.Size()),
		FileHash: hash,
		Signature: rpc_api.Signature{
			Address:   up.walletAddress,
			Pubkey:    wpk,
			Signature: hex.EncodeToString(sign),
		},
		ReqTime:         nowSec,
		DesiredTier:     2,
		AllowHigherTier: true,
		SequenceNumber:  sn,
	}, nil
}

func (up *Uploader) uploadDataMsg(hash, data, sn string) (rpc_api.ParamUploadData, error) {
	nowSec := time.Now().Unix()
	// signature
	sign, err := up.walletPrivateKey.Sign([]byte(msgutils.GetFileUploadWalletSignMessage(hash, up.walletAddress, sn, nowSec)))
	if err != nil {
		return rpc_api.ParamUploadData{}, err
	}
	wpk, err := fwtypes.WalletPubKeyToBech32(up.walletPublicKey)
	if err != nil {
		return rpc_api.ParamUploadData{}, err
	}

	return rpc_api.ParamUploadData{
		FileHash: hash,
		Data:     data,
		Signature: rpc_api.Signature{
			Address:   up.walletAddress,
			Pubkey:    wpk,
			Signature: hex.EncodeToString(sign),
		},
		ReqTime:        nowSec,
		SequenceNumber: sn,
	}, nil
}

func (up *Uploader) readWalletKeys(wallet, password string) error {
	if wallet == "" {
		up.walletAddress = up.findWallet(filepath.Join(setting.GetRootPath(), "./accounts/"))
	} else {
		up.walletAddress = wallet
	}
	if up.walletAddress == "" {
		return errors.New("wallet is empty")
	}

	keyjson, err := os.ReadFile(filepath.Join(setting.GetRootPath(), "./accounts/", up.walletAddress+".json"))
	if utils.CheckError(err) {
		return errors.Wrap(err, "getPublicKey ioutil.ReadFile")
	}

	key, err := fwtypes.DecryptKey(keyjson, password, true)
	if utils.CheckError(err) {
		return errors.Wrap(err, "getPublicKey DecryptKey")
	}
	up.walletPrivateKey = key.PrivateKey
	up.walletPublicKey = up.walletPrivateKey.PubKey()
	return nil
}

func (up *Uploader) findWallet(folder string) string {
	var files []string
	var file string

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		file = filepath.Base(path)

		if m, _ := filepath.Match("st1*", file); !info.IsDir() && filepath.Ext(path) == ".json" && m {
			// only catch the first wallet file
			if files == nil {
				files = append(files, file[:len(file)-len(filepath.Ext(file))])
			}
		}
		return nil
	})
	if err != nil {
		return ""
	}

	if files != nil {
		return files[0]
	}
	return ""
}
