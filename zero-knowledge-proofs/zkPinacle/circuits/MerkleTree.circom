pragma circom 2.1.6;

include "./MerkleTreeChecker.circom";
include "./eth_addr.circom";
include "../../circomlib/circuits/comparators.circom";

template EthereumAddress(n, k) {
 signal input privateKey[k];
 signal output hashedAddr;

 // PrivKeyToAddr will compute the address given the private key
 // Keys are encoded as (x, y) pairs with each coordinate being
 // encoded with k registers of n bits each
 component privToAddr = PrivKeyToAddr(n, k);
 for (var i = 0; i < k; i++) {
        privToAddr.privkey[i] <== privateKey[i];
 }

 component hasher = HashLeftRight();
 hasher.left <== privToAddr.addr;
 hasher.right <== 0;
 
 hashedAddr <== hasher.hash;
}

// Finds the total in the array
template Sum(n) {
    signal input in[n];
    signal output out;

    signal cumulative[n];
    cumulative[0] <== in[0];

    for (var i = 1; i < n; i++) {
        cumulative[i] <== cumulative[i - 1] + in[i];
    }

    out <== cumulative[n - 1];
}

template MerkleTree(n, k, levels) {
    signal input privateKey[k];
    signal input pathElements[levels];
    signal input pathIndices[levels];
    signal output hashedAddr;
    signal output root;

    component calculatePathElements = Sum(levels);
    component calculatePathIndices = Sum(levels);
    for (var j = 0; j < levels; j++) {
        calculatePathElements.in[j] <== pathElements[j];
        calculatePathIndices.in[j] <== pathIndices[j];
    }
    
    // Check if PathElements are all zeros
    component iszPathElements = IsZero();
    iszPathElements.in <== calculatePathElements.out;

    // Check if PathIndices are all zeros
    component iszPathIndices = IsZero();
    iszPathIndices.in <== calculatePathIndices.out;

    // Check if both PathElements and PathIndices total outputs are equal
    signal zkEthereumAddress <-- (iszPathElements.out == 1 && iszPathElements.out == 1) ? 1 : 0; 
        
    // Use EthereumAddress circuit
    component ethereumAddress = EthereumAddress(n, k);
    for (var i = 0; i < k; i++) {
        ethereumAddress.privateKey[i] <== privateKey[i];
    }
    hashedAddr <== ethereumAddress.hashedAddr;
    
    // Verifies that merkle proof is correct for given merkle root and a leaf
    // pathIndices input is an array of 0/1 selectors telling whether given pathElement
    // is on the left or right side of merkle path
    component merkleTree = MerkleTreeChecker(levels);
    merkleTree.leaf <== ethereumAddress.hashedAddr;
    for (var j = 0; j < levels; j++) {
       merkleTree.pathElements[j] <== pathElements[j];
       merkleTree.pathIndices[j] <== pathIndices[j];
    }
    root <-- (zkEthereumAddress == 1) ? 0 : merkleTree.root;
}

component main = MerkleTree(64, 4, 32);
