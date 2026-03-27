import { Product, TCG_SHORT } from '@/lib/types';

interface ProductTableRowProps {
  product: Product;
  onEdit: (p: Product) => void;
  onDelete: (id: string, name: string) => void;
}

export function ProductTableRow({ product: p, onEdit, onDelete }: ProductTableRowProps) {
  const tcgName = p.tcg.length <= 4 ? p.tcg.toUpperCase() : (TCG_SHORT[p.tcg] || p.tcg.substring(0, 3).toUpperCase());

  return (
    <tr key={p.id} onClick={() => onEdit(p)} className="cursor-pointer hover:bg-gold/10 hover:shadow-inner transition-all duration-200 group border-b border-ink-border/50 last:border-0 hover:z-10 relative">
      <td>
         <div className="flex items-center gap-3">
            <div className="w-10 h-14 bg-ink-border/20 rounded-sm shrink-0 overflow-hidden border border-ink-border/30 flex items-center justify-center relative group/img">
              {p.image_url ? (
                <img src={p.image_url} alt="" className="w-full h-full object-cover" />
              ) : (
                <span className="text-[8px] font-mono-stack text-text-muted">NO IMG</span>
              )}
              {p.image_url && (
                <div className="fixed left-full top-0 ml-4 z-[100] hidden group-hover/img:block pointer-events-none">
                   <div className="card p-1 bg-ink-surface shadow-2xl border-2 border-gold/50 scale-150 origin-left">
                      <img src={p.image_url} alt="" className="w-48 h-auto rounded-sm" />
                   </div>
                </div>
              )}
            </div>
            <div>
              <div className="flex items-center gap-2 mb-0.5">
                <span className="text-[10px] bg-ink-surface px-1.5 py-0.5 rounded border border-ink-border font-bold text-gold" title={p.tcg}>
                   {tcgName}
                </span>
                <span className="font-bold text-ink-deep leading-tight truncate max-w-[200px]">{p.name}</span>
              </div>
              <div className="text-[10px] font-mono-stack text-text-muted flex items-center gap-2">
                <span className="truncate max-w-[150px]">{p.set_name || 'N/A'}</span>
                {p.set_code && <span className="opacity-60 px-1 bg-ink-surface/50">[{p.set_code.toUpperCase()}]</span>}
              </div>
            </div>
         </div>
      </td>
      <td>
        <div className="flex flex-col gap-1 items-start">
           <span className="badge badge-secondary truncate max-w-[100px]" style={{ fontSize: '10px' }}>
              {p.category.toUpperCase()}
           </span>
           {p.card_treatment && p.card_treatment !== 'normal' && (
              <span className="text-[9px] font-mono-stack opacity-70 px-1 bg-gold/10 text-gold border border-gold/20 rounded-sm">
                 {p.card_treatment.replace(/_/g, ' ').toUpperCase()}
              </span>
           )}
        </div>
      </td>
      <td className="font-mono-stack font-bold">
        <span className={`px-2 py-0.5 rounded text-[11px] ${
          p.condition === 'NM' ? 'bg-nm-color' : 
          p.condition === 'LP' ? 'bg-lp-color' : 
          p.condition === 'MP' ? 'bg-mp-color' : 
          'bg-hp-color'
        } text-white`}>
          {p.condition}
        </span>
      </td>
      <td className="text-right">
        <div className="flex flex-col items-end">
          <span className="font-mono-stack font-bold text-ink-deep">
            {p.price ? `$${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}` : 'N/A'}
          </span>
          <span className="text-[9px] font-mono-stack text-text-muted">
            {p.price_source === 'manual' ? 'MANUAL' : (p.price_source === 'tcgplayer' ? 'TCGP' : 'MCK')}
          </span>
        </div>
      </td>
      <td className="text-center">
        <div className="flex flex-col items-center gap-1">
          <span className={`text-sm font-bold font-mono-stack ${p.stock <= 0 ? 'text-hp-color' : 'text-ink-deep'}`}>
            {p.stock}
          </span>
          <div className="flex flex-wrap gap-1 justify-center">
            {p.stored_in && p.stored_in.length > 0 ? p.stored_in.map((s, idx) => (
              <span key={s.stored_in_id || `loc-${idx}`} className="badge shrink-0" style={{ background: 'var(--kraft-light)', color: 'var(--kraft-dark)', fontSize: '8px', padding: '1px 3px' }}>
                {s.name}: {s.quantity}
              </span>
            )) : <span className="text-[8px] text-text-muted opacity-50 italic">unassigned</span>}
          </div>
        </div>
      </td>
      <td onClick={(e) => e.stopPropagation()}>
        <div className="flex justify-center">
          <button 
            onClick={() => onDelete(p.id, p.name)} 
            className="w-8 h-8 flex items-center justify-center text-hp-color hover:bg-hp-color/10 rounded-full transition-colors opacity-40 hover:opacity-100"
            title="Delete Product"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              <line x1="10" y1="11" x2="10" y2="17"></line>
              <line x1="14" y1="11" x2="14" y2="17"></line>
            </svg>
          </button>
        </div>
      </td>
    </tr>
  );
}

