## Normative Policy Translator — Setup and Testing

The **eFLINT-to-Smart-Contract Translator** converts normative policy specifications into executable Go chaincode for Hyperledger Fabric. The implementation supports five policy configurations covering admission control (P1), continuing authorisation (P2), derived-output governance (P3), retained duties (P4), and human assessment (P5).

This section explains how to generate the smart contracts, execute the functional tests, and reproduce the translation performance experiments.

### 1. Prerequisites

The following dependencies are required:

- **Python 3** for validation, NIR construction, chaincode generation, and benchmarking.
- **eFLINT REPL** (`eflint-repl`) for validating normative specifications and executing reference scenarios.
- **Go**, compatible with the Hyperledger Fabric Contract API.
- **Docker**, Docker Compose v2, and Hyperledger Fabric test-network dependencies for chaincode deployment.
- **Bash**, **Git**, and **jq** for execution scripts and Fabric transactions.

Install the official eFLINT implementation and ensure that its executable is available in your `PATH`.

Verify the environment:

```bash
python3 --version
eflint-repl --help
go version
docker --version
jq --version
```

Clone the repository and navigate to the translator directory:

```bash
git clone https://github.com/TUDTrust4DataLab/decentralised-ac.git
cd decentralised-ac/policy_translator
```

All commands below assume the current working directory is `policy_translator/`, unless otherwise specified.

### 2. Translator Structure

The main directories and files are:

| Path | Description |
|---|---|
| `normative/specifications/` | Source eFLINT policy definitions |
| `normative/scenarios/` | Combined specifications and execution scenarios |
| `normative/bindings/` | Initial ledger states for functional testing |
| `src/validator/` | eFLINT specification validation |
| `src/structure_builder/` | Normative Intermediate Representation (NIR) construction |
| `src/policy_builder/fabric/` | Go chaincode generation |
| `scripts/` | Functional tests and performance benchmarks |
| `generated/` | Generated NIR and chaincode artefacts |
| `results/smoke/` | Reference-interpreter and Fabric execution logs |
| `results/experiments/e3/` | Translation performance measurements and summaries |

### 3. Generating Smart Contracts

The translation pipeline consists of three stages:

1. **Validation:** Check the normative specification using the eFLINT reference interpreter.
2. **NIR Construction:** Convert the supported normative declarations into an intermediate representation.
3. **Chaincode Generation:** Generate the corresponding Go smart contract.

For example, generate the P1 contract:

```bash
# Validate the specification
python3 validate.py normative/scenarios/p1/p1-combined.eflint

# Build the NIR
python3 build_nir.py \
  normative/scenarios/p1/p1-combined.eflint \
  --output generated/nir/p1.nir.json

# Generate Go chaincode
python3 generate_chaincode.py \
  generated/nir/p1.nir.json \
  --output generated/fabric/p1
```

The generated contract is written to:

```text
generated/fabric/p1/main.go
```

To generate all five policy configurations:

```bash
for p in 1 2 3 4 5; do
  python3 validate.py \
    "normative/scenarios/p${p}/p${p}-combined.eflint"

  python3 build_nir.py \
    "normative/scenarios/p${p}/p${p}-combined.eflint" \
    --output "generated/nir/p${p}.nir.json"

  python3 generate_chaincode.py \
    "generated/nir/p${p}.nir.json" \
    --output "generated/fabric/p${p}"
done
```

The generated Go projects require the Hyperledger Fabric Contract API dependency. Before deploying a newly generated contract, initialise its Go module if necessary and resolve dependencies:

```bash
cd generated/fabric/p1

go mod init normative-p1
go mod tidy
go build ./...

cd ../../..
```

Repeat the Go module setup for other generated configurations, replacing `p1` with the corresponding policy identifier.

### 4. Functional Validation

The functional tests examine whether the generated contracts exhibit the expected normative decisions and state transitions under representative scenarios.

#### 4.1. Reference eFLINT Tests

The P1 suite includes positive and negative scenarios, covering missing approval, missing terms, incorrect purposes, incorrect sessions, and unauthorised approval attempts.

Execute the automated reference-interpreter tests:

```bash
python3 scripts/scenario_runner.py
```

Alternatively, execute the P1 scenario scripts directly:

```bash
bash scripts/run-p1-scenario.sh
```

The automated runner generates a summary at:

```text
results/smoke/p1/report.json
```

Individual execution logs are also saved under `results/smoke/p1/`.

For P2–P5, the repository includes reference scenarios that can be executed using `eflint-repl`. For example:

```bash
eflint-repl \
  normative/scenarios/p4/p4-combined.eflint \
  < normative/scenarios/p4/compliant.txt
```

The published reference results for all five configurations are available under `results/smoke/`.

#### 4.2. Hyperledger Fabric Setup

The Fabric tests require a running Hyperledger Fabric test network.

From the repository root:

