'use client';

import { useState, useEffect } from 'react';
import { TCG } from '@/lib/types';
import { adminFetchTCGs, adminCreateTCG, adminUpdateTCG, adminDeleteTCG } from '@/lib/api';

type Props = Record<string, never>;

export default function TCGManager({ }: Props) {
  const [tcgs, setTcgs] = useState<TCG[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isAdding, setIsAdding] = useState(false);
  const [newTcg, setNewTcg] = useState({ id: '', name: '' });
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const loadTCGs = async () => {
    setLoading(true);
    try {
      const data = await adminFetchTCGs();
      setTcgs(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTCGs();
  }, []);

  const handleToggle = async (tcg: TCG) => {
    try {
      await adminUpdateTCG(tcg.id, tcg.name, !tcg.is_active);
      setTcgs(prev => prev.map(t => t.id === tcg.id ? { ...t, is_active: !t.is_active } : t));
    } catch (err: unknown) {
      setError('Failed to toggle TCG: ' + (err instanceof Error ? err.message : String(err)));
    }
  };

  const handleRename = async (tcg: TCG) => {
    if (!editingName.trim() || editingName === tcg.name) {
      setEditingId(null);
      return;
    }

    try {
      await adminUpdateTCG(tcg.id, editingName, tcg.is_active);
      setTcgs(prev => prev.map(t => t.id === tcg.id ? { ...t, name: editingName } : t));
      setEditingId(null);
    } catch (err: unknown) {
      setError('Failed to rename TCG: ' + (err instanceof Error ? err.message : String(err)));
    }
  };

  const handleCreate = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    if (!newTcg.id || !newTcg.name) return;

    try {
      const id = newTcg.id.toLowerCase().replace(/[^a-z0-9]/g, '');
      const created = await adminCreateTCG(id, newTcg.name);
      setTcgs(prev => [...prev, created].sort((a, b) => a.name.localeCompare(b.name)));
      setNewTcg({ id: '', name: '' });
      setIsAdding(false);
    } catch (err: unknown) {
      setError('Failed to create TCG: ' + (err instanceof Error ? err.message : String(err)));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await adminDeleteTCG(id);
      setTcgs(prev => prev.filter(t => t.id !== id));
      setConfirmDeleteId(null);
    } catch (err: unknown) {
      setError('Failed to delete TCG: ' + (err instanceof Error ? err.message : String(err)));
      setConfirmDeleteId(null);
    }
  };

  if (loading && tcgs.length === 0) {
    return (
      <div className="p-8 text-center text-[#8b7355] animate-pulse font-bold">
        UNPACKING TCG REGISTRY...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-end mb-4">
        <div>
          <h3 className="text-2xl font-display text-ink-deep uppercase tracking-tighter m-0 leading-none">Registered Systems</h3>
          <p className="text-text-muted text-[10px] font-mono-stack uppercase font-bold mt-1 opacity-70">Enable or disable game systems from the warehouse</p>
        </div>
        <button
          onClick={() => setIsAdding(!isAdding)}
          className={`px-4 py-2 font-mono-stack text-[10px] font-bold uppercase transition-all flex items-center gap-2 ${
            isAdding 
              ? 'bg-red-50 text-red-600 border border-red-200 hover:bg-red-100' 
              : 'btn-secondary shadow-sm'
          }`}
        >
          {isAdding ? '✕ CANCEL' : '＋ REGISTER NEW TCG'}
        </button>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-xs font-mono-stack font-bold mb-4 rounded flex justify-between items-center animate-in fade-in zoom-in-95">
          <span>⚠️ {error}</span>
          <button onClick={() => setError('')} className="opacity-50 hover:opacity-100 transition-opacity">✕</button>
        </div>
      )}

      {isAdding && (
        <form onSubmit={handleCreate} className="p-6 bg-white border border-kraft-dark/20 shadow-xl shadow-kraft-dark/5 rounded-xl mb-8 animate-in slide-in-from-top-4 duration-300">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div>
              <label className="block text-[10px] font-mono-stack font-bold text-text-muted uppercase mb-2">Internal Slug (URL ID)</label>
              <input
                value={newTcg.id}
                onChange={e => setNewTcg(p => ({ ...p, id: e.target.value }))}
                placeholder="e.g. starwars, disney-lorcana"
                className="w-full bg-kraft-paper/30 border border-kraft-dark/20 p-3 text-ink-deep text-sm focus:border-gold focus:ring-1 focus:ring-gold outline-none rounded-lg font-bold"
                required
              />
            </div>
            <div>
              <label className="block text-[10px] font-mono-stack font-bold text-text-muted uppercase mb-2">Display Name</label>
              <input
                value={newTcg.name}
                onChange={e => setNewTcg(p => ({ ...p, name: e.target.value }))}
                placeholder="e.g. Star Wars: Unlimited"
                className="w-full bg-kraft-paper/30 border border-kraft-dark/20 p-3 text-ink-deep text-sm focus:border-gold focus:ring-1 focus:ring-gold outline-none rounded-lg font-bold"
                required
              />
            </div>
          </div>
          <button
            type="submit"
            className="w-full py-3 btn-primary shadow-lg shadow-gold/20 font-bold uppercase tracking-widest text-xs"
          >
            CONFIRM SYSTEM REGISTRATION
          </button>
        </form>
      )}

      <div className="grid grid-cols-1 gap-4">
        {tcgs.map(tcg => (
          <div
            key={tcg.id}
            className={`flex items-center justify-between p-4 bg-white border border-kraft-dark/20 shadow-sm rounded-xl transition-all hover:shadow-md hover:border-gold/30 ${
              !tcg.is_active ? 'opacity-60 grayscale-[0.8] bg-kraft-light/20' : ''
            }`}
          >
            <div className="flex items-center gap-5 flex-1 min-w-0">
              <div 
                className={`w-3 h-3 rounded-full shrink-0 shadow-sm ${tcg.is_active ? 'bg-emerald-500 ring-4 ring-emerald-500/10' : 'bg-red-400 ring-4 ring-red-400/10'}`}
              />
              
              {editingId === tcg.id ? (
                <div className="flex gap-2 flex-1 max-w-md animate-in fade-in scale-95 duration-200">
                  <input
                    autoFocus
                    value={editingName}
                    onChange={e => setEditingName(e.target.value)}
                    className="flex-1 bg-kraft-paper/20 border border-gold p-2 text-ink-deep text-sm outline-none rounded font-bold"
                    onKeyDown={e => {
                      if (e.key === 'Enter') handleRename(tcg);
                      if (e.key === 'Escape') setEditingId(null);
                    }}
                  />
                  <button onClick={() => handleRename(tcg)} className="btn-primary px-3 text-[9px] font-bold uppercase">SAVE</button>
                  <button onClick={() => setEditingId(null)} className="btn-secondary px-3 text-[9px] font-bold uppercase">CANCEL</button>
                </div>
              ) : (
                <div className="min-w-0">
                  <div className="flex items-center gap-3 mb-1">
                    <h4 className="font-display text-xl text-ink-deep uppercase tracking-tight leading-none m-0">{tcg.name}</h4>
                    {tcg.item_count !== undefined && (
                      <span className="px-2 py-0.5 rounded-full bg-gold/10 text-gold-dark text-[9px] font-mono-stack border border-gold/10 font-bold">
                        {tcg.item_count} PRODUCTS
                      </span>
                    )}
                  </div>
                  <p className="text-[10px] font-mono-stack text-text-muted font-bold opacity-60">SYSTEM ID: <span className="text-ink-deep">{tcg.id}</span></p>
                </div>
              )}
            </div>

            <div className="flex items-center gap-4 pl-4 border-l border-kraft-dark/10">
              {editingId !== tcg.id ? (
                confirmDeleteId === tcg.id ? (
                  <div className="flex items-center gap-2 animate-in fade-in slide-in-from-right-2 duration-200">
                    <span className="text-[10px] font-mono-stack font-bold text-red-600 uppercase mr-1">ARE YOU SURE?</span>
                    <button
                      onClick={() => handleDelete(tcg.id)}
                      className="px-3 py-1.5 bg-red-600 text-white text-[9px] font-bold uppercase hover:bg-red-700 transition-all rounded"
                    >
                      CONFIRM
                    </button>
                    <button
                      onClick={() => setConfirmDeleteId(null)}
                      className="px-3 py-1.5 border border-kraft-dark/20 text-text-muted text-[9px] font-bold uppercase hover:bg-kraft-paper transition-all rounded"
                    >
                      CANCEL
                    </button>
                  </div>
                ) : (
                  <>
                    <button
                      onClick={() => { setEditingId(tcg.id); setEditingName(tcg.name); setConfirmDeleteId(null); }}
                      className="p-2 text-text-muted hover:text-gold transition-all hover:bg-gold/5 rounded-full"
                      title="Rename TCG"
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                        <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 00 2 2h14a2 2 0 00 2-2v-7M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                      </svg>
                    </button>

                    <button
                      onClick={() => handleToggle(tcg)}
                      className={`px-4 py-1.5 text-[9px] font-mono-stack font-bold uppercase border transition-all rounded-md ${
                        tcg.is_active 
                          ? 'border-red-200 text-red-600 hover:bg-red-50' 
                          : 'border-emerald-200 text-emerald-600 hover:bg-emerald-50'
                      }`}
                    >
                      {tcg.is_active ? 'DISABLE' : 'ENABLE'}
                    </button>
                    
                    <button
                      onClick={() => { setConfirmDeleteId(tcg.id); setEditingId(null); }}
                      disabled={tcg.item_count ? tcg.item_count > 0 : false}
                      className={`p-2 transition-all rounded-full ${
                        tcg.item_count && tcg.item_count > 0 
                          ? 'opacity-10 cursor-not-allowed text-text-muted' 
                          : 'text-text-muted hover:text-red-600 hover:bg-red-50'
                      }`}
                      title={tcg.item_count && tcg.item_count > 0 ? "Cannot delete TCG with active products" : "Delete TCG"}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                        <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2M10 11v6M14 11v6" />
                      </svg>
                    </button>
                  </>
                )
              ) : null}
            </div>
          </div>
        ))}
        {tcgs.length === 0 && !loading && (
          <div className="py-20 text-center border-2 border-dashed border-kraft-dark/10 rounded-2xl">
            <p className="font-mono-stack text-xs text-text-muted uppercase font-bold tracking-widest opacity-40">No Systems Registered</p>
          </div>
        )}
      </div>
    </div>
  );
}
