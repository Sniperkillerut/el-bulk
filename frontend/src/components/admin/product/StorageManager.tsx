'use client';

import { StorageLocation, StoredIn } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

interface StorageManagerProps {
  productStorage: StorageLocation[];
  storageLocations: StoredIn[];
  onUpdateQty: (id: string, delta: number) => void;
  onSetQty: (id: string, qty: number) => void;
  onRemove: (id: string) => void;
  onAdd: (id: string) => void;
}

export default function StorageManager({
  productStorage,
  storageLocations,
  onUpdateQty,
  onSetQty,
  onRemove,
  onAdd
}: StorageManagerProps) {
  const { t } = useLanguage();
  const total = productStorage.reduce((acc, l) => acc + l.quantity, 0);

  return (
    <div className="pt-4" style={{ borderTop: '1px dashed var(--ink-border)' }}>
      <div className="flex justify-between items-center mb-3">
        <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>{t('components.admin.product.storage.title', 'STORAGE LOCATIONS')}</h3>
        <span className="text-xs font-bold bg-ink-surface px-2 py-1 text-gold rounded border border-ink-border">
          {t('pages.cart.drawer.total', 'TOTAL')}: {total}
        </span>
      </div>
      <div className="space-y-2 max-h-60 overflow-y-auto mb-4">
        {productStorage.length === 0 && <p className="text-xs text-text-muted italic text-center py-2">{t('components.admin.product.storage.empty', 'No storage assignments yet.')}</p>}
        {productStorage.map(loc => {
          const isPending = loc.name.toLowerCase() === 'pending';
          return (
          <div key={loc.stored_in_id} className="flex items-center justify-between gap-2 text-sm border-b border-ink-border/50 pb-2">
            <span className="truncate flex-1 font-semibold leading-tight min-w-0" title={loc.name}>
              {loc.name} {isPending && <span className="text-[10px] uppercase ml-1 px-1 py-0.5 rounded bg-hp-color/10 text-status-hp font-mono-stack">LOCKED</span>}
            </span>
            <div className={`flex items-center gap-0.5 shrink-0 ${isPending ? 'opacity-60 cursor-not-allowed' : ''}`}>
              <button
                onClick={() => onUpdateQty(loc.stored_in_id, -1)}
                className="w-6 h-6 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-hp-color transition-colors rounded-l-sm text-xs"
                disabled={isPending || loc.quantity <= 0}
              >−</button>
              <input
                type="number"
                value={loc.quantity === 0 ? '' : loc.quantity}
                min="0"
                onChange={e => onSetQty(loc.stored_in_id, parseInt(e.target.value) || 0)}
                className="px-1 py-0 text-center text-xs font-mono-stack border-y border-ink-border bg-white focus:outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                style={{ width: '50px', height: '24px', borderRadius: '0' }}
                placeholder="0"
                disabled={isPending}
              />
              <button
                onClick={() => onUpdateQty(loc.stored_in_id, 1)}
                className="w-6 h-6 flex items-center justify-center bg-ink-surface border border-ink-border hover:text-gold transition-colors rounded-r-sm text-xs"
                disabled={isPending}
              >+</button>
              {!isPending ? (
                <button
                  onClick={() => onRemove(loc.stored_in_id)}
                  className="w-8 h-6 flex items-center justify-center hover:text-hp-color opacity-30 hover:opacity-100 transition-opacity ml-1"
                  title="Remove"
                >
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                  </svg>
                </button>
              ) : (
                <div className="w-8 h-6 ml-1" />
              )}
            </div>
          </div>
        )})}
      </div>
      <div>
        <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">{t('components.admin.product.storage.quick_add', 'Quick Add Location')}</label>
        <select
          className="w-full text-xs px-2 h-10"
          style={{ padding: '0 8px' }}
          onChange={(e) => {
            const id = e.target.value;
            if (!id) return;
            onAdd(id);
            e.target.value = "";
          }}
        >
          <option value="">{t('components.admin.product.storage.select_placeholder', '-- Select Location --')}</option>
          {storageLocations
            .filter(l => l.name.toLowerCase() !== 'pending' && !productStorage.find(p => p.stored_in_id === l.id))
            .map(l => <option key={l.id} value={l.id}>{l.name}</option>)
          }
        </select>
      </div>
    </div>
  );
}
