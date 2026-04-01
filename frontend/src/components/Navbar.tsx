'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect } from 'react';
import { useCart } from '@/lib/CartContext';
import { useUser } from '@/context/UserContext';
import { TCG } from '@/lib/types';
import { fetchTCGs } from '@/lib/api';
import { useUI } from '@/context/UIContext';
import CartDrawer from './CartDrawer';
import Dropdown from './ui/Dropdown';
import ThemeSelector from './ui/ThemeSelector';

export default function Navbar() {
  const { totalItems, openCart, isOpen, closeCart } = useCart();
  const { user, loading: userLoading, logout } = useUser();
  const { foilEffectsEnabled, toggleFoilEffects } = useUI();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [tcgs, setTcgs] = useState<TCG[]>([]);

  useEffect(() => {
    fetchTCGs(true).then(setTcgs);
  }, []);

  const singlesItems = tcgs.map(tcg => ({
    label: `${tcg.name} Singles`,
    href: `/${tcg.id}/singles`,
  }));

  const sealedItems = tcgs.map(tcg => ({
    label: `${tcg.name} Sealed`,
    href: `/${tcg.id}/sealed`,
  }));

  return (
    <>
      <nav 
        id="main-navbar"
        data-theme-area="header"
        className="sticky top-0 z-40 bg-bg-header border-b border-border-main"
      >
        <div className="w-full px-4 sm:px-6 lg:px-8 flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-1 sm:gap-2 no-underline shrink-0">
            <div className="bg-accent-primary rounded-[4px] px-1.5 py-0.5 sm:px-2 sm:py-1">
              <span className="font-display text-xl sm:text-2xl text-text-on-accent leading-none">
                EL BULK
              </span>
            </div>
            <span className="hidden xs:block text-text-muted text-[10px] sm:text-[11px] font-mono whitespace-nowrap">
              TCG STORE
            </span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden md:flex items-center gap-6" data-theme-area="nav-links">
            <Dropdown 
              label={<Link href="/singles" className="no-underline text-sm font-medium text-white/80 hover:text-accent-primary transition-colors">Singles</Link>}
              items={singlesItems}
            />
            
            <Dropdown 
              label={<Link href="/sealed" className="no-underline text-sm font-medium text-white/80 hover:text-accent-primary transition-colors">Sealed</Link>}
              items={sealedItems}
            />

            <Link href="/accessories" className="no-underline text-sm font-medium text-white/80 hover:text-accent-primary transition-colors">
              Accessories
            </Link>
            <Link href="/store-exclusives" className="no-underline text-sm font-medium text-accent-primary hover:text-white transition-colors">
              Store Exclusives
            </Link>
            <Link href="/notices" className="no-underline text-sm font-medium text-white/80 hover:text-accent-primary transition-colors">
              Notices
            </Link>
            <Link href="/contact" className="no-underline text-sm font-medium text-white/80 hover:text-accent-primary transition-colors">
              Contact
            </Link>
            <Link href="/bounties" className="no-underline text-sm font-medium text-status-hp hover:opacity-80 transition-colors">
              🎯 Wanted Cards
            </Link>
            <Link href="/bulk" className="no-underline text-sm font-medium text-accent-primary hover:opacity-80 transition-colors">
              💰 Sell Your Bulk
            </Link>
          </div>

          {/* Cart + Mobile + User */}
          <div className="flex items-center gap-3">
            {/* User Auth */}
            {!userLoading && (
              <div className="relative group flex items-center">
                {user ? (
                  <div className="flex flex-col items-end group-hover:flex relative">
                    <div className="w-8 h-8 relative rounded-full border border-border-main overflow-hidden cursor-pointer">
                      <Image 
                        src={user.avatar_url || 'https://www.gravatar.com/avatar/?d=mp'} 
                        alt={user.first_name || 'User'}
                        fill
                        className="object-cover"
                        referrerPolicy="no-referrer"
                        sizes="32px"
                      />
                    </div>
                    <div className="absolute top-10 right-0 w-32 bg-bg-surface border border-border-main rounded-md shadow-xl p-2 hidden group-hover:block transition-all z-50">
                      <p className="text-xs text-text-muted truncate mb-2">{user.email}</p>
                      <button 
                        onClick={logout}
                        className="w-full text-left text-sm text-text-secondary hover:text-accent-primary-hover transition-colors bg-transparent border-none cursor-pointer p-0"
                      >
                        Logout
                      </button>
                    </div>
                  </div>
                ) : (
                  <Link 
                    href="/login"
                    className="text-sm font-medium text-white/80 transition-colors hover:text-accent-primary flex items-center gap-1 no-underline"
                  >
                    Login
                  </Link>
                )}
              </div>
            )}

            {/* Foil Toggle */}
            <button
              onClick={toggleFoilEffects}
              className="p-2 rounded-lg transition-all flex items-center justify-center bg-bg-surface border border-border-main cursor-pointer"
              style={{ 
                color: foilEffectsEnabled ? 'var(--accent-primary)' : 'var(--text-muted)',
                opacity: foilEffectsEnabled ? 1 : 0.6,
                transform: foilEffectsEnabled ? 'scale(1.1)' : 'scale(1)',
              }}
              aria-label="Toggle foil effects"
              title={foilEffectsEnabled ? "Disable foil effects" : "Enable foil effects"}
            >
              <span className="text-[1.1rem]">✨</span>
            </button>

            <ThemeSelector />

            <button
              id="cart-toggle"
              onClick={openCart}
              className="relative p-2 rounded-lg transition-colors bg-bg-surface border border-border-main cursor-pointer text-text-main"
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
              className="md:hidden p-2 rounded bg-transparent border-none text-white cursor-pointer"
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
          <div className="md:hidden px-4 py-4 max-h-[calc(100vh-64px)] overflow-y-auto bg-bg-surface border-t border-border-main" data-theme-area="mobile-nav">
            <div className="mb-4">
              <p className="font-mono text-[10px] text-status-hp font-bold mb-2 uppercase tracking-widest">Singles Inventory</p>
              <div className="grid grid-cols-2 gap-2">
                {tcgs.map(tcg => (
                  <Link key={`s-${tcg.id}`} href={`/${tcg.id}/singles`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-bg-page/30 rounded-sm border border-border-main/20 text-text-secondary no-underline">
                    {tcg.name}
                  </Link>
                ))}
              </div>
            </div>
            
            <div className="mb-4">
              <p className="font-mono text-[10px] text-status-hp font-bold mb-2 uppercase tracking-widest">Sealed Product</p>
              <div className="grid grid-cols-2 gap-2">
                {tcgs.map(tcg => (
                  <Link key={`se-${tcg.id}`} href={`/${tcg.id}/sealed`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-bg-page/30 rounded-sm border border-border-main/20 text-text-secondary no-underline">
                    {tcg.name}
                  </Link>
                ))}
              </div>
            </div>
            
            <hr className="border-border-main my-4" />
            <div className="flex flex-col gap-2">
              <Link href="/accessories" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-text-secondary hover:text-accent-primary transition-colors no-underline">
                Accessories
              </Link>
              <Link href="/store-exclusives" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-accent-primary hover:text-accent-primary-hover transition-colors no-underline">
                Store Exclusives
              </Link>
              <Link href="/contact" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-text-secondary hover:text-accent-primary transition-colors no-underline">
                Contact
              </Link>
              <Link href="/bounties" onClick={() => setMobileOpen(false)}
                className="block py-3 text-center border-2 border-status-hp rounded-sm mt-4 font-bold text-status-hp no-underline">
                🎯 WANTED CARDS
              </Link>
              <Link href="/bulk" onClick={() => setMobileOpen(false)}
                className="block py-3 text-center border-2 border-accent-primary rounded-sm mt-2 font-bold text-accent-primary no-underline">
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
