import { describe, it, expect, vi } from 'vitest';

// Simulating the API call with a fixed delay
const adminUpdateClientRequestStatusMock = vi.fn(async (id: string, status: string) => {
  await new Promise(resolve => setTimeout(resolve, 50));
  return { id, status };
});

describe('BountiesPage Optimization Benchmark', () => {
  const toFulfill = Array.from({ length: 10 }, (_, i) => ({ id: `req-${i}` }));

  it('Sequential execution (Baseline)', async () => {
    const start = performance.now();
    for (const req of toFulfill) {
      await adminUpdateClientRequestStatusMock(req.id, 'solved');
    }
    const end = performance.now();
    const duration = end - start;
    console.log(`Sequential duration: ${duration.toFixed(2)}ms`);
    expect(duration).toBeGreaterThan(500); // 10 * 50ms
  });

  it('Parallel execution (Optimized)', async () => {
    const start = performance.now();
    await Promise.all(toFulfill.map(req => adminUpdateClientRequestStatusMock(req.id, 'solved')));
    const end = performance.now();
    const duration = end - start;
    console.log(`Parallel duration: ${duration.toFixed(2)}ms`);
    expect(duration).toBeLessThan(150); // Should be close to 50ms, giving some buffer
  });
});
