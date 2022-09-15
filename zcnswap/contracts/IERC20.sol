pragma solidity 0.8.9;
pragma abicoder v2;

interface IERC20 {
    function transferFrom(
        address sender,
        address recipient,
        uint256 amount
    ) external returns (bool);

    function approve(address spender, uint256 tokens)
        external
        returns (bool success);
}
