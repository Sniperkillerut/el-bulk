'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useCart } from '@/lib/CartContext';
import { KNOWN_TCGS, TCG_SHORT } from '@/lib/types';
import CartDrawer from './CartDrawer';

export default function Navbar() {
  const { totalItems, openCart, isOpen, closeCart } = useCart();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [singlesDropOpen, setSinglesDropOpen] = useState(false);
  const [sealedDropOpen, setSealedDropOpen] = useState(false);

  return (
    <>
      <nav style={{ background: 'var(--ink-navy)', borderBottom: '1px solid var(--ink-border)' }}
        className="sticky top-0 z-40">
        <div className="centered-container px-4 flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-1 sm:gap-2 no-underline shrink-0">
            <div style={{ background: 'var(--gold)', borderRadius: '4px', padding: '2px 6px sm:8px' }}>
              <span className="font-display text-xl sm:text-2xl" style={{ color: 'var(--ink-deep)', lineHeight: 1 }}>
                EL BULK
              </span>
            </div>
            <span className="hidden xs:block" style={{ color: 'var(--text-muted)', fontSize: '0.6rem sm:0.7rem', fontFamily: 'Space Mono, monospace' }}>
              TCG STORE
            </span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden md:flex items-center gap-6">
            {/* Singles Dropdown */}
            <div className="relative" onMouseLeave={() => setSinglesDropOpen(false)}>
              <Link
                href="/singles"
                onMouseEnter={() => setSinglesDropOpen(true)}
                className="flex items-center gap-1 text-sm font-medium transition-colors hover:text-gold-dark"
                style={{ color: 'var(--text-secondary)', textDecoration: 'none', background: 'none', border: 'none', cursor: 'pointer' }}
              >
                Singles
                <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
                  <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round"/>
                </svg>
              </Link>
              {singlesDropOpen && (
                <div
                  className="absolute top-full left-0 pt-1"
                  style={{ minWidth: '180px', zIndex: 50 }}
                >
                  <div className="rounded-sm shadow-xl" 
                    style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)' }}>
                    {KNOWN_TCGS.map(tcg => (
                      <Link
                        key={tcg}
                        href={`/${tcg}/singles`}
                        onClick={() => setSinglesDropOpen(false)}
                        className="block px-4 py-2 text-sm transition-colors hover:bg-neutral-100"
                        style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}
                      >
                        {TCG_SHORT[tcg]} Singles
                      </Link>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Sealed Dropdown */}
            <div className="relative" onMouseLeave={() => setSealedDropOpen(false)}>
              <Link
                href="/sealed"
                onMouseEnter={() => setSealedDropOpen(true)}
                className="flex items-center gap-1 text-sm font-medium transition-colors hover:text-gold-dark"
                style={{ color: 'var(--text-secondary)', textDecoration: 'none', background: 'none', border: 'none', cursor: 'pointer' }}
              >
                Sealed
                <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
                  <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round"/>
                </svg>
              </Link>
              {sealedDropOpen && (
                <div
                  className="absolute top-full left-0 pt-1"
                  style={{ minWidth: '180px', zIndex: 50 }}
                >
                  <div className="rounded-sm shadow-xl" 
                    style={{ background: 'var(--ink-surface)', border: '1px solid var(--ink-border)' }}>
                    {KNOWN_TCGS.map(tcg => (
                      <Link
                        key={tcg}
                        href={`/${tcg}/sealed`}
                        onClick={() => setSealedDropOpen(false)}
                        className="block px-4 py-2 text-sm transition-colors hover:bg-neutral-100"
                        style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}
                      >
                        {TCG_SHORT[tcg]} Sealed
                      </Link>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <Link href="/accessories" className="text-sm font-medium transition-colors hover:text-gold-dark"
              style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
              Accessories
            </Link>
            <Link href="/contact" className="text-sm font-medium transition-colors hover:text-gold-dark"
              style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
              Contact
            </Link>
            <Link href="/bulk" className="text-sm font-medium transition-colors hover:opacity-80"
              style={{ color: 'var(--gold-dark)', textDecoration: 'none' }}>
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
          <div style={{ background: 'var(--ink-surface)', borderTop: '1px solid var(--ink-border)' }} 
            className="md:hidden px-4 py-4 max-h-[calc(100vh-64px)] overflow-y-auto">
            <div className="mb-4">
              <p className="font-mono-stack text-[10px] text-hp-color font-bold mb-2 uppercase tracking-widest">Singles Inventory</p>
              <div className="grid grid-cols-2 gap-2">
                {KNOWN_TCGS.map(tcg => (
                  <Link key={`s-${tcg}`} href={`/${tcg}/singles`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-kraft-light/30 rounded-sm border border-kraft-mid/20" 
                    style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                    {TCG_SHORT[tcg]}
                  </Link>
                ))}
              </div>
            </div>
            
            <div className="mb-4">
              <p className="font-mono-stack text-[10px] text-hp-color font-bold mb-2 uppercase tracking-widest">Sealed Product</p>
              <div className="grid grid-cols-2 gap-2">
                {KNOWN_TCGS.map(tcg => (
                  <Link key={`se-${tcg}`} href={`/${tcg}/sealed`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-kraft-light/30 rounded-sm border border-kraft-mid/20" 
                    style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                    {TCG_SHORT[tcg]}
                  </Link>
                ))}
              </div>
            </div>
            
            <hr style={{ borderColor: 'var(--ink-border)', margin: '1rem 0' }} />
            <div className="flex flex-col gap-2">
              <Link href="/accessories" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium" style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                Accessories
              </Link>
              <Link href="/contact" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium" style={{ color: 'var(--text-secondary)', textDecoration: 'none' }}>
                Contact
              </Link>
              <Link href="/bulk" onClick={() => setMobileOpen(false)}
                className="block py-3 text-center border-2 border-gold-dark rounded-sm mt-2 font-bold" 
                style={{ color: 'var(--gold-dark)', textDecoration: 'none' }}>
                💰 SELL YOUR BULK
              </Link>
            </div>
          </div>
        )}
      </nav>

      <CartDrawer isOpen={isOpen} onClose={closeCart} />
    </>
  );
}
