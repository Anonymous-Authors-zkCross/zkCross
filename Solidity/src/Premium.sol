pragma solidity ^0.4.24;

contract HealthPremium {

    address owner = msg.sender;           // The owner of the contract

    // The struct of the insurance policy
    struct Policy {       
        address ID;   
        address receiver;            
        address signer;                       // the public key and the address of the signer
        bytes signature;                      // the authority's signature

        bytes32 hashSecData;
        bytes32 hashCoefficient;
        uint32  premium;
        bytes   proof;
        bool isValid;                       // The flag of the policy(valid), in order to debug
    }
	
    Policy public policy;                   // Policy instance

    bool public locked = false;             // use for mutex

    // ----------------- modifier ----------------------------
    // The function body is inserted where the special symbol "_;" in the definition of a modifier appears.
    
    // This means that if the owner calls this function, the function is executed and otherwise, an exception is thrown.
    modifier onlyOwner {
        require(msg.sender == owner);
        _;
    }

    // This like the onlyOwner.
    modifier onlyReceiver {
        require(msg.sender == policy.receiver);
        _;
    }

    // This like the onlyOwner.
    modifier onlyOthers {
        require(msg.sender != owner);
        _;
    }
    
    // This is a mutex, only one account can call this function in the same time.
    modifier mutex() {
        require(locked == false);
        locked = true;
        _;
        locked = false;
    }

    // "address" converts to "uint", use to debug.
    function addressToUint(address addr) internal pure returns (uint res) {
        res = uint(addr);
    }

    constructor() public payable {}

    function  initPolicy(address recvAddr, address sigAddr, bytes sig, bytes32 hashData, bytes32 hashCoeff, uint32 result, bytes proofData) 
        onlyOwner
        payable
        public  
    {
        policy.ID = owner;
        policy.receiver = recvAddr;
        policy.signer = sigAddr;
        policy.signature = sig;
        policy.hashSecData = hashData;
        policy.hashCoefficient = hashCoeff;
        policy.premium = result;
        policy.proof = proofData;
        policy.isValid = false;
    }

    // ---------- get-value -----------------
    // there functions use for trueffle test.
    function getPolicyID() public returns (address) {
        return policy.ID;
    }
    function getPolicyReceiver() public returns (address) {
        return policy.receiver;
    }
    function getSigner() public returns (address) {
        return policy.signer;
    }
    function getSignature() public returns (bytes) {
        return policy.signature;
    }
    function getHashSecData() public returns (bytes32) {
        return policy.hashSecData;
    }
    function getHashCoefficient() public returns (bytes32) {
        return policy.hashCoefficient;
    }
    function getPremium() public returns (uint32) {
        return policy.premium;
    }
    function getProof() public returns (bytes) {
        return policy.proof;
    }
    function getPolicyIsValid() public returns (bool) {
        return policy.isValid;
    }
    
    // ----------user-opration------------------
    
    // print the info of the bill.
    //function printBill() onlyOwner returns (Bill[10] ){
	//    return (bill);
    //}

    // get the contract balance
    function getBalance() view public returns (uint){
        return address(this).balance;
    }

    // get the contract address
    function getContractAddrees() view public returns (address){
        return this;
    }

    // get given address balance
    function getBalance(address addr) view public returns (uint){
        return addr.balance;
    }
    
    //fallback function
    //Contracts that receive Ether directly (without a function call, i.e. using send or transfer) 
    //but do not define a fallback function throw an exception, sending back the Ether (this was different before Solidity v0.4.0). 
    //So if you want your contract to receive Ether, you have to implement a fallback function.
    function() payable external{ 
        uint x;
        x = 1; 
    }
     
    // Owner sends money to the smart contract.
    function fundDeposit() onlyReceiver payable public returns (bool success){ 
        //msg.sender => someone who call the contract.
        //msg.value => the money (wei)
        emit Deposit(msg.sender, msg.value);
        address(this).transfer(msg.value);
        return true;	
    }
    
    // deduct money from smart contract.
    function deduct() onlyReceiver payable public returns (bool){
        require(verifyProof() == true);
        bool success = false;
        if (getBalance() < policy.premium) {
            success = false;
            policy.isValid = false;
        } else {
            policy.receiver.transfer(policy.premium);
            emit Deduct(policy.receiver, policy.premium);
            success = true;
            policy.isValid = true;
        }
        return success;
    }
	
    //----------operate-event-----------
    event Deposit(address _from, uint _amount);  // so you can log these events
    event Deduct(address _to, uint _amount);

    //-------------Operation--------------
    
    /// destroy the contract and reclaim the leftover funds.
    function kill() public {
        require(msg.sender == owner);
        selfdestruct(msg.sender);
    }

    // verifyProof methods.
    function verifyProof() public returns( bool ){
        require(recoverSigner(policy.hashSecData, policy.signature) == policy.signer);

       //gyh require(verProof(policy.proof, policy.hashSecData, policy.hashCoefficient, policy.premium) == 1);

        return true;

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
