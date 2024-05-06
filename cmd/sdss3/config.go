package main

type Config struct {
	Connectivity ConnectivityConfig `toml:"connectivity"`
	Keys         KeysConfig         `toml:"keys"`
}

type ConnectivityConfig struct {
	RpcMode     string `toml:"rpc_mode"`
	HttpRpcUrl  string `toml:"http_rpc_url"`
	IpcEndpoint string `toml:"ipc_endpoint"`
}

type KeysConfig struct {
	WalletAddress  string `toml:"wallet_address" comment:"Address of the stratos wallet. Eg: \"stxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\""`
	WalletPassword string `toml:"wallet_password"`
}
