'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect } from 'react';
import { useCart } from '@/lib/CartContext';
import { useUser } from '@/context/UserContext';
import { TCG } from '@/lib/types';
import { fetchTCGs } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';
import { LineIcons } from './ui/LineIcons';
import CartDrawer from './CartDrawer';
import Dropdown from './ui/Dropdown';
import ThemeSelector from './ui/ThemeSelector';

const NavLink = ({ href, icon: Icon, label, colorClass = "", onClick }: { href: string, icon: React.ComponentType<React.SVGProps<SVGSVGElement>>, label: string, colorClass?: string, onClick?: () => void }) => (
  <Link href={href} onClick={onClick} className={`flex flex-col items-center gap-1 no-underline group transition-all text-text-main`}>
    <div className={`p-1 rounded-md group-hover:bg-text-main/5 transition-colors ${colorClass}`}>
      <Icon className="w-5 h-5 stroke-[1.5]" />
    </div>
    <span className="text-[10px] font-bold uppercase tracking-tighter opacity-70 group-hover:opacity-100 transition-opacity">
      {label}
    </span>
  </Link>
);

export default function Navbar() {
  const { totalItems, openCart, isOpen, closeCart } = useCart();
  const { user, loading: userLoading, logout } = useUser();
  const { locale, setLocale, t, availableLocales, hideSelector, getLocaleDisplay } = useLanguage();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [tcgs, setTcgs] = useState<TCG[]>([]);
  const [isPulsing, setIsPulsing] = useState(false);

  useEffect(() => {
    fetchTCGs(true).then(setTcgs);
  }, []);

  useEffect(() => {
    if (totalItems > 0) {
      setIsPulsing(true);
      const timer = setTimeout(() => setIsPulsing(false), 600);
      return () => clearTimeout(timer);
    }
  }, [totalItems]);

  const singlesItems = [
    { label: t('pages.nav.main.view_all', 'View All Singles'), href: '/singles' },
    ...tcgs.map(tcg => ({
      label: `${tcg.name} ${t('pages.nav.main.singles', 'Singles')}`,
      href: `/${tcg.id}/singles`,
    }))
  ];

  const sealedItems = [
    { label: t('pages.nav.main.view_all_sealed', 'View All Sealed'), href: '/sealed' },
    ...tcgs.map(tcg => ({
      label: `${tcg.name} ${t('pages.nav.main.sealed', 'Sealed')}`,
      href: `/${tcg.id}/sealed`,
    }))
  ];

  return (
    <>
      <nav
        id="main-navbar"
        className="sticky top-0 z-sticky bg-bg-page/80 backdrop-blur-md border-b border-border-main/20"
      >
        <div className="w-full px-4 sm:px-6 lg:px-8 flex items-center justify-between h-14 max-w-[var(--content-max-width)] mx-auto">
          {/* Logo Section */}
          <Link href="/" className="flex flex-col no-underline shrink-0 group">
            <span className="font-display text-2xl text-text-main leading-none tracking-tight group-hover:opacity-80 transition-opacity">
              EL BULK
            </span>
          </Link>

          {/* Unified Desktop Nav */}
          <div className="hidden lg:flex items-center gap-4 xl:gap-6">
            <div className="flex items-center gap-4 pr-4 border-r border-border-plum/30">
              <Dropdown
                trigger={<Link href="/singles" className="flex flex-col items-center gap-1 bg-transparent border-none cursor-pointer group text-text-main no-underline">
                  <div className="p-1 rounded-md group-hover:bg-text-main/5 transition-colors">
                    <LineIcons.Singles />
                  </div>
                  <span className="text-[10px] font-bold uppercase tracking-tighter opacity-70 group-hover:opacity-100">{t('pages.nav.main.singles', 'Singles')}</span>
                </Link>}
                items={singlesItems}
              />
              <Dropdown
                trigger={<Link href="/sealed" className="flex flex-col items-center gap-1 bg-transparent border-none cursor-pointer group text-text-main no-underline">
                  <div className="p-1 rounded-md group-hover:bg-text-main/5 transition-colors">
                    <LineIcons.Sealed />
                  </div>
                  <span className="text-[10px] font-bold uppercase tracking-tighter opacity-70 group-hover:opacity-100">{t('pages.nav.main.sealed', 'Sealed')}</span>
                </Link>}
                items={sealedItems}
              />
              <NavLink href="/accessories" icon={LineIcons.Accessories} label={t('pages.nav.main.accessories', 'Accessories')} />
            </div>

            <div className="flex items-center gap-4 px-4 border-r border-border-plum/30">
              <NavLink href="/store-exclusives" icon={LineIcons.Exclusives} label={t('pages.nav.main.store_exclusives', 'Exclusives')} />
              <NavLink href="/notices" icon={LineIcons.News} label={t('pages.nav.main.notices', 'News')} />
              <NavLink href="/contact" icon={LineIcons.Contact} label={t('pages.nav.main.contact', 'Contact')} />
            </div>

            <div className="flex items-center gap-4 pl-4">
              <NavLink href="/bounties" icon={LineIcons.Bounties} label={t('pages.nav.main.wanted', 'Bounties')} colorClass="text-accent-rose" />
            </div>
          </div>

          {/* Utilities */}
          <div className="flex items-center gap-2 sm:gap-4">
            {!userLoading && (
              user ? (
                <Dropdown
                  align="end"
                  trigger={
                    <div className="flex items-center gap-2 cursor-pointer group">
                      <div className="w-8 h-8 relative rounded-full border border-border-main/50 overflow-hidden shadow-sm group-hover:border-text-main transition-colors">
                        <Image
                          src={user.avatar_url || 'https://www.gravatar.com/avatar/?d=mp'}
                          alt={user.first_name || 'User'}
                          fill
                          className="object-cover"
                          referrerPolicy="no-referrer"
                          sizes="32px"
                        />
                      </div>
                    </div>
                  }
                >
                  <div className="p-2 w-48 bg-bg-kraft border border-border-plum rounded-sm shadow-xl">
                    <div className="mb-2 px-2 py-1 border-b border-border-main/10">
                      <p className="text-[9px] font-mono text-text-main uppercase tracking-widest">{user.first_name} {user.last_name}</p>
                      <p className="text-[10px] text-text-secondary truncate font-medium">{user.email}</p>
                    </div>
                    <Link
                      href="/profile"
                      className="block w-full text-left text-sm text-ink-plum hover:bg-ink-plum/5 rounded px-2 py-2 transition-colors no-underline mb-1"
                    >
                      {t('pages.nav.user.profile', 'My Profile')}
                    </Link>
                    <button
                      onClick={logout}
                      className="w-full text-left text-sm text-ink-plum/70 hover:text-accent-rose hover:bg-accent-rose/5 rounded px-2 py-2 transition-colors bg-transparent border-none cursor-pointer"
                    >
                      {t('pages.nav.user.logout', 'Logout')}
                    </button>
                  </div>
                </Dropdown>
              ) : (
                <Link
                  href="/login"
                  className="flex flex-col items-center gap-1 no-underline group text-text-main"
                >
                  <div className="p-1 rounded-md group-hover:bg-text-main/5 transition-colors">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                  </div>
                  <span className="text-[9px] font-bold uppercase opacity-70">{t('pages.nav.user.login', 'Login')}</span>
                </Link>
              )
            )}

            {!hideSelector && (
              <Dropdown
                align="end"
                trigger={
                  <button
                    className="p-1.5 rounded-md transition-all flex flex-col items-center justify-center bg-transparent border border-border-main/30 cursor-pointer text-text-main group hover:border-text-main min-w-[36px]"
                    title={t('pages.nav.tooltips.change_lang', 'Change language')}
                  >
                    <span className="text-[10px] font-bold uppercase tracking-widest">{locale}</span>
                  </button>
                }
              >
                <div className="p-1 min-w-[140px] bg-bg-kraft border border-border-plum rounded-sm shadow-xl flex flex-col gap-0.5">
                  {availableLocales.map(loc => {
                    const item = getLocaleDisplay(loc);
                    
                    return (
                      <button
                        key={loc}
                        onClick={() => setLocale(loc)}
                        className={`flex items-center gap-3 w-full text-left text-xs px-3 py-2 rounded transition-colors bg-transparent border-none cursor-pointer ${
                          locale === loc ? 'text-ink-plum font-bold bg-ink-plum/10' : 'text-ink-plum hover:bg-ink-plum/5'
                        }`}
                      >
                        <span className="text-base leading-none inline-block w-5">{item.icon}</span>
                        <span className="font-medium">{item.label}</span>
                        {locale === loc && <div className="ml-auto w-1.5 h-1.5 rounded-full bg-ink-plum" />}
                      </button>
                    );
                  })}
                </div>
              </Dropdown>
            )}

            <ThemeSelector />

            <button
              id="cart-toggle"
              onClick={isOpen ? closeCart : openCart}
              className={`relative p-2 rounded-md transition-all bg-transparent border border-border-main/30 cursor-pointer text-text-main hover:border-text-main group ${isPulsing ? 'animate-package-pulse border-accent-primary scale-110 shadow-lg' : ''}`}
              aria-label={isOpen ? t('pages.cart.drawer.close', 'Close cart') : t('pages.cart.drawer.title', 'Open cart')}
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z" />
                <line x1="3" y1="6" x2="21" y2="6" />
                <path d="M16 10a4 4 0 01-8 0" />
              </svg>
              {totalItems > 0 && (
                <span className={`absolute -top-1.5 -right-1.5 bg-accent-rose text-white text-[9px] font-bold w-4 h-4 flex items-center justify-center rounded-full shadow-sm transition-transform duration-300 ${isPulsing ? 'scale-125' : ''}`}>
                  {totalItems}
                </span>
              )}
            </button>

            {/* Mobile hamburger */}
            <button
              className="lg:hidden p-2 rounded bg-transparent border-none text-text-main cursor-pointer"
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
          <div className="lg:hidden px-4 py-4 max-h-[calc(100vh-64px)] overflow-y-auto bg-bg-page border-t border-border-main" data-theme-area="mobile-nav">
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
              <NavLink href="/singles" icon={LineIcons.Singles} label={t('pages.nav.main.singles', 'Singles')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/sealed" icon={LineIcons.Sealed} label={t('pages.nav.main.sealed', 'Sealed')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/accessories" icon={LineIcons.Accessories} label={t('pages.nav.main.accessories', 'Accessories')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/store-exclusives" icon={LineIcons.Exclusives} label={t('pages.nav.main.store_exclusives', 'Exclusives')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/notices" icon={LineIcons.News} label={t('pages.nav.main.notices', 'News')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/contact" icon={LineIcons.Contact} label={t('pages.nav.main.contact', 'Contact')} onClick={() => setMobileOpen(false)} />
              <NavLink href="/bounties" icon={LineIcons.Bounties} label={t('pages.nav.main.wanted', 'Bounties')} colorClass="text-accent-rose" onClick={() => setMobileOpen(false)} />
            </div>
          </div>
        )}
      </nav>

      <CartDrawer isOpen={isOpen} onClose={closeCart} />
    </>
  );
}
