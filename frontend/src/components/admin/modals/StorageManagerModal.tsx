'use client';

import { StoredIn } from '@/lib/types';
import { useState } from 'react';
import Modal from '@/components/ui/Modal';

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
    if (newName.trim().toLowerCase() === 'pending') {
      alert("The name 'pending' is reserved for system use.");
      return;
    }
    await onCreate(newName);
    setNewName('');
  };

  const handleUpdate = async (id: string) => {
    if (!editingName.trim()) return;
    await onUpdate(id, editingName);
    setEditingId(null);
  };

  return (
    <Modal 
      isOpen={true} 
      onClose={onClose} 
      title="STORAGE LOCATIONS"
      maxWidth="max-w-2xl"
    >
      <div className="p-4 md:p-8">
        <div className="flex flex-col sm:flex-row gap-2 mb-6">
          <input 
            type="text" 
            placeholder="New Location Name (e.g. Binder A)" 
            value={newName} 
            onChange={e => setNewName(e.target.value)} 
            className="flex-1 bg-white border-border-main" 
          />
          <button onClick={handleCreate} className="btn-primary px-6">ADD</button>
        </div>

        <div className="space-y-2 max-h-96 overflow-y-auto pr-2 custom-scrollbar">
          {storageLocations.map(loc => (
            <div key={loc.id} className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 p-3 border border-border-main bg-bg-surface/50">
              {editingId === loc.id ? (
                <div className="flex gap-2 flex-1 mr-4">
                  <input 
                    type="text" 
                    value={editingName} 
                    onChange={e => setEditingName(e.target.value)} 
                    className="flex-1 py-1 bg-white border-border-main" 
                  />
                  <button onClick={() => handleUpdate(loc.id)} className="btn-primary px-3 py-1 text-xs">SAVE</button>
                  <button onClick={() => setEditingId(null)} className="btn-secondary px-3 py-1 text-xs">CANCEL</button>
                </div>
              ) : (
                <>
                  <div className="flex items-center gap-3">
                    <span className="font-semibold text-lg text-text-main">{loc.name}</span>
                    <span className="text-xs font-mono-stack text-text-muted bg-bg-page px-2 py-0.5 rounded border border-border-main">
                      {loc.item_count || 0} items
                    </span>
                  </div>
                  <div className="flex gap-2">
                    {loc.name !== 'pending' ? (
                      <>
                        <button 
                          onClick={() => { setEditingId(loc.id); setEditingName(loc.name); }} 
                          className="btn-secondary px-3 py-1 text-xs"
                        >EDIT</button>
                        <button 
                          onClick={() => onDelete(loc.id, loc.name, loc.item_count || 0)} 
                          className="px-3 py-1 text-xs border border-hp-color text-hp-color hover:bg-hp-color hover:text-white transition-colors" 
                          style={{ borderRadius: 4 }}
                        >DELETE</button>
                      </>
                    ) : (
                      <span className="text-[10px] font-mono-stack text-gold-dark uppercase font-bold px-2 py-1 bg-gold/10 border border-gold/20 rounded-md">
                        🔒 SYSTEM PROTECTED
                      </span>
                    )}
                  </div>
                </>
              )}
            </div>
          ))}
          {storageLocations.length === 0 && <p className="text-center text-text-muted py-8">No storage locations configured.</p>}
        </div>
      </div>
    </Modal>
  );
}
