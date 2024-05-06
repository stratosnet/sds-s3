---
title: SDS S3 Command Handbook
description: List of commands for Operation between SDS and AWS S3.
---

# SDS S3 Migrate Command Handbook

The `sdss3 migrate` command migrates a bucket or a file from AWS S3 to sds network. It first downloads the file from the S3 by the
given bucket or file key then uploads it to the SDS network. The command tool needs to communicate with a SDS resource node to interact with the network. For setting up a
SDS resource node please refer to  [Setup and run a SDS Resource Node](https://docs.thestratos.org/docs-resource-node/setup-and-run-a-sds-resource-node). 
For generating wallet, please refer to [Generate/Recover wallet](https://docs.thestratos.org/docs-resource-node/setup-and-run-a-sds-resource-node)

### Flags
``` { .yaml .no-copy }
sdss3 migrate -h

Usage:
  sdss3 migrate [flags]

Flags:
  -h, --help                   help for migrate
      --httpRpcUrl string      http rpc url
      --ipcEndpoint string     ipc endpoint path
  -p, --password string        wallet password
  -m, --rpcMode string         use http rpc or ipc
  -w, --walletAddress string   wallet address

Global Flags:
  -r, --home string   path for the workspace 
```
There are two modes to communicate to a SDS resource node, and it could be switched by the --rpcMode flag

- `httpRpc` mode is to send RPC request over http. In this mode the `httpRpcUrl` flag must point to the rpc port of the
  resource node
   ``` shell
   sdss3 migrate --rpcMode httpRpc --httpRpcUrl http://<node url>:<node rpc port>
   ```
- `ipc` mode is to send PRC requests over IPC (Inter-process communication). The path to the ipc endpoint is set
  by the flag `ipcEndpoint`. The default path will be used when flag is not set.
   ```  shell
   sdss3 migrate --rpcMode ipc --ipcEndpoint <path to the ipc endpoint>
   ```
### Parameters
```shell
sdss3 migrate [flags] <bucket> <filename>
```
- `bucket` is the bucket to be downloaded from S3.  
- `filename` is an optional parameter. When it is given, only the specific file will be migrated to SDS instead of the whole bucket
to the SDS network.

### Config File
All the parameters could be pre-defined in the config file `config.toml` placed in the folder `config` under the home path 
(defined by -r or --home flag).
#### Template
```toml
[connectivity]
rpc_mode='httpRpc or ipc'
http_rpc_url='http://<node url>:<node rpc port>'
ipc_endpoint='path to the ipc endpoint'

[keys]
wallet_address = 'wallet address stxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'
wallet_password = 'wallet password'
```

### Folder Structure
Folder structure under the home path

| Folder   | Content                                                           |
|----------|-------------------------------------------------------------------|
| accounts | wallet files Eg: "stxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.json" |
| config   | config file "config.toml"                                          |

### AWS Credentials File
AWS credential needs to be stored in the shared configuration file (~/.aws/config)
#### Template
```
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-2
```

### S3 Migrate
The `sdss3 migrate` command migrates a file or a bucket from AWS S3 to SDS network. 

