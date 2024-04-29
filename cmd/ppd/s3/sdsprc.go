package s3

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
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

func UploadToSds(requester Requester, filePah string) error {
	fileName := filePah
	hash := file.GetFileHash(fileName, "")
	utils.Log("- start uploading the file:", fileName)

	// compose reqOzone for the SN
	paramReqGetOzone, err := reqOzone()
	if err != nil {
		return errors.Wrap(err, "failed to create request message")
	}

	utils.Log("- request get ozone (method: user_requestGetOzone)")

	var resOzone rpc_api.GetOzoneResult
	err = requester(paramReqGetOzone, &resOzone, "requestGetOzone")
	if err != nil {
		return errors.Wrap(err, "failed to send upload file request")
	}

	// compose request file upload params
	paramsFile, err := reqUploadMsg(fileName, hash, resOzone.SequenceNumber)
	if err != nil {
		return errors.Wrap(err, "failed to create request message")
	}

	utils.Log("- request uploading file (method: user_requestUpload)")

	var res rpc_api.Result
	err = requester(paramsFile, &res, "requestUpload")
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
		paramsData, err := uploadDataMsg(hash, encoded, resOzone.SequenceNumber)
		if err != nil {
			return errors.Wrap(err, "failed to prepare upload data request")
		}
		utils.Log("- request upload date (method: user_uploadData)")

		err = requester(paramsData, &res, "uploadData")
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

func List(requester Requester, page uint64) (*rpc_api.FileListResult, error) {
	params, err := reqListMsg(page)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create list request")
	}
	var res rpc_api.FileListResult
	err = requester(params, &res, "requestList")
	if err != nil {
		return nil, errors.Wrap(err, "failed to send list request")
	}
	return &res, nil
}

func reqOzone() (*rpc_api.ParamReqGetOzone, error) {
	// wallet address
	ret := readWalletKeys(WalletAddress)
	if !ret {
		return nil, errors.New("failed reading key file")
	}
	return &rpc_api.ParamReqGetOzone{
		WalletAddr: WalletAddress,
	}, nil
}

func reqUploadMsg(filePath, hash, sn string) (*rpc_api.ParamReqUploadFile, error) {
	// file size
	info, err := file.GetFileInfo(filePath)
	if info == nil || err != nil {
		return nil, errors.New("failed to get file information")
	}
	fileName := info.Name()
	// wallet address
	ret := readWalletKeys(WalletAddress)
	if !ret {
		return nil, errors.New("failed reading key file")
	}
	nowSec := time.Now().Unix()
	// signature
	sign, err := WalletPrivateKey.Sign([]byte(msgutils.GetFileUploadWalletSignMessage(hash, WalletAddress, sn, nowSec)))
	if err != nil {
		return nil, err
	}
	wpk, err := fwtypes.WalletPubKeyToBech32(WalletPublicKey)
	if err != nil {
		return nil, err
	}

	return &rpc_api.ParamReqUploadFile{
		FileName: fileName,
		FileSize: int(info.Size()),
		FileHash: hash,
		Signature: rpc_api.Signature{
			Address:   WalletAddress,
			Pubkey:    wpk,
			Signature: hex.EncodeToString(sign),
		},
		ReqTime:         nowSec,
		DesiredTier:     2,
		AllowHigherTier: true,
		SequenceNumber:  sn,
	}, nil
}

func uploadDataMsg(hash, data, sn string) (rpc_api.ParamUploadData, error) {
	// wallet address
	ret := readWalletKeys(WalletAddress)
	if !ret {
		return rpc_api.ParamUploadData{}, errors.New("failed reading key file")
	}
	nowSec := time.Now().Unix()
	// signature
	sign, err := WalletPrivateKey.Sign([]byte(msgutils.GetFileUploadWalletSignMessage(hash, WalletAddress, sn, nowSec)))
	if err != nil {
		return rpc_api.ParamUploadData{}, err
	}
	wpk, err := fwtypes.WalletPubKeyToBech32(WalletPublicKey)
	if err != nil {
		return rpc_api.ParamUploadData{}, err
	}

	return rpc_api.ParamUploadData{
		FileHash: hash,
		Data:     data,
		Signature: rpc_api.Signature{
			Address:   WalletAddress,
			Pubkey:    wpk,
			Signature: hex.EncodeToString(sign),
		},
		ReqTime:        nowSec,
		SequenceNumber: sn,
	}, nil
}

func reqListMsg(page uint64) (*rpc_api.ParamReqFileList, error) {
	//wallet address
	ret := readWalletKeys(WalletAddress)
	if !ret {
		return nil, errors.New("failed reading key file")
	}
	nowSec := time.Now().Unix()
	// signature
	sign, err := WalletPrivateKey.Sign([]byte(msgutils.FindMyFileListWalletSignMessage(WalletAddress, nowSec)))
	if err != nil {
		return nil, err
	}
	wpk, err := fwtypes.WalletPubKeyToBech32(WalletPublicKey)
	if err != nil {
		return nil, err
	}
	return &rpc_api.ParamReqFileList{
		Signature: rpc_api.Signature{
			Address:   WalletAddress,
			Pubkey:    wpk,
			Signature: hex.EncodeToString(sign),
		},
		PageId:  page,
		ReqTime: nowSec,
	}, nil
}

func readWalletKeys(wallet string) bool {
	if wallet == "" {
		WalletAddress = findWallet(filepath.Join(setting.GetRootPath(), "./accounts/"))
	} else {
		WalletAddress = wallet
	}
	if WalletAddress == "" {
		return false
	}

	keyjson, err := os.ReadFile(filepath.Join(setting.GetRootPath(), "./accounts/", WalletAddress+".json"))
	if utils.CheckError(err) {
		utils.ErrorLog("getPublicKey ioutil.ReadFile", err)
		return false
	}

	key, err := fwtypes.DecryptKey(keyjson, WalletPassword, true)
	if utils.CheckError(err) {
		utils.ErrorLog("getPublicKey DecryptKey", err)
		return false
	}
	WalletPrivateKey = key.PrivateKey
	WalletPublicKey = WalletPrivateKey.PubKey()
	return true
}

func findWallet(folder string) string {
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
