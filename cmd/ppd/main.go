package main

import (
	"fmt"
	"os"
	"sds-s3/cmd/ppd/s3"

	"github.com/spf13/cobra"

	"github.com/stratosnet/sds/cmd/common"
	"github.com/stratosnet/sds/framework/utils"
	"github.com/stratosnet/sds/pp/setting"
)

func main() {

	rootCmd := getRootCmd()
	verCmd := getVersionCmd()
	s3Cmd := getS3()

	rootCmd.AddCommand(verCmd)
	rootCmd.AddCommand(s3Cmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "ppd",
		Short:             "resource node",
		PersistentPreRunE: common.RootPreRunE,
	}

	dir, err := os.Getwd()
	if err != nil {
		utils.ErrorLog("failed to get working directory")
		panic(err)
	}

	rootCmd.PersistentFlags().StringP(common.Home, "r", dir, "path for the node")
	rootCmd.PersistentFlags().StringP(common.Config, "c", common.DefaultConfigPath, "configuration file path ")
	return rootCmd
}

func getS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "s3migrate",
		Short:   "migrate s3 file to sds",
		PreRunE: s3.PreRunE,
		Run:     s3.S3Migrate,
	}

	cmd.PersistentFlags().StringP(s3.RpcModeFlag, "m", "ipc", "use http rpc or ipc")
	cmd.PersistentFlags().String(s3.PasswordFlag, "", "wallet password")
	cmd.PersistentFlags().StringP(s3.IpfsPortFlag, "p", "6798", "port")
	cmd.PersistentFlags().StringP(s3.IpcEndpoint, "", "", "ipc endpoint path")
	cmd.PersistentFlags().StringP(s3.HttpRpcUrl, "", s3.HttpRpcDefaultUrl, "http rpc url")
	return cmd
}

func getVersionCmd() *cobra.Command {

	version := setting.Version
	cmd := &cobra.Command{
		Use:   "version",
		Short: "get version of the build",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
	return cmd
}
