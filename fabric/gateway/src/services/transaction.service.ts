import type { Gateway } from '@hyperledger/fabric-gateway';

export class TransactionService {
  constructor(private readonly gateway: Gateway) {}

  private getContract(
    channelName: string,
    chaincodeName: string,
    contractName?: string,
  ) {
    const network = this.gateway.getNetwork(channelName);

    return contractName
      ? network.getContract(chaincodeName, contractName)
      : network.getContract(chaincodeName);
  }

  async evaluate(
    channelName: string,
    chaincodeName: string,
    contractName: string,
    transactionName: string,
    args: string[] = [],
  ): Promise<Uint8Array> {
    const contract = this.getContract(
      channelName,
      chaincodeName,
      contractName,
    );

    return contract.evaluateTransaction(
      transactionName,
      ...args,
    );
  }

  async submit(
    channelName: string,
    chaincodeName: string,
    contractName: string,
    transactionName: string,
    args: string[] = [],
  ): Promise<Uint8Array> {
    const contract = this.getContract(
      channelName,
      chaincodeName,
      contractName,
    );

    return contract.submitTransaction(
      transactionName,
      ...args,
    );
  }
}