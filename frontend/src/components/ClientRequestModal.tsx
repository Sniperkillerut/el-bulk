'use client';

import { createClientRequest } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import { useForm } from '@/hooks/useForm';
import { useLanguage } from '@/context/LanguageContext';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
  const { t } = useLanguage();
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

  const onSubmit = async (data: Record<string, string>) => {
    await createClientRequest(data as any); // Casting here if createClientRequest expects a specific type but we use a generic record
    onSuccess();
  };

  return (
    <Modal isOpen={true} onClose={onClose} title={t('components.client_request_modal.title', 'Request a card')}>
      <p className="text-xs text-text-muted mb-6 uppercase tracking-widest leading-relaxed">
        {t('components.client_request_modal.desc', "Can't find what you need? Tell us the details and we'll start the hunt!")}
      </p>

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded mb-4">
          {error}
        </div>
      )}
      
      <form onSubmit={(e) => { e.preventDefault(); handleSubmit(onSubmit); }} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.name_label', 'Your Name *')}</label>
            <input 
              name="customer_name"
              type="text" 
              className="w-full text-sm" 
              required 
              value={form.customer_name} 
              onChange={handleChange} 
              placeholder={t('components.client_request_modal.form.name_placeholder', 'John Doe')}
            />
          </div>
          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.contact_label', 'Contact Info *')}</label>
            <input 
              name="customer_contact"
              type="text" 
              className="w-full text-sm" 
              required 
              value={form.customer_contact} 
              onChange={handleChange} 
              placeholder={t('components.client_request_modal.form.contact_placeholder', 'Phone or Instagram')}
            />
          </div>
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.card_label', 'Card Name *')}</label>
          <input 
            name="card_name"
            type="text" 
            className="w-full text-sm" 
            required 
            value={form.card_name} 
            onChange={handleChange} 
            placeholder={t('components.client_request_modal.form.card_placeholder', 'e.g. Sheoldred, the Apocalypse')}
          />
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.set_label', 'Specific Set (Optional)')}</label>
          <input 
            name="set_name"
            type="text" 
            className="w-full text-sm" 
            value={form.set_name} 
            onChange={handleChange} 
            placeholder={t('components.client_request_modal.form.set_placeholder', 'e.g. Dominaria United')} 
          />
        </div>
        
        <div>
          <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.details_label', 'Additional Details')}</label>
          <textarea 
            name="details"
            className="w-full text-sm resize-none" 
            rows={3}
            value={form.details} 
            onChange={handleChange} 
            placeholder={t('components.client_request_modal.form.details_placeholder', 'Condition, foil, language, etc...')} 
          />
        </div>
        
        <Button 
          type="submit" 
          loading={submitting} 
          fullWidth 
          size="lg" 
          className="mt-4"
        >
          {t('components.client_request_modal.form.submit_btn', 'SUBMIT MISSION')}
        </Button>
      </form>
    </Modal>
  );
}
