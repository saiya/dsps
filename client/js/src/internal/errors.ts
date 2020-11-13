export class UnreachableCaseError extends Error {
  constructor(value: never) {
    super(`UnreachableCaseError(${value})`);
  }
}

export function unreachableCaseError(value: never): never {
  throw new UnreachableCaseError(value);
}
