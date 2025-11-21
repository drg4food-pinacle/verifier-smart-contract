pragma circom 2.1.6;

include "../../circomlib/circuits/mimcsponge.circom";

template CommitmentHasher() {
    signal input nullifier;
    signal input secret;
    signal output commitment;
    signal output nullifierHash;

    component commitmentHasher = MiMCSponge(2, 220, 1);
    component nullifierHasher = MiMCSponge(2, 220, 1);

    commitmentHasher.ins[0] <== nullifier;
    commitmentHasher.ins[1] <== secret;
    commitmentHasher.k <== 0;

    nullifierHasher.ins[0] <== nullifier;
    nullifierHasher.ins[1] <== 0;
    nullifierHasher.k <== 0;

    commitment <== commitmentHasher.outs[0];
    nullifierHash <== nullifierHasher.outs[0];
}
