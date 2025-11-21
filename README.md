# PINACLE 
Optimising food aid with AI-driven nutrition planning and privacy-preserving digital identity – turning surplus food into personalised wellbeing for vulnerable communities.

## About the project

Pinacle is a responsible digital solution designed to make food aid more efficient, transparent, and fair - connecting the logistics of food donation with the personal realities of nutrition and wellbeing. It demonstrates how intelligent technology and privacy-preserving data design can help food banks plan, distribute, and report in ways that are both operationally sound and socially responsible.

The project has developed an integrated, AI-driven platform that connects food banks, volunteers, and recipients through three interoperable tools: a mobile app for citizens, a web dashboard for food-bank operators, and a privacy-preserving backend that aligns donated food inventories with individual dietary profiles. Together, these components enable food-aid organisations to plan and distribute food more efficiently while offering recipients guidance that is healthy, culturally relevant, and respectful of their privacy.

## Unique Value

What sets PINACLE apart is its community impact, and the personalized nutrition matched on food donations and profile needs, backed by privacy-first architecture. Unlike generic food platforms, PINACLE delivers dietary recommendations aligned with everyone's health needs while efficiently allocating food donations. The use of DLT, Verifiable Credentials, and ZKPs ensures robust data, data privacy, user trust and compliance with EU regulations such as GDPR. Its dual focus—optimizing foodbanks' operations and empowering individuals—combined with user co-creation, creates scalable impact. The system’s modular design also facilitates replication across European contexts.

## Partners

Konnecta 

Co2gether (Food Bank of Western Greece)

Sapienza University of Rome

## The Verifier Smart Contract

The verifier smart contract developed within the PINACLE project enable the validation of Zero Knowledge Proofs (ZKPs) on-chain, ensuring privacy-preserving and secure access control to sensitive functionalities of the platform. These contracts allow decentralized verification of identity proofs without disclosing any personal data, thus supporting GDPR compliance and strengthening trust between food recipients, intermediaries, and food banks. By contributing this tool, PINACLE adds a reusable and scalable building block to the DRG4Food Toolbox for projects requiring privacy-preserving role-based access control.


## Installation
### Contract Compilation & ABI Generation

This project uses a Go tool to compile smart contracts and generate Go bindings.

#### 🛠 Compile Contracts

Run the following commands:

```bash
cd go-contracts
go run cmd/main.go compile
cp ../contracts/mimc/mimc.json ../contracts/bin
go run cmd/main.go abigen
```

#### 🚀 Deploying Contracts

To deploy the contracts, you only need to set **three environment variables** in the `deployer/.env` file:

```bash
GETH_NODE_URL=         # RPC URL of the Ethereum node, must be a valid URL (following standard URL formatting rules)
GETH_NODE_KEYSTORE=    # Path to the keystore directory, must exists
GETH_NODE_PASSWORD=    # Password for the keystore
```
and run the following commands:

```bash
cd deployer
go run cmd/deploy/deploy.go
```

#### 🔎 Running the ZKP Test

To run the Zero-Knowledge Proof test:

```bash
go run cmd/pinacle/pinacle.go
```

#### ⚠️ Important Notice About ZKP Files

Due to the large size of .zkey proving keys and verification keys, they are not included in the repository.
You must download them separately and update only these **three environment variables** in the `deployer/.env` file:

```bash
ZK_WASM_FILENAME=                # Path to the circuit .wasm file
ZK_ZKEY_FILENAME=                # Path to the proving key .zkey file
ZK_VERIFICATION_KEY_FILENAME=    # Path to the verification key JSON file
```


## Licensing

Distributed under the Apache 2.0 License. See [LICENSE](https://github.com/drg4food-pinacle/verifier-smart-contract/blob/main/LICENSE) for more information.

## Contact

Website, https://drg4foodtoolbox.eu/project/pinacle/

LinkedIn, https://www.linkedin.com/showcase/pinacle-project/

Email: info@konnecta.io

## Acknowledgements

The PINACLE project has indirectly received funding from the European Union’s Horizon Europe research and innovation action programme, via the DRG4FOOD – Open Call #2 issued and executed under the DRG4FOOD project (Grant Agreement no. 101086523)

