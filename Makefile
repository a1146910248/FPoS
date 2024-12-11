.PHONY: generate-contract

generate-contract:
    solc --abi --bin contracts/L2Contract.sol -o contracts/build
    abigen --abi=contracts/build/L2Contract.abi --bin=contracts/build/L2Contract.bin --pkg=ethereum --out=core/ethereum/l2_contract.go