import { FormState } from '../../ProductEditModal';
import { PriceSource, Settings, StorageLocation, StoredIn } from '@/lib/types';
import StorageManager from '../StorageManager';

interface PricingTabProps {
  form: FormState;
  settings: Settings | undefined;
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
  productStorage,
  storageLocations,
  onUpdate,
  onSourceChange,
  onUpdateStoreQty,
  onSetStoreQty,
  onRemoveStorage,
  onAddStorage
}: PricingTabProps) {
  return (
    <div className="space-y-6">
      {/* Pricing */}
      <div>
        <div className="flex justify-between items-center mb-3">
          <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>PRICING</h3>
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
            <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE SOURCE *</label>
            <select value={form.price_source} onChange={e => onSourceChange(e.target.value as PriceSource)}>
              <option value="manual">Manual Override (COP)</option>
              <option value="tcgplayer">External: TCGPlayer (USD)</option>
              <option value="cardmarket">External: Cardmarket (EUR)</option>
            </select>
          </div>
          {form.price_source === 'manual' ? (
            <div>
              <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>PRICE (COP) *</label>
              <input 
                type="number" 
                step="100" 
                min="0" 
                value={form.price_cop_override} 
                onChange={e => onUpdate({ price_cop_override: e.target.value ? parseFloat(e.target.value) : '' })} 
              />
            </div>
          ) : (
            <div>
              <label className="text-[10px] font-mono-stack mb-1 block" style={{ color: 'var(--text-muted)' }}>
                REFERENCE PRICE ({form.price_source === 'tcgplayer' ? 'USD' : 'EUR'}) *
              </label>
              <input 
                type="number" 
                step="0.01" 
                min="0" 
                value={form.price_reference}
                onChange={e => onUpdate({ price_reference: e.target.value ? parseFloat(e.target.value) : '' })}
                style={{ 
                  color: form.price_reference === 0 ? 'var(--hp-color)' : 'inherit', 
                  borderColor: form.price_reference === 0 ? 'var(--hp-color)' : 'var(--ink-border)' 
                }} 
              />
            </div>
          )}
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
    </div>
  );
}