interface ProductTableProps {
  products: Product[];
  sortKey: string;
  sortDir: 'asc' | 'desc';
  onSort: (key: string) => void;
  onEdit: (p: Product) => void;
  onDelete: (id: string, name: string) => void;
  loading: boolean;
}

export default function ProductTable({
  products,
  sortKey,
  sortDir,
  onSort,
  onEdit,
  onDelete,
  loading
}: ProductTableProps) {
  const renderSortIcon = (key: string) => {
    if (sortKey !== key) return <span className="opacity-20 ml-1">⇅</span>;
    return <span className="ml-1 text-gold">{sortDir === 'asc' ? '↑' : '↓'}</span>;
  };

  return (
    <div className="card no-tilt p-0 overflow-hidden relative" style={{ minHeight: '400px' }}>
      {loading && (
        <div className="absolute inset-0 z-10 bg-white/60 backdrop-blur-[1px] flex items-center justify-center">
          <div className="flex flex-col items-center">
             <div className="w-12 h-12 border-4 border-kraft-dark border-t-gold rounded-full animate-spin mb-2" />
             <span className="font-mono-stack text-xs font-bold text-kraft-dark uppercase tracking-widest">Scanning Catalog...</span>
          </div>
        </div>
      )}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr>
              <th onClick={() => onSort('name')} className="cursor-pointer hover:bg-ink-surface transition-colors">
                 <div className="flex items-center">PRODUCT {renderSortIcon('name')}</div>
              </th>
              <th onClick={() => onSort('category')} title="Category / Treatment" className="cursor-pointer hover:bg-ink-surface transition-colors">
                 <div className="flex items-center">TYPE {renderSortIcon('category')}</div>
              </th>
              <th onClick={() => onSort('condition')} className="w-24 text-center cursor-pointer hover:bg-ink-surface transition-colors">
                 <div className="flex items-center justify-center">CND {renderSortIcon('condition')}</div>
              </th>
              <th onClick={() => onSort('price')} className="w-32 text-right cursor-pointer hover:bg-ink-surface transition-colors">
                 <div className="flex items-center justify-end">PRICE {renderSortIcon('price')}</div>
              </th>
              <th onClick={() => onSort('stock')} className="w-28 text-center cursor-pointer hover:bg-ink-surface transition-colors">
                 <div className="flex items-center justify-center">STOCK {renderSortIcon('stock')}</div>
              </th>
              <th className="w-20 text-center">CMD</th>
            </tr>
          </thead>
          <tbody>
            {products.map(p => (
              <ProductTableRow key={p.id} product={p} onEdit={onEdit} onDelete={onDelete} />
            ))}
            {!loading && products.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-20 bg-ink-surface/30">
                  <div className="flex flex-col items-center opacity-30">
                    <span className="text-6xl mb-4">📭</span>
                    <p className="font-display text-2xl tracking-tight">NO PRODUCTS FOUND</p>
                    <p className="text-xs font-mono-stack">Try adjusting your scanner filters</p>
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
