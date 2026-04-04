'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import {
  adminFetchOrders, adminFetchOrderDetail, adminUpdateOrder, adminCompleteOrder,
  adminDownloadAccountingCSV
} from '@/lib/api';
import {
  OrderWithCustomer, OrderDetail,
  ORDER_STATUS_LABELS, PAYMENT_METHODS, FOIL_LABELS, TREATMENT_LABELS, StorageLocation
} from '@/lib/types';
import CardImage from '@/components/CardImage';

interface Props {
  initialOrderId?: string | null;
}

export default function OrdersPanel({ initialOrderId }: Props) {
  // List state
  const [orders, setOrders] = useState<OrderWithCustomer[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [loading, setLoading] = useState(true);

  // Accounting export state
  const [exportDates, setExportDates] = useState({ start: '', end: '' });
  const [exporting, setExporting] = useState(false);

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

  // Complete modal state
  const [showCompleteModal, setShowCompleteModal] = useState(false);
  const [decrements, setDecrements] = useState<Record<string, Record<string, number>>>({});
  const [completing, setCompleting] = useState(false);
  const [completeError, setCompleteError] = useState('');

  // Mobile master-detail toggle
  const [mobileShowDetail, setMobileShowDetail] = useState(false);

  const loadOrders = useCallback(async () => {
    // No longer setting loading synchronously at start to avoid cascaded renders
    try {
      const res = await adminFetchOrders({ status: statusFilter, search: debouncedSearch, page, page_size: 20 });
      setOrders(res.orders);
      setTotal(res.total);
    } catch (err) {
      console.error('Failed to load orders:', err);
    }
    setLoading(false);
  }, [statusFilter, debouncedSearch, page]);

  useEffect(() => {
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
      return;
    }

    setLoadingDetail(true);
    try {
      const d = await adminFetchOrderDetail(id);
      setDetail(d);
      setDetailsCache(prev => ({ ...prev, [id]: d }));
      // Init item edits
      const edits: Record<string, number> = {};
      d.items.forEach(i => { edits[i.id] = i.quantity; });
      setItemEdits(edits);
    } catch { }
    setLoadingDetail(false);
  }, [detailsCache]);

  useEffect(() => {
    if (initialOrderId) {
      const timer = setTimeout(() => {
        setItemEdits({});
        selectOrder(initialOrderId);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [initialOrderId, selectOrder]);

  const handleSaveChanges = async () => {
    if (!detail) return;
    setSaving(true);
    try {
      const items = Object.entries(itemEdits).map(([id, quantity]) => ({ id, quantity }));
      const updated = await adminUpdateOrder(detail.order.id, { items });
      setDetail(updated);
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

  const handleStatusChange = async (status: string) => {
    if (!detail) return;
    setSaving(true);
    try {
      const updated = await adminUpdateOrder(detail.order.id, { status });
      setDetail(updated);
      setDetailsCache(prev => ({ ...prev, [updated.order.id]: updated }));
      // Update list in-place
      setOrders(prev => prev.map(o => o.id === updated.order.id ? { ...o, status: updated.order.status } : o));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : 'Failed to update status');
    }
    setSaving(false);
  };

  const openCompleteModal = () => {
    if (!detail) return;
    const initial: Record<string, Record<string, number>> = {};
    detail.items.forEach(item => {
      if (item.quantity > 0 && item.product_id) {
        initial[item.product_id] = {};
        // Pre-fill: auto-assign from first storage location with enough stock
        let remaining = item.quantity;
        item.stored_in.forEach(loc => {
          if (remaining > 0) {
            const take = Math.min(remaining, loc.quantity);
            initial[item.product_id!][loc.stored_in_id] = take;
            remaining -= take;
          }
        });
      }
    });
    setDecrements(initial);
    setCompleteError('');
    setShowCompleteModal(true);
  };

  const handleComplete = async () => {
    if (!detail) return;
    setCompleting(true);
    setCompleteError('');

    // Validate: total decrements per product must match order quantity
    for (const item of detail.items) {
      if (item.quantity <= 0 || !item.product_id) continue;
      const productDecs = decrements[item.product_id] || {};
      const totalDec = Object.values(productDecs).reduce((s, v) => s + v, 0);
      if (totalDec !== item.quantity) {
        setCompleteError(`${item.product_name}: asignaste ${totalDec} de ${item.quantity} necesarios.`);
        setCompleting(false);
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
      const updated = await adminCompleteOrder(detail.order.id, decArr);
      setDetail(updated);
      setShowCompleteModal(false);
      // Update list in-place
      setOrders(prev => prev.map(o => o.id === updated.order.id ? { ...o, status: updated.order.status } : o));
    } catch (e: unknown) {
      setCompleteError(e instanceof Error ? e.message : 'Error al completar');
    }
    setCompleting(false);
  };

  const setDecrement = (productId: string, storedInId: string, qty: number) => {
    setDecrements(prev => ({
      ...prev,
      [productId]: { ...prev[productId], [storedInId]: Math.max(0, qty) }
    }));
  };

  const totalPages = Math.ceil(total / 20);
  const hasEdits = detail ? detail.items.some(i => itemEdits[i.id] !== i.quantity) : false;

  const handleExport = async () => {
    setExporting(true);
    try {
      await adminDownloadAccountingCSV({
        start_date: exportDates.start ? `${exportDates.start}T00:00:00Z` : undefined,
        end_date: exportDates.end ? `${exportDates.end}T23:59:59Z` : undefined,
      });
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Error exporting accounting data');
    }
    setExporting(false);
  };

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex flex-col lg:flex-row flex-1 min-h-0">
        {/* Left: Orders List */}
        <div className={`w-full lg:w-[420px] flex-shrink-0 flex flex-col overflow-hidden border-r border-border-main/20 bg-bg-page/5 ${mobileShowDetail ? 'hidden lg:flex' : 'flex'}`}>
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
                  {s ? (ORDER_STATUS_LABELS[s] || s) : 'Todas'}
                </button>
              ))}
            </div>

            {/* Accounting Export UI */}
            <div className="mt-2 p-2 rounded bg-gold/5 border border-gold/20">
              <p className="text-[10px] font-mono-stack text-gold-dark mb-1 font-bold uppercase tracking-wider">Accounting Export (CSV)</p>
              <div className="flex gap-2 items-end">
                <div className="flex-1">
                  <label className="text-[8px] font-mono-stack text-text-muted block uppercase">Start</label>
                  <input 
                    type="date" 
                    value={exportDates.start} 
                    onChange={e => setExportDates(prev => ({ ...prev, start: e.target.value }))}
                    className="w-full text-[10px] p-1 bg-white border-kraft-dark/20"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-[8px] font-mono-stack text-text-muted block uppercase">End</label>
                  <input 
                    type="date" 
                    value={exportDates.end} 
                    onChange={e => setExportDates(prev => ({ ...prev, end: e.target.value }))}
                    className="w-full text-[10px] p-1 bg-white border-kraft-dark/20"
                  />
                </div>
                <button 
                  onClick={handleExport}
                  disabled={exporting}
                  className="btn-primary text-[10px] p-1.5 px-3 whitespace-nowrap shadow-sm"
                >
                  {exporting ? '...' : 'CSV'}
                </button>
              </div>
            </div>
          </div>

          {/* Orders list */}
          <div className="flex-1 overflow-y-auto">
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
                        {ORDER_STATUS_LABELS[o.status] || o.status}
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
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 p-2" style={{ borderTop: '1px solid var(--border-main)' }}>
              <button onClick={() => { setPage(p => Math.max(1, p - 1)); setLoading(true); }} disabled={page === 1} className="btn-secondary" style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem', opacity: page === 1 ? 0.4 : 1 }}>←</button>
              <span className="text-xs font-mono-stack flex items-center" style={{ color: 'var(--text-muted)' }}>{page}/{totalPages}</span>
              <button onClick={() => { setPage(p => Math.min(totalPages, p + 1)); setLoading(true); }} disabled={page === totalPages} className="btn-secondary" style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem', opacity: page === totalPages ? 0.4 : 1 }}>→</button>
            </div>
          )}
        </div>

        {/* Right: Order Detail */}
        <div className={`flex-1 overflow-y-auto p-4 md:p-8 bg-white ${!mobileShowDetail ? 'hidden lg:block' : 'block'}`}>
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
                    disabled={saving || detail.order.status === 'completed'}
                    className="bg-surface border-dark text-gold rounded px-2 py-1 outline-none focus:border-gold transition-colors"
                    style={{ fontSize: '0.85rem', padding: '0.3rem 0.6rem', minWidth: 120 }}
                  >
                    {Object.entries(ORDER_STATUS_LABELS)
                      .filter(([k]) => k !== 'completed' || detail.order.status === 'completed')
                      .map(([k, v]) => (
                        <option key={k} value={k}>{v}</option>
                      ))}
                  </select>
                  {detail.order.status !== 'completed' && detail.order.status !== 'cancelled' && (
                    <button onClick={openCompleteModal} className="btn-primary" style={{ fontSize: '0.85rem', padding: '0.35rem 1rem' }}>
                      ✓ COMPLETAR
                    </button>
                  )}
                </div>
              </div>

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
                </div>
                <div className="cardbox p-3">
                  <h4 className="text-xs font-mono-stack mb-2" style={{ color: 'var(--text-muted)' }}>PAGO</h4>
                  <p className="font-semibold">{PAYMENT_METHODS[detail.order.payment_method] || detail.order.payment_method}</p>
                  <p className="price text-xl mt-1">${detail.order.total_cop.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</p>
                  {detail.order.notes && (
                    <p className="text-xs mt-2" style={{ color: 'var(--text-muted)', fontStyle: 'italic' }}>📝 {detail.order.notes}</p>
                  )}
                </div>
              </div>

              <div className="divider" />

              {/* Items */}
              <h4 className="font-display text-xl mb-3">PRODUCTOS ({detail.items.length})</h4>
              <div className="space-y-2">
                {detail.items.map(item => {
                  const qty = itemEdits[item.id] ?? item.quantity;
                  const isZero = qty === 0;
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
                            disabled={detail.order.status === 'completed'}
                            className="w-5 h-5 flex items-center justify-center text-xs"
                            style={{ background: 'var(--border-main)', border: 'none', borderRadius: 2, cursor: detail.order.status === 'completed' ? 'not-allowed' : 'pointer' }}>−</button>
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
                            disabled={detail.order.status === 'completed'}
                            className="w-10 text-center text-sm font-mono-stack"
                            style={{ height: 20, padding: '0 2px' }}
                          />
                          <button
                            onClick={() => setItemEdits(prev => ({ ...prev, [item.id]: Math.min(item.stock, (prev[item.id] ?? item.quantity) + 1) }))}
                            disabled={detail.order.status === 'completed' || qty >= item.stock}
                            className="w-5 h-5 flex items-center justify-center text-xs"
                            style={{ background: 'var(--border-main)', border: 'none', borderRadius: 2, cursor: (detail.order.status === 'completed' || qty >= item.stock) ? 'not-allowed' : 'pointer' }}>+</button>
                        </div>
                        {isZero && <span className="text-[8px] font-mono-stack" style={{ color: 'var(--status-hp)' }}>REMOVIDO</span>}
                      </div>

                      {/* Price */}
                      <div className="text-right flex-shrink-0">
                        <span className="text-[9px] font-mono-stack block" style={{ color: 'var(--text-muted)' }}>
                          ${item.unit_price_cop.toLocaleString('en-US', { maximumFractionDigits: 0 })} c/u
                        </span>
                        <span className="price text-sm">
                          ${(item.unit_price_cop * qty).toLocaleString('en-US', { maximumFractionDigits: 0 })}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Save changes button */}
              {hasEdits && detail.order.status !== 'completed' && (
                <div className="mt-4 flex gap-3">
                  <button onClick={handleSaveChanges} disabled={saving} className="btn-primary flex-1" style={{ fontSize: '0.9rem' }}>
                    {saving ? 'GUARDANDO...' : 'GUARDAR CAMBIOS'}
                  </button>
                  <button onClick={() => {
                    const edits: Record<string, number> = {};
                    detail.items.forEach(i => { edits[i.id] = i.quantity; });
                    setItemEdits(edits);
                  }} className="btn-secondary" style={{ fontSize: '0.9rem' }}>DESCARTAR</button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Complete Order Modal */}
      {showCompleteModal && detail && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center px-4"
          style={{ background: 'rgba(0,0,0,0.9)', backdropFilter: 'blur(3px)' }}>
          <div className="card max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto">
            <h3 className="font-display text-3xl mb-1">COMPLETAR ORDEN</h3>
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

                    {item.stored_in.length === 0 ? (
                      <p className="text-xs italic" style={{ color: 'var(--status-hp)' }}>⚠ Sin ubicaciones de almacenamiento</p>
                    ) : (
                      <div className="space-y-2">
                        {item.stored_in.map((loc: StorageLocation) => {
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

            {completeError && (
              <p className="text-sm font-mono-stack mt-3" style={{ color: 'var(--status-hp)' }}>{completeError}</p>
            )}

            <div className="mt-6 p-4 border-2 border-dashed border-status-hp/30 rounded-lg bg-status-hp/5">
              <p className="text-sm font-semibold text-status-hp mb-3 uppercase tracking-wider text-center">
                ⚠ ¿Estás seguro? Esto bloqueará la orden y no se podrá editar después.
              </p>
              <div className="flex gap-3">
                <button onClick={handleComplete} disabled={completing} className="btn-primary flex-1 py-3 bg-status-hp hover:bg-status-hp/90">
                  {completing ? 'PROCESANDO...' : '✓ CONFIRMAR Y COMPLETAR ORDEN'}
                </button>
                <button onClick={() => setShowCompleteModal(false)} className="btn-secondary px-6 py-3">CANCELAR</button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
