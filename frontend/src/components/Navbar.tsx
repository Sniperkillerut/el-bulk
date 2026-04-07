'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect } from 'react';
import { useCart } from '@/lib/CartContext';
import { useUser } from '@/context/UserContext';
import { TCG } from '@/lib/types';
import { fetchTCGs } from '@/lib/api';
import { useUI } from '@/context/UIContext';
import { useLanguage } from '@/context/LanguageContext';
import CartDrawer from './CartDrawer';
import Dropdown from './ui/Dropdown';
import ThemeSelector from './ui/ThemeSelector';

export default function Navbar() {
  const { totalItems, openCart, isOpen, closeCart } = useCart();
  const { user, loading: userLoading, logout } = useUser();
  const { foilEffectsEnabled, toggleFoilEffects } = useUI();
  const { locale, setLocale, t, availableLocales } = useLanguage();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [tcgs, setTcgs] = useState<TCG[]>([]);

  useEffect(() => {
    fetchTCGs(true).then(setTcgs);
  }, []);

  const singlesItems = tcgs.map(tcg => ({
    label: `${tcg.name} ${t('pages.nav.main.singles', 'Singles')}`,
    href: `/${tcg.id}/singles`,
  }));

  const sealedItems = tcgs.map(tcg => ({
    label: `${tcg.name} ${t('pages.nav.main.sealed', 'Sealed')}`,
    href: `/${tcg.id}/sealed`,
  }));

  return (
    <>
      <nav
        id="main-navbar"
        data-theme-area="header"
        className="sticky top-0 z-[var(--z-sticky)] bg-bg-header border-b border-border-main"
      >
        <div
          className="w-full px-4 sm:px-6 lg:px-8 flex items-center justify-between h-16 max-w-[var(--content-max-width)] mx-auto"
        >
          {/* Logo */}
          <Link href="/" className="flex items-center gap-1 sm:gap-2 no-underline shrink-0">
            <div className="bg-accent-primary rounded-[4px] px-1.5 py-0.5 sm:px-2 sm:py-1">
              <span className="font-display text-xl sm:text-2xl text-text-on-accent leading-none">
                EL BULK
              </span>
            </div>
            <span className="hidden xs:block text-text-muted text-[10px] sm:text-[11px] font-mono whitespace-nowrap">
              {t('pages.nav.main.tcg_store', 'TCG STORE')}
            </span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden md:flex items-center gap-6" data-theme-area="nav-links">
            <Dropdown
              label={<Link href="/singles" className="no-underline text-sm font-medium text-text-on-header/80 hover:text-accent-primary transition-colors">{t('pages.nav.main.singles', 'Singles')}</Link>}
              items={singlesItems}
            />

            <Dropdown
              label={<Link href="/sealed" className="no-underline text-sm font-medium text-text-on-header/80 hover:text-accent-primary transition-colors">{t('pages.nav.main.sealed', 'Sealed')}</Link>}
              items={sealedItems}
            />

            <Link href="/accessories" className="no-underline text-sm font-medium text-text-on-header/80 hover:text-accent-header transition-colors">
              {t('pages.nav.main.accessories', 'Accessories')}
            </Link>
            <Link href="/store-exclusives" className="no-underline text-sm font-medium text-accent-header hover:text-text-on-header transition-colors">
              {t('pages.nav.main.store_exclusives', 'Store Exclusives')}
            </Link>
            <Link href="/notices" className="no-underline text-sm font-medium text-text-on-header/80 hover:text-accent-primary transition-colors">
              {t('pages.nav.main.notices', 'Notices')}
            </Link>
            <Link href="/contact" className="no-underline text-sm font-medium text-text-on-header/80 hover:text-accent-primary transition-colors">
              {t('pages.nav.main.contact', 'Contact')}
            </Link>
            <Link href="/bounties" className="no-underline text-sm font-medium text-status-hp-header hover:opacity-80 transition-colors">
              🎯 {t('pages.nav.main.wanted', 'Wanted Cards')}
            </Link>
            <Link href="/bulk" className="no-underline text-sm font-medium text-accent-header hover:opacity-80 transition-colors">
              💰 {t('pages.nav.main.bulk', 'Sell Your Bulk')}
            </Link>
          </div>

          {/* Cart + Mobile + User */}
          <div className="flex items-center gap-3">
            {/* User Auth */}
            {!userLoading && (
              user ? (
                <Dropdown
                  align="end"
                  trigger={
                    <div className="w-8 h-8 relative rounded-full border border-border-main overflow-hidden cursor-pointer shadow-md hover:border-accent-primary transition-colors">
                      <Image
                        src={user.avatar_url || 'https://www.gravatar.com/avatar/?d=mp'}
                        alt={user.first_name || 'User'}
                        fill
                        className="object-cover"
                        referrerPolicy="no-referrer"
                        sizes="32px"
                      />
                    </div>
                  }
                >
                  <div className="p-2 w-48">
                    <div className="mb-2 px-2 py-1 border-b border-border-main/30">
                      <p className="text-[9px] font-mono text-text-muted uppercase tracking-widest">{user.first_name} {user.last_name}</p>
                      <p className="text-[10px] text-accent-primary truncate font-medium">{user.email}</p>
                    </div>
                    <Link
                      href="/profile"
                      className="block w-full text-left text-sm text-text-main hover:text-accent-primary hover:bg-accent-primary/5 rounded px-2 py-2 transition-colors no-underline mb-1"
                    >
                      {t('pages.nav.user.profile', 'My Profile')}
                    </Link>
                    <button
                      onClick={logout}
                      className="w-full text-left text-sm text-text-secondary hover:text-red-400 hover:bg-red-400/5 rounded px-2 py-2 transition-colors bg-transparent border-none cursor-pointer"
                    >
                      {t('pages.nav.user.logout', 'Logout')}
                    </button>
                  </div>
                </Dropdown>
              ) : (
                <Link
                  href="/login"
                  className="text-sm font-medium text-text-on-header/80 transition-colors hover:text-accent-primary flex items-center gap-1 no-underline"
                >
                  {t('pages.nav.user.login', 'Login')}
                </Link>
              )
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
              aria-label={t('pages.nav.tooltips.foil_toggle', 'Toggle foil effects')}
              title={foilEffectsEnabled
                ? t('pages.nav.tooltips.foil_disable', 'Disable foil effects')
                : t('pages.nav.tooltips.foil_enable', 'Enable foil effects')}
            >
              <span className="text-[1.1rem]">✨</span>
            </button>

            <Dropdown
              align="end"
              trigger={
                <button
                  className="p-2 rounded-lg transition-all flex items-center justify-center bg-bg-surface border border-border-main cursor-pointer"
                  title={t('pages.nav.tooltips.change_lang', 'Change language')}
                >
                  <span className="text-xs font-bold uppercase">{locale}</span>
                </button>
              }
            >
              <div className="p-1 w-32">
                {availableLocales.map(loc => (
                  <button
                    key={loc}
                    onClick={() => setLocale(loc)}
                    className={`block w-full text-left text-xs px-3 py-2 rounded transition-colors bg-transparent border-none cursor-pointer ${locale === loc ? 'text-accent-primary font-bold bg-accent-primary/10' : 'text-text-main hover:bg-bg-page'
                      }`}
                  >
                    {loc === 'en' ? '🇺🇸 English' : loc === 'es' ? '🇪🇸 Español' : loc.toUpperCase()}
                  </button>
                ))}
              </div>
            </Dropdown>

            <ThemeSelector />

            <button
              id="cart-toggle"
              onClick={isOpen ? closeCart : openCart}
              className="relative p-2 rounded-lg transition-colors bg-bg-surface border border-border-main cursor-pointer text-text-main"
              aria-label={isOpen ? t('pages.cart.drawer.close', 'Close cart') : t('pages.cart.drawer.title', 'Open cart')}
              title={isOpen ? t('pages.cart.drawer.close', 'Close cart') : t('pages.cart.drawer.title', 'Open cart')}
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z" />
                <line x1="3" y1="6" x2="21" y2="6" />
                <path d="M16 10a4 4 0 01-8 0" />
              </svg>
              {totalItems > 0 && (
                <span className="cart-badge">{totalItems}</span>
              )}
            </button>

            {/* Mobile hamburger */}
            <button
              className="md:hidden p-2 rounded bg-transparent border-none text-text-on-header cursor-pointer"
              onClick={() => setMobileOpen(o => !o)}
              aria-label={t('pages.nav.mobile.toggle_menu', 'Toggle menu')}
            >
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                {mobileOpen
                  ? <><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></>
                  : <><line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="12" x2="21" y2="12" /><line x1="3" y1="18" x2="21" y2="18" /></>
                }
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile menu */}
        {mobileOpen && (
          <div className="md:hidden px-4 py-4 max-h-[calc(100vh-64px)] overflow-y-auto bg-bg-surface border-t border-border-main" data-theme-area="mobile-nav">
            <div className="mb-4">
              <p className="font-mono text-[10px] text-status-hp font-bold mb-2 uppercase tracking-widest">{t('pages.nav.mobile.inventory_title', 'Singles Inventory')}</p>
              <div className="grid grid-cols-2 gap-2">
                {tcgs.map(tcg => (
                  <Link key={`s-${tcg.id}`} href={`/${tcg.id}/singles`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-bg-page/30 rounded-sm border border-border-main/20 text-text-secondary no-underline">
                    {tcg.name} {t('pages.nav.main.singles', 'Singles')}
                  </Link>
                ))}
              </div>
            </div>

            <div className="mb-4">
              <p className="font-mono text-[10px] text-status-hp font-bold mb-2 uppercase tracking-widest">{t('pages.nav.mobile.sealed_title', 'Sealed Product')}</p>
              <div className="grid grid-cols-2 gap-2">
                {tcgs.map(tcg => (
                  <Link key={`se-${tcg.id}`} href={`/${tcg.id}/sealed`} onClick={() => setMobileOpen(false)}
                    className="block py-2 px-3 text-xs bg-bg-page/30 rounded-sm border border-border-main/20 text-text-secondary no-underline">
                    {tcg.name} {t('pages.nav.main.sealed', 'Sealed')}
                  </Link>
                ))}
              </div>
            </div>

            <hr className="border-border-main my-4" />
            <div className="flex flex-col gap-2">
              <Link href="/accessories" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-text-secondary hover:text-accent-header transition-colors no-underline">
                {t('pages.nav.main.accessories', 'Accessories')}
              </Link>
              <Link href="/store-exclusives" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-accent-header hover:text-accent-header transition-colors no-underline">
                {t('pages.nav.main.store_exclusives', 'Store Exclusives')}
              </Link>
              <Link href="/contact" onClick={() => setMobileOpen(false)}
                className="block py-2 text-sm font-medium text-text-secondary hover:text-accent-header transition-colors no-underline">
                {t('pages.nav.main.contact', 'Contact')}
              </Link>
              <Link href="/bounties" onClick={() => setMobileOpen(false)}
                className="block py-3 text-center border-2 border-status-hp-header rounded-sm mt-4 font-bold text-status-hp-header no-underline">
                🎯 {t('pages.nav.main.wanted', 'WANTED CARDS')}
              </Link>
              <Link href="/bulk" onClick={() => setMobileOpen(false)}
                className="block py-3 text-center border-2 border-accent-header rounded-sm mt-2 font-bold text-accent-header no-underline">
                💰 {t('pages.nav.main.bulk', 'SELL YOUR BULK')}
              </Link>
            </div>
          </div>
        )}
      </nav>

      <CartDrawer isOpen={isOpen} onClose={closeCart} />
    </>
  );
}
