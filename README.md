# Terra Oracle Feeder

This repo contains the Go implementation of [terra-money/oracle-feeder](https://github.com/terra-money/oracle-feeder) with some additional features related to the [Alliance Protocol](https://github.com/terra-money/alliance-protocol).

## Overview

The Oracle has two components: 

- [`price-server`](cmd/price-server/): requests pricing and data for the [Alliance Protocol](https://github.com/terra-money/alliance-protocol).
   
- [`feeder`](cmd/feeder/): requests data from the price-server, signs it, and submits it on chain.

## Price Server entry points

- **`GET:/health`**: checks if the price server is up and reachable.
  
  Response:
  
    ```
    OK
    ```

- **`GET:/latest`**: requests the latest prices for the configured tokens from different data sources. Returns the value of each token in USD. Additionally, it adds the timestamp when the response has been created.

   Response: 

    ```JSON
    {
        "created_at": "2023-08-10T09:22:39Z",
        "prices": [
            {
                "denom": "LUNA",
                "price": 0.5586257361595627
            },
            {
                "denom": "BTC",
                "price": 29574.81793975724
            },
            {
                "denom": "ETH",
                "price": 1853.0755933960982
            }
        ]
    }
    ```

- **`GET:/alliance/protocol`**: builds the object needed by the Alliance Oracle smart contract given different data sources. 

  Response: 

    ```JSON
    {
        "update_chains_info": {
            "chains_info": {
                "luna_price": "0.558580061990141200",
                "protocols_info": [
                    {
                        "chain_id": "migaloo-1",
                        "native_token": {
                            "denom": "uwhale",
                            "token_price": "0.016258975361478255",
                            "annual_provisions": "23700941391808.640000000000000000"
                        },
                        "luna_alliances": [
                            {
                                "ibc_denom": "ibc/05238E98A143496C8AF2B6067BABC84503909ECE9E45FBCBAC2CBA5C889FD82A",
                                "rebase_factor": "1.210223120423448490",
                                "normalized_reward_weight": "0.023809523809523810",
                                "annual_take_rate": "0.009999998624824108",
                                "total_lsd_staked": "95485.002736000000000000"
                            }
                        ],
                        "chain_alliances_on_phoenix": [
                            {
                                "ibc_denom": "ibc/B3F639855EE7478750CC8F82072307ED6E131A8EFF20345E1D136B50C4E5EC36",
                                "rebase_factor": "1.044440454507987116"
                            }
                        ]
                    }
                ]
            }
        }
    }
    ```


- **`GET:/alliance/rebalance`**: queries Terra's blockchain for information on where the Alliance Hub smart contract delegates its stake and checks if the validators are still compliant with the [Alliance Hub validator rules](https://github.com/terra-money/oracle-feeder-go/blob/main/internal/provider/alliance/alliance_validators.go#L139-L147). If any validators currently staked to are not compliant or if they have more stake than they should, this endpoint is in charge of rebalancing the stake between all the other eligible validators. 

   Response: 

    ```JSON
    {
    "alliance_redelegate": {
        "redelegations": [
        {
            "amount": "499749",
            "dst_validator": "terravaloper1zdpgj8am5nqqvht927k3etljyl6a52kwqndjz2",
            "src_validator": "terravaloper1nqp2hmrynlsu6jcv6j9t5r2qgyvtvna58t2erh"
        }
        ]
    }
    }
    ```

- **`GET:/alliance/delegations`**: queries Terra's blockchain for information on the current set of validators and the amount of unstaked coins that the Alliance protocol has available. The endpoint applies the [Alliance Hub's rules](https://github.com/terra-money/oracle-feeder-go/blob/main/internal/provider/alliance/alliance_validators.go#L98-L106) and creates an initial delegation message to stake the coins. 

   Response: 

    ```JSON
    {
        "alliance_delegate": {
            "delegations": [
                {
                    "validator": "terravaloper1q8w4u2wyhx574m70gwe8km5za2ptanny9mnqy3",
                    "amount": "33333333333"
                }
            ]
        }
    }
    ```

## Feeder CLI

The Feeder CLI receives a single argument from the following list and performs the specified action. 

Examples of how to use the feeder are available in the [Makefile](./Makefile)
(e.g. `go run ./cmd/feeder/feeder.go alliance-initial-delegation`).

- **`alliance-initial-delegation`**: initiates a REST request to the `GET:/alliance/delegations` endpoint, signs the data, and submits it on chain to the Alliance Hub smart contract to perform the initial delegations.

- **`alliance-oracle-feeder`**: initiates a REST request to the `GET:/alliance/protocol` endpoint, signs the data, and submits it on chain to the Alliance Oracle smart contract. This data provides the contract with information about the markets and available assets to be delegated.

- **`alliance-rebalance-feeder`**: initiates a REST request to the endpoint `GET:/alliance/rebalance`, signs the data, and submits it on chain to the Alliance Hub smart contract. The stake is then rebalanced between the compliant validators according to the rules of Alliance Hub.

- **`alliance-update-rewards`**: creates an [update_rewards execute message](https://github.com/terra-money/alliance-protocol/blob/main/packages/alliance-protocol/src/alliance_protocol.rs#L37), signs the message, and submits it on chain.

- **`alliance-rebalance-emissions`**: creates a [rebalance_emissions execute message](https://github.com/terra-money/alliance-protocol/blob/main/packages/alliance-protocol/src/alliance_protocol.rs#L37), signs the message, and submits it on chain.

## Installation

1. Install [Go 1.20+](https://golang.org/).

2. Define the following variables in the `params.json` file. These variables can also be found in [.env.example](.env.example):

    **Note:** Keep in mind that the mnemonic below must be the same as `controller_addr` and `controller` from the deployed smart contracts on chain.

    ```sh
    # Port for the price server to listen
    PRICE_SERVER_PORT=8532
    # URL used to retreive the prices from
    PRICE_SERVER_URL=http://localhost:8532
    # Used by the feeder to derive the private key and signs the transactions
    MNEMONIC=...
    # Used for the feeder to submit the transactions on chain
    NODE_GRPC_URL=pisco-grpc.terra.dev:443
    # LCD URL used by the feeder to query the chain
    TERRA_LCD_URL=https://pisco-lcd.terra.dev
    # Address of the oracle contract where the alliance data is stored
    ORACLE_ADDRESS=terra19cs0qcqpa5n7wlqjlhvfe235kc20g52enwr0d5cdar5zkmza7skqv54070
    # Address of the alliance hub contract
    ALLIANCE_HUB_CONTRACT_ADDRESS=terra1q95pe55eea0akft0xezak2s50l4vkkquve5emw7gzw65a7ptdl8qel50ea
    # How many retries to do if the feeder fails to submit the transaction
    FEEDER_RETRIES=4
    # Chain ID where the oracle contract is deployed
    CHAIN_ID=pisco-1
    # Station API (https://station.terra.money/)
    STATION_API=https://pisco-api.terra.dev
    # Minimum amount of blocks to be considered as a rebalancing validator
    BLOCKS_TO_BE_SENIOR_VALIDATOR=100000
    # Minimum amount of votes  for proposals to be a rebalancing validator
    VOTE_ON_PROPOSALS_TO_BE_SENIOR_VALIDATOR=3
    ```

3. Use the price server start command:

    ```sh
    $ make start-price-server
    ```

4. To feed the contracts with data, use the following commands:
    ```sh
    $ make start-alliance-initial-delegation
    $ make start-alliance-oracle-feeder
    $ make start-alliance-rebalance-feeder
    $ make start-alliance-update-rewards
    $ make start-alliance-rebalance-emissions
    ```
