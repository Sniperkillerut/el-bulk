'use client';

import { createClientRequest } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import { useForm } from '@/hooks/useForm';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
  const {
    form,
    handleChange,
    submitting,
    error,
    handleSubmit
  } = useForm({
    customer_name: '',
    customer_contact: '',
    card_name: '',
    set_name: '',
    details: ''
  });

  const onSubmit = async (data: any) => {
    await createClientRequest(data);
    onSuccess();
  };

  return (
    <Modal isOpen={true} onClose={onClose} title="Request a card">
      <p className="text-xs text-text-muted mb-6 uppercase tracking-widest leading-relaxed">
        Can't find what you need? Tell us the details and we'll start the hunt!
      </p>

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded mb-4">
          {error}
        </div>
      )}
      
      <form onSubmit={(e) => { e.preventDefault(); handleSubmit(onSubmit); }} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Your Name *</label>
            <input 
              name="customer_name"
              type="text" 
              className="w-full text-sm" 
              required 
              value={form.customer_name} 
              onChange={handleChange} 
              placeholder="John Doe"
            />
          </div>
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Contact Info *</label>
            <input 
              name="customer_contact"
              type="text" 
              className="w-full text-sm" 
              required 
              value={form.customer_contact} 
              onChange={handleChange} 
              placeholder="Phone or Instagram"
            />
          </div>
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Card Name *</label>
          <input 
            name="card_name"
            type="text" 
            className="w-full text-sm" 
            required 
            value={form.card_name} 
            onChange={handleChange} 
            placeholder="e.g. Sheoldred, the Apocalypse"
          />
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Specific Set (Optional)</label>
          <input 
            name="set_name"
            type="text" 
            className="w-full text-sm" 
            value={form.set_name} 
            onChange={handleChange} 
            placeholder="e.g. Dominaria United" 
          />
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">Additional Details</label>
          <textarea 
            name="details"
            className="w-full text-sm resize-none" 
            rows={3}
            value={form.details} 
            onChange={handleChange} 
            placeholder="Condition, foil, language, etc..." 
          />
        </div>
        
        <Button 
          type="submit" 
          loading={submitting} 
          fullWidth 
          size="lg" 
          className="mt-4"
        >
          SUBMIT MISSION
        </Button>
      </form>
    </Modal>
  );
}
