# CIPHER Fabric Gateway

The **CIPHER Fabric Gateway** provides an integration layer between CIPHER services and a Hyperledger Fabric network.

The gateway is implemented in **TypeScript** and will expose an HTTP API for interacting with Hyperledger Fabric.

## Planned Features

- Submit transactions to chaincodes.
- Evaluate read-only chaincode transactions.
- Deploy and manage Go chaincodes using the Hyperledger Fabric chaincode lifecycle.
- Provide a simple API for CIPHER services that do not directly integrate with a Fabric SDK.

Future versions will also support additional Hyperledger Fabric network administration operations.

## Technology Stack

- Node.js
- TypeScript
- Hyperledger Fabric Gateway SDK
- Hyperledger Fabric 2.5

## Requirements

- Node.js 20+
- npm
- Access to a running Hyperledger Fabric network

## Getting Started

Clone the repository and install the dependencies:

```bash
cd fabric/gateway

npm install
```

The project is currently under development.

## Project Structure

The initial project will follow the structure:

```text
gateway/
├── src/
│   ├── fabric/
│   ├── routes/
│   ├── services/
│   ├── config/
│   └── index.ts
├── test/
├── .env.example
├── package.json
├── tsconfig.json
└── README.md
```

The responsibilities of these modules will evolve as the Fabric integration and API are implemented.

## Licence

This software is developed as part of the CIPHER project.