// Based on https://github.com/tornadocash/tornado-core/blob/master/contracts/MerkleTreeWithHistory.sol

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

interface IHasher {
    function MiMCSponge(
        uint256 in_xL,
        uint256 in_xR,
        uint256 k
    ) external pure returns (uint256 xL, uint256 xR);
}

contract MerkleTreeWithHistory {
    /*
     ** Variables
     */
    uint256 private constant FIELD_SIZE =
        21888242871839275222246405745257275088548364400416034343698204186575808495617;

    // Precomputed: Equals to keccak256("tornado") % FIELD_SIZE
    uint256 private constant ZERO_VALUE =
        21663839004416932945382355908790599225266501822907911457504978515578255421292;

    // Zero Leaf
    uint256 private constant ZERO_INPUT = 0;

    // Mimc Hasher Interface
    IHasher private immutable hasher;

    // Merkle Key Index
    uint32 internal merkleKeyIndex = 0;

    // Max Subtrees Allowed
    uint32 internal immutable MAXIMUM_ALLOWED_SUBTREES = 3;
    uint32 internal immutable MAXIMUM_ALLOWED_LEVELS = 32;

    /*
     ** Structs
     */
    struct Tree {
        uint32 nextIndex;
        uint32 levels;
        mapping(uint256 => uint256) filledSubtrees;
    }

    struct MerkleProof {
        uint256[] pathElements;
        uint256[] pathIndices;
    }

    /*
     ** Mappings (move to private or internal after conclusion)
     */
    mapping(uint32 => mapping(uint32 => mapping(address => MerkleProof)))
        internal merkleProofs;
    mapping(uint32 => mapping(uint32 => mapping(uint256 => bool)))
        internal roots;
    mapping(uint32 => mapping(uint32 => Tree)) internal merkleTrees;
    mapping(uint32 => uint256) private zeroValues;
    mapping(uint32 => uint32) private pow2Values;

    /*
     ** Modifiers
     */
    // Check if the TreeId is valid
    modifier validTree(uint32 _treeId, uint32 index) {
        require(
            _treeId >= index && _treeId < merkleKeyIndex,
            "Invalid Tree Detected"
        );
        _;
    }

    // Check if the SubtreeId is valid
    modifier validSubtree(uint32 _subtreeId) {
        require(
            _subtreeId >= 0 && _subtreeId < MAXIMUM_ALLOWED_SUBTREES,
            "Invalid Subtree Detected. Subtrees should be [0, 3)"
        );
        _;
    }

    // Check if Level is valid
    modifier validLevel(uint32 _level) {
        require(
            _level > 0 && _level <= MAXIMUM_ALLOWED_LEVELS,
            "Invalid Level Detected. Levels should be (0, 32]"
        );
        _;
    }

    // Check if leaf is valid
    modifier validInput(uint256 _input) {
        require(_input != ZERO_INPUT, "Invalid Leaf/Root Detected");
        _;
    }

    /*
     ** Events (to be removed)
     */
    /*event TreeEvent(string msg, uint32 i, uint32 j, bytes32 filledSubtrees, uint32 index, uint32 level);
 event Test(uint32 i, string _tree, uint32 _subtrees, uint32 _levels);*/
    //event Insert(uint32 tree, uint32 subtree, uint256 leaf, uint256 root, /*uint256[] filledsubtrees,*/ uint256[] pathElements, uint256[] pathIndices, uint32 index);

    /*
     ** Constructor
     */
    constructor(
        uint32 _trees,
        uint32[] memory _subtrees,
        uint32[] memory _levels,
        IHasher _hasher
    ) {
        require(
            _trees == _subtrees.length && _subtrees.length == _levels.length,
            "Length of Trees, Subtrees and Levels mismatch"
        );
        hasher = _hasher; // Contract Address of the hasher

        initZeros(); // Populates zeros in the HashMap
        initPowers(); // Populates powersof2 values in the HashMap

        for (uint32 i; i < _trees; i++) {
            createTreeWithSubtrees(_subtrees[i], _levels[i]);
        }
    }

    /*
     ** @dev Creates a new Tree with multiple Subtrees
     */
    function createTreeWithSubtrees(
        uint32 _subtrees,
        uint32 _levels
    ) internal validLevel(_levels) returns (uint32) {
        require(
            _subtrees > 0 && _subtrees <= MAXIMUM_ALLOWED_SUBTREES,
            "Maximum Allowed Subtrees are 3"
        );

        // Current Merkle Tree Index
        uint32 i = merkleKeyIndex;
        for (uint32 j; j < _subtrees; j++) {
            for (uint32 k; k < _levels; k++) {
                // Store trees
                merkleTrees[i][j].filledSubtrees[k] = zeros(k);
            }
            merkleTrees[i][j].nextIndex;
            merkleTrees[i][j].levels = _levels;

            // Store Root
            roots[i][j][zeros(_levels - 1)] = true;
        }
        merkleKeyIndex += 1;
        return i;
    }

    /*
     ** @dev Hash 2 tree leaves, returns MiMC(_left, _right)
     */
    function hashLeftRight(
        uint256 _left,
        uint256 _right
    ) internal view returns (uint256) {
        require(_left < FIELD_SIZE, "_left should be inside the field");
        require(_right < FIELD_SIZE, "_right should be inside the field");
        uint256 R = _left;
        uint256 C;
        (R, C) = hasher.MiMCSponge(R, C, 0);
        R = addmod(R, _right, FIELD_SIZE);
        (R, C) = hasher.MiMCSponge(R, C, 0);
        return R;
    }

    /*
     ** @dev Insert a new leaf in the Merkle Tree
     */
    function _insert(
        uint32 _tree,
        uint32 _subtree,
        uint256 _leaf
    )
        internal
        validTree(_tree, 0)
        validSubtree(_subtree)
        validLevel(merkleTrees[_tree][_subtree].levels)
        validInput(_leaf)
        returns (MerkleProof memory)
    {
        // Fetch the tree
        // Ensures that subtree exists in the tree
        uint32 _levels = merkleTrees[_tree][_subtree].levels;
        uint32 _nextIndex = merkleTrees[_tree][_subtree].nextIndex;
        require(
            _nextIndex <= pow2(_levels),
            "Merkle tree is full. No more leaves can be added"
        );

        uint32 currentIndex = _nextIndex;
        uint256 currentLevelHash = _leaf;

        uint256 left;
        uint256 right;

        uint256[] memory pathElements = new uint256[](_levels);
        uint256[] memory pathIndices = new uint256[](_levels);

        for (uint32 i; i < _levels; i++) {
            if (currentIndex % 2 == 0) {
                pathElements[i] = zeros(i);
                pathIndices[i] = 0; // Right Path
                left = currentLevelHash;
                right = zeros(i);
                merkleTrees[_tree][_subtree].filledSubtrees[
                    i
                ] = currentLevelHash;
            } else {
                pathElements[i] = merkleTrees[_tree][_subtree].filledSubtrees[
                    i
                ];
                pathIndices[i] = 1; // Left Path
                left = merkleTrees[_tree][_subtree].filledSubtrees[i];
                right = currentLevelHash;
            }
            // Hash the data using MIMC
            currentLevelHash = hashLeftRight(left, right);
            //currentIndex /= 2;
            currentIndex = currentIndex >> 1;
        }

        // Store root in roots map
        roots[_tree][_subtree][currentLevelHash] = true;

        require(
            pathElements.length == _levels && pathIndices.length == _levels,
            "Invalid pathElements or pathIndices length Detected."
        );

        //emit Insert(_tree, _subtree, _leaf, currentLevelHash, /*filledsubtrees,*/ pathElements, pathIndices, _nextIndex);

        // Increase the index
        merkleTrees[_tree][_subtree].nextIndex = _nextIndex + 1;

        // Return Merkle Proof
        return MerkleProof(pathElements, pathIndices);
    }

    /*
     ** @dev Whether the root is present in the root history
     */
    function isKnownRoot(
        uint32 _tree,
        uint32 _subtree,
        uint256 _root
    )
        internal
        view
        validTree(_tree, 0)
        validSubtree(_subtree)
        validLevel(merkleTrees[_tree][_subtree].levels)
        validInput(_root)
        returns (bool)
    {
        return roots[_tree][_subtree][_root];
    }

    /*
     ** @dev Provides precomputed powers of 2. Up to 32 levels
     */
    function pow2(uint32 i) private view returns (uint32) {
        require(i > 0 && i <= 32, "Index out of bounds");
        return pow2Values[i];
    }

    function initPowers() private {
        pow2Values[1] = 2;
        pow2Values[2] = 4;
        pow2Values[3] = 8;
        pow2Values[4] = 16;
        pow2Values[5] = 32;
        pow2Values[6] = 64;
        pow2Values[7] = 128;
        pow2Values[8] = 256;
        pow2Values[9] = 512;
        pow2Values[10] = 1024;
        pow2Values[11] = 2048;
        pow2Values[12] = 4096;
        pow2Values[13] = 8192;
        pow2Values[14] = 16384;
        pow2Values[15] = 32768;
        pow2Values[16] = 65536;
        pow2Values[17] = 131072;
        pow2Values[18] = 262144;
        pow2Values[19] = 524288;
        pow2Values[20] = 1048576;
        pow2Values[21] = 2097152;
        pow2Values[22] = 4194304;
        pow2Values[23] = 8388608;
        pow2Values[24] = 16777216;
        pow2Values[25] = 33554432;
        pow2Values[26] = 67108864;
        pow2Values[27] = 134217728;
        pow2Values[28] = 268435456;
        pow2Values[29] = 536870912;
        pow2Values[30] = 1073741824;
        pow2Values[31] = 2147483648;
        // Integers in Solidity are restricted to a certain range.
        // For example, with uint32, this is 0 up to 2**32 - 1.
        // Thus, 2**32 will not fit into a uint32 variable.
        // So by subtracting 1, (2**32 - 1 = 4294967296 - 1 = 4294967295) the
        // variable fits in a uint32 variable.
        pow2Values[32] = 4294967295;
    }

    /*
     ** @dev Provides Zero (Empty) elements for a MiMC MerkleTree.
     **      Up to 32 levels
     */
    function zeros(uint32 i) private view returns (uint256) {
        require(i >= 0 && i < 32, "Index out of bounds");
        return zeroValues[i];
    }

    // Initialize zeroValues mapping with precomputed values
    function initZeros() private {
        zeroValues[
            0
        ] = 21663839004416932945382355908790599225266501822907911457504978515578255421292;
        zeroValues[
            1
        ] = 16923532097304556005972200564242292693309333953544141029519619077135960040221;
        zeroValues[
            2
        ] = 7833458610320835472520144237082236871909694928684820466656733259024982655488;
        zeroValues[
            3
        ] = 14506027710748750947258687001455876266559341618222612722926156490737302846427;
        zeroValues[
            4
        ] = 4766583705360062980279572762279781527342845808161105063909171241304075622345;
        zeroValues[
            5
        ] = 16640205414190175414380077665118269450294358858897019640557533278896634808665;
        zeroValues[
            6
        ] = 13024477302430254842915163302704885770955784224100349847438808884122720088412;
        zeroValues[
            7
        ] = 11345696205391376769769683860277269518617256738724086786512014734609753488820;
        zeroValues[
            8
        ] = 17235543131546745471991808272245772046758360534180976603221801364506032471936;
        zeroValues[
            9
        ] = 155962837046691114236524362966874066300454611955781275944230309195800494087;
        zeroValues[
            10
        ] = 14030416097908897320437553787826300082392928432242046897689557706485311282736;
        zeroValues[
            11
        ] = 12626316503845421241020584259526236205728737442715389902276517188414400172517;
        zeroValues[
            12
        ] = 6729873933803351171051407921027021443029157982378522227479748669930764447503;
        zeroValues[
            13
        ] = 12963910739953248305308691828220784129233893953613908022664851984069510335421;
        zeroValues[
            14
        ] = 8697310796973811813791996651816817650608143394255750603240183429036696711432;
        zeroValues[
            15
        ] = 9001816533475173848300051969191408053495003693097546138634479732228054209462;
        zeroValues[
            16
        ] = 13882856022500117449912597249521445907860641470008251408376408693167665584212;
        zeroValues[
            17
        ] = 6167697920744083294431071781953545901493956884412099107903554924846764168938;
        zeroValues[
            18
        ] = 16572499860108808790864031418434474032816278079272694833180094335573354127261;
        zeroValues[
            19
        ] = 11544818037702067293688063426012553693851444915243122674915303779243865603077;
        zeroValues[
            20
        ] = 18926336163373752588529320804722226672465218465546337267825102089394393880276;
        zeroValues[
            21
        ] = 11644142961923297861823153318467410719458235936926864848600378646368500787559;
        zeroValues[
            22
        ] = 14452740608498941570269709581566908438169321105015301708099056566809891976275;
        zeroValues[
            23
        ] = 7578744943370928628486790984031172450284789077258575411544517949960795417672;
        zeroValues[
            24
        ] = 5265560722662711931897489036950489198497887581819190855722292641626977795281;
        zeroValues[
            25
        ] = 731223578478205522266734242762040379509084610212963055574289967577626707020;
        zeroValues[
            26
        ] = 20461032451716111710758703191059719329157552073475405257510123004109059116697;
        zeroValues[
            27
        ] = 21109115181850306325376985763042479104020288670074922684065722930361593295700;
        zeroValues[
            28
        ] = 81188535419966333443828411879788371791911419113311601242851320922268145565;
        zeroValues[
            29
        ] = 7369375930008366466575793949976062119589129382075515225587339510228573090855;
        zeroValues[
            30
        ] = 14128481056524536957498216347562587505734220138697483515041882766627531681467;
        zeroValues[
            31
        ] = 20117374654854068065360091929240690644953205021847304657748312176352011708876;
    }
}
