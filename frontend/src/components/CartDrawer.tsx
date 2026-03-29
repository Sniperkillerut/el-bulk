'use client';

import { useCart } from '@/lib/CartContext';
import Link from 'next/link';
import CardImage from './CardImage';

export default function CartDrawer({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) {
  const { items, removedItems, totalItems, totalPrice, removeItem, restoreItem, permanentRemove, updateQty } = useCart();

  return (
    <>
      {/* Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-50"
          style={{ background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(2px)' }}
          onClick={onClose}
        />
      )}

      {/* Drawer */}
      <div
        className="fixed right-0 top-0 h-full z-50 flex flex-col"
        style={{
          width: 'min(400px, 100vw)',
          background: 'var(--ink-surface)',
          borderLeft: '1px solid var(--ink-border)',
          transform: isOpen ? 'translateX(0)' : 'translateX(100%)',
          transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-5" style={{ borderBottom: '1px solid var(--ink-border)' }}>
          <div>
            <h2 className="font-display text-2xl text-gold">YOUR CART</h2>
            <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
              {totalItems} item{totalItems !== 1 ? 's' : ''}
            </p>
          </div>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', padding: '4px' }}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        {/* Items */}
        <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-3">
          {items.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full gap-4" style={{ color: 'var(--text-muted)' }}>
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" opacity="0.4">
                <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 01-8 0"/>
              </svg>
              <p className="text-center text-sm">Your cart is empty.<br />Go find some cards.</p>
            </div>
          ) : (
            items.map(item => (
              <div key={item.product.id} className="card p-3 flex gap-3">
                {/* Image */}
                <div style={{ width: 48, flexShrink: 0 }}>
                  <CardImage imageUrl={item.product.image_url} name={item.product.name} tcg={item.product.tcg} height={64} enableHover={true} enableModal={true} />
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate" style={{ color: 'var(--text-primary)' }}>
                    {item.product.name}
                  </p>
                  {item.product.set_name && (
                    <p className="text-xs truncate" style={{ color: 'var(--text-muted)' }}>{item.product.set_name}</p>
                  )}
                    <div className="flex items-center gap-2 mt-2">
                      <button
                        onClick={() => updateQty(item.product.id, item.quantity - 1)}
                        style={{ width: 24, height: 24, background: 'var(--ink-border)', border: 'none', borderRadius: 3, color: 'var(--text-primary)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                      >−</button>
                      <span className="text-sm w-6 text-center" style={{ fontFamily: 'Space Mono, monospace' }}>{item.quantity}</span>
                      <button
                        onClick={() => updateQty(item.product.id, item.quantity + 1)}
                        disabled={item.quantity >= item.product.stock}
                        style={{ 
                          width: 24, height: 24, 
                          background: 'var(--ink-border)', border: 'none', borderRadius: 3, 
                          color: 'var(--text-primary)', 
                          cursor: item.quantity >= item.product.stock ? 'not-allowed' : 'pointer', 
                          opacity: item.quantity >= item.product.stock ? 0.5 : 1,
                          display: 'flex', alignItems: 'center', justifyContent: 'center' 
                        }}
                      >+</button>
                      <span className="text-[10px] opacity-40 ml-1" style={{ fontFamily: 'Space Mono, monospace' }}>
                        {item.product.stock} dispon.
                      </span>
                    </div>
                </div>

                {/* Price + remove */}
                <div className="flex flex-col items-end justify-between">
                  <span className="price text-sm">${(item.product.price * item.quantity).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
                  <button
                    onClick={() => removeItem(item.product.id)}
                    style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', padding: 2 }}
                    title="Remove"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/>
                    </svg>
                  </button>
                </div>
              </div>
            ))
          )}

          {/* Removed Items Cache */}
          {removedItems.length > 0 && (
            <div className="mt-6 pt-6" style={{ borderTop: '1px dashed var(--ink-border)' }}>
              <h3 className="text-[10px] font-mono-stack mb-3 tracking-widest text-muted uppercase">Removed Items</h3>
              <div className="flex flex-col gap-2">
                {removedItems.map(item => (
                  <div key={item.product.id} className="flex items-center gap-3 opacity-50 hover:opacity-100 transition-opacity p-1">
                    <div style={{ width: 32 }}>
                      <CardImage imageUrl={item.product.image_url} name={item.product.name} tcg={item.product.tcg} height={40} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-xs font-semibold truncate text-primary">{item.product.name}</p>
                    </div>
                    <div className="flex items-center gap-1">
                      <button 
                        onClick={() => restoreItem(item.product.id)}
                        className="text-[9px] font-mono-stack px-2 py-1 bg-ink-border rounded hover:bg-gold hover:text-black transition-colors uppercase"
                      >
                        Restore
                      </button>
                      <button 
                        onClick={() => permanentRemove(item.product.id)}
                        className="p-1 text-hp-color"
                        title="Delete permanently"
                      >
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        {items.length > 0 && (
          <div className="p-5" style={{ borderTop: '1px solid var(--ink-border)' }}>
            <div className="flex justify-between items-center mb-4">
              <span style={{ fontFamily: 'Bebas Neue, sans-serif', letterSpacing: '0.05em', fontSize: '1.1rem' }}>TOTAL</span>
               <span className="price text-xl">${totalPrice.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP</span>
            </div>
            <div style={{ background: 'var(--ink-card)', border: '1px dashed var(--ink-border)', borderRadius: 6, padding: '0.75rem 1rem', marginBottom: '0.75rem' }}>
              <p className="text-xs text-center" style={{ color: 'var(--text-muted)' }}>
                📦 Revisa tu orden y completa los datos de envío.
              </p>
            </div>
            <Link href="/checkout" onClick={onClose} className="btn-primary text-center w-full block">
              PROCEDER AL CHECKOUT →
            </Link>
          </div>
        )}
      </div>
    </>
  );
}
