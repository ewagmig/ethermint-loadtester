# loadtester

`loadtester` is a tool designed for load testing and automated account setup on Ethereum-compatible (EVM-based) blockchain networks. It supports stress-testing node performance with simulated transaction scenarios and provides an easy way to populate a genesis file with multiple accounts for initial network setup.

## Features
- **offchain_feeding**
  - Automatically creates multiple accounts with a specified balance (e.g., 1 ETH) on the genesis file of an Ethermint-based chain (e.g., Canto).
  - Facilitates the generation of new accounts and updates the genesis file, making it ideal for network initialization or testing purposes.

- **evmtx**
  - Sends multiple Ethereum transactions via JSON-RPC to simulate various load scenarios on an Ethereum node.
  - Configurable transaction rates and scenarios make it suitable for performance benchmarking.

## Getting Started

Before using `loadtester`, ensure you have a configured network and an accessible genesis file for your chain.

### offchain_feeding

1: **Set Up the Network and Prepare the Genesis File**

Ensure the network is initialized, and the genesis file is accessible.

2: **Configure `offchain_feeding` in `config.toml`**

Edit the relevant section in the `config.toml` file:
```toml
[offchain_feeding]
acc_num = 1000000 # Number of accounts to create
bech_prefix = "basechain" # Bech32 prefix for the chain
genesis_loc = "/home/ubuntu/.basechaind/config/genesis.json" # Location of the genesis file
denom = "abasecoin" # Denomination used in the chain
```
   
3: **Run the `offchain_feeding` Command**

Execute the command to create accounts and update the genesis file:
```shell
$ loadtester offchain_feeding
```

At this stage, the genesis file should contain 1M accounts, each with 1 abasecoin.

4: **Distribute the Updated Genesis File**
Copy the updated genesis file to the node’s home directory. If you have multiple validators, ensure all nodes receive the same genesis file.

5: **Start the Nodes**
Launch the nodes using the updated genesis file.

### evmtx
1: **Configure evmtx in config.toml**

Edit the evmtx section of the config.toml file to specify the load testing parameters:

```toml
[evmtx]
gas_limit = 200000
gas_price = 0
sending_amt = 0
chain_id = 9000
duration = "10m"
tpu = 1500 # Transactions per time unit
time_unit = "300ms" # Send 1500 transactions every 300ms
acc_num = 1000000 # Number of accounts to use for sending transactions
key_algo = "eth_secp256k1"
node_num = 1
node_idx = 0
scenario = "eth_transfer_to_self" # Scenario to run
```

Available scenarios:
- eth_transfer_to_random
  - Uses the accounts created by offchain_feeding as senders.
  - Randomly generates recipients for the transactions.
- eth_transfer_to_known
  - Splits the accounts into two halves: one for senders and one for recipients.
- eth_transfer_to_self
  - Uses the same set of accounts as both senders and recipients.

2: **Run evmtx Command**

Execute the load testing with the specified configuration:
```shell
$ loadtester evmtx
```

