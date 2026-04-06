'use client';

import { FormState } from '../types';
import { CustomCategory, PriceSource, Settings, StorageLocation, StoredIn } from '@/lib/types';
import StorageManager from '../StorageManager';
import { useLanguage } from '@/context/LanguageContext';

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

  return (
    <div className="space-y-6">
      {/* Pricing */}
      <div>
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>{t('components.admin.product_modal.pricing.title', 'PRICING')}</h3>
          <div className="text-xs font-mono-stack px-2 py-1 rounded" style={{ background: 'var(--ink-surface)', color: 'var(--gold)' }}>
            {form.price_source === 'tcgplayer' && `(x ${settings?.usd_to_cop_rate || 0} COP)`}
            {form.price_source === 'cardmarket' && `(x ${settings?.eur_to_cop_rate || 0} COP)`}
            {form.price_source !== 'manual' && typeof form.price_reference === 'number' && (
              <span className="ml-2 font-bold text-sm" style={{ color: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--gold)' }}>
                = ${(form.price_reference * (form.price_source === 'tcgplayer' ? (settings?.usd_to_cop_rate || 0) : (settings?.eur_to_cop_rate || 0))).toLocaleString('en-US', { maximumFractionDigits: 0 })} COP
              </span>
            )}
          </div>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>{t('components.admin.product_modal.pricing.source_label', 'PRICE SOURCE *')}</label>
            <select value={form.price_source} onChange={e => onSourceChange(e.target.value as PriceSource)}>
              <option value="manual">{t('components.admin.product_modal.pricing.source_manual', 'Manual Override (COP)')}</option>
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
              <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>
                {t('components.admin.product_modal.pricing.ref_price_label', 'REFERENCE PRICE ({currency}) *', { currency: form.price_source === 'tcgplayer' ? 'USD' : 'EUR' })}
              </label>
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
              <input 
                type="text" 
                value={form.image_url} 
                onChange={e => onUpdate({ image_url: e.target.value })} 
                placeholder="https://..." 
              />
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
