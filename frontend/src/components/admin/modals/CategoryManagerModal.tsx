'use client';

import { CustomCategory } from '@/lib/types';
import { useState } from 'react';
import { CategoryIcon, COLLECTION_ICONS } from '@/components/CategoryIcon';

interface CategoryManagerModalProps {
  categories: CustomCategory[];
  onCreate: (name: string, data?: Partial<CustomCategory>) => Promise<void>;
  onUpdate: (id: string, name: string, data?: Partial<CustomCategory>) => Promise<void>;
  onDelete: (id: string, name: string) => Promise<void>;
  onClose: () => void;
}

export default function CategoryManagerModal({
  categories,
  onCreate,
  onUpdate,
  onDelete,
  onClose
}: CategoryManagerModalProps) {
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<Partial<CustomCategory>>({});

  const handleCreate = async () => {
    if (!newName.trim()) return;
    await onCreate(newName);
    setNewName('');
  };

  const startEditing = (cat: CustomCategory) => {
    setEditingId(cat.id);
    setEditForm({
      name: cat.name,
      bg_color: cat.bg_color || '#b8860b',
      text_color: cat.text_color || '#ffffff',
      icon: cat.icon || 'none'
    });
  };

  const handleUpdate = async (id: string) => {
    if (!editForm.name?.trim()) return;
    await onUpdate(id, editForm.name, editForm);
    setEditingId(null);
  };

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
      style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
      <div className="card max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)' }}>
        <div className="flex items-center justify-between mb-8">
          <h2 className="font-display text-4xl m-0 uppercase tracking-tighter">COLLECTIONS & CATS</h2>
          <button onClick={onClose} className="text-text-muted hover:text-text-primary text-xl">✕</button>
        </div>
        
        <div className="flex gap-2 mb-6">
          <input 
            type="text" 
            placeholder="New Collection Name (e.g. Commander Staples)" 
            value={newName} 
            onChange={e => setNewName(e.target.value)} 
            className="flex-1 bg-white border-kraft-dark/30" 
          />
          <button onClick={handleCreate} className="btn-primary px-6">ADD</button>
        </div>

        <div className="space-y-3 max-h-[500px] overflow-y-auto pr-2 custom-scrollbar">
          {categories.map(cat => {
            const isEditing = editingId === cat.id;
            return (
              <div key={cat.id} className={`p-4 border transition-all ${isEditing ? 'border-accent-primary bg-accent-primary/5 shadow-lg shadow-gold/10' : 'border-kraft-dark/20 bg-white/40'}`}>
                {isEditing ? (
                  <div className="space-y-4">
                    <div className="flex gap-2 items-center">
                      <input 
                        type="text" 
                        value={editForm.name} 
                        onChange={e => setEditForm({ ...editForm, name: e.target.value })} 
                        className="flex-1 py-2 bg-white border-kraft-dark/30 text-base font-bold" 
                      />
                      <div className="flex gap-1">
                        <button onClick={() => handleUpdate(cat.id)} className="btn-primary px-4 py-2 text-xs font-black uppercase tracking-widest">SAVE</button>
                        <button onClick={() => setEditingId(null)} className="btn-secondary px-4 py-2 text-xs font-black uppercase tracking-widest">CANCEL</button>
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4 pt-2 border-t border-kraft-dark/10">
                      <div className="space-y-1">
                        <label className="text-[10px] font-mono-stack uppercase font-bold text-text-muted opacity-60">Background & Text</label>
                        <div className="flex items-center gap-3">
                           <div className="flex items-center gap-2 bg-white/60 p-1.5 rounded border border-kraft-dark/20">
                             <input type="color" value={editForm.bg_color} onChange={e => setEditForm({...editForm, bg_color: e.target.value})} className="w-8 h-8 rounded border-none cursor-pointer" />
                             <input type="text" value={editForm.bg_color} onChange={e => setEditForm({...editForm, bg_color: e.target.value})} className="w-16 p-0 border-none bg-transparent text-[10px] font-mono uppercase" />
                           </div>
                           <div className="flex items-center gap-2 bg-white/60 p-1.5 rounded border border-kraft-dark/20">
                             <input type="color" value={editForm.text_color} onChange={e => setEditForm({...editForm, text_color: e.target.value})} className="w-8 h-8 rounded border-none cursor-pointer" />
                             <input type="text" value={editForm.text_color} onChange={e => setEditForm({...editForm, text_color: e.target.value})} className="w-16 p-0 border-none bg-transparent text-[10px] font-mono uppercase" />
                           </div>
                        </div>
                      </div>
                      <div className="space-y-1">
                        <label className="text-[10px] font-mono-stack uppercase font-bold text-text-muted opacity-60">Badge Icon</label>
                        <div className="flex flex-wrap gap-1.5">
                          {COLLECTION_ICONS.map(i => (
                            <button
                              key={i.name}
                              type="button"
                              onClick={() => setEditForm({ ...editForm, icon: i.name })}
                              className={`w-8 h-8 flex items-center justify-center rounded border transition-colors ${editForm.icon === i.name ? 'bg-accent-primary text-text-on-accent border-accent-primary' : 'bg-white border-kraft-dark/20 text-text-muted hover:border-accent-primary'}`}
                              title={i.label}
                            >
                              {i.name === 'none' ? <span className="text-[8px] font-mono">Ø</span> : i.svg}
                            </button>
                          ))}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center justify-center py-4 bg-ink-border/5 rounded-md border border-dashed border-kraft-dark/20">
                       <span className="text-[10px] font-mono uppercase text-text-muted mr-3">PREVIEW:</span>
                       <span className="badge shadow-md" style={{ background: editForm.bg_color, color: editForm.text_color, fontSize: '0.65rem', padding: '0.2rem 0.65rem', fontWeight: 'bold', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                         <CategoryIcon icon={editForm.icon} />
                         {editForm.name?.toUpperCase()}
                       </span>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="flex flex-col">
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-lg">{cat.name}</span>
                          <span className="badge" style={{ 
                            background: cat.bg_color || 'var(--accent-primary)', 
                            color: cat.text_color || 'var(--text-on-accent)',
                            fontSize: '0.6rem',
                            padding: '0.1rem 0.4rem',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.2rem',
                            fontWeight: 'bold'
                          }}>
                            <CategoryIcon icon={cat.icon} />
                            {cat.name.toUpperCase()}
                          </span>
                        </div>
                        <span className="text-[10px] font-mono-stack text-text-muted opacity-60 uppercase font-bold tracking-widest mt-0.5">
                          {cat.item_count || 0} Products Linked
                        </span>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <button 
                        onClick={() => startEditing(cat)} 
                        className="btn-secondary px-4 py-1.5 text-xs font-bold"
                      >EDIT</button>
                      <button 
                        onClick={() => onDelete(cat.id, cat.name)} 
                        className="px-4 py-1.5 text-xs font-bold border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors rounded"
                      >DELETE</button>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
          {categories.length === 0 && <p className="text-center text-text-muted py-8 font-mono-stack">VOID DETECTED // NO COLLECTIONS DEFINED</p>}
        </div>
      </div>
    </div>
  );
}
