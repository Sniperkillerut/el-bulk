'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useCart } from '@/lib/CartContext';
import { KNOWN_TCGS, TCG_SHORT } from '@/lib/types';
import CartDrawer from './CartDrawer';

export default function Navbar() {
  const { totalItems, openCart, isOpen, closeCart } = useCart();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [tcgDropOpen, setTcgDropOpen] = useState(false);

  return (
    <>
      <nav style={{ background: 'var(--ink-navy)', borderBottom: '1px solid var(--ink-border)' }}
        className="sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2 no-underline">
            <div style={{ background: 'var(--gold)', borderRadius: '4px', padding: '2px 8px' }}>
              <span className="font-display text-2xl" style={{ color: 'var(--ink-deep)', lineHeight: 1 }}>
                EL BULK
              </span>
            </div>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.7rem', fontFamily: 'Space Mono, monospace' }}>
              TCG STORE
            </span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden md:flex items-center gap-6">
            {/* TCG Dropdown */}
            <div className="relative">
              <button
                onClick={() => setTcgDropOpen(o => !o)}
                className="flex items-center gap-1 text-sm font-medium transition-colors"
                style={{ color: 'var(--text-secondary)', background: 'none', border: 'none', cursor: 'pointer' }}
              >
                Singles & Sealed
                <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
                  <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round"/>
                </svg>
              </button>
              {tcgDropOpen && (
                <div
                  className="absolute top-8 left-0 rounded-lg shadow-xl"
                  style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)', minWidth: '180px', zIndex: 50 }}
                  onMouseLeave={() => setTcgDropOpen(false)}
                >
                  {KNOWN_TCGS.map(tcg => (
                    <div key={tcg}>
                      <Link
                        href={`/${tcg}/singles`}
                        onClick={() => setTcgDropOpen(false)}
                        className="block px-4 py-2 text-sm transition-colors"
                        style={{ color: 'var(--text-secondary)' }}
                      >
                        {TCG_SHORT[tcg]} Singles
                      </Link>
                      <Link
                        href={`/${tcg}/sealed`}
                        onClick={() => setTcgDropOpen(false)}
                        className="block px-4 py-2 text-sm transition-colors"
                        style={{ color: 'var(--text-secondary)' }}
                      >
                        {TCG_SHORT[tcg]} Sealed
                      </Link>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <Link href="/accessories" className="text-sm font-medium transition-colors"
              style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
              Accessories
            </Link>
            <Link href="/bulk" className="text-sm font-medium transition-colors"
              style={{ color: 'var(--gold)', textDecoration: 'none' }}>
              💰 Sell Your Bulk
            </Link>
          </div>

          {/* Cart + Mobile */}
          <div className="flex items-center gap-3">
            <button
              id="cart-toggle"
              onClick={openCart}
              className="relative p-2 rounded-lg transition-colors"
              style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)', cursor: 'pointer', color: 'var(--text-primary)' }}
              aria-label="Open cart"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z"/>
                <line x1="3" y1="6" x2="21" y2="6"/>
                <path d="M16 10a4 4 0 01-8 0"/>
              </svg>
              {totalItems > 0 && (
                <span className="cart-badge">{totalItems}</span>
              )}
            </button>

            {/* Mobile hamburger */}
            <button
              className="md:hidden p-2 rounded"
              style={{ background: 'none', border: 'none', color: 'var(--text-primary)', cursor: 'pointer' }}
              onClick={() => setMobileOpen(o => !o)}
              aria-label="Toggle menu"
            >
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                {mobileOpen
                  ? <><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></>
                  : <><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></>
                }
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile menu */}
        {mobileOpen && (
          <div style={{ background: 'var(--ink-surface)', borderTop: '1px solid var(--ink-border)' }} className="md:hidden px-4 py-4">
            <div className="font-display text-sm mb-2" style={{ color: 'var(--text-muted)' }}>SINGLES & SEALED</div>
            {KNOWN_TCGS.map(tcg => (
              <div key={tcg} className="mb-1">
                <Link href={`/${tcg}/singles`} onClick={() => setMobileOpen(false)}
                  className="block py-1 text-sm" style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                  {TCG_SHORT[tcg]} Singles
                </Link>
                <Link href={`/${tcg}/sealed`} onClick={() => setMobileOpen(false)}
                  className="block py-1 text-sm" style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                  {TCG_SHORT[tcg]} Sealed
                </Link>
              </div>
            ))}
            <hr style={{ borderColor: 'var(--ink-border)', margin: '0.75rem 0' }} />
            <Link href="/accessories" onClick={() => setMobileOpen(false)}
              className="block py-2 text-sm" style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
              Accessories
            </Link>
            <Link href="/bulk" onClick={() => setMobileOpen(false)}
              className="block py-2 text-sm font-semibold" style={{ color: 'var(--gold)', textDecoration: 'none' }}>
              💰 Sell Your Bulk
            </Link>
          </div>
        )}
      </nav>

      <CartDrawer isOpen={isOpen} onClose={closeCart} />
    </>
  );
}
