import { Product, TCG_SHORT, FOIL_LABELS, TREATMENT_LABELS, resolveLabel } from '@/lib/types';
import CardImage from '@/components/CardImage';
import { useLanguage } from '@/context/LanguageContext';

interface ProductTableRowProps {
  product: Product;
  selected: boolean;
  onSelect: (id: string, selected: boolean) => void;
  onEdit: (p: Product) => void;
  onDelete: (id: string, name: string) => void;
}

export function ProductTableRow({ product: p, selected, onSelect, onEdit, onDelete }: ProductTableRowProps) {
  const { t, locale } = useLanguage();
  const tcgName = p.tcg.length <= 4 ? p.tcg.toUpperCase() : (TCG_SHORT[p.tcg] || p.tcg.substring(0, 3).toUpperCase());

  const formatUpdated = (dateStr: string) => {
    const d = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffHrs = diffMs / (1000 * 60 * 60);

    if (diffHrs < 24) {
      if (diffMs < 60000) return t('pages.common.dates.just_now', 'Just now');
      const mins = Math.floor(diffMs / 60000);
      if (mins < 60) return t('pages.common.dates.mins_ago', '{mins}m ago', { mins });
      return t('pages.common.dates.hours_ago', '{hours}h ago', { hours: Math.floor(diffHrs) });
    }
    
    return d.toLocaleDateString(locale === 'es' ? 'es-ES' : 'en-US', { month: 'short', day: 'numeric' }) + ' ' + 
           d.toLocaleTimeString(locale === 'es' ? 'es-ES' : 'en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
  };

  return (
    <tr key={p.id} onClick={() => onEdit(p)}
      className="cursor-pointer transition-all duration-200 group border-b border-ink-border/30 last:border-0 relative hover:bg-gold/[0.03]"
    >
      <td className="relative overflow-hidden pl-2 pr-0 w-8">
        <div className="flex items-center justify-center h-full">
          <input
            type="checkbox"
            checked={selected}
            onChange={(e) => onSelect(p.id, e.target.checked)}
            onClick={(e) => e.stopPropagation()}
            className="w-4 h-4 rounded border-ink-border/30 text-gold focus:ring-gold bg-white cursor-pointer"
          />
        </div>
      </td>
      <td className="relative overflow-hidden">
        {/* Hover Highlight Bar */}
        <div className="absolute left-0 top-0 bottom-0 w-1 bg-gold scale-y-0 group-hover:scale-y-100 transition-transform duration-200 origin-center" />

        <div className="flex items-center gap-3 py-1">
          <div className="w-9 h-12 sm:w-10 sm:h-14 shrink-0 overflow-hidden relative group/img shadow-sm">
            <CardImage
              imageUrl={p.image_url}
              name={p.name}
              tcg={p.tcg}
              foilTreatment={p.foil_treatment}
              height="100%"
              enableHover={true}
              enableModal={true}
            />
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-0.5">
              <span className="text-[9px] bg-ink-surface px-1.5 py-0.5 rounded border border-ink-border font-bold text-gold tracking-tight" title={p.tcg}>
                {tcgName}
              </span>
              <span className="font-bold text-ink-deep leading-tight truncate max-w-[180px] group-hover:text-gold transition-colors">{p.name}</span>
              <span 
                className="px-1.5 py-0.5 rounded text-[8px] font-bold text-white shadow-sm shrink-0 uppercase"
                style={{ 
                  backgroundColor: p.condition === 'NM' ? 'var(--status-nm)' :
                                   p.condition === 'LP' ? 'var(--status-lp)' :
                                   p.condition === 'MP' ? 'var(--status-mp)' :
                                   p.condition === 'HP' ? 'var(--status-hp)' : 'var(--status-dmg)'
                }}
              >
                {p.condition}
              </span>
              {/* Edit Icon that appears on hover */}
              <span className="opacity-0 group-hover:opacity-100 transition-opacity ml-1 text-gold">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                </svg>
              </span>
            </div>
            <div className="text-[10px] font-mono-stack text-text-muted flex items-center gap-2">
              <span className="truncate max-w-[150px] opacity-70">{p.set_name || t('pages.common.labels.na', 'N/A')}</span>
              {p.set_code && (
                <span className="opacity-70 px-1 bg-ink-surface/50 rounded-sm font-bold">
                  [{p.set_code.toUpperCase()}] {p.collector_number && `#${p.collector_number}`}
                </span>
              )}
            </div>
          </div>
        </div>
      </td>
      <td className="hidden md:table-cell">
        <div className="flex flex-col gap-1 items-start">
          <span className="badge badge-secondary truncate max-w-[100px] border-ink-border/50 text-text-muted" style={{ fontSize: '9px' }}>
            {p.category.toUpperCase()}
          </span>
        </div>
      </td>
      <td className="hidden lg:table-cell">
        <div className="flex flex-wrap gap-1 max-w-[160px]">
          {p.card_treatment && p.card_treatment !== 'normal' && (
            <span className="text-[9px] font-mono-stack px-1.5 py-0.5 bg-gold/5 text-gold border border-gold/20 rounded-sm font-bold uppercase tracking-tighter">
              {resolveLabel(p.card_treatment, TREATMENT_LABELS)}
            </span>
          )}
          {p.promo_type && p.promo_type !== 'none' && p.promo_type !== '' && (
            <span className="text-[9px] font-mono-stack px-1.5 py-0.5 bg-ink-deep/5 text-ink-deep border border-ink-deep/20 rounded-sm font-bold uppercase tracking-tighter">
              {p.promo_type.split(',').map(s => resolveLabel(s.trim(), {})).join(' / ')}
            </span>
          )}
          {p.foil_treatment && p.foil_treatment !== 'non_foil' && (
            <span className="text-[9px] font-mono-stack px-1.5 py-0.5 bg-hp-color/5 text-hp-color border border-hp-color/20 rounded-sm font-bold uppercase tracking-tighter">
              {FOIL_LABELS[p.foil_treatment] || resolveLabel(p.foil_treatment, {})}
            </span>
          )}
        </div>
      </td>
      <td className="hidden xl:table-cell text-center">
        {p.rarity ? (
          <span className={`text-[9px] px-2 py-0.5 rounded-full font-bold uppercase ${
            p.rarity === 'mythic' ? 'bg-hp-color/10 text-hp-color border border-hp-color/20' :
            p.rarity === 'rare' ? 'bg-gold/10 text-gold border border-gold/20' :
            p.rarity === 'uncommon' ? 'bg-ink-deep/10 text-ink-deep border border-ink-deep/20' :
            'bg-text-muted/10 text-text-muted border border-text-muted/20'
          }`}>
            {p.rarity}
          </span>
        ) : <span className="text-[10px] text-text-muted opacity-30">—</span>}
      </td>
      <td className="text-right whitespace-nowrap">
        <div className="flex flex-col items-end">
          <span className="font-mono-stack font-bold text-ink-deep group-hover:text-gold transition-colors">
            {p.price ? `$${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}` : t('pages.common.labels.na', 'N/A')}
          </span>
          <span className="text-[9px] font-mono-stack text-text-muted opacity-50 uppercase tracking-tighter">
            {p.price_source === 'manual' ? t('components.admin.bounty_modal.source_manual', 'MANUAL') : (p.price_source === 'tcgplayer' ? 'TCGP' : 'MCK')}
          </span>
        </div>
      </td>
      <td className="text-center">
        <div className="flex flex-col items-center gap-1">
          <span className={`text-sm font-bold font-mono-stack ${p.stock <= 0 ? 'text-hp-color' : 'text-ink-deep'}`}>
            {p.stock}
          </span>
          <div className="hidden sm:flex flex-wrap gap-1 justify-center">
            {p.stored_in && p.stored_in.length > 0 ? p.stored_in.slice(0, 2).map((s, idx) => (
              <span key={s.stored_in_id || `loc-${idx}`} className="badge shrink-0 shadow-sm" style={{ background: 'var(--kraft-light)', color: 'var(--kraft-dark)', fontSize: '8px', padding: '1px 3px', border: 'none' }}>
                {s.name}: {s.quantity}
              </span>
            )) : null}
          </div>
        </div>
      </td>
      <td className="hidden sm:table-cell text-center">
        <div className="flex flex-col items-center">
           <span className="text-[10px] font-mono-stack font-bold text-text-muted opacity-80 whitespace-nowrap">
             {formatUpdated(p.updated_at || p.created_at)}
           </span>
        </div>
      </td>
      <td onClick={(e) => e.stopPropagation()} className="hidden md:table-cell">
        <div className="flex justify-center">
          <button
            onClick={() => onDelete(p.id, p.name)}
            className="w-8 h-8 flex items-center justify-center text-hp-color hover:bg-hp-color/10 rounded-full transition-all opacity-20 hover:opacity-100 hover:scale-110"
            title={t('pages.admin.inventory.delete_product_tooltip', 'Delete Product')}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
          </button>
        </div>
      </td>
    </tr>
  );
}

interface ProductTableProps {
  products: Product[];
  selectedIds: string[];
  onSelect: (id: string, selected: boolean) => void;
  onSelectAll: (selected: boolean) => void;
  sortKey: string;
  sortDir: 'asc' | 'desc';
  onSort: (key: string) => void;
  onEdit: (p: Product) => void;
  onDelete: (id: string, name: string) => void;
  loading: boolean;
}

export default function ProductTable({
  products,
  selectedIds,
  onSelect,
  onSelectAll,
  sortKey,
  sortDir,
  onSort,
  onEdit,
  onDelete,
  loading
}: ProductTableProps) {
  const { t } = useLanguage();
  const renderSortIcon = (key: string) => {
    if (sortKey !== key) return <span className="opacity-20 ml-1">⇅</span>;
    return <span className="ml-1 text-gold">{sortDir === 'asc' ? '↑' : '↓'}</span>;
  };

  return (
    <div className="p-2 overflow-hidden relative" style={{ minHeight: '400px' }}>
      {loading && (
        <div className="absolute inset-0 z-10 bg-white/60 backdrop-blur-[1px] flex items-center justify-center">
          <div className="flex flex-col items-center">
            <div className="w-12 h-12 border-4 border-kraft-dark border-t-gold rounded-full animate-spin mb-2" />
            <span className="font-mono-stack text-xs font-bold text-kraft-dark uppercase tracking-widest">{t('pages.admin.inventory.scanning_catalog', 'Scanning Catalog...')}</span>
          </div>
        </div>
      )}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr>
              <th className="w-8 pl-2 pr-0">
                <div className="flex items-center justify-center">
                  <input
                    type="checkbox"
                    checked={products.length > 0 && products.every(p => selectedIds.includes(p.id))}
                    onChange={(e) => onSelectAll(e.target.checked)}
                    className="w-4 h-4 rounded border-ink-border/30 text-gold focus:ring-gold bg-white cursor-pointer"
                  />
                </div>
              </th>
              <th onClick={() => onSort('name')} className="cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center">{t('pages.admin.inventory.table.product', 'PRODUCT')} {renderSortIcon('name')}</div>
              </th>
              <th onClick={() => onSort('category')} className="hidden md:table-cell cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center">{t('pages.admin.inventory.table.type', 'TYPE')} {renderSortIcon('category')}</div>
              </th>
              <th className="hidden lg:table-cell min-w-[130px]">
                <div className="flex items-center">{t('pages.admin.inventory.table.variant', 'VARIANT')}</div>
              </th>
              <th onClick={() => onSort('rarity')} className="hidden xl:table-cell w-24 text-center cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center justify-center">{t('pages.admin.inventory.table.rarity', 'RARITY')} {renderSortIcon('rarity')}</div>
              </th>
              <th onClick={() => onSort('price')} className="w-24 sm:w-32 text-right cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center justify-end">{t('pages.admin.inventory.table.price', 'PRICE')} {renderSortIcon('price')}</div>
              </th>
              <th onClick={() => onSort('stock')} className="w-20 sm:w-24 text-center cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center justify-center">{t('pages.admin.inventory.table.stock', 'STOCK')} {renderSortIcon('stock')}</div>
              </th>
              <th onClick={() => onSort('updated_at')} className="hidden sm:table-cell w-24 sm:w-32 text-center cursor-pointer hover:bg-ink-surface transition-colors">
                <div className="flex items-center justify-center">{t('pages.admin.inventory.table.updated', 'UPDATED')} {renderSortIcon('updated_at')}</div>
              </th>
              <th className="hidden md:table-cell w-16 text-center">{t('pages.admin.inventory.table.cmd', 'CMD')}</th>
            </tr>
          </thead>
          <tbody>
            {products.map(p => (
              <ProductTableRow 
                key={p.id} 
                product={p} 
                selected={selectedIds.includes(p.id)}
                onSelect={onSelect}
                onEdit={onEdit} 
                onDelete={onDelete} 
              />
            ))}
            {!loading && products.length === 0 && (
              <tr>
                <td colSpan={8} className="text-center py-20 bg-ink-surface/30">
                  <div className="flex flex-col items-center opacity-30">
                    <span className="text-6xl mb-4">📭</span>
                    <p className="font-display text-2xl tracking-tight">{t('pages.admin.inventory.no_products', 'NO PRODUCTS FOUND')}</p>
                    <p className="text-xs font-mono-stack">{t('pages.admin.inventory.no_products_hint', 'Try adjusting your scanner filters')}</p>
                  </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
