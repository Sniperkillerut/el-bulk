'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useCart } from '@/lib/CartContext';
import { useUser } from '@/context/UserContext';
import { createOrder, fetchPublicSettings } from '@/lib/api';
import { PAYMENT_METHODS, FOIL_LABELS, TREATMENT_LABELS, PublicSettings, CartItem } from '@/lib/types';
import CardImage from '@/components/CardImage';
import { useLanguage } from '@/context/LanguageContext';

const validateOrder = (
  items: CartItem[],
  t: (key: string, defaultMessage: string) => string,
  setError: (msg: string) => void
): boolean => {
  if (items.length === 0) {
    setError(t('pages.checkout.summary.empty_cart', 'Your cart is empty.'));
    return false;
  }
  return true;
};

type CheckoutForm = {
  first_name: string;
  last_name: string;
  phone: string;
  email: string;
  id_number: string;
  address: string;
  payment_method: string;
  is_local_pickup: boolean;
  notes: string;
};

const processOrder = async (
  form: CheckoutForm,
  items: CartItem[],
  router: ReturnType<typeof useRouter>,
  clearCart: () => void,
  t: (key: string, defaultMessage: string) => string,
  setSubmitting: (val: boolean) => void,
  setError: (msg: string) => void
) => {
  setSubmitting(true);
  setError('');
  try {
    const result = await createOrder({
      ...form,
      items: items.map((i: CartItem) => ({ product_id: i.product.id, quantity: i.quantity })),
    });
    clearCart();
    router.push(`/order/${result.order_number}`);
  } catch (e: unknown) {
    setError(e instanceof Error ? e.message : t('pages.checkout.error.create_failed', 'Error creating the order.'));
  } finally {
    setSubmitting(false);
  }
};

