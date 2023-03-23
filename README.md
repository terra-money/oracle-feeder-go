# Terra Oracle Feeder

This is the Go implementation of [terra-money/oracle-feeder](https://github.com/terra-money/oracle-feeder)

This contains the Oracle feeder software that is used for periodically submitting oracle votes for the exchange rate of the different assets offered by the oracle chain. This implementation can be used as-is, but also can serve as a reference for creating your own custom oracle feeder.

## Overview

This solution has 2 components:

1. [`price-server`](cmd/price-server/)

   - Obtain information from various data sources (exchanges, forex data, etc),
   - Model the data,
   - Enable a url to query the data,

2. [`feeder`](cmd/feeder/)

   - Reads the exchange rates data from a data source (`price-server`)
   - Periodically submits vote and prevote messages following the oracle voting procedure

## Prerequisites

- Install [Go 1.20+](https://golang.org/)

## Instructions

1. Clone this repository

```sh
git clone https://github.com/terra-money/oracle-feeder-go
cd oracle-feeder-go
```

2. Configure and launch `price-server`, following instructions [here](price-server/).

```sh
go build cmd/price-server/price_server.go
./price_server
```

3. Configure and launch `feeder`, following instructions [here](cmd/feeder/).

```sh
cd feeder
go build cmd/feeder/feeder.go

TODO ...
```
