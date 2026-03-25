'use client';

import { useState, useEffect } from 'react';
import { TCG } from '@/lib/types';
import { adminFetchTCGs, adminCreateTCG, adminUpdateTCG, adminDeleteTCG } from '@/lib/api';

interface Props {
  token: string;
}

export default function TCGManager({ token }: Props) {
  const [tcgs, setTcgs] = useState<TCG[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isAdding, setIsAdding] = useState(false);
  const [newTcg, setNewTcg] = useState({ id: '', name: '' });

  const loadTCGs = async () => {
    setLoading(true);
    try {
      const data = await adminFetchTCGs(token);
      setTcgs(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTCGs();
  }, [token]);

  const handleToggle = async (tcg: TCG) => {
    try {
      await adminUpdateTCG(token, tcg.id, tcg.name, !tcg.is_active);
      setTcgs(prev => prev.map(t => t.id === tcg.id ? { ...t, is_active: !t.is_active } : t));
    } catch (err: any) {
      setError('Failed to toggle TCG: ' + err.message);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTcg.id || !newTcg.name) return;

    try {
      const created = await adminCreateTCG(token, newTcg.id.toLowerCase().replace(/[^a-z0-9]/g, ''), newTcg.name);
      setTcgs(prev => [...prev, created].sort((a, b) => a.name.localeCompare(b.name)));
      setNewTcg({ id: '', name: '' });
      setIsAdding(false);
    } catch (err: any) {
      setError('Failed to create TCG: ' + err.message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this TCG? This will fail if there are products assigned to it.')) return;

    try {
      await adminDeleteTCG(token, id);
      setTcgs(prev => prev.filter(t => t.id !== id));
    } catch (err: any) {
      setError('Failed to delete TCG: ' + err.message);
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
      <div className="flex justify-between items-center mb-4">
        <div>
          <h3 className="text-xl font-black text-[#d4c3b3] uppercase tracking-tighter">TCG Registry</h3>
          <p className="text-[#8b7355] text-xs font-bold uppercase">Enable or disable game systems from the warehouse</p>
        </div>
        <button
          onClick={() => setIsAdding(!isAdding)}
          className="px-4 py-2 bg-[#3c2a21] text-[#d4c3b3] border-2 border-[#d4c3b3]/20 font-black text-xs uppercase hover:bg-[#d4c3b3] hover:text-[#3c2a21] transition-all"
        >
          {isAdding ? 'CANCEL' : '+ ADD NEW TCG'}
        </button>
      </div>

      {error && (
        <div className="p-3 bg-[#b04b4b]/20 border-2 border-[#b04b4b] text-[#ffbaba] text-xs font-bold mb-4">
          ⚠️ {error}
          <button onClick={() => setError('')} className="float-right">✕</button>
        </div>
      )}

      {isAdding && (
        <form onSubmit={handleCreate} className="p-4 bg-[#2a1e17] border-2 border-[#3c2a21] mb-6 animate-in slide-in-from-top-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-[10px] font-black text-[#8b7355] uppercase mb-1">Internal Slug (URL ID)</label>
              <input
                value={newTcg.id}
                onChange={e => setNewTcg(p => ({ ...p, id: e.target.value }))}
                placeholder="e.g. starwars, disney-lorcana"
                className="w-full bg-[#1a1614] border-2 border-[#3c2a21] p-2 text-[#d4c3b3] text-sm focus:border-[#d4c3b3] outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-[10px] font-black text-[#8b7355] uppercase mb-1">Display Name</label>
              <input
                value={newTcg.name}
                onChange={e => setNewTcg(p => ({ ...p, name: e.target.value }))}
                placeholder="e.g. Star Wars: Unlimited"
                className="w-full bg-[#1a1614] border-2 border-[#3c2a21] p-2 text-[#d4c3b3] text-sm focus:border-[#d4c3b3] outline-none"
                required
              />
            </div>
          </div>
          <button
            type="submit"
            className="w-full py-2 bg-[#8b7355] text-white font-black uppercase text-xs hover:bg-[#d4c3b3] hover:text-[#3c2a21] transition-all"
          >
            REGISTER GAME SYSTEM
          </button>
        </form>
      )}

      <div className="grid grid-cols-1 gap-3">
        {tcgs.map(tcg => (
          <div
            key={tcg.id}
            className={`flex items-center justify-between p-4 border-2 transition-all ${
              tcg.is_active ? 'bg-[#1a1614] border-[#3c2a21]' : 'bg-[#1a1614]/50 border-[#3c2a21]/50 grayscale'
            }`}
          >
            <div className="flex items-center gap-4">
              <div 
                className={`w-3 h-3 rounded-full ${tcg.is_active ? 'bg-[#7eb07e] shadow-[0_0_10px_rgba(126,176,126,0.5)]' : 'bg-[#b04b4b]'}`}
              />
              <div>
                <h4 className="font-black text-[#d4c3b3] uppercase tracking-tight">{tcg.name}</h4>
                <p className="text-[10px] font-mono text-[#8b7355]">SLUG: {tcg.id}</p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => handleToggle(tcg)}
                className={`px-3 py-1.5 text-[10px] font-black uppercase border-2 transition-all ${
                  tcg.is_active 
                    ? 'border-[#b04b4b] text-[#b04b4b] hover:bg-[#b04b4b] hover:text-white' 
                    : 'border-[#7eb07e] text-[#7eb07e] hover:bg-[#7eb07e] hover:text-white'
                }`}
              >
                {tcg.is_active ? 'DISABLE' : 'ENABLE'}
              </button>
              
              <button
                onClick={() => handleDelete(tcg.id)}
                className="p-1.5 text-[#8b7355] hover:text-[#b04b4b] transition-colors"
                title="Delete TCG"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2M10 11v6M14 11v6" />
                </svg>
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
