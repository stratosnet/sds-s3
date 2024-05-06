package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/stratosnet/sds/cmd/common"
	"github.com/stratosnet/sds/framework/utils"
)

func main() {
	rootCmd := getRootCmd()
	s3Cmd := getS3()
	rootCmd.AddCommand(s3Cmd)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "sdss3",
		Short:             "sds s3 command tool",
		PersistentPreRunE: common.RootPreRunE,
	}

	dir, err := os.Getwd()
	if err != nil {
		utils.ErrorLog("failed to get working directory")
		panic(err)
	}

	rootCmd.PersistentFlags().StringP(Home, "r", dir, "path for the workspace")
	return rootCmd
}

func getS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate",
		Short:   "migrate s3 file to sds",
		PreRunE: PreRunE,
		Run:     S3Migrate,
	}

	cmd.PersistentFlags().StringP(RpcModeFlag, "m", "", "use http rpc or ipc")
	cmd.PersistentFlags().StringP(WalletAddressFlag, "w", "", "wallet address")
	cmd.PersistentFlags().StringP(PasswordFlag, "p", "", "wallet password")
	cmd.PersistentFlags().StringP(IpcEndpoint, "", "", "ipc endpoint path")
	cmd.PersistentFlags().StringP(HttpRpcUrl, "", "", "http rpc url")
	return cmd
}
