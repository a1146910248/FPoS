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
    uint256 public currentRandomNumber;
    address public owner;

    event BlockSubmitted(
        uint64 indexed height,
        bytes32 blockHash,
        bytes32 stateRoot,
        uint256 timestamp
    );

    event RandomNumberUpdated(
        uint256 oldValue,
        uint256 newValue,
        uint64 blockHeight
    );

    event StateReset(
        uint64 lastHeight,
        uint256 timestamp
    );

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    constructor() {
        owner = msg.sender;
        currentRandomNumber = uint256(
            keccak256(
                abi.encodePacked(
                    block.timestamp,
                    block.number,
                    blockhash(block.number - 1)
                )
            )
        );
    }

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

        // 更新随机数
        uint256 oldRandomNumber = currentRandomNumber;
        currentRandomNumber = uint256(
            keccak256(
                abi.encodePacked(
                    blockHash,
                    stateRoot,
                    block.timestamp,
                    block.number,
                    currentRandomNumber
                )
            )
        );

        emit BlockSubmitted(
            height,
            blockHash,
            stateRoot,
            block.timestamp
        );

        emit RandomNumberUpdated(
            oldRandomNumber,
            currentRandomNumber,
            height
        );
    }

    // 获取当前随机数
    function getRandomNumber() external view returns (uint256) {
        return currentRandomNumber;
    }

    // 重置所有状态
    function resetState() external onlyOwner {
        uint64 lastHeight = latestHeight;

        // 清空所有区块记录
        for(uint64 i = 0; i <= latestHeight; i++) {
            delete blocks[i];
        }

        // 重置高度
        latestHeight = 0;

        // 重新生成随机数
        currentRandomNumber = uint256(
            keccak256(
                abi.encodePacked(
                    block.timestamp,
                    block.number,
                    blockhash(block.number - 1)
                )
            )
        );

        emit StateReset(lastHeight, block.timestamp);
    }

    // 转移所有权
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "New owner cannot be zero address");
        owner = newOwner;
    }

    // 获取指定区间的区块信息
    function getBlockRange(uint64 fromHeight, uint64 toHeight)
    external
    view
    returns (BlockInfo[] memory)
    {
        require(fromHeight <= toHeight, "Invalid height range");
        require(toHeight <= latestHeight, "Height out of range");

        uint64 count = toHeight - fromHeight + 1;
        BlockInfo[] memory result = new BlockInfo[](count);

        for(uint64 i = 0; i < count; i++) {
            result[i] = blocks[fromHeight + i];
        }

        return result;
    }
}