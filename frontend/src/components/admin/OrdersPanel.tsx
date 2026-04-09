'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import Link from 'next/link';
import {
  adminFetchOrders, adminFetchOrderDetail, adminUpdateOrder, adminConfirmOrder,
  adminFetchProducts, adminRestoreOrderStock, adminFetchStorage
} from '@/lib/api';
import {
  OrderWithCustomer, OrderDetail,
  ORDER_STATUS_LABELS, PAYMENT_METHODS, FOIL_LABELS, TREATMENT_LABELS, StorageLocation,
  Product, StoredIn
} from '@/lib/types';
import CardImage from '@/components/CardImage';
import { useLanguage } from '@/context/LanguageContext';

interface Props {
  initialOrderId?: string | null;
}

export default function OrdersPanel({ initialOrderId }: Props) {
  const { t } = useLanguage();
  // List state
  const [orders, setOrders] = useState<OrderWithCustomer[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [statusFilter, setStatusFilter] = useState('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);


  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Detail state
  const [detail, setDetail] = useState<OrderDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [saving, setSaving] = useState(false);
  const [itemEdits, setItemEdits] = useState<Record<string, number>>({});
  const [detailsCache, setDetailsCache] = useState<Record<string, OrderDetail>>({});
  const [paymentMethodEdit, setPaymentMethodEdit] = useState<string>('');
  const [shippingCopEdit, setShippingCopEdit] = useState<number>(0);

  const canEditMetadata = detail?.order.status === 'pending' || detail?.order.status === 'confirmed';
  const canEditInventory = detail?.order.status === 'pending';
  const isCompleted = detail?.order.status === 'completed' || detail?.order.status === 'cancelled';

  // Confirm modal state
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [decrements, setDecrements] = useState<Record<string, Record<string, number>>>({});
  const [confirming, setConfirming] = useState(false);
  const [confirmError, setConfirmError] = useState('');
  const [mobileShowDetail, setMobileShowDetail] = useState(false);
  const handledInitialId = useRef<string | null>(null);

  // Restore modal state (for cancelled orders)
  const [showRestoreModal, setShowRestoreModal] = useState(false);
  const [increments, setIncrements] = useState<Record<string, Record<string, number>>>({});
  const [restoring, setRestoring] = useState(false);
  const [restoreError, setRestoreError] = useState('');

  // Add items functionality
  const [productSearch, setProductSearch] = useState('');
  const [searchResults, setSearchResults] = useState<Product[]>([]);
  const [searchingItems, setSearchingItems] = useState(false);
  const [stagedItems, setStagedItems] = useState<{ product: Product; quantity: number; unit_price_cop: number }[]>([]);
  const [deletedIds, setDeletedIds] = useState<string[]>([]);
  const [allStorage, setAllStorage] = useState<StoredIn[]>([]);

  const loadOrders = useCallback(async () => {
    // No longer setting loading synchronously at start to avoid cascaded renders
    try {
      const res = await adminFetchOrders({ status: statusFilter, search: debouncedSearch, page, page_size: pageSize });
      setOrders(res.orders || []);
      setTotal(res.total || 0);
    } catch (err) {
      console.error('Failed to load orders:', err);
    }
    setLoading(false);
  }, [statusFilter, debouncedSearch, page, pageSize]);

  useEffect(() => {
    adminFetchStorage().then(setAllStorage).catch(console.error);
    const timer = setTimeout(() => {
      loadOrders();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadOrders]);

  const selectOrder = useCallback(async (id: string) => {
    // No longer setting state synchronously here to avoid cascaded renders when called from effect
    if (detailsCache[id]) {
      setDetail(detailsCache[id]);
      const edits: Record<string, number> = {};
      detailsCache[id].items.forEach(i => { edits[i.id] = i.quantity; });
      setItemEdits(edits);
      setPaymentMethodEdit(detailsCache[id].order.payment_method);
      setShippingCopEdit(detailsCache[id].order.shipping_cop);
      setLoadingDetail(false);
      return;
    }

    setLoadingDetail(true);
    try {
      const d = await adminFetchOrderDetail(id);
      setDetail(d);
      setDetailsCache(prev => ({ ...prev, [id]: d }));
      // Init item edits
      const edits: Record<string, number> = {};
      setItemEdits(edits);
      setPaymentMethodEdit(d.order.payment_method);
      setShippingCopEdit(d.order.shipping_cop);
    } catch { }
    setLoadingDetail(false);
  }, [detailsCache]);

  useEffect(() => {
    if (initialOrderId && handledInitialId.current !== initialOrderId) {
      handledInitialId.current = initialOrderId;
      const timer = setTimeout(() => {
        setItemEdits({});
        setStagedItems([]);
        setDeletedIds([]);
        selectOrder(initialOrderId);
        setMobileShowDetail(true);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [initialOrderId, selectOrder]);

  const searchProducts = useCallback(async (q: string) => {
    if (!q.trim()) {
      setSearchResults([]);
      return;
    }
    setSearchingItems(true);
    try {
      const res = await adminFetchProducts({ search: q, page_size: 5 });
      setSearchResults(res.products);
    } catch (e) {
      console.error(e);
    }
    setSearchingItems(false);
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      searchProducts(productSearch);
    }, 300);
    return () => clearTimeout(timer);
  }, [productSearch, searchProducts]);

  const addStagedItem = (p: Product) => {
    setStagedItems(prev => {
      return [...prev, { product: p, quantity: 1, unit_price_cop: p.price }];
    });
    setProductSearch('');
    setSearchResults([]);
  };

  const removeStagedItem = (productId: string) => {
    setStagedItems(prev => prev.filter(i => i.product.id !== productId));
  };

  const updateStagedItem = (productId: string, data: Partial<{ quantity: number; unit_price_cop: number }>) => {
    setStagedItems(prev => prev.map(i => i.product.id === productId ? { ...i, ...data } : i));
  };

  const handleSaveChanges = async () => {
    if (!detail) return;
    setSaving(true);
    try {
      const items = Object.entries(itemEdits).map(([id, quantity]) => ({ id, quantity }));
      const added_items = stagedItems.map(si => ({
        product_id: si.product.id,
        quantity: si.quantity,
        unit_price_cop: si.unit_price_cop
      }));
      const deleted_ids = deletedIds;
      const payment_method = paymentMethodEdit !== detail.order.payment_method ? paymentMethodEdit : undefined;
      const shipping_cop = shippingCopEdit !== detail.order.shipping_cop ? shippingCopEdit : undefined;
      const updated = await adminUpdateOrder(detail.order.id, { 
        items, 
        added_items, 
        deleted_ids,
        payment_method,
        shipping_cop
      });
      setDetail(updated);
      setStagedItems([]);
      setDeletedIds([]);
      setDetailsCache(prev => ({ ...prev, [updated.order.id]: updated }));
      // Update list in-place
      setOrders(prev => prev.map(o => o.id === updated.order.id ? { 
        ...o, 
        total_cop: updated.order.total_cop,
        item_count: updated.items.length 
      } : o));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : 'Failed to save');
    }
    setSaving(false);
  };

  const handleStatusChange = async (status: string, trackingInfo?: { tracking_number?: string, tracking_url?: string }) => {
    if (!detail) return;

    // Issue 3 & 4: Alert for final changes
    if (status === 'completed' || status === 'cancelled') {
        const msg = status === 'completed' 
            ? '¿Estás seguro de marcar como COMPLETADA? Esta acción es final y la orden se bloqueará.'
            : '¿Estás seguro de CANCELAR esta orden? Esta acción es final.';
        if (!window.confirm(msg)) return;
    }

    setSaving(true);
    try {
      const updated = await adminUpdateOrder(detail.order.id, { status, ...trackingInfo });
      setDetail(updated);
      setDetailsCache(prev => ({ ...prev, [updated.order.id]: updated }));
      // Update list in-place
      setOrders(prev => prev.map(o => o.id === updated.order.id ? { ...o, status: updated.order.status } : o));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : 'Failed to update status');
    }
    setSaving(false);
  };

  const openConfirmModal = () => {
    if (!detail) return;
    const initial: Record<string, Record<string, number>> = {};
    detail.items.forEach(item => {
      if (item.quantity > 0 && item.product_id) {
        initial[item.product_id] = {};
        // Pre-fill: auto-assign from first physical storage location with enough stock
        let remaining = item.quantity;
        item.stored_in.forEach(loc => {
          if (loc.name.toLowerCase() !== 'pending' && remaining > 0) {
            const take = Math.min(remaining, loc.quantity);
            initial[item.product_id!][loc.stored_in_id] = take;
            remaining -= take;
          }
        });
      }
    });
    setDecrements(initial);
    setConfirmError('');
    setShowConfirmModal(true);
  };

  const openRestoreModal = () => {
    if (!detail) return;
    const initial: Record<string, Record<string, number>> = {};
    detail.items.forEach(item => {
      if (item.quantity > 0 && item.product_id) {
        initial[item.product_id] = {};
      }
    });
    setIncrements(initial);
    setRestoreError('');
    setShowRestoreModal(true);
  };

  const handleConfirm = async () => {
    if (!detail) return;
    setConfirming(true);
    setConfirmError('');

    // Validate: total decrements per product must match order quantity
    for (const item of detail.items) {
      if (item.quantity <= 0 || !item.product_id) continue;
      const productDecs = decrements[item.product_id] || {};
      const totalDec = Object.values(productDecs).reduce((s, v) => s + v, 0);
      if (totalDec !== item.quantity) {
        setConfirmError(`${item.product_name}: asignaste ${totalDec} de ${item.quantity} necesarios.`);
        setConfirming(false);
        return;
      }
    }

    // Build decrement array
    const decArr: { product_id: string; stored_in_id: string; quantity: number }[] = [];
    Object.entries(decrements).forEach(([pid, locs]) => {
      Object.entries(locs).forEach(([sid, qty]) => {
        if (qty > 0) {
          decArr.push({ product_id: pid, stored_in_id: sid, quantity: qty });
        }
      });
    });

    try {
      const updated = await adminConfirmOrder(detail.order.id, decArr);
      setDetail(updated);
      setShowConfirmModal(false);
      setOrders(prev => prev.map(o => o.id === updated.order.id ? { ...o, status: updated.order.status } : o));
    } catch (e: unknown) {
      setConfirmError(e instanceof Error ? e.message : 'Error al confirmar orden');
    }
    setConfirming(false);
  };

  const handleRestoreStock = async () => {
    if (!detail) return;
    setRestoring(true);
    setRestoreError('');
    
    // Validate: total increments per product must match order quantity exactly
    for (const item of detail.items) {
      if (item.quantity <= 0 || !item.product_id) continue;
      const productIncs = increments[item.product_id] || {};
      const totalAssigned = Object.values(productIncs).reduce((s, v) => s + v, 0);
      
      if (totalAssigned !== item.quantity) {
        setRestoreError(`${item.product_name}: debes restaurar todos los productos (${totalAssigned} de ${item.quantity} asignados).`);
        setRestoring(false);
        return;
      }
    }

    const incArr: { product_id: string; stored_in_id: string; quantity: number }[] = [];
    Object.entries(increments).forEach(([pid, locs]) => {
      Object.entries(locs).forEach(([sid, qty]) => {
        if (qty > 0) {
          incArr.push({ product_id: pid, stored_in_id: sid, quantity: qty });
        }
      });
    });

    if (incArr.length === 0) {
      setRestoreError('Debe asignar al menos una cantidad a una ubicación');
      setRestoring(false);
      return;
    }

    try {
      const updated = await adminRestoreOrderStock(detail.order.id, incArr);
      setDetail(updated);
      setShowRestoreModal(false);
      setDetailsCache(prev => ({ ...prev, [updated.order.id]: updated }));
      alert('Inventario restaurado exitosamente');
    } catch (e: unknown) {
      setRestoreError(e instanceof Error ? e.message : 'Error al restaurar inventario');
    }
    setRestoring(false);
  };

  const setDecrement = (productId: string, storedInId: string, qty: number) => {
    setDecrements(prev => ({
      ...prev,
      [productId]: { ...prev[productId], [storedInId]: Math.max(0, qty) }
    }));
  };

  const setIncrement = (productId: string, storedInId: string, qty: number) => {
    setIncrements(prev => ({
      ...prev,
      [productId]: { ...prev[productId], [storedInId]: Math.max(0, qty) }
    }));
  };

  const totalPages = Math.ceil(total / pageSize);
  const hasEdits = detail ? (detail.items.some(i => itemEdits[i.id] !== i.quantity) || stagedItems.length > 0 || deletedIds.length > 0) : false;


  return (
    <div className="flex flex-col flex-1 min-h-0 w-full">
      <div className="flex flex-col lg:flex-row flex-1 min-h-0">
        {/* Left: Orders List */}
        <div className={`flex-1 lg:flex-none w-full lg:w-[420px] lg:flex-shrink-0 flex flex-col min-h-0 overflow-hidden border-r border-border-main/20 bg-bg-page/5 ${mobileShowDetail ? 'hidden lg:flex' : 'flex'}`}>
          {/* Filters */}
          <div className="p-3 flex flex-col gap-2" style={{ borderBottom: '1px solid var(--border-main)' }}>
            <input
              type="search"
              placeholder="Buscar por # orden, nombre, teléfono..."
              value={search}
              onChange={e => { setSearch(e.target.value); setPage(1); setLoading(true); }}
              style={{ fontSize: '0.85rem' }}
            />
            <div className="flex gap-1 flex-wrap">
              {['', 'pending', 'confirmed', 'completed', 'cancelled'].map(s => (
                <button
                  key={s}
                  onClick={() => { setStatusFilter(s); setPage(1); setLoading(true); }}
                  className="badge cursor-pointer transition-colors"
                  style={{
                    padding: '3px 8px',
                    background: statusFilter === s ? 'var(--text-main)' : 'transparent',
                    color: statusFilter === s ? 'var(--btn-primary-text)' : 'var(--text-secondary)',
                    border: `1px solid ${statusFilter === s ? 'var(--text-main)' : 'var(--border-main)'}`,
                  }}
                >
                  {s ? t(`pages.order.status.${s}`, ORDER_STATUS_LABELS[s] || s) : t('pages.common.labels.all', 'Todas')}
                </button>
              ))}
            </div>

          </div>

          {/* Orders list */}
          <div className="flex-1 overflow-y-auto min-h-0 overscroll-contain" style={{ WebkitOverflowScrolling: 'touch' }}>
            {loading ? (
              <div className="p-6 text-center" style={{ color: 'var(--text-muted)' }}>Cargando...</div>
            ) : orders.length === 0 ? (
              <div className="p-6 text-center" style={{ color: 'var(--text-muted)' }}>No se encontraron órdenes.</div>
            ) : (
              orders.map(o => (
                <div
                  key={o.id}
                  onClick={() => { setItemEdits({}); setLoadingDetail(true); selectOrder(o.id); setMobileShowDetail(true); }}
                  className="p-3 cursor-pointer transition-colors"
                  style={{
                    borderBottom: '1px solid var(--border-main)',
                    background: detail?.order.id === o.id ? 'var(--bg-page)' : 'transparent',
                  }}
                  onMouseEnter={e => { if (detail?.order.id !== o.id) e.currentTarget.style.background = 'var(--bg-surface)'; }}
                  onMouseLeave={e => { if (detail?.order.id !== o.id) e.currentTarget.style.background = 'transparent'; }}
                >
                  <div className="flex justify-between items-start">
                    <div>
                      <span className="font-mono-stack text-sm font-bold">{o.order_number}</span>
                      <p className="text-xs" style={{ color: 'var(--text-secondary)' }}>{o.customer_name}</p>
                    </div>
                    <div className="text-right">
                      <span className={`badge ${o.status === 'completed' ? 'badge-nm' : o.status === 'cancelled' ? 'badge-hp' : o.status === 'confirmed' ? 'badge-lp' : ''}`}
                        style={{ fontSize: '0.6rem' }}>
                        {t(`pages.order.status.${o.status}`, ORDER_STATUS_LABELS[o.status] || o.status)}
                      </span>
                    </div>
                  </div>
                  <div className="flex justify-between items-center mt-1">
                    <span className="text-[10px] font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                      {new Date(o.created_at).toLocaleDateString('es-CO', { day: '2-digit', month: 'short', year: '2-digit' })}
                      {' · '}{o.item_count} item{o.item_count !== 1 ? 's' : ''}
                    </span>
                    <span className="price text-sm">${o.total_cop.toLocaleString('en-US', { maximumFractionDigits: 0 })}</span>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Pagination */}
          {totalPages > 0 && (
            <div className="flex flex-col sm:flex-row justify-between items-center gap-2 p-3 bg-bg-page/10" style={{ borderTop: '1px solid var(--border-main)' }}>
              <div className="flex items-center gap-2">
                <span className="text-[10px] font-mono-stack text-text-muted uppercase">Mostrar</span>
                <select 
                  value={pageSize}
                  onChange={e => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                    setLoading(true);
                  }}
                  className="bg-transparent border-0 font-mono-stack text-xs font-bold text-gold-dark outline-none cursor-pointer hover:text-hp-color transition-colors"
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </div>

              {totalPages > 1 && (
                <div className="flex items-center gap-3">
                  <button 
                    onClick={() => { setPage(p => Math.max(1, p - 1)); setLoading(true); }} 
                    disabled={page === 1} 
                    className="btn-secondary" 
                    style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem', opacity: page === 1 ? 0.4 : 1 }}
                  >
                    ←
                  </button>
                  <span className="text-xs font-mono-stack font-bold" style={{ color: 'var(--text-muted)' }}>
                    {page} / {totalPages}
                  </span>
                  <button 
                    onClick={() => { setPage(p => Math.min(totalPages, p + 1)); setLoading(true); }} 
                    disabled={page === totalPages} 
                    className="btn-secondary" 
                    style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem', opacity: page === totalPages ? 0.4 : 1 }}
                  >
                    →
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Right: Order Detail */}
        <div className={`flex-1 overflow-y-auto p-4 md:p-8 bg-white ${mobileShowDetail ? 'flex flex-col' : 'hidden lg:flex lg:flex-col'} min-h-0 overscroll-contain`} style={{ WebkitOverflowScrolling: 'touch' }}>
          {!detail && !loadingDetail && (
            <div className="flex items-center justify-center h-full" style={{ color: 'var(--text-muted)' }}>
              <div className="text-center">
                <div className="text-5xl opacity-30 mb-3">📋</div>
                <p className="font-display text-xl">Selecciona una orden</p>
              </div>
            </div>
          )}

          {loadingDetail && (
            <div className="flex items-center justify-center h-full" style={{ color: 'var(--text-muted)' }}>Cargando...</div>
          )}

          {detail && !loadingDetail && (
            <>
              {/* Mobile back button */}
              <button 
                onClick={() => { setMobileShowDetail(false); setDetail(null); }}
                className="lg:hidden flex items-center gap-2 text-xs font-mono-stack text-gold-dark mb-4 hover:text-hp-color transition-colors"
              >
                ← BACK TO ORDERS
              </button>

              {/* Order header */}
              <div className="flex flex-col sm:flex-row justify-between gap-3 mb-4">
                <div>
                  <h3 className="font-display text-3xl">{detail.order.order_number}</h3>
                  <p className="text-xs font-mono-stack" style={{ color: 'var(--text-muted)' }}>
                    {new Date(detail.order.created_at).toLocaleString('es-CO')}
                    {detail.order.completed_at && ` · Completada: ${new Date(detail.order.completed_at).toLocaleString('es-CO')}`}
                  </p>
                </div>
                <div className="flex items-start gap-2">
                  <select
                    value={detail.order.status}
                    onChange={e => handleStatusChange(e.target.value)}
                    disabled={saving || isCompleted}
                    className="bg-surface border-dark text-gold rounded px-2 py-1 outline-none focus:border-gold transition-colors"
                    style={{ fontSize: '0.85rem', padding: '0.3rem 0.6rem', minWidth: 120 }}
                  >
                    {Object.entries(ORDER_STATUS_LABELS)
                      .map(([k, v]) => {
                        // Issue 1: confirmed is now a post-confirmation state
                        const isPostConfirmation = k === 'confirmed' || k === 'ready_for_pickup' || k === 'shipped' || k === 'completed';
                        const isOriginalPending = k === 'pending';
                        
                        // Disable going back to pending if already confirmed
                        const disableBackToPending = isOriginalPending && (detail.order.status !== 'pending' && detail.order.status !== 'cancelled');
                        
                        // Hide post-confirmation states if currently pending
                        const hidePostConfirmation = isPostConfirmation && detail.order.status === 'pending';
                        
                        if (hidePostConfirmation) return null;

                        return (
                          <option key={k} value={k} disabled={disableBackToPending}>
                            {t(`pages.order.status.${k}`, v)}
                          </option>
                        );
                      })}
                  </select>
                  {detail.order.status === 'pending' && (
                    <button onClick={openConfirmModal} className="btn-primary" style={{ fontSize: '0.85rem', padding: '0.35rem 1rem' }}>
                      ✓ CONFIRMAR
                    </button>
                  )}
                  {detail.order.status === 'cancelled' && detail.order.confirmed_at && (
                    detail.order.inventory_restored ? (
                      <span className="px-3 py-1.5 rounded bg-status-nm/10 text-status-nm border border-status-nm/20 text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5">
                        <span className="text-xs">✓</span> {t('pages.order.inventory_restored', 'Inventario Restaurado')}
                      </span>
                    ) : (
                      <button onClick={openRestoreModal} className="btn-primary" style={{ fontSize: '0.85rem', padding: '0.35rem 1rem', background: 'var(--status-nm)' }}>
                        ♻ {t('pages.order.restore_inventory', 'RESTAURAR INVENTARIO')}
                      </button>
                    )
                  )}
                </div>
              </div>

              {/* Tracking Info (if shipped or local pickup) */}
              {(detail.order.status === 'shipped' || detail.order.is_local_pickup || detail.order.tracking_number) && (
                <div className="cardbox p-4 mb-4 bg-gold/5 border-gold/20">
                  <h4 className="text-xs font-mono-stack mb-3 text-gold-dark font-bold uppercase tracking-widest">
                    {detail.order.is_local_pickup ? 'ENTREGA EN TIENDA' : 'INFORMACIÓN DE ENVÍO'}
                  </h4>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="text-[10px] font-mono-stack text-text-muted block uppercase mb-1">Guía / Tracking Number</label>
                      <div className="flex gap-2">
                        <input 
                          type="text" 
                          defaultValue={detail.order.tracking_number || ''}
                          onBlur={e => handleStatusChange(detail.order.status, { tracking_number: e.target.value })}
                          disabled={isCompleted}
                          className={`flex-1 text-sm p-1.5 border-kraft-dark/20 ${
                            isCompleted ? 'bg-zinc-100 text-text-muted cursor-not-allowed opacity-70' : 'bg-white'
                          }`}
                          placeholder="Ej: ABC12345"
                        />
                      </div>
                    </div>
                    <div>
                      <label className="text-[10px] font-mono-stack text-text-muted block uppercase mb-1">Link de Seguimiento</label>
                      <input 
                        type="url" 
                        defaultValue={detail.order.tracking_url || ''}
                        onBlur={e => handleStatusChange(detail.order.status, { tracking_url: e.target.value })}
                        disabled={isCompleted}
                        className={`w-full text-sm p-1.5 border-kraft-dark/20 ${
                          isCompleted ? 'bg-zinc-100 text-text-muted cursor-not-allowed opacity-70' : 'bg-white'
                        }`}
                        placeholder="https://servientrega.com/..."
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Customer + Payment info */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-4">
                <div className="cardbox p-3">
                  <h4 className="text-xs font-mono-stack mb-2" style={{ color: 'var(--text-muted)' }}>CLIENTE</h4>
                  <Link href={`/admin/clients/${detail.customer.id}`} className="font-semibold text-gold-dark hover:underline transition-all block mb-1">
                    {detail.customer.first_name} {detail.customer.last_name} →
                  </Link>
                  <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                    📱 <a 
                      href={`https://wa.me/${detail.customer.phone.replace(/\D/g, '')}`} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      className="text-gold-dark hover:underline font-bold transition-all"
                    >
                      {detail.customer.phone}
                    </a>
                  </p>
                  {detail.customer.email && (
                    <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                      ✉ <a href={`mailto:${detail.customer.email}`} className="text-gold-dark hover:underline transition-all">
                        {detail.customer.email}
                      </a>
                    </p>
                  )}
                  {detail.customer.address && <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>📍 {detail.customer.address}</p>}
                  {detail.customer.id_number && <p className="text-xs font-mono-stack mt-1" style={{ color: 'var(--text-muted)' }}>CC: {detail.customer.id_number}</p>}
                  
                  {/* WhatsApp Notification Button */}
                  {detail.whatsapp_url && (
                    <div className="mt-4">
                      <a 
                        href={detail.whatsapp_url} 
                        target="_blank" 
                        rel="noopener noreferrer"
                        className="btn-primary w-full py-2 flex items-center justify-center gap-2 bg-green-600 hover:bg-green-700 text-white"
                        style={{ fontSize: '0.8rem' }}
                      >
                        <span className="text-xl">💬</span>
                        NOTIFICAR POR WHATSAPP
                      </a>
                    </div>
                  )}
                </div>
                <div className={`cardbox p-3 transition-all duration-300 ${!canEditMetadata ? 'opacity-70 grayscale-[0.8] bg-kraft-paper/30' : ''}`}>
                  <h4 className="text-xs font-mono-stack mb-2 flex items-center gap-2" style={{ color: 'var(--text-muted)' }}>
                    PAGO {!canEditMetadata && <span className="text-[10px] bg-red-100 text-red-600 px-1 rounded flex items-center gap-0.5">🔒 BLOQUEADO</span>}
                  </h4>
                  <div className="mb-3">
                    <select 
                      value={paymentMethodEdit}
                      onChange={e => setPaymentMethodEdit(e.target.value)}
                      disabled={!canEditMetadata}
                      className={`w-full border border-kraft-dark/30 rounded p-1 text-sm font-semibold outline-none focus:border-gold transition-colors ${
                        !canEditMetadata ? 'bg-zinc-100 cursor-not-allowed text-text-muted' : 'bg-white'
                      }`}
                    >
                      {Object.entries(PAYMENT_METHODS).map(([key, label]) => (
                        <option key={key} value={key}>{label}</option>
                      ))}
                    </select>
                  </div>
                  <div className="flex justify-between items-center mt-2">
                    <span className="text-xs font-mono-stack text-text-muted uppercase">Subtotal</span>
                    <span className="text-sm font-semibold">${detail.order.subtotal_cop.toLocaleString('en-US', { maximumFractionDigits: 0 })}</span>
                  </div>
                  <div className="flex justify-between items-center mt-2 pt-2 border-t border-kraft-dark/10">
                    <span className="text-xs font-mono-stack text-text-muted uppercase">Envío</span>
                    <div className="flex items-center gap-1">
                      <span className="text-sm font-semibold">$</span>
                      <input 
                        type="number"
                        value={shippingCopEdit}
                        onChange={e => setShippingCopEdit(Number(e.target.value) || 0)}
                        disabled={!canEditMetadata}
                        className={`w-24 text-right border border-kraft-dark/30 rounded p-1 text-sm font-semibold outline-none focus:border-gold transition-colors ${
                          !canEditMetadata ? 'bg-zinc-100 cursor-not-allowed text-text-muted' : 'bg-white'
                        }`}
                      />
                    </div>
                  </div>
                  <div className="price text-xl mt-3 border-t-2 border-gold/20 pt-2 flex justify-between items-center">
                    <span className="text-xs font-mono-stack">TOTAL</span>
                    <span>
                      ${(detail.order.total_cop + (shippingCopEdit - detail.order.shipping_cop)).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
                    </span>
                  </div>
                  {detail.order.notes && (
                    <p className="text-xs mt-2" style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>📝 {detail.order.notes}</p>
                  )}
                </div>
              </div>

              <div className="divider" />

              {/* Items */}
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-display text-xl">PRODUCTOS ({detail.items.length})</h4>
                {detail.order.status !== 'pending' && (
                  <span className="flex items-center gap-1.5 px-2 py-1 rounded bg-hp-color/10 text-status-hp text-[10px] font-bold uppercase tracking-wider border border-status-hp/20">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
                    Inventario Bloqueado
                  </span>
                )}
              </div>
              <div className="space-y-2">
                {detail.items.map(item => {
                  const qty = itemEdits[item.id] ?? item.quantity;
                  const isDeleted = deletedIds.includes(item.id);
                  const isZero = qty === 0 || isDeleted;
                  const badges: string[] = [];
                  if (item.condition) badges.push(item.condition);
                  if (item.foil_treatment && item.foil_treatment !== 'non_foil') badges.push(FOIL_LABELS[item.foil_treatment] || item.foil_treatment);
                  if (item.card_treatment && item.card_treatment !== 'normal') badges.push(TREATMENT_LABELS[item.card_treatment] || item.card_treatment);

                  return (
                    <div key={item.id} className="flex gap-3 p-3 border border-border-main rounded transition-opacity"
                      style={{ opacity: isZero ? 0.4 : 1, background: isZero ? 'var(--bg-surface)' : 'var(--bg-card)', overflow: 'visible' }}>
                      {/* Thumbnail */}
                      <div style={{ width: 44, flexShrink: 0 }}>
                        <CardImage imageUrl={item.image_url} name={item.product_name} tcg="mtg" foilTreatment={item.foil_treatment || 'non_foil'} height={60} enableHover={true} enableModal={true} />
                      </div>

                      {/* Product info */}
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold truncate">{item.product_name}</p>
                        {item.product_set && <p className="text-[10px]" style={{ color: 'var(--text-muted)' }}>{item.product_set}</p>}
                        {badges.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {badges.map((b, i) => <span key={i} className="badge" style={{ fontSize: '0.5rem', padding: '0 3px' }}>{b}</span>)}
                          </div>
                        )}
                        {/* Storage info */}
                        {item.stored_in.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {item.stored_in.map((s: StorageLocation) => (
                              <span key={s.stored_in_id} className="text-[9px] px-1.5 py-0.5 rounded"
                                style={{ background: 'var(--bg-page)', border: '1px solid var(--border-main)', color: 'var(--text-secondary)' }}>
                                📦 {s.name}: {s.quantity}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>

                      {/* Stock info */}
                      <div className="flex flex-col items-center gap-1">
                        <span className="text-[9px] font-mono-stack" style={{ color: 'var(--text-muted)' }}>STOCK</span>
                        <span className="text-sm font-mono-stack font-bold"
                          style={{ color: item.stock === 0 ? 'var(--status-hp)' : item.stock < 3 ? 'var(--status-mp)' : 'var(--text-main)' }}>
                          {item.stock}
                        </span>
                      </div>

                      {/* Quantity controls */}
                      <div className="flex flex-col items-center gap-1">
                        <span className="text-[9px] font-mono-stack" style={{ color: 'var(--text-muted)' }}>QTY</span>
                        <div className="flex items-center gap-1">
                          <button
                            onClick={() => setItemEdits(prev => ({ ...prev, [item.id]: Math.max(0, (prev[item.id] ?? item.quantity) - 1) }))}
                            disabled={!canEditInventory}
                            className={`w-5 h-5 flex items-center justify-center text-xs transition-opacity ${!canEditInventory ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}`}
                            style={{ background: 'var(--border-main)', border: 'none', borderRadius: 2 }}>−</button>
                          <input
                            type="number"
                            value={qty}
                            min={0}
                            max={item.stock}
                            onChange={e => {
                              const val = parseInt(e.target.value) || 0;
                              setItemEdits(prev => ({ ...prev, [item.id]: Math.min(item.stock, Math.max(0, val)) }));
                            }}
                            onBlur={e => {
                              const val = parseInt(e.target.value) || 0;
                              if (val > item.stock) {
                                setItemEdits(prev => ({ ...prev, [item.id]: item.stock }));
                              }
                            }}
                            disabled={!canEditInventory}
                            className={`w-10 text-center text-sm font-mono-stack transition-colors ${
                              !canEditInventory ? 'bg-zinc-100 text-text-muted cursor-not-allowed' : 'bg-white'
                            }`}
                            style={{ height: 20, padding: '0 2px' }}
                          />
                          <button
                            onClick={() => setItemEdits(prev => ({ ...prev, [item.id]: Math.min(item.stock, (prev[item.id] ?? item.quantity) + 1) }))}
                            disabled={!canEditInventory || qty >= item.stock || isDeleted}
                            className={`w-5 h-5 flex items-center justify-center text-xs transition-opacity ${
                              (!canEditInventory || qty >= item.stock || isDeleted) ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'
                            }`}
                            style={{ background: 'var(--border-main)', border: 'none', borderRadius: 2 }}>+</button>
                        </div>
                        {isZero && <span className="text-[8px] font-mono-stack" style={{ color: 'var(--status-hp)' }}>{isDeleted ? 'ELIMINADO' : 'REMOVIDO'}</span>}
                      </div>
 
                      {/* Price */}
                      <div className="text-right flex-shrink-0 flex flex-col justify-between">
                        <button 
                          onClick={() => {
                            if (isDeleted) {
                              setDeletedIds(prev => prev.filter(id => id !== item.id));
                            } else {
                              setDeletedIds(prev => [...prev, item.id]);
                            }
                          }}
                          disabled={detail.order.status === 'completed'}
                          className={`p-1 rounded transition-colors ${isDeleted ? 'text-status-nm hover:text-status-nm/80' : 'text-hp-color/40 hover:text-hp-color'}`}
                          title={isDeleted ? 'Restaurar' : 'Eliminar de la orden'}
                        >
                          {isDeleted ? (
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path><path d="M3 3v5h5"></path></svg>
                          ) : (
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                          )}
                        </button>
                        <div>
                          <span className="text-[9px] font-mono-stack block" style={{ color: 'var(--text-muted)' }}>
                            ${item.unit_price_cop.toLocaleString('en-US', { maximumFractionDigits: 0 })} c/u
                          </span>
                          <span className="price text-sm">
                            ${(item.unit_price_cop * (isDeleted ? 0 : qty)).toLocaleString('en-US', { maximumFractionDigits: 0 })}
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Add New Product Search */}
              {detail.order.status === 'pending' && (
                <div className="mt-6 mb-4 cardbox p-4 bg-ink-surface/30 border-dashed border-gold/30">
                  <h4 className="text-xs font-mono-stack mb-3 text-gold-dark font-bold uppercase tracking-widest flex items-center gap-2">
                    <span className="text-lg">🎁</span> AGREGAR PRODUCTO / REGALO
                  </h4>
                  <div className="relative">
                    <input 
                      type="text"
                      value={productSearch}
                      onChange={e => setProductSearch(e.target.value)}
                      placeholder="Buscar producto por nombre..."
                      disabled={!canEditInventory}
                      className={`w-full text-sm p-3 border-kraft-dark/40 rounded shadow-inner transition-colors ${
                        !canEditInventory ? 'bg-zinc-100 text-text-muted cursor-not-allowed' : 'bg-white'
                      }`}
                    />
                    {searchingItems && (
                      <div className="absolute right-3 top-3">
                         <div className="animate-spin h-4 w-4 border-2 border-gold border-t-transparent rounded-full"></div>
                      </div>
                    )}
                    {searchResults.length > 0 && (
                      <div className="absolute left-0 right-0 top-full mt-1 bg-white border border-kraft-dark/30 rounded shadow-xl z-50 max-h-60 overflow-y-auto">
                        {searchResults.map(p => (
                          <button
                            key={p.id}
                            onClick={() => addStagedItem(p)}
                            className="w-full flex items-center gap-3 p-2 hover:bg-gold/10 transition-colors border-b border-kraft-dark/10"
                          >
                            <div className="w-8 h-10 border border-kraft-dark/20 rounded overflow-hidden flex-shrink-0">
                               <CardImage imageUrl={p.image_url} name={p.name} tcg="mtg" foilTreatment={p.foil_treatment || 'non_foil'} height={40} enableHover={false} />
                            </div>
                            <div className="text-left flex-1 min-w-0">
                              <p className="text-xs font-bold truncate leading-tight">{p.name}</p>
                              <p className="text-[10px] text-text-muted">{p.set_name} · {p.stock} en stock</p>
                            </div>
                            <div className="text-xs font-mono-stack text-gold-dark font-bold">
                              ${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Staged Items Table */}
              {stagedItems.length > 0 && (
                <div className="mb-6 space-y-2">
                   <h5 className="text-[10px] font-mono-stack text-text-muted uppercase tracking-widest font-bold">Por agregar:</h5>
                   {stagedItems.map(si => (
                     <div key={si.product.id} className="flex gap-3 p-3 border-2 border-status-nm/30 rounded bg-status-nm/5">
                        <div className="w-8 flex-shrink-0">
                           <CardImage imageUrl={si.product.image_url} name={si.product.name} tcg="mtg" foilTreatment={si.product.foil_treatment || 'non_foil'} height={40} />
                        </div>
                        <div className="flex-1">
                          <p className="text-sm font-bold leading-tight">{si.product.name}</p>
                          <p className="text-[10px] text-text-muted">{si.product.set_name}</p>
                        </div>
                        <div className="flex flex-col items-center">
                          <span className="text-[8px] text-text-muted">PESO ($)</span>
                          <input 
                             type="number"
                             value={si.unit_price_cop}
                             onChange={e => updateStagedItem(si.product.id, { unit_price_cop: Number(e.target.value) || 0 })}
                             className="w-20 text-xs p-1 text-center font-mono-stack outline-none border border-kraft-dark/20 rounded"
                          />
                        </div>
                        <div className="flex flex-col items-center">
                          <span className="text-[8px] text-text-muted">CANT</span>
                          <input 
                             type="number"
                             value={si.quantity}
                             onChange={e => updateStagedItem(si.product.id, { quantity: Math.min(si.product.stock, Number(e.target.value) || 1) })}
                             className="w-12 text-xs p-1 text-center font-mono-stack outline-none border border-kraft-dark/20 rounded"
                          />
                        </div>
                        <button onClick={() => removeStagedItem(si.product.id)} className="text-hp-color transition-colors self-center p-1">
                           <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                        </button>
                     </div>
                   ))}
                </div>
              )}

              {/* Save changes button */}
              {hasEdits && detail.order.status !== 'completed' && detail.order.status !== 'confirmed' && detail.order.status !== 'cancelled' && (
                <div className="mt-4 flex gap-3">
                  <button onClick={handleSaveChanges} disabled={saving} className="btn-primary flex-1" style={{ fontSize: '0.9rem' }}>
                    {saving ? 'GUARDANDO...' : 'GUARDAR CAMBIOS'}
                  </button>
                  <button onClick={() => {
                    const edits: Record<string, number> = {};
                    detail.items.forEach(i => { edits[i.id] = i.quantity; });
                    setItemEdits(edits);
                    setStagedItems([]);
                    setDeletedIds([]);
                  }} className="btn-secondary" style={{ fontSize: '0.9rem' }}>DESCARTAR</button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Confirm Order Modal */}
      {showConfirmModal && detail && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.9)', backdropFilter: 'blur(3px)' }}>
          <div className="card max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
            <h3 className="font-display text-3xl mb-1">CONFIRMAR ORDEN</h3>
            <p className="text-xs font-mono-stack mb-4" style={{ color: 'var(--text-muted)' }}>
              Selecciona de qué ubicación descontar el stock para cada producto.
            </p>

            <div className="space-y-4">
              {detail.items.filter(i => i.quantity > 0 && i.product_id).map(item => {
                const productDecs = decrements[item.product_id!] || {};
                const totalAssigned = Object.values(productDecs).reduce((s, v) => s + v, 0);
                const isComplete = totalAssigned === item.quantity;

                return (
                  <div key={item.id} className="cardbox p-4" style={{ borderColor: isComplete ? 'var(--status-nm)' : 'var(--border-main)' }}>
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <p className="font-semibold text-sm">{item.product_name}</p>
                        {item.product_set && <p className="text-[10px]" style={{ color: 'var(--text-muted)' }}>{item.product_set}</p>}
                      </div>
                      <div className="text-right">
                        <span className="text-xs font-mono-stack" style={{ color: isComplete ? 'var(--status-nm)' : 'var(--status-mp)' }}>
                          {totalAssigned} / {item.quantity}
                        </span>
                      </div>
                    </div>

                    {item.stored_in.filter(l => l.name.toLowerCase() !== 'pending').length === 0 ? (
                      <p className="text-xs italic" style={{ color: 'var(--status-hp)' }}>⚠ Sin ubicaciones físicas de almacenamiento</p>
                    ) : (
                      <div className="space-y-2">
                        {/* pending row */}
                        <div className="flex items-center justify-between gap-3 opacity-60 bg-status-hp/10 px-2 py-1.5 rounded border border-status-hp/20">
                           <div className="flex-1">
                             <span className="text-sm font-semibold">pending</span>
                             <span className="text-[10px] uppercase ml-2 px-1 py-0.5 rounded bg-hp-color/20 text-status-hp font-mono-stack">AUTO</span>
                           </div>
                           <div className="flex items-center gap-1">
                             <button disabled className="w-6 h-6 flex items-center justify-center text-xs opacity-50 bg-ink-surface border border-ink-border rounded-l-sm cursor-not-allowed">−</button>
                             <input
                               type="number" 
                               value={Math.max(0, item.quantity - totalAssigned)}
                               disabled
                               className="w-12 px-1 py-0 text-center text-xs font-mono-stack border-y border-ink-border bg-white focus:outline-none opacity-50 cursor-not-allowed [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                               style={{ height: 24, padding: '0 2px' }}
                             />
                             <button disabled className="w-6 h-6 flex items-center justify-center text-xs opacity-50 bg-ink-surface border border-ink-border rounded-r-sm cursor-not-allowed">+</button>
                           </div>
                        </div>

                        {item.stored_in.filter((loc: StorageLocation) => loc.name.toLowerCase() !== 'pending').map((loc: StorageLocation) => {
                          const val = productDecs[loc.stored_in_id] || 0;
                          return (
                            <div key={loc.stored_in_id} className="flex items-center justify-between gap-3">
                              <div className="flex-1">
                                <span className="text-sm font-semibold">{loc.name}</span>
                                <span className="text-xs font-mono-stack ml-2" style={{ color: 'var(--text-muted)' }}>
                                  (disponible: {loc.quantity})
                                </span>
                              </div>
                              <div className="flex items-center gap-1">
                                <button
                                  onClick={() => setDecrement(item.product_id!, loc.stored_in_id, val - 1)}
                                  className="w-6 h-6 flex items-center justify-center text-xs"
                                  style={{ background: 'var(--ink-border)', border: 'none', borderRadius: 2, cursor: 'pointer' }}
                                  disabled={val <= 0}>−</button>
                                <input
                                  type="number" min={0} max={loc.quantity}
                                  value={val || ''}
                                  onChange={e => setDecrement(item.product_id!, loc.stored_in_id, Math.min(parseInt(e.target.value) || 0, loc.quantity))}
                                  className="w-12 text-center text-sm font-mono-stack"
                                  style={{ height: 24, padding: '0 2px' }}
                                  placeholder="0"
                                />
                                <button
                                  onClick={() => setDecrement(item.product_id!, loc.stored_in_id, Math.min(val + 1, loc.quantity))}
                                  className="w-6 h-6 flex items-center justify-center text-xs"
                                  style={{ background: 'var(--ink-border)', border: 'none', borderRadius: 2, cursor: 'pointer' }}
                                  disabled={val >= loc.quantity}>+</button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {confirmError && (
              <p className="text-sm font-mono-stack mt-3" style={{ color: 'var(--status-hp)' }}>{confirmError}</p>
            )}

            <div className="mt-6 p-4 border-2 border-dashed border-status-hp/30 rounded-lg bg-status-hp/5">
              <p className="text-sm font-semibold text-status-hp mb-3 uppercase tracking-wider text-center">
                ⚠ ¿Estás seguro? Esto bloqueará el stock asignado.
              </p>
              <div className="flex gap-3">
                <button onClick={handleConfirm} disabled={confirming} className="btn-primary flex-1 py-3 bg-status-hp hover:bg-status-hp/90">
                  {confirming ? 'PROCESANDO...' : '✓ CONFIRMAR ORDEN'}
                </button>
                <button onClick={() => setShowConfirmModal(false)} className="btn-secondary px-6 py-3">CANCELAR</button>
              </div>
            </div>
          </div>
        </div>
      )}
      {/* Restore Stock Modal */}
      {showRestoreModal && detail && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.9)', backdropFilter: 'blur(3px)' }}>
          <div className="card max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
            <h3 className="font-display text-3xl mb-1">RESTAURAR INVENTARIO</h3>
            <p className="text-xs font-mono-stack mb-4" style={{ color: 'var(--text-muted)' }}>
              Selecciona a qué ubicaciones devolver los productos de esta orden cancelada.
            </p>

            <div className="space-y-4">
              {detail.items.filter(i => i.quantity > 0 && i.product_id).map(item => {
                const productIncs = increments[item.product_id!] || {};
                const totalAssigned = Object.values(productIncs).reduce((s, v) => s + v, 0);
                const isComplete = totalAssigned === item.quantity;

                return (
                  <div key={item.id} className="cardbox p-4" style={{ borderColor: isComplete ? 'var(--status-nm)' : 'var(--border-main)' }}>
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <p className="font-semibold text-sm">{item.product_name}</p>
                        {item.product_set && <p className="text-[10px]" style={{ color: 'var(--text-muted)' }}>{item.product_set}</p>}
                      </div>
                      <div className="text-right">
                        <span className="text-xs font-mono-stack" style={{ color: isComplete ? 'var(--status-nm)' : 'var(--status-mp)' }}>
                          {totalAssigned} / {item.quantity}
                        </span>
                      </div>
                    </div>

                    <div className="space-y-2">
                        {allStorage
                          .filter(as => as.name.toLowerCase() !== 'pending' && (item.stored_in.some(si => si.stored_in_id === as.id) || productIncs[as.id] !== undefined))
                          .map(as => {
                            const si = item.stored_in.find(s => s.stored_in_id === as.id);
                            const val = productIncs[as.id] || 0;
                            return (
                              <div key={as.id} className="flex items-center justify-between gap-3">
                                <div className="flex-1">
                                  <span className="text-sm font-semibold">{as.name}</span>
                                  <span className="text-xs font-mono-stack ml-2" style={{ color: 'var(--text-muted)' }}>
                                    (actual: {si?.quantity || 0})
                                  </span>
                                </div>
                                <div className="flex items-center gap-1">
                                  <button
                                    onClick={() => setIncrement(item.product_id!, as.id, val - 1)}
                                    className="w-6 h-6 flex items-center justify-center text-xs"
                                    style={{ background: 'var(--ink-border)', border: 'none', borderRadius: 2, cursor: 'pointer' }}
                                    disabled={val <= 0}>−</button>
                                  <input
                                    type="number" min={0} max={item.quantity}
                                    value={val || ''}
                                    onChange={e => {
                                      const newVal = Math.max(0, parseInt(e.target.value) || 0);
                                      const otherTotal = totalAssigned - val;
                                      const allowed = Math.max(0, item.quantity - otherTotal);
                                      setIncrement(item.product_id!, as.id, Math.min(newVal, allowed));
                                    }}
                                    className="w-12 text-center text-sm font-mono-stack"
                                    style={{ height: 24, padding: '0 2px' }}
                                    placeholder="0"
                                  />
                                  <button
                                    onClick={() => {
                                      if (totalAssigned < item.quantity) {
                                        setIncrement(item.product_id!, as.id, val + 1);
                                      }
                                    }}
                                    className="w-6 h-6 flex items-center justify-center text-xs"
                                    style={{ 
                                      background: 'var(--ink-border)', 
                                      border: 'none', 
                                      borderRadius: 2, 
                                      cursor: totalAssigned >= item.quantity ? 'not-allowed' : 'pointer',
                                      opacity: totalAssigned >= item.quantity ? 0.5 : 1
                                    }}
                                    disabled={totalAssigned >= item.quantity}
                                  >+</button>
                                </div>
                              </div>
                            );
                          })}
                        
                        {/* Add Location Dropdown */}
                        <div className="mt-2 pt-2 border-t border-border-main/10 flex justify-end">
                          <select 
                            className="text-[10px] bg-transparent border-none text-gold-dark font-bold cursor-pointer outline-none hover:text-hp-color transition-colors"
                            value=""
                            onChange={e => {
                              if (e.target.value) {
                                setIncrement(item.product_id!, e.target.value, 0);
                              }
                            }}
                          >
                            <option value="">+ {t('pages.order.add_location', 'AÑADIR UBICACIÓN')}</option>
                            {allStorage
                              .filter(as => as.name.toLowerCase() !== 'pending' && !item.stored_in.some(si => si.stored_in_id === as.id) && productIncs[as.id] === undefined)
                              .map(as => (
                                <option key={as.id} value={as.id}>{as.name}</option>
                              ))
                            }
                          </select>
                        </div>
                    </div>
                  </div>
                );
              })}
            </div>

            {restoreError && (
              <p className="text-sm font-mono-stack mt-3" style={{ color: 'var(--status-hp)' }}>{restoreError}</p>
            )}

            <div className="mt-6 p-4 border-2 border-dashed border-status-nm/30 rounded-lg bg-status-nm/5">
              <p className="text-sm font-semibold text-status-nm mb-3 uppercase tracking-wider text-center">
                Confirmar devolución de productos al inventario físico.
              </p>
              <div className="flex gap-3">
                <button onClick={handleRestoreStock} disabled={restoring} className="btn-primary flex-1 py-3" style={{ background: 'var(--status-nm)' }}>
                  {restoring ? 'RESTAURANDO...' : '✓ RESTAURAR INVENTARIO'}
                </button>
                <button onClick={() => setShowRestoreModal(false)} className="btn-secondary px-6 py-3">CANCELAR</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
