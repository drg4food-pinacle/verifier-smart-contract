#! /bin/bash
set -e
##########################################################
#########################[COLORS] ########################
##########################################################
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if no arguments are provided
if [ $# -eq 0 ]; then
  echo -e "${RED}Usage: ./setup.sh [FLAG] [VALUE]              ${NC}"
  echo -e "${RED}Options:                                      ${NC}"
  echo -e "${RED}  --circom CIRCOM_FILE  Specify a circom file ${NC}"
  exit 1
fi

# Initialize variables
CIRCOM_FILE=
POWER_OF_TAU=

# Process command-line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --circom)
      shift
      CIRCOM_FILE="$1"
      ;;
    --power)
      shift
      POWER_OF_TAU="$1"
      ;;
    *)
      echo "Invalid option: $1"
      exit 1
      ;;
  esac
  shift
done

# Validate circom file
if [ -z "${CIRCOM_FILE}" ]; then
  echo -e "${RED}Error: --circom flag is required.${NC}"
  exit 1
fi

if [ ! -f "${CIRCOM_FILE}" ]; then
    echo -e "${RED}Circom file does not exists${NC}"
    exit 1
fi

if [[ ${CIRCOM_FILE} =~ *.circom ]]; then
    echo -e "${RED}Circom file must have .circom filetype${NC}"
    exit 1
fi

# Validate power of tau
if [ -z "${POWER_OF_TAU}" ]; then
  echo -e "${RED}Error: --power flag is required.${NC}"
  exit 1
fi

if ! [[ "${POWER_OF_TAU}" =~ ^[0-9]+$ ]]; then
  echo -e "${RED}Error: Power of tau must be a valid integer.${NC}"
  exit 1
fi

CIRCOM_BASENAME=$(basename ${CIRCOM_FILE})
CIRCOM_FULLPATH=$(readlink -f "${CIRCOM_FILE}")
CIRCOM_FILENAME=$(echo ${CIRCOM_BASENAME} | awk -F '.' '{print $1}')

if ! command -v circom &> /dev/null; then
    echo -e "${RED}Circom is not Installed.${NC}"
    exit 1
fi

if ! command -v snarkjs &> /dev/null; then
    echo -e "${RED}SnarkJS is not Installed.${NC}"
    exit 1
fi

if [ ! -d "${PWD}/powersOfTau" ]; then
 echo -e "${RED}Directory ${PWD}/powersOfTau DOES NOT exists${NC}"
 echo -e "${GREEN}Creating missing directory${NC}"
 mkdir -p ${PWD}/powersOfTau || true
fi

if [ ! -f "${PWD}/powersOfTau/pot${POWER_OF_TAU}_final.ptau" ]; then
 pushd ${PWD}/powersOfTau
  echo -e "${GREEN}Start a new powers of tau ceremony${NC}"
  snarkjs powersoftau new bn128 ${POWER_OF_TAU} pot${POWER_OF_TAU}_0000.ptau -v

  echo -e "${GREEN}Contribute to the ceremony${NC}"
  echo -e "${GREEN}First Contribution${NC}"
  snarkjs powersoftau contribute pot${POWER_OF_TAU}_0000.ptau pot${POWER_OF_TAU}_0001.ptau --name="First contribution" -v -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;)

  echo -e "${GREEN}Second Contribution${NC}"
  snarkjs powersoftau contribute pot${POWER_OF_TAU}_0001.ptau pot${POWER_OF_TAU}_0002.ptau --name="Second contribution" -v -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;)

  echo -e "${GREEN}Third Contribution${NC}"
  snarkjs powersoftau export challenge pot${POWER_OF_TAU}_0002.ptau challenge_0003
  snarkjs powersoftau challenge contribute bn128 challenge_0003 response_0003 -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;)
  snarkjs powersoftau import response pot${POWER_OF_TAU}_0002.ptau response_0003 pot${POWER_OF_TAU}_0003.ptau -n="Third contribution name"

  echo -e "${GREEN}Verify the protocol${NC}"
  snarkjs powersoftau verify pot${POWER_OF_TAU}_0003.ptau

  echo -e "${GREEN}Apply random beacon${NC}"
  snarkjs powersoftau beacon pot${POWER_OF_TAU}_0003.ptau pot${POWER_OF_TAU}_beacon.ptau 0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f 10 -n="Final Beacon"

  echo -e "${GREEN}Prepare phase 2${NC}"
  snarkjs powersoftau prepare phase2 pot${POWER_OF_TAU}_beacon.ptau pot${POWER_OF_TAU}_final.ptau -v

  echo -e "${GREEN}Verify the final ptau${NC}"
  snarkjs powersoftau verify pot${POWER_OF_TAU}_final.ptau

  echo -e "${GREEN}Cleaning unnecessary files${NC}"
  rm pot${POWER_OF_TAU}_0000.ptau pot${POWER_OF_TAU}_0001.ptau pot${POWER_OF_TAU}_0002.ptau challenge_0003 response_0003 pot${POWER_OF_TAU}_0003.ptau pot${POWER_OF_TAU}_beacon.ptau || true
 popd
else
 echo -e "${RED}Powers of Tau Ceremony for power ${POWER_OF_TAU} has already been performed.${NC}"
fi

