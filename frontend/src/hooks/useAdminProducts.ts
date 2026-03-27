import { useState, useEffect, useCallback } from 'react';
import { Product, TCG_SHORT } from '@/lib/types';
import { adminFetchProducts } from '@/lib/api';

export function useAdminProducts(token: string) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(25);
  const [search, setSearch] = useState('');
  const [tcgFilter, setTcgFilter] = useState('');
  const [storageFilter, setStorageFilter] = useState('');
  const [sortKey, setSortKey] = useState('created_at');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const fetchProducts = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const data = await adminFetchProducts(token, {
        page, 
        page_size: pageSize, 
        search, 
        tcg: tcgFilter, 
        storage_id: storageFilter, 
        sort_by: sortKey, 
        sort_dir: sortDir
      });
      setProducts(data.products || []);
      setTotal(data.total || 0);
    } catch (e) {
      console.error('Failed to fetch products:', e);
    } finally {
      setLoading(false);
    }
  }, [token, page, pageSize, search, tcgFilter, storageFilter, sortKey, sortDir]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleSort = (key: string) => {
    if (sortKey === key) setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    else { setSortKey(key); setSortDir('asc'); }
    setPage(1);
  };

  return {
    products, loading, total, page, pageSize, 
    search, setSearch, tcgFilter, setTcgFilter, 
    storageFilter, setStorageFilter, sortKey, sortDir,
    setPage, handleSort, refresh: fetchProducts
  };
}
