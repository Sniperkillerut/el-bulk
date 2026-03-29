'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useCart } from '@/lib/CartContext';
import { createOrder } from '@/lib/api';
import { PAYMENT_METHODS, FOIL_LABELS, TREATMENT_LABELS } from '@/lib/types';
import CardImage from '@/components/CardImage';

export default function CheckoutPage() {
  const router = useRouter();
  const { items, removedItems, totalPrice, updateQty, removeItem, restoreItem, permanentRemove, clearCart } = useCart();
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const [form, setForm] = useState({
    first_name: '',
    last_name: '',
    phone: '',
    email: '',
    id_number: '',
    address: '',
    payment_method: 'cash',
    notes: '',
  });

  const set = (key: string, val: string) => setForm(f => ({ ...f, [key]: val }));

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    if (items.length === 0) {
      setError('Tu carrito está vacío.');
      return;
    }

    setSubmitting(true);
    setError('');
    try {
      const result = await createOrder({
        ...form,
        items: items.map(i => ({ product_id: i.product.id, quantity: i.quantity })),
      });
      clearCart();
      router.push(`/order/${result.order_number}`);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Error al crear la orden.');
    } finally {
      setSubmitting(false);
    }
  };

  if (items.length === 0) {
    return (
      <div className="centered-container px-4 py-16 text-center">
        <h1 className="font-display text-5xl mb-4">CHECKOUT</h1>
        <p style={{ color: 'var(--text-muted)', fontSize: '1.1rem' }}>
          Tu carrito está vacío. Agrega productos antes de continuar.
        </p>
        <button onClick={() => router.push('/')} className="btn-primary mt-6">
          ← VOLVER A LA TIENDA
        </button>
      </div>
    );
  }

  return (
    <div className="centered-container px-4 py-8">
      <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>EL BULK / CHECKOUT</p>
      <h1 className="font-display text-5xl mb-2">FINALIZAR COMPRA</h1>
      <div className="gold-line mb-8" />

      <form onSubmit={handleSubmit} className="flex flex-col lg:flex-row gap-8">
        {/* Left: Customer Form */}
        <div className="flex-1">
          <div className="card p-6">
            <h2 className="font-display text-2xl mb-4">INFORMACIÓN DE CONTACTO</h2>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>NOMBRE *</label>
                <input type="text" value={form.first_name} onChange={e => set('first_name', e.target.value)} placeholder="Juan" required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>APELLIDO *</label>
                <input type="text" value={form.last_name} onChange={e => set('last_name', e.target.value)} placeholder="Pérez" required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>TELÉFONO / WHATSAPP *</label>
                <input type="tel" value={form.phone} onChange={e => set('phone', e.target.value)} placeholder="3001234567" required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>EMAIL *</label>
                <input type="email" value={form.email} onChange={e => set('email', e.target.value)} placeholder="correo@ejemplo.com" required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>CÉDULA / ID *</label>
                <input type="number" value={form.id_number} onChange={e => set('id_number', e.target.value)} placeholder="1234567890" required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>DIRECCIÓN *</label>
                <input type="text" value={form.address} onChange={e => set('address', e.target.value)} placeholder="Cra 1 # 2-3" required />
              </div>
            </div>

            <div className="divider" />

            <h3 className="font-display text-xl mb-3">MÉTODO DE PAGO</h3>
            <div className="flex flex-wrap gap-2">
              {Object.entries(PAYMENT_METHODS).map(([key, label]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => set('payment_method', key)}
                  className="badge transition-colors cursor-pointer"
                  style={{
                    fontSize: '0.85rem',
                    padding: '0.5rem 1rem',
                    background: form.payment_method === key ? 'var(--ink-deep)' : 'var(--ink-surface)',
                    color: form.payment_method === key ? '#fff' : 'var(--text-secondary)',
                    border: `2px solid ${form.payment_method === key ? 'var(--ink-deep)' : 'var(--ink-border)'}`,
                  }}
                >
                  {label}
                </button>
              ))}
            </div>

            <div className="mt-4">
              <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>NOTAS (OPCIONAL)</label>
              <textarea value={form.notes} onChange={e => set('notes', e.target.value)} rows={2} placeholder="Instrucciones especiales..." />
            </div>
          </div>
        </div>

        {/* Right: Order Summary */}
        <div className="w-full lg:w-[420px] flex-shrink-0">
          <div className="card p-6 sticky top-4">
            <h2 className="font-display text-2xl mb-4">RESUMEN DEL PEDIDO</h2>
            <p className="text-xs font-mono-stack mb-4" style={{ color: 'var(--text-muted)' }}>
              {items.reduce((s, i) => s + i.quantity, 0)} ARTÍCULO{items.length !== 1 ? 'S' : ''}
            </p>

            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-1">
              {items.map(item => {
                const p = item.product;
                const badges: string[] = [];
                if (p.condition) badges.push(p.condition);
                if (p.foil_treatment && p.foil_treatment !== 'non_foil') badges.push(FOIL_LABELS[p.foil_treatment] || p.foil_treatment);
                if (p.card_treatment && p.card_treatment !== 'normal') badges.push(TREATMENT_LABELS[p.card_treatment] || p.card_treatment);

                return (
                  <div key={p.id} className="flex gap-3 pb-3" style={{ borderBottom: '1px solid var(--ink-border)' }}>
                    <div style={{ width: 52, flexShrink: 0 }}>
                      <CardImage imageUrl={p.image_url} name={p.name} tcg={p.tcg} height={70} enableHover={true} enableModal={true} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-semibold truncate">{p.name}</p>
                      {p.set_name && <p className="text-[10px] truncate" style={{ color: 'var(--text-muted)' }}>{p.set_name}</p>}
                      {badges.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {badges.map((b, i) => <span key={i} className="badge" style={{ fontSize: '0.55rem', padding: '1px 4px' }}>{b}</span>)}
                        </div>
                      )}
                      <div className="flex items-center gap-2 mt-2">
                        <button
                          type="button"
                          onClick={() => updateQty(p.id, item.quantity - 1)}
                          style={{ width: 22, height: 22, background: 'var(--ink-border)', border: 'none', borderRadius: 2, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.9rem' }}
                        >−</button>
                        <span className="text-sm font-mono-stack w-5 text-center">{item.quantity}</span>
                        <button
                          type="button"
                          onClick={() => updateQty(p.id, item.quantity + 1)}
                          disabled={item.quantity >= p.stock}
                          style={{ width: 22, height: 22, background: 'var(--ink-border)', border: 'none', borderRadius: 2, cursor: item.quantity >= p.stock ? 'not-allowed' : 'pointer', opacity: item.quantity >= p.stock ? 0.4 : 1, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.9rem' }}
                        >+</button>
                        <span className="text-[10px] font-mono-stack ml-1" style={{ color: 'var(--text-muted)' }}>
                          {p.stock} dispon.
                        </span>
                        <button
                          type="button"
                          onClick={() => removeItem(p.id)}
                          className="ml-auto p-1 opacity-60 hover:opacity-100 transition-opacity"
                          title="Eliminar"
                          style={{ color: 'var(--hp-color)', background: 'none', border: 'none', cursor: 'pointer' }}
                        >
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <polyline points="3 6 5 6 21 6" /><path d="M19 6l-1 14H6L5 6" /><path d="M10 11v6M14 11v6" /><path d="M9 6V4h6v2" />
                          </svg>
                        </button>
                      </div>
                    </div>
                    <div className="text-right flex-shrink-0">
                      <span className="price text-sm">${(p.price * item.quantity).toLocaleString('en-US', { maximumFractionDigits: 0 })}</span>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="divider" />

            <div className="flex justify-between items-center mb-4">
              <span className="font-display text-xl">TOTAL</span>
              <span className="price text-2xl">${totalPrice.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
            </div>

            {error && <p className="text-sm font-mono-stack mb-3" style={{ color: 'var(--hp-color)' }}>{error}</p>}

            <button
              type="submit"
              disabled={submitting}
              className="btn-primary w-full py-3 text-lg"
              style={{ opacity: submitting ? 0.7 : 1 }}
            >
              {submitting ? 'PROCESANDO...' : 'CONFIRMAR PEDIDO →'}
            </button>

            <p className="text-[10px] font-mono-stack text-center mt-3" style={{ color: 'var(--text-muted)' }}>
              Al confirmar, un asesor se pondrá en contacto contigo para coordinar la entrega.
            </p>

            {/* Removed Items Cache */}
            {removedItems.length > 0 && (
              <div className="mt-8 opacity-60 grayscale hover:opacity-100 hover:grayscale-0 transition-all duration-300">
                <h3 className="font-display text-xl mb-4 text-muted">PRODUCTOS ELIMINADOS</h3>
                <div className="space-y-2">
                  {removedItems.map(item => (
                    <div key={item.product.id} className="card p-3 flex items-center gap-4 bg-ink-surface/30">
                      <div style={{ width: 40 }}>
                        <CardImage imageUrl={item.product.image_url} name={item.product.name} tcg={item.product.tcg} height={50} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold truncate">{item.product.name}</p>
                        <p className="text-[10px] text-muted truncate">{item.product.set_name}</p>
                      </div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          onClick={() => restoreItem(item.product.id)}
                          className="badge cursor-pointer hover:bg-gold hover:text-black transition-colors"
                          style={{ fontSize: '0.7rem' }}
                        >
                          REAGREGAR
                        </button>
                        <button
                          type="button"
                          onClick={() => permanentRemove(item.product.id)}
                          className="p-1 text-hp-color"
                          title="Eliminar permanentemente"
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <polyline points="3 6 5 6 21 6" /><path d="M19 6l-1 14H6L5 6" /><path d="M10 11v6M14 11v6" /><path d="M9 6V4h6v2" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </form>
    </div>
  );
}