export default function CheckoutPage() {
  const router = useRouter();
  const { items, removedItems, totalPrice, updateQty, removeItem, restoreItem, permanentRemove, clearCart } = useCart();
  const { user, loading, loginWithGoogle } = useUser();
  const { t } = useLanguage();
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
    is_local_pickup: false,
    notes: '',
  });

  const [settings, setSettings] = useState<PublicSettings | undefined>();

  useEffect(() => {
    fetchPublicSettings().then(setSettings).catch(console.error);
  }, []);

  useEffect(() => {
    if (user) {
      setForm(f => ({
        ...f,
        first_name: f.first_name || user.first_name || '',
        last_name: f.last_name || user.last_name || '',
        email: f.email || user.email || '',
        phone: f.phone || user.phone || '',
        id_number: f.id_number || user.id_number || '',
        address: f.address || user.address || '',
      }));
    }
  }, [user]);

  const set = (key: string, val: string) => setForm(f => ({ ...f, [key]: val }));

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    if (!validateOrder(items, t, setError)) {
      return;
    }

    await processOrder(form, items, router, clearCart, t, setSubmitting, setError);
  };

  if (loading) {
    return (
      <div className="centered-container py-32 text-center flex flex-col items-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent-primary mb-6"></div>
        <p className="text-text-muted font-mono text-xs tracking-widest uppercase animate-pulse">
          {t('pages.checkout.loading_auth', 'Verifying secure session...')}
        </p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="centered-container px-4 py-20 text-center flex flex-col items-center max-w-2xl">
        <div className="w-20 h-20 bg-accent-primary/10 rounded-full flex items-center justify-center text-4xl mb-8 border border-accent-primary/20 shadow-lg shadow-accent-primary/5">
          🔒
        </div>
        <h1 className="font-display text-5xl mb-6">{t('pages.checkout.auth_required.title', 'SIGN IN REQUIRED')}</h1>
        <p className="text-text-muted mb-10 text-lg leading-relaxed max-w-md">
          {t('pages.checkout.auth_required.desc', 'For your security and to track your orders, you must be logged in to finalize your purchase.')}
        </p>
        <div className="flex flex-col sm:flex-row gap-4 w-full">
          <button 
            onClick={loginWithGoogle} 
            className="flex-1 py-4 bg-accent-primary hover:bg-accent-primary-hover text-text-on-accent font-bold rounded-lg transition-all shadow-xl shadow-accent-primary/20 active:scale-95"
          >
             {t('pages.auth.login.google', 'Login with Google')}
          </button>
          {/* TODO: Facebook login hidden until OAuth is functional
          <button 
            onClick={loginWithFacebook} 
            className="flex-1 py-4 bg-[#1877F2]/10 hover:bg-[#1877F2]/20 text-[#1877F2] border border-[#1877F2]/30 font-bold rounded-lg transition-all active:scale-95"
          >
             {t('pages.auth.login.facebook', 'Login with Facebook')}
          </button>
          */}
        </div>
        <button 
          onClick={() => router.push('/')} 
          className="mt-12 text-[10px] font-mono text-text-muted hover:text-accent-primary uppercase tracking-[0.2em] transition-all border-b border-transparent hover:border-accent-primary pb-1"
        >
          {t('pages.checkout.buttons.cancel_back', '← Return to Armory')}
        </button>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="centered-container px-4 py-16 text-center">
        <h1 className="font-display text-5xl mb-4">{t('pages.checkout.page.title_short', 'CHECKOUT')}</h1>
        <p style={{ color: 'var(--text-muted)', fontSize: '1.1rem' }}>
          {t('pages.checkout.error.empty', 'Your cart is empty. Add products before continuing.')}
        </p>
        <button onClick={() => router.push('/')} className="btn-primary mt-6">
          {t('pages.checkout.buttons.back', '← BACK TO STORE')}
        </button>
      </div>
    );
  }

  return (
    <div className="centered-container px-4 py-8">
      <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>EL BULK / {t('pages.checkout.page.title_short', 'CHECKOUT')}</p>
      <h1 className="font-display text-5xl mb-2">{t('pages.checkout.page.title', 'FINALIZE PURCHASE')}</h1>
      <div className="gold-line mb-8" />

      <form onSubmit={handleSubmit} className="flex flex-col lg:flex-row gap-8">
        {/* Left: Customer Form */}
        <div className="flex-1">
          <div className="card p-6">
            <h2 className="font-display text-2xl mb-4">{t('pages.checkout.section.contact', 'CONTACT INFORMATION')}</h2>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.first_name', 'FIRST NAME')} *</label>
                <input type="text" value={form.first_name} onChange={e => set('first_name', e.target.value)} placeholder={t('pages.checkout.placeholders.first_name', 'Juan')} required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.last_name', 'LAST NAME')} *</label>
                <input type="text" value={form.last_name} onChange={e => set('last_name', e.target.value)} placeholder={t('pages.checkout.placeholders.last_name', 'Perez')} required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.phone', 'PHONE / WHATSAPP')} *</label>
                <input type="tel" value={form.phone} onChange={e => set('phone', e.target.value)} placeholder={t('pages.checkout.placeholders.phone', '3001234567')} required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.email', 'EMAIL')} *</label>
                <input type="email" value={form.email} onChange={e => set('email', e.target.value)} placeholder={t('pages.checkout.placeholders.email', 'correo@ejemplo.com')} required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.id_number', 'ID NUMBER / CEDULA')} *</label>
                <input type="number" value={form.id_number} onChange={e => set('id_number', e.target.value)} placeholder={t('pages.checkout.placeholders.id', '1234567890')} required />
              </div>
              <div>
                <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.address', 'ADDRESS')} *</label>
                <input type="text" value={form.address} onChange={e => set('address', e.target.value)} placeholder={t('pages.checkout.placeholders.address', 'Cra 1 # 2-3')} required />
              </div>
            </div>

            <div className="divider" />

            <h3 className="font-display text-xl mb-3">{t('pages.checkout.section.delivery', 'DELIVERY METHOD')}</h3>
            <div className="flex gap-4 mb-6">
              <button
                type="button"
                onClick={() => setForm(f => ({ ...f, is_local_pickup: false }))}
                className={`flex-1 p-4 rounded-lg border-2 transition-all text-left ${!form.is_local_pickup ? 'border-accent-primary bg-accent-primary/5' : 'border-ink-border bg-ink-surface'}`}
              >
                <div className="flex items-center gap-3">
                  <span className="text-2xl">🚚</span>
                  <div>
                    <p className="font-bold text-sm">{t('pages.checkout.delivery.shipping', 'SHIPPING')}</p>
                    <p className="text-[10px] text-text-muted">{t('pages.checkout.delivery.shipping_desc', 'Reliable delivery to your address')}</p>
                  </div>
                </div>
              </button>
              <button
                type="button"
                onClick={() => setForm(f => ({ ...f, is_local_pickup: true }))}
                className={`flex-1 p-4 rounded-lg border-2 transition-all text-left ${form.is_local_pickup ? 'border-accent-primary bg-accent-primary/5' : 'border-ink-border bg-ink-surface'}`}
              >
                <div className="flex items-center gap-3">
                  <span className="text-2xl">⚓</span>
                  <div>
                    <p className="font-bold text-sm">{t('pages.checkout.delivery.pickup', 'LOCAL PICKUP')}</p>
                    <p className="text-[10px] text-text-muted">{t('pages.checkout.delivery.pickup_desc', 'Pick up at our store in Bogotá')}</p>
                  </div>
                </div>
              </button>
            </div>

            <div className="divider" />

            <h3 className="font-display text-xl mb-3">{t('pages.checkout.section.payment', 'PAYMENT METHOD')}</h3>
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
                  {t(`pages.checkout.payment_methods.${key}`, label)}
                </button>
              ))}
            </div>

            <div className="mt-4">
              <label className="text-xs font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('pages.checkout.form.notes', 'NOTES (OPTIONAL)')}</label>
              <textarea value={form.notes} onChange={e => set('notes', e.target.value)} rows={2} placeholder={t('pages.checkout.form.notes_placeholder', 'Special instructions...')} />
            </div>
          </div>
        </div>

        {/* Right: Order Summary */}
        <div className="w-full lg:w-[420px] flex-shrink-0">
          <div className="card p-6 sticky top-4">
            <div className="flex justify-between items-center mb-4">
              <h2 className="font-display text-2xl">{t('pages.checkout.section.summary', 'ORDER SUMMARY')}</h2>
              <button
                type="button"
                onClick={() => {
                  if (window.confirm(t('pages.checkout.summary.vaciar_confirm', 'Are you sure you want to empty your cart?'))) {
                    clearCart();
                  }
                }}
                className="flex items-center gap-1 text-[10px] font-mono-stack hover:opacity-80 transition-all uppercase tracking-wider"
                style={{ color: 'var(--hp-color)', background: 'none', border: 'none', cursor: 'pointer' }}
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <polyline points="3 6 5 6 21 6" /><path d="M19 6l-1 14H6L5 6" /><path d="M10 11v6M14 11v6" /><path d="M9 6V4h6v2" />
                </svg>
                {t('pages.checkout.summary.vaciar', 'EMPTY CART')}
              </button>
            </div>
            <p className="text-xs font-mono-stack mb-4" style={{ color: 'var(--text-muted)' }}>
              {items.reduce((s, i) => s + i.quantity, 0)} {items.reduce((s, i) => s + i.quantity, 0) === 1 
                ? t('pages.checkout.summary.item', 'ITEM') 
                : t('pages.checkout.summary.items', 'ITEMS')}
            </p>

            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-1">
              {items.map(item => {
                const p = item.product;
                const badges: string[] = [];
                if (p.condition) badges.push(t(`pages.product.condition.${p.condition.toLowerCase()}`, p.condition));
                if (p.foil_treatment && p.foil_treatment !== 'non_foil') badges.push(t(`pages.product.finish.${p.foil_treatment}`, FOIL_LABELS[p.foil_treatment] || p.foil_treatment));
                if (p.card_treatment && p.card_treatment !== 'normal') badges.push(t(`pages.product.version.${p.card_treatment}`, TREATMENT_LABELS[p.card_treatment] || p.card_treatment));

                return (
                  <div key={p.id} className="flex gap-3 pb-3" style={{ borderBottom: '1px solid var(--ink-border)' }}>
                    <div style={{ width: 52, flexShrink: 0 }}>
                      <CardImage imageUrl={p.image_url} name={p.name} tcg={p.tcg} foilTreatment={p.foil_treatment} height={70} enableHover={true} enableModal={true} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-semibold truncate">{p.name}</p>
                      {p.set_name && <p className="text-[10px] truncate" style={{ color: 'var(--text-muted)' }}>{p.set_name}</p>}
                      {(p.cart_count ?? 0) > 0 && (
                        <p className="text-[9px] font-mono mt-0.5 opacity-60" style={{ color: 'var(--gold)' }}>
                          ● {(p.cart_count ?? 0) === 1 
                              ? t('pages.product.cart_users_has', '{count} OTHER USER HAS THIS IN THEIR CART', { count: (p.cart_count ?? 0) })
                              : t('pages.product.cart_users_have', '{count} OTHER USERS HAVE THIS IN THEIR CART', { count: (p.cart_count ?? 0) })}
                        </p>
                      )}
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
                          {p.stock} {t('pages.product.status.available', 'available')}
                        </span>
                        <button
                          type="button"
                          onClick={() => removeItem(p.id)}
                          className="ml-auto p-1 opacity-60 hover:opacity-100 transition-opacity"
                          title={t('pages.common.buttons.remove', 'Remove')}
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

            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-text-muted">{t('pages.checkout.summary.subtotal', 'SUBTOTAL')}</span>
                <span>${totalPrice.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-text-muted">{t('pages.checkout.summary.shipping', 'SHIPPING')}</span>
                <span>{form.is_local_pickup ? t('pages.checkout.summary.free', 'FREE') : `$${(settings?.flat_shipping_fee_cop || 0).toLocaleString()}`}</span>
              </div>
            </div>

            <div className="divider" />

            <div className="flex justify-between items-center mb-4">
              <span className="font-display text-xl">{t('pages.checkout.total', 'TOTAL')}</span>
              <span className="price text-2xl">
                ${(totalPrice + (form.is_local_pickup ? 0 : (settings?.flat_shipping_fee_cop || 0))).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
              </span>
            </div>

            {error && <p className="text-sm font-mono-stack mb-3" style={{ color: 'var(--hp-color)' }}>{error}</p>}

            <button
              type="submit"
              disabled={submitting}
              className="btn-primary w-full py-3 text-lg"
              style={{ opacity: submitting ? 0.7 : 1 }}
            >
              {submitting 
                ? t('pages.checkout.buttons.processing', 'PROCESSING...') 
                : t('pages.checkout.buttons.confirm', 'CONFIRM ORDER →')}
            </button>

            <p className="text-[10px] font-mono-stack text-center mt-3" style={{ color: 'var(--text-muted)' }}>
              {t('pages.checkout.footer.notice', 'Upon confirmation, an advisor will contact you to coordinate delivery.')}
            </p>

            {/* Removed Items Cache */}
            {removedItems.length > 0 && (
              <div className="mt-8 opacity-60 grayscale hover:opacity-100 hover:grayscale-0 transition-all duration-300">
                <h3 className="font-display text-xl mb-4 text-muted">{t('pages.checkout.section.removed', 'REMOVED PRODUCTS')}</h3>
                <div className="space-y-2">
                  {removedItems.map(item => (
                    <div key={item.product.id} className="card p-3 flex items-center gap-4 bg-ink-surface/30">
                      <div style={{ width: 40 }}>
                        <CardImage imageUrl={item.product.image_url} name={item.product.name} tcg={item.product.tcg} foilTreatment={item.product.foil_treatment} height={50} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold truncate">{item.product.name}</p>
                        <p className="text-[10px] text-muted truncate">{item.product.set_name}</p>
                        {(item.product.cart_count ?? 0) > 0 && (
                          <p className="text-[9px] font-mono mt-0.5 opacity-60" style={{ color: 'var(--gold)' }}>
                            ● {(item.product.cart_count ?? 0) === 1 
                                ? t('pages.product.cart_users_has', '{count} OTHER USER HAS THIS IN THEIR CART', { count: (item.product.cart_count ?? 0) })
                                : t('pages.product.cart_users_have', '{count} OTHER USERS HAVE THIS IN THEIR CART', { count: (item.product.cart_count ?? 0) })}
                          </p>
                        )}
                      </div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          onClick={() => restoreItem(item.product.id)}
                          className="badge cursor-pointer hover:bg-gold hover:text-black transition-colors"
                          style={{ fontSize: '0.7rem' }}
                        >
                          {t('pages.checkout.buttons.re_add', 'RE-ADD')}
                        </button>
                        <button
                          type="button"
                          onClick={() => permanentRemove(item.product.id)}
                          className="p-1 text-hp-color"
                          title={t('pages.checkout.buttons.delete_perm', 'Permanently remove')}
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
