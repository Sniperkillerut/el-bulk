'use client';

import { Product, TCG_SHORT } from '@/lib/types';
import CardImage from '@/components/CardImage';

interface ProductTableRowProps {
  product: Product;
  onEdit: (p: Product) => void;
  onDelete: (id: string, name: string) => void;
}

export function ProductTableRow({ product: p, onEdit, onDelete }: ProductTableRowProps) {
  const tcgName = p.tcg.length <= 4 ? p.tcg.toUpperCase() : (TCG_SHORT[p.tcg] || p.tcg.substring(0, 3).toUpperCase());

  return (
    <tr key={p.id} onClick={() => onEdit(p)} 
      className="cursor-pointer transition-all duration-200 group border-b border-ink-border/30 last:border-0 relative hover:bg-gold/[0.03]"
    >
      <td className="relative overflow-hidden">
        {/* Hover Highlight Bar */}
        <div className="absolute left-0 top-0 bottom-0 w-1 bg-gold scale-y-0 group-hover:scale-y-100 transition-transform duration-200 origin-center" />
        
        <div className="flex items-center gap-3 py-1">
            <div className="w-10 h-14 shrink-0 overflow-hidden relative group/img shadow-sm">
              <CardImage 
                imageUrl={p.image_url} 
                name={p.name} 
                tcg={p.tcg} 
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
                <span className="font-bold text-ink-deep leading-tight truncate max-w-[220px] group-hover:text-gold transition-colors">{p.name}</span>
                {/* Edit Icon that appears on hover */}
                <span className="opacity-0 group-hover:opacity-100 transition-opacity ml-1 text-gold">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                  </svg>
                </span>
              </div>
              <div className="text-[10px] font-mono-stack text-text-muted flex items-center gap-2">
                <span className="truncate max-w-[150px] opacity-70">{p.set_name || 'N/A'}</span>
                {p.set_code && <span className="opacity-40 px-1 bg-ink-surface/50 rounded-sm">[{p.set_code.toUpperCase()}]</span>}
              </div>
            </div>
         </div>
      </td>
      <td>
        <div className="flex flex-col gap-1 items-start">
           <span className="badge badge-secondary truncate max-w-[100px] border-ink-border/50 text-text-muted" style={{ fontSize: '9px' }}>
              {p.category.toUpperCase()}
           </span>
           {p.card_treatment && p.card_treatment !== 'normal' && (
              <span className="text-[9px] font-mono-stack opacity-80 px-1 bg-gold/5 text-gold border border-gold/10 rounded-sm">
                 {p.card_treatment.replace(/_/g, ' ').toUpperCase()}
              </span>
           )}
        </div>
      </td>
      <td className="font-mono-stack">
        <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
          p.condition === 'NM' ? 'bg-nm-color/80' : 
          p.condition === 'LP' ? 'bg-lp-color/80' : 
          p.condition === 'MP' ? 'bg-mp-color/80' : 
          'bg-hp-color/80'
        } text-white shadow-sm`}>
          {p.condition}
        </span>
      </td>
      <td className="text-right">
        <div className="flex flex-col items-end">
          <span className="font-mono-stack font-bold text-ink-deep group-hover:text-gold transition-colors">
            {p.price ? `$${p.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}` : 'N/A'}
          </span>
          <span className="text-[9px] font-mono-stack text-text-muted opacity-50 uppercase tracking-tighter">
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
              <span key={s.stored_in_id || `loc-${idx}`} className="badge shrink-0 shadow-sm" style={{ background: 'var(--kraft-light)', color: 'var(--kraft-dark)', fontSize: '8px', padding: '1px 3px', border: 'none' }}>
                {s.name}: {s.quantity}
              </span>
            )) : <span className="text-[8px] text-text-muted opacity-30 italic">unassigned</span>}
          </div>
        </div>
      </td>
      <td onClick={(e) => e.stopPropagation()}>
        <div className="flex justify-center">
          <button 
            onClick={() => onDelete(p.id, p.name)} 
            className="w-8 h-8 flex items-center justify-center text-hp-color hover:bg-hp-color/10 rounded-full transition-all opacity-20 hover:opacity-100 hover:scale-110"
            title="Delete Product"
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
    <div className="card p-0 overflow-hidden relative" style={{ minHeight: '400px' }}>
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
