---
title: SDS S3 Command Handbook
description: List of commands for Operation between SDS and AWS S3.
---

# SDS S3 Migrate Command Handbook

## S3 Migrate
The `ppd s3migrate` command migrates a bucket or a file from AWS S3 to sds network. The command tool needs to communicate with a SDS resource node to interact with the network. For setting up a
SDS resource node please refer to  [Setup and run a SDS Resource Node](../setup-and-run-a-sds-resource-node/)

``` { .yaml .no-copy }
ppd s3migrate -h

Usage:
  ppd s3migrate [flags]

Flags:
  -h, --help                 help for s3migrate
      --httpRpcUrl string    http rpc url (default "http://127.0.0.1:9301")
      --ipcEndpoint string   ipc endpoint path
  -p, --password string      wallet password
  -m, --rpcMode string       use http rpc or ipc (default "ipc")

Global Flags:
  -r, --home string   path for the workspace (default "<root directory of your command tool>")
```
<br>

Please make sure sds account files are put in `accounts` folder under the home path (defined by -r or --home flag)

There are two modes to communicate to a SDS resource node, and it could be switched by the --rpcMode flag

- `httpRpc` mode is to send RPC request over http. In this mode the `httpRpcUrl` flag must point to the rpc port of the
  resource node
   ``` shell
   ppd s3migrate --rpcMode httpRpc --httpRpcUrl http://<node url>:<node rpc port>
   ```
- `ipc` mode is to send PRC requests over IPC (Inter-process communication). The path to the ipc endpoint is set
  by the flag `ipcEndpoint`. The default path will be used when flag is not set.
   ```  shell
   ppd s3migrate --rpcMode ipc --ipcEndpoint <path to the ipc endpoint>
   ```
<br>

## AWS Credentials File
AWS credential needs to be stored in the shared configuration file (~/.aws/config) in the following example format

```
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-2
```

## S3 Migrate
The `ppd s3migrate` command migrates a file or a bucket from AWS S3 to SDS network. It first downloads the file from the S3 by the
given bucket or file key then uploads it to the SDS network.

```shell
ppd s3migrate <bucket> <filename>
```
`bucket` is the bucket to be downloaded from S3.  
`filename` is an optional parameter. When it is given, only the specific file will be migrated instead of the whole bucket
to the SDS network.
