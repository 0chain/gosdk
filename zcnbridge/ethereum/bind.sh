#!/bin/sh
# shellcheck disable=SC2164
cd ./erc20
#solc --abi ERC20.sol | awk '/JSON ABI/{x=1;next}x' > erc20.abi
#solc --bin ERC20.sol | awk '/Binary:/{x=1;next}x' > erc20.bin
#abigen --bin=erc20.bin --abi=erc20.abi --pkg=erc20 --out=erc20.go
abigen --pkg erc20 --sol ERC20.sol --out ./erc20.go
