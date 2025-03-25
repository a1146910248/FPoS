// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract DACEnhancedL2Contract {
    struct BlockInfo {
        uint64 height;
        bytes32 blockHash;
        bytes32 stateRoot;
        uint256 timestamp;
        // 新增DAC状态根
        bytes32 dacAccountRoot;
        bytes32 dacTxRoot;
        bool hasDAC;  // 标记是否包含DAC状态
    }

    // DAC证明结构
    struct DACProof {
        bytes32 accountRoot;
        bytes32 txRoot;
        bytes[] proof;
        address submitter;
        uint256 timestamp;
        bool isVerified;
    }

    // DAC成员结构
    struct DACMember {
        address addr;
        uint256 stake;
        bool active;
    }

    mapping(uint64 => BlockInfo) public blocks;
    mapping(uint64 => DACProof) public dacProofs;
    mapping(address => DACMember) public dacMembers;

    address[] public activeDACMembers;
    uint64 public latestHeight;
    uint256 public currentRandomNumber;
    address public owner;
    uint256 public dacMemberCount;
    bytes32 public latestDACAccountRoot;
    bytes32 public latestDACTxRoot;

    // 保留现有事件
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

    // 新增DAC相关事件
    event DACStateUpdated(
        uint64 indexed height,
        bytes32 accountRoot,
        bytes32 txRoot,
        address submitter
    );

    event DACProofSubmitted(
        uint64 indexed height,
        bytes32 accountRoot,
        address submitter,
        bool verified
    );

    event DACMemberAdded(
        address indexed member,
        uint256 stake
    );

    event DACMemberRemoved(
        address indexed member
    );

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    modifier onlyDAC() {
        require(dacMembers[msg.sender].active, "Only DAC members can call this function");
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

    // 保留原有submitBlock功能，但增加DAC状态参数
    function submitBlock(
        uint64 height,
        bytes32 blockHash,
        bytes32 stateRoot,
        bytes32 dacAccountRoot,
        bytes32 dacTxRoot
    ) external {
        require(height == latestHeight + 1, "Invalid block height");

        bool hasDAC = dacAccountRoot != bytes32(0) && dacTxRoot != bytes32(0);

        blocks[height] = BlockInfo({
            height: height,
            blockHash: blockHash,
            stateRoot: stateRoot,
            timestamp: block.timestamp,
            dacAccountRoot: dacAccountRoot,
            dacTxRoot: dacTxRoot,
            hasDAC: hasDAC
        });

        if (hasDAC) {
            latestDACAccountRoot = dacAccountRoot;
            latestDACTxRoot = dacTxRoot;
        }

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

        if (hasDAC) {
            emit DACStateUpdated(
                height,
                dacAccountRoot,
                dacTxRoot,
                msg.sender
            );
        }

        emit RandomNumberUpdated(
            oldRandomNumber,
            currentRandomNumber,
            height
        );
    }

    // 提交DAC证明
    function submitDACProof(
        uint64 height,
        bytes32 accountRoot,
        bytes32 txRoot,
        bytes[] memory proof
    ) external onlyDAC {
        require(blocks[height].height == height, "Block does not exist");
        require(blocks[height].hasDAC, "Block has no DAC state");
        require(blocks[height].dacAccountRoot == accountRoot, "Account root mismatch");
        require(blocks[height].dacTxRoot == txRoot, "Transaction root mismatch");

        // 存储证明
        dacProofs[height] = DACProof({
            accountRoot: accountRoot,
            txRoot: txRoot,
            proof: proof,
            submitter: msg.sender,
            timestamp: block.timestamp,
            isVerified: true  // 默认为已验证，实际项目中应该进行验证
        });

        emit DACProofSubmitted(
            height,
            accountRoot,
            msg.sender,
            true
        );
    }

    // 验证账户是否存在于DAC状态(简化版)
    function verifyAccount(
        uint64 height,
        address account,
        bytes[] memory proof,
        bytes memory accountData
    ) public view returns (bool) {
        require(blocks[height].hasDAC, "Block has no DAC state");
        require(dacProofs[height].isVerified, "DAC proof not verified");

        bytes32 leaf = keccak256(abi.encodePacked(account, accountData));
        bytes32 root = blocks[height].dacAccountRoot;

        // 这里应该实现默克尔证明验证逻辑
        // 简化版只检查proof数组长度
        return proof.length > 0;
    }

    // 验证交易是否存在于DAC状态(简化版)
    function verifyTransaction(
        uint64 height,
        bytes32 txHash,
        bytes[] memory proof
    ) public view returns (bool) {
        require(blocks[height].hasDAC, "Block has no DAC state");
        require(dacProofs[height].isVerified, "DAC proof not verified");

        bytes32 root = blocks[height].dacTxRoot;

        // 这里应该实现默克尔证明验证逻辑
        // 简化版只检查proof数组长度
        return proof.length > 0;
    }

    // 添加DAC成员
    function addDACMember(address member, uint256 stake) external onlyOwner {
        require(!dacMembers[member].active, "Member already active");
        require(stake > 0, "Stake must be positive");

        dacMembers[member] = DACMember({
            addr: member,
            stake: stake,
            active: true
        });

        activeDACMembers.push(member);
        dacMemberCount++;

        emit DACMemberAdded(member, stake);
    }

    // 移除DAC成员
    function removeDACMember(address member) external onlyOwner {
        require(dacMembers[member].active, "Member not active");

        dacMembers[member].active = false;

        // 从活跃列表中移除
        for (uint i = 0; i < activeDACMembers.length; i++) {
            if (activeDACMembers[i] == member) {
                // 将最后一个元素移到当前位置，然后删除最后一个元素
                activeDACMembers[i] = activeDACMembers[activeDACMembers.length - 1];
                activeDACMembers.pop();
                break;
            }
        }

        dacMemberCount--;

        emit DACMemberRemoved(member);
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
            delete dacProofs[i];
        }

        // 重置高度
        latestHeight = 0;
        latestDACAccountRoot = bytes32(0);
        latestDACTxRoot = bytes32(0);

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

    // 获取最新的DAC状态根
    function getLatestDACRoots() external view returns (bytes32, bytes32) {
        return (latestDACAccountRoot, latestDACTxRoot);
    }

    // 获取指定高度的DAC证明
    function getDACProof(uint64 height) external view returns (DACProof memory) {
        require(blocks[height].hasDAC, "Block has no DAC state");
        return dacProofs[height];
    }

    // 获取所有活跃DAC成员
    function getAllActiveDACMembers() external view returns (address[] memory) {
        return activeDACMembers;
    }

    // 检查地址是否是活跃的DAC成员
    function isActiveDACMember(address member) external view returns (bool) {
        return dacMembers[member].active;
    }
}