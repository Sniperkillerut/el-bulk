'use client';

import { StoredIn } from '@/lib/types';
import { useState } from 'react';

interface StorageManagerModalProps {
  storageLocations: StoredIn[];
  onCreate: (name: string) => Promise<void>;
  onUpdate: (id: string, name: string) => Promise<void>;
  onDelete: (id: string, name: string, count: number) => Promise<void>;
  onClose: () => void;
}

export default function StorageManagerModal({
  storageLocations,
  onCreate,
  onUpdate,
  onDelete,
  onClose
}: StorageManagerModalProps) {
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');

  const handleCreate = async () => {
    if (!newName.trim()) return;
    await onCreate(newName);
    setNewName('');
  };

  const handleUpdate = async (id: string) => {
    if (!editingName.trim()) return;
    await onUpdate(id, editingName);
    setEditingId(null);
  };

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center px-4"
      style={{ background: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(5px)' }}>
      <div className="card max-w-2xl w-full p-8" style={{ background: 'var(--ink-surface)', border: '4px solid var(--kraft-dark)' }}>
        <div className="flex items-center justify-between mb-8">
          <h2 className="font-display text-4xl m-0">STORAGE LOCATIONS</h2>
          <button onClick={onClose} className="text-text-muted hover:text-text-primary text-xl">✕</button>
        </div>
        
        <div className="flex gap-2 mb-6">
          <input 
            type="text" 
            placeholder="New Location Name (e.g. Binder A)" 
            value={newName} 
            onChange={e => setNewName(e.target.value)} 
            className="flex-1 bg-white" 
          />
          <button onClick={handleCreate} className="btn-primary px-6">ADD</button>
        </div>

        <div className="space-y-2 max-h-96 overflow-y-auto pr-2">
          {storageLocations.map(loc => (
            <div key={loc.id} className="flex items-center justify-between p-3 border border-kraft-dark bg-kraft-light/10">
              {editingId === loc.id ? (
                <div className="flex gap-2 flex-1 mr-4">
                  <input 
                    type="text" 
                    value={editingName} 
                    onChange={e => setEditingName(e.target.value)} 
                    className="flex-1 py-1 bg-white" 
                  />
                  <button onClick={() => handleUpdate(loc.id)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                  <button onClick={() => setEditingId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                </div>
              ) : (
                <>
                  <div className="flex items-center gap-3">
                    <span className="font-semibold text-lg">{loc.name}</span>
                    <span className="text-xs font-mono-stack text-text-muted bg-kraft-light px-2 py-0.5 rounded border border-kraft-dark">
                      {loc.item_count || 0} items
                    </span>
                  </div>
                  <div className="flex gap-2">
                    <button 
                      onClick={() => { setEditingId(loc.id); setEditingName(loc.name); }} 
                      className="btn-secondary px-3 py-1 text-xs"
                    >EDIT</button>
                    <button 
                      onClick={() => onDelete(loc.id, loc.name, loc.item_count || 0)} 
                      className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" 
                      style={{ borderRadius: 4 }}
                    >DELETE</button>
                  </div>
                </>
              )}
            </div>
          ))}
          {storageLocations.length === 0 && <p className="text-center text-text-muted py-8">No storage locations configured.</p>}
        </div>
      </div>
    </div>
  );
}
