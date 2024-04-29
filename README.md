---
title: SDS S3 Command Handbook
description: List of commands for Operation between SDS and AWS S3.
---

# SDS IPFS Command Handbook

## IPFS Client
The `ppd s3migrate` command migrates a bucket or a file from AWS S3 to sds network. The command tool needs to communicate with a SDS resource node to interact with the network. For setting up a
SDS resource node please refer to  [Setup and run a SDS Resource Node](../setup-and-run-a-sds-resource-node/)

``` { .yaml .no-copy }
ppd s3migrate -h

Usage:
  ppd s3migrate [flags] [parameters]

Flags:
  -h, --help                 help for ipfs
      --httpRpcUrl string    http rpc url (default "http://127.0.0.1:9301")
      --ipcEndpoint string   ipc endpoint path
  -p, --port string          port (default "6798")
  -m, --rpcMode string       use httpRpc or ipc (default "ipc")
      --password string      password to wallet

Global Flags:
  -r, --home string     path for the node (default "<root directory of your resource node>")
```

<br>

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

## S3 Migrate
The `ppd s3migrate` command migrates a file or a bucket from AWS S3 to SDS network. It first downloads the file from the S3 by the
given bucket or file key then uploads it to the SDS network.

```shell
ppd s3migrate <bucket> <filename>
```
`bucket` is the bucket to be downloaded from S3.  
`filename` is an optional parameter. When it is given, only the specific file will be migrated instead of the whole bucket
to the SDS network.
