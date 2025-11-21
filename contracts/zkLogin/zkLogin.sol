// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;
import "../MerkleTree/MerkleTreeWithHistory.sol";

// Voting Proof Verifier
interface IFoodBankVerifier {
    function verifyProof(
        uint[2] calldata _pA,
        uint[2][2] calldata _pB,
        uint[2] calldata _pC,
        uint[2] calldata _pubSignals
    ) external pure returns (bool r);
}

// This contract was designed and created for user (end-user and foodbanks) as Authentication Proxy
contract zkLogin is MerkleTreeWithHistory {
    /**
     ** Variables
     */
    IFoodBankVerifier private immutable foodBankVerifier; //ZKVoting Verifier

    // TreeIDs for Foodbanks and Users (RESERVED)
    uint32 private immutable FOODBANKS = 0;
    uint32 private immutable USERS = 1;

    /**
     ** Structs
     */
    struct Groth16Proof {
        uint256[2] pi_a; // Array representing the 'a' part of the proof
        uint256[2][2] pi_b; // Array of arrays representing the 'b' part of the proof
        uint256[2] pi_c; // Array representing the 'c' part of the proof
    }

    /**
     ** Mappings
     */
    mapping(uint32 => mapping(uint32 => mapping(uint256 => bool)))
        internal isExists; // Duplicate protection of leafs
    mapping(uint256 => uint256[]) private foodBankUsers;
    mapping(address => bool) private blacklist; // Revocation List

    /**
     ** Modifiers
     */
    // Protects against some weird attacks
    modifier validAddress(address _addr) {
        require(_addr != address(0), "Zero Address Detected");
        _;
    }

    // Verify Ethereum Address ZKP
    modifier validEthereumAddressZKP(
        address _user,
        Groth16Proof calldata _proof,
        uint256[2] calldata _publicSignals
    ) {
        // Verify that 1 public signal is zero
        require(
            _publicSignals[1] == 0,
            "zkEthereumAddress: Invalid Public Signals"
        );
        // Verify Proof (zkEthereumAddress)
        require(
            foodBankVerifier.verifyProof(
                _proof.pi_a,
                _proof.pi_b,
                _proof.pi_c,
                _publicSignals
            ),
            "zkEthereumAddress: Invalid Proofs"
        );
        // Verify that the hashedAddress of the transaction originator is the same as in
        // circuit's output
        require(
            hashAddress(_user) == _publicSignals[0],
            "zkEthereumAddress: Unauthorized Access"
        );
        _;
    }

    // Verify Merkle Tree ZKP
    modifier validMerkleTreeZKP(
        uint32 _treeId,
        uint32 _subtreeId,
        address _user,
        Groth16Proof calldata _proof,
        uint256[2] calldata _publicSignals
    ) {
        // Verify that 0 and 1 public signals are zeroes
        require(
            _publicSignals[0] != 0 && _publicSignals[1] != 0,
            "zkMerkleTree: Invalid Public Signals"
        );
        // Verify Proofs (zkMerkleTree)
        require(
            foodBankVerifier.verifyProof(
                _proof.pi_a,
                _proof.pi_b,
                _proof.pi_c,
                _publicSignals
            ),
            "zkMerkleTree: Invalid Proofs"
        );
        // Verify that the hashedAddress of the transaction originator is the same as in
        // circuit's output
        require(
            hashAddress(_user) == _publicSignals[0],
            "zkMerkleTree: Unauthorized Access"
        );
        // Verify that the root calculated from the circuit is known
        require(
            isKnownRoot(_treeId, _subtreeId, _publicSignals[1]),
            "zkMerkleTree: Unknown Root Detected"
        );
        _;
    }

    // Check if user is blacklisted
    modifier validAccount(address _user) {
        require(!blacklist[_msgSender()], "Blacklisted User Detected");
        _;
    }

    /**
     ** Events
     */
    // Event structure for logging contract interactions
    // event ContractInteraction(
    //     address indexed userAddress,
    //     string functionName,
    //     bytes data
    // );

    /***
     ** @author Constructor
     ** @notice Only Contract Owner
     ** @dev
     **     1) _trees: Number of trees in the Merkle tree
     **     2) _subtrees: Array of subtree sizes.
     **     3) _levels: Array of Merkle tree levels
     **     4) _hasher: Hasher contract address
     **     5) _foodBankVerifier: Voting verifier contract address
     **     6) _foodBanks: Array of initial food bank addresses
     */
    constructor(
        uint32 _trees,
        uint32[] memory _subtrees,
        uint32[] memory _levels,
        IHasher _hasher,
        IFoodBankVerifier _foodBankVerifier,
        address[] memory _foodBanks
    )
        validAddress(_msgSender())
        validAddress(address(_hasher))
        validAddress(address(_foodBankVerifier))
        MerkleTreeWithHistory(_trees, _subtrees, _levels, _hasher)
    {
        // Contract Address of the Voting Verifier
        foodBankVerifier = _foodBankVerifier;

        // Check if the array is empty
        uint256 foodBanksLength = _foodBanks.length;
        require(foodBanksLength != 0, "No FoodBanks' addresses presented");

        // Add initial addresses to the Merkle tree
        for (uint256 i; i < foodBanksLength; i++) {
            // Check if the address is empty
            if (_foodBanks[i] == address(0)) {
                continue;
            }
            // Register FoodBank
            _registerUser(FOODBANKS, 0, _foodBanks[i]);
            //emit ContractInteraction(_msgSender(), "constructor", abi.encodePacked(addressToBigint(_foodBanks[i])));
        }
    }

    /***
     ** @dev Terminates a Food Bank Account
     ** @notice Only Food Bank
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function terminateFoodBank(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals
    )
        external
        validAddress(_msgSender())
        validAccount(_msgSender())
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        returns (bool)
    {
        // Terminate the Account
        terminateAccount(_msgSender());
        // Delete FoodBank's Merkle Proofs
        deleteMerkleProof(FOODBANKS, 0);
        return true;
    }

    /***
     ** @dev Terminates a User Account
     ** @notice Only USERS
     ** @param
     **   1) _userMerkleProof: Zero Knowledge Merkle Proofs of the User (transactor)
     **   2) _userMerklePublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function terminateUser(
        Groth16Proof calldata _userMerkleProof,
        uint256[2] calldata _userMerklePublicSignals
    )
        external
        validAddress(_msgSender())
        validAccount(_msgSender())
        validMerkleTreeZKP(
            USERS,
            0,
            _msgSender(),
            _userMerkleProof,
            _userMerklePublicSignals
        )
        returns (bool)
    {
        // Terminate the Account
        terminateAccount(_msgSender());
        // Delete User's Merkle Proofs
        deleteMerkleProof(USERS, 0);
        return true;
    }

    /***
     ** @dev Verifies Proofs
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the FoodBank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     **   3) _user: The address of the User
     **   4) _userMerkleProof: Zero Knowledge Merkle Proofs of the Voting (_user)
     **   5) _userMerklePublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function verifyProof(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals,
        address _user,
        Groth16Proof calldata _userMerkleProof,
        uint256[2] calldata _userMerklePublicSignals
    )
        external
        view
        validAddress(_msgSender())
        validAddress(_user)
        validAccount(_msgSender())
        validAccount(_user)
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        validMerkleTreeZKP(
            USERS,
            0,
            _user,
            _userMerkleProof,
            _userMerklePublicSignals
        )
        returns (bool)
    {
        return true;
    }

    /**
     ** @dev Return Merkle Proofs for Food Banks
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankEthereumAddressProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankEthereumAddressPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) MerkleProof Struct (PathElements and PathIndices)
     */
    function fetchFoodBankMerkleProofs(
        Groth16Proof calldata _foodBankEthereumAddressProof,
        uint256[2] calldata _foodBankEthereumAddressPublicSignals
    )
        external
        view
        validAddress(_msgSender())
        validAccount(_msgSender())
        validEthereumAddressZKP(
            _msgSender(),
            _foodBankEthereumAddressProof,
            _foodBankEthereumAddressPublicSignals
        )
        returns (MerkleProof memory)
    {
        return fetchMerkleProof(FOODBANKS, 0);
    }

    /**
     ** @dev Delete Merkle Proofs for Food Banks
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function deleteFoodBankMerkleProofs(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals
    )
        external
        validAddress(_msgSender())
        validAccount(_msgSender())
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        returns (bool)
    {
        return deleteMerkleProof(FOODBANKS, 0);
    }

    /**
     ** @dev Return Merkle Proofs for Users
     ** @notice Only USERS
     ** @param
     **   1) _userEthereumAddressProof: Zero Knowledge Merkle Proofs of the User (transactor)
     **   2) _userEthereumAddressPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) MerkleProof Struct (PathElements and PathIndices)
     */
    function fetchUserMerkleProofs(
        Groth16Proof calldata _userEthereumAddressProof,
        uint256[2] calldata _userEthereumAddressPublicSignals
    )
        external
        view
        validAddress(_msgSender())
        validAccount(_msgSender())
        validEthereumAddressZKP(
            _msgSender(),
            _userEthereumAddressProof,
            _userEthereumAddressPublicSignals
        )
        returns (MerkleProof memory)
    {
        return fetchMerkleProof(USERS, 0);
    }

    /**
     ** @dev Delete Merkle Proofs for Users
     ** @notice Only USERS
     ** @param
     **   1) _userMerkleProof: Zero Knowledge Merkle Proofs of the User (transactor)
     **   2) _userPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function deleteUserMerkleProofs(
        Groth16Proof calldata _userMerkleProof,
        uint256[2] calldata _userPublicSignals
    )
        external
        validAddress(_msgSender())
        validAccount(_msgSender())
        validMerkleTreeZKP(
            USERS,
            0,
            _msgSender(),
            _userMerkleProof,
            _userPublicSignals
        )
        returns (bool)
    {
        return deleteMerkleProof(USERS, 0);
    }

    /**
     ** @dev Register a new Food Bank
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     **   3) _newFoodBank: The address of the new Food Bank
     **   4) _newFoodBankEthereumAddressProof: Zero Knowledge Ethereum Address Proofs of the new Food Bank
     **   5) _newFoodBankPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function registerFoodBank(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals,
        address _newFoodBank,
        Groth16Proof calldata _newFoodBankEthereumAddressProof,
        uint256[2] calldata _newFoodBankPublicSignals
    )
        external
        validAddress(_msgSender())
        validAddress(_newFoodBank)
        validAccount(_msgSender())
        validAccount(_newFoodBank)
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        validEthereumAddressZKP(
            _newFoodBank,
            _newFoodBankEthereumAddressProof,
            _newFoodBankPublicSignals
        )
        returns (bool)
    {
        return _registerUser(FOODBANKS, 0, _newFoodBank);
    }

    /**
     ** @dev Register a new User
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     **   3) _newUser: The address of the new User
     **   4) _newUserEthereumAddressProof: Zero Knowledge Ethereum Address Proofs of the new User
     **   5) _newUserPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) SUCCESS (true) or FAILED (false)
     */
    function registerUser(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals,
        address _newUser,
        Groth16Proof calldata _newUserEthereumAddressProof,
        uint256[2] calldata _newUserPublicSignals
    )
        external
        validAddress(_msgSender())
        validAddress(_newUser)
        validAccount(_msgSender())
        validAccount(_newUser)
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        validEthereumAddressZKP(
            _newUser,
            _newUserEthereumAddressProof,
            _newUserPublicSignals
        )
        returns (bool)
    {
        bool success = _registerUser(USERS, 0, _newUser);
        // Add the user to the food bank's list
        foodBankUsers[_foodBankPublicSignals[0]].push(hashAddress(_newUser));
        return success;
    }

    /**
     ** @dev Retrieves the array of hashed users
     ** @notice Only Food Banks
     ** @param
     **   1) _foodBankMerkleProof: Zero Knowledge Merkle Proofs of the Food Bank (transactor)
     **   2) _foodBankPublicSignals: Array representing the public signals (Lenght = 2)
     ** @return
     **   1) Array (uint256[]) of hashed addresses of the users
     */
    function fetchUsersAsFoodBank(
        Groth16Proof calldata _foodBankMerkleProof,
        uint256[2] calldata _foodBankPublicSignals
    )
        external
        view
        validAddress(_msgSender())
        validAccount(_msgSender())
        validMerkleTreeZKP(
            FOODBANKS,
            0,
            _msgSender(),
            _foodBankMerkleProof,
            _foodBankPublicSignals
        )
        returns (uint256[] memory)
    {
        // Ensure that only registered food banks can call this function
        require(
            foodBankUsers[_foodBankPublicSignals[0]].length > 0,
            "Not a registered food bank or no users found"
        );

        return foodBankUsers[_foodBankPublicSignals[0]];
    }

    /**
     ** Helper Functions
     */
    // RegisterUser registers a new User (FoodBank or User)
    function _registerUser(
        uint32 _treeId,
        uint32 _subtreeId,
        address _newUser
    )
        private
        validAddress(_msgSender())
        validAddress(_newUser)
        validAccount(_msgSender())
        validAccount(_newUser)
        validTree(_treeId, 0)
        validSubtree(_subtreeId)
        returns (bool)
    {
        // Convert address to BigInt and Hash leaf using mimc converting to uint256
        uint256 hashedAddress = hashAddress(_newUser);
        // Check if user is already registered
        require(
            !isExists[_treeId][_subtreeId][hashedAddress],
            "User is already Registered"
        );
        isExists[_treeId][_subtreeId][hashedAddress] = true;
        // Insert leaf to Merkle Tree and store indexes
        merkleProofs[_treeId][_subtreeId][_newUser] = _insert(
            _treeId,
            _subtreeId,
            hashedAddress
        );
        return true;
    }

    // Terminate a User Account
    function terminateAccount(
        address _user
    ) private validAddress(_msgSender()) returns (bool) {
        // Blacklist the Account
        blacklist[_user] = true;
        return true;
    }

    // MerkleProof returns the PathElements and PathIndices
    function fetchMerkleProof(
        uint32 _treeId,
        uint32 _subtreeId
    )
        private
        view
        validAddress(_msgSender())
        validTree(_treeId, 0)
        validSubtree(_subtreeId)
        validLevel(merkleTrees[_treeId][_subtreeId].levels)
        returns (MerkleProof memory)
    {
        return merkleProofs[_treeId][_subtreeId][_msgSender()];
    }

    // Delete Merkle Proofs
    function deleteMerkleProof(
        uint32 _treeId,
        uint32 _subtreeId
    )
        private
        validAddress(_msgSender())
        validTree(_treeId, 0)
        validSubtree(_subtreeId)
        validLevel(merkleTrees[_treeId][_subtreeId].levels)
        returns (bool)
    {
        delete merkleProofs[_treeId][_subtreeId][_msgSender()];
        return true;
    }

    // Retrieves the address of the transactor
    function _msgSender() internal view virtual returns (address) {
        return msg.sender;
    }

    // Hash an Address
    function hashAddress(address _addr) internal view returns (uint256) {
        return hashLeftRight(addressToBigint(_addr), 0);
    }

    // Covert address to BigInt
    function addressToBigint(address _addr) internal pure returns (uint256) {
        // Convert the address to a uint160, then to a uint256
        return uint256(uint160(_addr));
    }
}
