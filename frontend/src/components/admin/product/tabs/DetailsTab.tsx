import { FormState } from '../../ProductEditModal';
import { CustomCategory } from '@/lib/types';

interface DetailsTabProps {
  form: FormState;
  isMTGSingles: boolean;
  categories: CustomCategory[];
  onUpdate: (updates: Partial<FormState>) => void;
}

export default function DetailsTab({
  form,
  isMTGSingles,
  categories,
  onUpdate
}: DetailsTabProps) {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Basic Details */}
        <div className="space-y-4">
          <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>BASIC DETAILS</h3>
          
          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Description</label>
            <textarea 
              rows={4} 
              value={form.description} 
              onChange={e => onUpdate({ description: e.target.value })} 
              placeholder="Product description..."
            />
          </div>

          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Collections / Categories</label>
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
              {categories.length === 0 && <span className="text-[10px] text-text-muted italic">No custom collections defined.</span>}
            </div>
          </div>
          
          <div>
            <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Image URL</label>
            <input 
              type="text" 
              value={form.image_url} 
              onChange={e => onUpdate({ image_url: e.target.value })} 
              placeholder="https://..." 
            />
          </div>
        </div>

        {/* MTG Metadata */}
        {isMTGSingles && (
          <div className="space-y-4">
            <h3 className="text-xs font-mono-stack uppercase" style={{ color: 'var(--text-muted)', letterSpacing: '0.1em' }}>MTG METADATA</h3>
            
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Language</label>
                <input type="text" value={form.language} onChange={e => onUpdate({ language: e.target.value })} />
              </div>
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Rarity</label>
                <input type="text" value={form.rarity} onChange={e => onUpdate({ rarity: e.target.value })} />
              </div>
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Colors</label>
                <input type="text" value={form.color_identity} onChange={e => onUpdate({ color_identity: e.target.value })} />
              </div>
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">CMC</label>
                <input type="number" value={form.cmc} onChange={e => onUpdate({ cmc: e.target.value === '' ? '' : Number(e.target.value) })} />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-x-4 gap-y-2 py-2 border-y border-ink-border/30">
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.is_legendary} onChange={e => onUpdate({ is_legendary: e.target.checked })} />
                LEGENDARY
              </label>
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.is_historic} onChange={e => onUpdate({ is_historic: e.target.checked })} />
                HISTORIC
              </label>
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.is_land} onChange={e => onUpdate({ is_land: e.target.checked })} />
                LAND
              </label>
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.is_basic_land} onChange={e => onUpdate({ is_basic_land: e.target.checked })} />
                BASIC LAND
              </label>
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.full_art} onChange={e => onUpdate({ full_art: e.target.checked })} />
                FULL ART
              </label>
              <label className="flex items-center gap-2 text-xs font-mono-stack cursor-pointer">
                <input type="checkbox" checked={form.textless} onChange={e => onUpdate({ textless: e.target.checked })} />
                TEXTLESS
              </label>
            </div>

            <div>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Type Line</label>
              <input type="text" value={form.type_line} onChange={e => onUpdate({ type_line: e.target.value })} />
            </div>

            <div>
              <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Oracle Text</label>
              <textarea rows={3} value={form.oracle_text} onChange={e => onUpdate({ oracle_text: e.target.value })} />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Artist</label>
                <input type="text" value={form.artist} onChange={e => onUpdate({ artist: e.target.value })} />
              </div>
              <div>
                <label className="text-[10px] font-mono-stack mb-1 block uppercase text-text-muted">Border Color</label>
                <input type="text" value={form.border_color} onChange={e => onUpdate({ border_color: e.target.value })} />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
