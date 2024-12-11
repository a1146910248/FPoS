// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract L2Contract {
    struct BlockInfo {
        uint64 height;
        bytes32 blockHash;
        bytes32 stateRoot;
        uint256 timestamp;
    }

    mapping(uint64 => BlockInfo) public blocks;
    uint64 public latestHeight;

    event BlockSubmitted(
        uint64 indexed height,
        bytes32 blockHash,
        bytes32 stateRoot,
        uint256 timestamp
    );

    function submitBlock(
        uint64 height,
        bytes32 blockHash,
        bytes32 stateRoot
    ) external {
        require(height == latestHeight + 1, "Invalid block height");

        blocks[height] = BlockInfo({
            height: height,
            blockHash: blockHash,
            stateRoot: stateRoot,
            timestamp: block.timestamp
        });

        latestHeight = height;

        emit BlockSubmitted(
            height,
            blockHash,
            stateRoot,
            block.timestamp
        );
    }
}