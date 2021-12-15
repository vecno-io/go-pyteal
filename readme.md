# go-pyteal

Functionality to aid in the development and testing of PyTeal contracts.

## Brief

The networking and account packages aims to steam line the management of network nodes for development testing of the application. At the same time, they provided a unified interface to go from devops on to testing and publishing without the need to adjust deployment routines. (With the exception that funding of accounts is only available when running the development network)

## Requirements

- Linux or macOS
- Golang version 1.17.0 or higher
- Python 3. The scripts assumes the Python executable is called `python3`.
- The [Algorand Node software][algorand-install]. A private network is used, hence there is no need to sync up MainNet or TestNet. `goal` is assumed to be in the PATH.

### Installation

Once you have [installed Go][golang-install], run this command
to install the `go-pyteal` package:

    go get github.com/vecno-io/go-pyteal

### Roadmap

- Update the network module to remove the dependency on goal.
- Implement proper account managment using ether the SDK or KDM.

[algorand-install]: https://developer.algorand.org/docs/run-a-node/setup/install/
[golang-install]: http://golang.org/doc/install.html
[sv]: http://semver.org/
