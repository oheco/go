import go from 'libgoharmony.so';

export async function verifyGo(): Promise<string> {
  const value = go.add(20, 22);
  const sum = await go.sumAsync(10000);
  if (value !== 42 || sum !== 50005000) {
    throw new Error('Go N-API result mismatch');
  }
  return `Go returned ${value}; asynchronous sum: ${sum}`;
}