```bash
cd fabric
./install-fabric.sh

cd fabric-samples/test-network

./network.sh up createChannel
```

The default channel is `mychannel`.

#### 4.3. Deploy the Generated Chaincode

From `fabric/fabric-samples/test-network`, deploy the generated P1 contract:

```bash
./network.sh deployCC \
  -c mychannel \
  -ccn normative-p1 \
  -ccp ../../../policy_translator/generated/fabric/p1 \
  -ccl go \
  -ccv 1.0 \
  -ccs 1
```

Other configurations can be deployed by replacing the chaincode name and generated path with the corresponding policy identifier.

P4 and P5 use separate chaincode deployments for their alternative execution paths, allowing each scenario to operate with independent ledger state.

#### 4.4. Execute Fabric Functional Tests

Return to the translator directory:

```bash
cd ../../../policy_translator
```

Run the P1–P3 functional tests:

```bash
bash scripts/run-p1-fabric.sh
bash scripts/run-p2-fabric.sh
bash scripts/run-p3-fabric.sh
```

For P4, execute the compliant and deadline-breach paths using separately deployed chaincode instances:

```bash
CHAINCODE=normative-p4-compliant \
MODE=compliant \
bash scripts/run-p4-fabric.sh

CHAINCODE=normative-p4-breach \
MODE=breach \
bash scripts/run-p4-fabric.sh
```

Similarly, P5 provides two human-assessment scenarios:

```bash
CHAINCODE=normative-p5-acceptance \
MODE=acceptance \
bash scripts/run-p5-fabric.sh

CHAINCODE=normative-p5-response \
MODE=response \
bash scripts/run-p5-fabric.sh
```

Each test checks expected decisions and ledger states, printing `PASS` when a checkpoint matches its expected result. Any failed assertion terminates the script with a non-zero exit status.

**Important:** Functional scripts initialise and modify persistent ledger state. They are intended for fresh chaincode deployments and should not be rerun against an already initialised instance without resetting its state or deploying a new instance.

The existing execution logs are available in `results/smoke/` for inspection and comparison.

### 5. Reproducing Translation Performance Experiments (E3)

The E3 experiments measure the computational performance of the translation pipeline across the five policy configurations.

Each configuration is evaluated using:

- 10 warm-up executions, excluded from the measurements.
- 500 measured translation executions.
- Stage-level timings for eFLINT validation, NIR construction, and Go chaincode generation.
- End-to-end translation time.

To reproduce the experiments:

```bash
python3 scripts/measure_translation.py
python3 scripts/measure_translation_p2.py
python3 scripts/measure_translation_p3.py
python3 scripts/measure_translation_p4.py
python3 scripts/measure_translation_p5.py
```

Each script produces:

```text
results/experiments/e3/pN/
├── runs.csv
└── summary.json
```

Here, `pN` represents the evaluated policy (`p1` through `p5`).

**Note:** Running these scripts overwrites the corresponding published `runs.csv` and `summary.json` files. Preserve the original results or use a separate repository clone when reproducing the experiments.

#### 5.1. Consolidating Results

After executing the benchmarks and generating the NIR files for all configurations, consolidate the measurements:

```bash
python3 scripts/consolidate_e3.py
```

This generates:

```text
results/experiments/e3/
├── comparison.csv
├── comparison.json
├── p1/
├── p2/
├── p3/
├── p4/
└── p5/
```

The consolidated results provide descriptive statistics, including mean execution time, standard deviation, median, 95th and 99th percentiles, and approximate 95% confidence intervals.

The confidence intervals are calculated using the normal approximation for the sample mean.

#### 5.2. Published Results

The repository includes the original experimental measurements, enabling independent inspection without rerunning the benchmarks.

The published experiment consists of **2,500 measured translations**, with 500 executions per policy configuration.

Mean translation times ranged from approximately **214.50 ms to 548.29 ms** across the five configurations. eFLINT validation accounted for approximately 99.8% of the total processing time, while NIR construction and Go chaincode generation together remained below 1 ms on average for every evaluated configuration.

These measurements exclude Go compilation, chaincode deployment, smart contract execution, and blockchain transaction processing.

Execution times may vary depending on the operating system, hardware, installed tool versions, and background system activity.

### 6. Reproducibility and Evaluation Scope

The published artefacts support the reproduction and inspection of two evaluation dimensions:

- **Functional validation:** Representative normative decisions and state transitions produced by the generated Fabric contracts.
- **Translation performance:** Computational costs associated with specification validation, NIR construction, and code generation.

The functional scenarios provide evidence of expected behaviour within the supported eFLINT subset.

These experiments do not evaluate end-to-end enforcement of external data operations. Such verification requires integrating the generated contracts with the components responsible for protecting resources and observing their responses to authorisation decisions.

For reproducibility, researchers should record the repository commit, software versions, hardware specifications, and experimental environment when reporting new results.