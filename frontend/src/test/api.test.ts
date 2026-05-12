import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiFetch, fetchProducts, fetchProduct, clearCached } from '../lib/api';

describe('api.ts', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    global.fetch = mockFetch;
    mockFetch.mockReset();
    clearCached();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('apiFetch', () => {
    it('should successfully fetch data', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: 'success' }),
      } as Response);

      const result = await apiFetch<{ data: string }>('/test');
      expect(result.data).toBe('success');
      expect(mockFetch).toHaveBeenCalledWith(expect.stringContaining('/test'), expect.any(Object));
    });

    it('should pass parameters as query string', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      } as Response);

      await apiFetch('/test', { params: { search: 'foo', page: 1 } });
      const callArgs = mockFetch.mock.calls[0];
      const url = callArgs[0] as string;
      expect(url).toContain('search=foo');
      expect(url).toContain('page=1');
    });

    it('should handle 204 no content', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
      } as Response);

      const result = await apiFetch('/test');
      expect(result).toEqual({});
    });

    it('should throw an error if response is not ok', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        clone: () => ({
          json: async () => ({ error: 'Invalid input' })
        }),
        url: 'http://localhost/test'
      } as unknown as Response);

      await expect(apiFetch('/test')).rejects.toThrow('Invalid input');
    });
  });

  describe('fetchProducts', () => {
    it('should call apiFetch with correct path and params', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ items: [], total: 0 }),
      } as Response);

      await fetchProducts({ page: 2 });
      const callArgs = mockFetch.mock.calls[0];
      expect(callArgs[0]).toContain('/api/products?page=2');
    });
  });

  describe('fetchProduct', () => {
    it('should call apiFetch with correct path', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ id: '123' }),
      } as Response);

      const result = await fetchProduct('123');
      expect(result).toEqual({ id: '123' });
      const callArgs = mockFetch.mock.calls[0];
      expect(callArgs[0]).toContain('/api/products/123');
    });
  });
});