rm -rf ${PWD}/circuits/build/${CIRCOM_FILENAME}
if [ ! -d "${PWD}/circuits/build/${CIRCOM_FILENAME}" ]; then
 echo -e "${RED}Directory ${PWD}/circuits/build/${CIRCOM_FILENAME} DOES NOT exists${NC}"
 echo -e "${GREEN}Creating missing directory${NC}"
 mkdir -p ${PWD}/circuits/build/${CIRCOM_FILENAME} || true
 #cp -a $(dirname ${CIRCOM_FULLPATH})/* ${PWD}/circuits/build/${CIRCOM_FILENAME}
fi

if [ ! -f "${PWD}/circuits/build/${CIRCOM_FILENAME}/${CIRCOM_FILENAME}.r1cs" ]; then
 pushd ${PWD}/circuits/build/${CIRCOM_FILENAME}
  echo -e "${GREEN}Compile Circom File${NC}"
  circom ${CIRCOM_FULLPATH} --r1cs --wasm --sym #--c --wat

  echo -e "${GREEN}View information about the circuit${NC}"
  snarkjs r1cs info ${CIRCOM_FILENAME}.r1cs

  echo -e "${GREEN}Print the constraints${NC}"
  snarkjs r1cs print ${CIRCOM_FILENAME}.r1cs ${CIRCOM_FILENAME}.sym

  # echo -e "${GREEN}Export r1cs to json${NC}"
  # snarkjs r1cs export json ${CIRCOM_FILENAME}.r1cs ${CIRCOM_FILENAME}.r1cs.json
 popd
else
 echo -e "${RED}Circuit is already compiled.${NC}"
fi

if [ ! -d "${PWD}/circuits/build/${CIRCOM_FILENAME}/keys" ]; then
 echo -e "${RED}Directory ${PWD}/circuits/build/${CIRCOM_FILENAME}/keys DOES NOT exists${NC}"
 echo -e "${GREEN}Creating missing directory${NC}"
 mkdir -p ${PWD}/circuits/build/${CIRCOM_FILENAME}/keys || true
fi

echo -e "${GREEN}Setup MPC Zero Knowledge Proof SNARKS${NC}"
if [ ! -f "${PWD}/circuits/build/${CIRCOM_FILENAME}/keys/verification_keys.json" ]; then
 pushd ${PWD}/circuits/build/${CIRCOM_FILENAME}/keys
  echo -e "${GREEN}Generate a zero-knowledge key${NC}"
  snarkjs groth16 setup ../${CIRCOM_FILENAME}.r1cs ../../../../powersOfTau/pot${POWER_OF_TAU}_final.ptau ${CIRCOM_FILENAME}_0000.zkey

  echo -e "${GREEN}Contribute to the MPC trusted setup${NC}"
  echo -e "${GREEN}First Contribution${NC}"
  snarkjs zkey contribute ${CIRCOM_FILENAME}_0000.zkey ${CIRCOM_FILENAME}_0001.zkey --name="First Contributor Name" -v -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;)

  echo -e "${GREEN}Second Contribution${NC}"
  snarkjs zkey contribute ${CIRCOM_FILENAME}_0001.zkey ${CIRCOM_FILENAME}_0002.zkey --name="Second Contribution Name" -v -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;) 

  echo -e "${GREEN}Third Contribution${NC}"
  snarkjs zkey export bellman ${CIRCOM_FILENAME}_0002.zkey  challenge_phase2_0003
  snarkjs zkey bellman contribute bn128 challenge_phase2_0003 response_phase2_0003 -e=$(cat /proc/sys/kernel/random/uuid | sed 's/[-]//g' | head -c 40; echo;)
  snarkjs zkey import bellman ${CIRCOM_FILENAME}_0002.zkey response_phase2_0003 ${CIRCOM_FILENAME}_0003.zkey -n="Third Contribution Name"

  echo -e "${GREEN}Verify the latest zkey${NC}"
  snarkjs zkey verify ../${CIRCOM_FILENAME}.r1cs ../../../../powersOfTau/pot${POWER_OF_TAU}_final.ptau ${CIRCOM_FILENAME}_0003.zkey

  echo -e "${GREEN}Apply random beacon${NC}"
  snarkjs zkey beacon ${CIRCOM_FILENAME}_0003.zkey ${CIRCOM_FILENAME}_final.zkey 0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f 10 -n="Final Beacon phase2"

  echo -e "${GREEN}Verify the final zkey${NC}"
  snarkjs zkey verify ../${CIRCOM_FILENAME}.r1cs ../../../../powersOfTau/pot${POWER_OF_TAU}_final.ptau ${CIRCOM_FILENAME}_final.zkey

  echo -e "${GREEN}Export the verification key${NC}"
  snarkjs zkey export verificationkey ${CIRCOM_FILENAME}_final.zkey verification_key.json

  echo -e "${GREEN}Cleaning unnecessary files${NC}"
  rm ${CIRCOM_FILENAME}_0000.zkey ${CIRCOM_FILENAME}_0001.zkey ${CIRCOM_FILENAME}_0002.zkey challenge_phase2_0003 response_phase2_0003 ${CIRCOM_FILENAME}_0003.zkey
 popd
else
  echo -e "${RED}zkSnark Trusted Ceremony has already been performed.${NC}"
fi

echo -e "${GREEN}zkSnarks MPC Trusted Ceremony has concluded. Proceed with Prove and Verification${NC}"