pragma solidity ^0.4.24;

contract ReceiverPays {
    address owner = msg.sender;

    constructor() public payable {}

    function claimPayment(uint256 amount, uint256 nonce, bytes signature) public {

        // this recreates the message that was signed on the client
        bytes32 message = prefixed(keccak256(abi.encodePacked(msg.sender, amount, nonce, this)));

        require(recoverSigner(message, signature) == owner);

        msg.sender.transfer(amount);
    }

    /// destroy the contract and reclaim the leftover funds.
    function kill() public {
        require(msg.sender == owner);
        selfdestruct(msg.sender);
    }

    /// signature methods.
    function splitSignature(bytes sig)
        internal
        pure
        returns (uint8 v, bytes32 r, bytes32 s)
    {
        require(sig.length == 65);

        assembly {
            // first 32 bytes, after the length prefix.
            r := mload(add(sig, 32))
            // second 32 bytes.
            s := mload(add(sig, 64))
            // final byte (first byte of the next 32 bytes).
            v := byte(0, mload(add(sig, 96)))
        }

        return (v, r, s);
    }

    function recoverSigner(bytes32 message, bytes sig)
        internal
        pure
        returns (address)
    {
        (uint8 v, bytes32 r, bytes32 s) = splitSignature(sig);

        return ecrecover(message, v, r, s);
    }

    /// builds a prefixed hash to mimic the behavior of eth_sign.
    function prefixed(bytes32 hash) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", hash));
    }
}

// address owner = msg.sender;

// mapping(uint256 => bool) usedNonces;
// function claimPayment(uint256 amount, uint256 nonce, bytes signature) 
// splitSignature(bytes sig) returns (uint8 v, bytes32 r, bytes32 s)
// recoverSigner(bytes32 message, bytes sig) returns (address)
// prefixed(bytes32 hash) internal pure returns (bytes32) 
// kill() 
// constructor() public payable {}


// abi=[{"constant":false,"inputs":[{"name":"amount","type":"uint256"}, {"name":"nonce","type":"uint256"}, {"name":"signature","type":"bytes"}], "name":"claimPayment","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"sig","type":"bytes"}], "name":"splitSignature", "outputs":[{"name":"v","type":"uint8"}, {"name":"r","type":"bytes32"}, {"name":"s","type":"bytes32"}],"payable":false,"type":"function"},{"constant":true,"input":[{"name":"message","type":"bytes32"},{"name":"sig","type":"bytes"}],"name":"recoverSigner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"hash","type":"bytes32"}],"name":"prefixed","outputs":[{"name":"","type":"bytes32"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"kill","outputs":[],"payable":false,"type":"function"},{"inputs":[],"payable":true,"type":"constructor"}]

