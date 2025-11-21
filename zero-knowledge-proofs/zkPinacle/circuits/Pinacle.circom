pragma circom 2.2.1;

include "./CommitmentHasher.circom";
include "./MerkleTreeChecker.circom";
include "./eth_addr.circom";
include "../../circomlib/circuits/comparators.circom";
include "../../circomlib/circuits/mimcsponge.circom";

template EthereumAddress(n, k) {
    signal input privateKey[k];
    signal output address;  
    // PrivKeyToAddr will compute the address given the private key
    // Keys are encoded as (x, y) pairs with each coordinate being
    // encoded with k registers of n bits each
    component privToAddr = PrivKeyToAddr(n, k);
    for (var i = 0; i < k; i++) {
        privToAddr.privkey[i] <== privateKey[i];
    }
    address <== privToAddr.addr;
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

template Pinacle(n, k, levels) {
    signal input privateKey[k];
    signal input pathElements[levels];
    signal input pathIndices[levels];
    signal output hashedAddr;
    signal output root;

    // Calculate PathElements and PathIndices
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
    signal zkEthereumAddress <-- (iszPathElements.out == 1 && iszPathIndices.out == 1) ? 1 : 0;
    
    // Use EthereumAddress circuit
    component ethereumAddress = EthereumAddress(n, k);
    for (var i = 0; i < k; i++) {
        ethereumAddress.privateKey[i] <== privateKey[i];
    }
  
    // Commitment Hasher will compute the MIMC hash of the provided
    // secret and the computed from the previous step nullifier (address)
    component hasher = CommitmentHasher();
    hasher.nullifier <== ethereumAddress.address;
    hasher.secret <== 0;
    
    // Verifies that merkle proof is correct for given merkle root and a leaf
    // pathIndices input is an array of 0/1 selectors telling whether given pathElement
    // is on the left or right side of merkle path
    component merkleTree = MerkleTreeChecker(levels);
    merkleTree.leaf <== hasher.nullifierHash;
    for (var j = 0; j < levels; j++) {
       merkleTree.pathElements[j] <== pathElements[j];
       merkleTree.pathIndices[j] <== pathIndices[j];
    }

    hashedAddr <== hasher.nullifierHash;
    root <-- (zkEthereumAddress == 1) ? 0 : merkleTree.root;
}

component main = Pinacle(64, 4, 32);
