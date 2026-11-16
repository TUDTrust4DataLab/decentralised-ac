import {
  ConflictError,
  NotFoundError,
} from './http-error.js';

export function translateFabricError(
  error: unknown,
): Error {

  const message =
    error instanceof Error
      ? error.message
      : String(error);

  if (
    message.includes('does not exist') ||
    message.includes('has no assigned policy')
  ) {
    return new NotFoundError(message);
  }

  if (
    message.includes('already exists') ||
    message.includes('already inactive') ||
    message.includes('cannot be updated') ||
    message.includes('cannot assign inactive policy')
  ) {
    return new ConflictError(message);
  }

  return error instanceof Error
    ? error
    : new Error(message);
}