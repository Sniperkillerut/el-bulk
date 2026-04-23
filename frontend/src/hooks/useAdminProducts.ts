import { useState, useEffect, useCallback } from 'react';
import { Product } from '@/lib/types';
import { adminFetchProducts } from '@/lib/api';

export function useAdminProducts() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [search, setSearch] = useState('');
  const [tcgFilter, setTcgFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [storageFilter, setStorageFilter] = useState('');
  const [onlyDuplicates, setOnlyDuplicates] = useState(false);
  const [sortKey, setSortKey] = useState('created_at');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const [queryTime, setQueryTime] = useState(0);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      const data = await adminFetchProducts({
        page, 
        page_size: pageSize, 
        search, 
        tcg: tcgFilter, 
        category: categoryFilter,
        storage_id: storageFilter, 
        only_duplicates: onlyDuplicates,
        sort_by: sortKey, 
        sort_dir: sortDir
      });
      setProducts(data.products || []);
      setTotal(data.total || 0);
      setQueryTime(data.query_time_ms || 0);
    } catch (e) {
      console.error('Failed to fetch products:', e);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, search, tcgFilter, categoryFilter, storageFilter, onlyDuplicates, sortKey, sortDir]);

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
    categoryFilter, setCategoryFilter,
    storageFilter, setStorageFilter, 
    onlyDuplicates, setOnlyDuplicates,
    sortKey, sortDir,
    queryTime,
    setPage, setPageSize, handleSort, refresh: fetchProducts
  };
}
