'use client';

import { FormState } from '../types';
import { CustomCategory, PriceSource, Settings, StorageLocation, StoredIn } from '@/lib/types';
import StorageManager from '../StorageManager';
import { useLanguage } from '@/context/LanguageContext';
import { useState } from 'react';
import { adminFetchExternalPrice } from '@/lib/api';

interface PricingTabProps {
  form: FormState;
  settings: Settings | undefined;
  categories: CustomCategory[];
  productStorage: StorageLocation[];
  storageLocations: StoredIn[];
  onUpdate: (update: Partial<FormState>) => void;
  onSourceChange: (src: PriceSource) => void;
  onUpdateStoreQty: (id: string, delta: number) => void;
  onSetStoreQty: (id: string, qty: number) => void;
  onRemoveStorage: (id: string) => void;
  onAddStorage: (id: string) => void;
}

export default function PricingTab({
	form,
	settings,
	categories,
	productStorage,
	storageLocations,
	onUpdate,
	onSourceChange,
	onUpdateStoreQty,
	onSetStoreQty,
	onRemoveStorage,
	onAddStorage
}: PricingTabProps) {
	const { t } = useLanguage();
	const [isUploading, setIsUploading] = useState(false);
	const [fetchingPrice, setFetchingPrice] = useState(false);

	const fetchSuggestedPrice = async () => {
		if (!form.name || form.price_source === 'manual') return;
		setFetchingPrice(true);
		try {
			const res = await adminFetchExternalPrice(form.name, form.set_code || '', form.set_name || '', form.collector_number || '', form.foil_treatment || '', form.card_treatment || '', form.price_source, form.scryfall_id);
			onUpdate({ price_reference: res.price });
		} catch (err) {
			console.error('Failed to fetch suggested price:', err);
			alert(`Failed to fetch ${form.price_source} price. The edition name might not match perfectly. Check console for details.`);
			onUpdate({ price_reference: 0 });
		} finally {
			setFetchingPrice(false);
		}
	};

	const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0];
		if (!file) return;

		// Simple local check
		if (file.size > 5 * 1024 * 1024) {
			alert(t('components.admin.product_modal.pricing.upload_size_error', 'File too large (Max 5MB)'));
			return;
		}

		const formData = new FormData();
		formData.append('file', file);

		setIsUploading(true);
		try {
			const res = await fetch('/api/admin/upload', {
				method: 'POST',
				body: formData,
			});

			if (!res.ok) {
				const error = await res.text();
				throw new Error(error || 'Upload failed');
			}

			const data = await res.json();
			onUpdate({ image_url: data.url });
			// No global toast system found, using alert for now
			alert(t('components.admin.product_modal.pricing.upload_success', 'Image uploaded and linked!'));
		} catch (err: unknown) {
			const message = err instanceof Error ? err.message : 'Unknown error';
			console.error('Upload error:', err);
			alert(t('components.admin.product_modal.pricing.upload_error', 'Failed to upload image: {error}', { error: message }));
		} finally {
			setIsUploading(false);
			// Reset input
			e.target.value = '';
		}
	};

  return (
    <div className="space-y-6">
      {/* Pricing */}
      <div>
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>{t('components.admin.product_modal.pricing.title', 'PRICING')}</h3>
          <div className="text-xs font-mono-stack px-2 py-1 rounded" style={{ background: 'var(--ink-surface)', color: 'var(--gold)' }}>
            {(form.price_source === 'tcgplayer' || form.price_source === 'cardkingdom') && `(x ${form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0) : (settings?.ck_to_cop_rate || settings?.usd_to_cop_rate || 0)} COP)`}
            {form.price_source === 'cardmarket' && `(x ${settings?.eur_to_cop_rate || 0} COP)`}
            {form.price_source !== 'manual' && typeof form.price_reference === 'number' && (
              <span className="ml-2 font-bold text-sm" style={{ color: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--gold)' }}>
                = ${(form.price_reference * (
                  form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0) : 
                  form.price_source === 'cardkingdom' ? (settings?.ck_to_cop_rate || settings?.usd_to_cop_rate || 0) : 
                  (settings?.eur_to_cop_rate || 0)
                )).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
              </span>
            )}
          </div>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.pricing.source_label', 'PRICE SOURCE *')}</label>
            <select value={form.price_source} onChange={e => onSourceChange(e.target.value as PriceSource)}>
              <option value="manual">{t('components.admin.product_modal.pricing.source_manual', 'Manual Override (COP)')}</option>
              <option value="cardkingdom">{t('components.admin.product_modal.pricing.source_cardkingdom', 'External: CardKingdom (USD)')}</option>
              <option value="tcgplayer">{t('components.admin.product_modal.pricing.source_tcgplayer', 'External: TCGPlayer (USD)')}</option>
              <option value="cardmarket">{t('components.admin.product_modal.pricing.source_cardmarket', 'External: Cardmarket (EUR)')}</option>
            </select>
          </div>
          {form.price_source === 'manual' ? (
            <div>
              <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.pricing.price_cop_label', 'PRICE (COP) *')}</label>
              <input type="number" step="0.01" value={form.price_cop_override ?? ''} onChange={e => onUpdate({ price_cop_override: e.target.value === '' ? '' : Number(e.target.value) })} />
            </div>
          ) : (
            <div>
              <div className="flex justify-between items-center mb-1">
                <label className="text-[10px] font-mono-stack block" style={{ color: 'var(--text-muted)' }}>
                  {t('components.admin.product_modal.pricing.ref_price_label', 'REFERENCE PRICE ({currency}) *', { currency: (form.price_source === 'tcgplayer' || form.price_source === 'cardkingdom') ? 'USD' : 'EUR' })}
                </label>
                {form.id && (
                  <button 
                    type="button"
                    onClick={fetchSuggestedPrice}
                    disabled={fetchingPrice}
                    className="text-[9px] font-bold text-gold hover:underline uppercase tracking-tighter"
                  >
                    {fetchingPrice ? '...' : t('components.admin.product_modal.pricing.suggested_btn', 'FETCH SUGGESTED')}
                  </button>
                )}
              </div>
              <input type="number" step="0.01" value={form.price_reference ?? ''} onChange={e => onUpdate({ price_reference: e.target.value === '' ? '' : Number(e.target.value) })} style={{ 
                  color: form.price_reference === 0 ? 'var(--hp-color)' : 'inherit', 
                  borderColor: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--ink-border)' 
                }} 
              />
            </div>
          )}
          <div>
            <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.pricing.cost_basis_label', 'COST BASIS (COP)')}</label>
            <input 
              type="number" 
              step="1" 
              value={form.cost_basis_cop ?? ''} 
              onChange={e => onUpdate({ cost_basis_cop: e.target.value === '' ? '' : Number(e.target.value) })} 
              placeholder="0"
            />
          </div>
        </div>
      </div>

      <StorageManager 
        productStorage={productStorage}
        storageLocations={storageLocations}
        onUpdateQty={onUpdateStoreQty}
        onSetQty={onSetStoreQty}
        onRemove={onRemoveStorage}
        onAdd={onAddStorage}
      />

      <div className="pt-6 border-t border-ink-border/50 space-y-6">
        <h3 className="text-xs font-mono-stack uppercase text-text-muted letter-spacing-widest">{t('components.admin.product_modal.pricing.basic_title', 'Basic Details')}</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.pricing.description_label', 'Description')}</label>
              <textarea 
                rows={4} 
                value={form.description} 
                onChange={e => onUpdate({ description: e.target.value })} 
                placeholder={t('components.admin.product_modal.pricing.description_placeholder', 'Product description...')}
              />
            </div>
            <div>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.pricing.image_url_label', 'Image URL')}</label>
              <div className="flex gap-2">
                <input 
                  type="text" 
                  value={form.image_url} 
                  onChange={e => onUpdate({ image_url: e.target.value })} 
                  placeholder="https://..." 
                  className="flex-1"
                />
                <label className={`shrink-0 w-10 h-10 flex items-center justify-center rounded border transition-all cursor-pointer ${isUploading ? 'bg-ink-border animate-pulse' : 'bg-white border-ink-border hover:border-gold hover:text-gold'}`} title={t('components.admin.product_modal.pricing.upload_btn_tooltip', 'Upload to Cloud')}>
                  {isUploading ? (
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                  ) : (
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                      <polyline points="17 8 12 3 7 8"></polyline>
                      <line x1="12" y1="3" x2="12" y2="15"></line>
                    </svg>
                  )}
                  <input type="file" className="hidden" accept="image/*" onChange={handleUpload} disabled={isUploading} />
                </label>
              </div>
            </div>
          </div>

          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product_modal.pricing.collections_label', 'Collections / Categories')}</label>
            <div className="flex flex-wrap gap-2 p-3 border border-ink-border bg-ink-surface rounded-sm min-h-[42px]">
              {categories.map(cat => (
                <label key={cat.id} className="flex items-center gap-2 px-2 py-1 bg-kraft-light/20 border border-kraft-dark/30 rounded cursor-pointer hover:bg-kraft-light/40 transition-colors">
                  <input 
                    type="checkbox" 
                    checked={form.category_ids.includes(cat.id)}
                    onChange={e => {
                      const ids = e.target.checked 
                        ? [...form.category_ids, cat.id]
                        : form.category_ids.filter(id => id !== cat.id);
                      onUpdate({ category_ids: ids });
                    }}
                  />
                  <span className="text-xs font-mono-stack">{cat.name}</span>
                </label>
              ))}
              {categories.length === 0 && <span className="text-[10px] text-text-muted italic">{t('components.admin.product_modal.pricing.no_collections', 'No custom collections defined.')}</span>}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
