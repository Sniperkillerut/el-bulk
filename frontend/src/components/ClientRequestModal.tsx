'use client';

import { createClientRequest } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import { useForm } from '@/hooks/useForm';
import { useLanguage } from '@/context/LanguageContext';
import { useUser } from '@/context/UserContext';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
  const router = useRouter();
  const { t } = useLanguage();
  const { user } = useUser();
  const {
    form,
    handleChange,
    setFieldValue,
    submitting,
    error,
    handleSubmit
  } = useForm({
    customer_name: user ? `${user.first_name} ${user.last_name || ''}`.trim() : '',
    customer_contact: user ? (user.email || user.phone || '') : '',
    card_name: '',
    set_name: '',
    details: ''
  });

  useEffect(() => {
    if (user && !form.customer_name) {
      setFieldValue('customer_name', `${user.first_name} ${user.last_name || ''}`.trim());
      setFieldValue('customer_contact', user.email || user.phone || '');
    }
  }, [user, setFieldValue, form.customer_name]);

  const onSubmit = async (data: Record<string, string>) => {
    await createClientRequest(data as unknown as import('@/lib/types').ClientRequestInput);
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
        {!user && (
          <div className="bg-accent-primary/10 border border-accent-primary/20 p-3 rounded-md mb-2 flex items-center justify-between gap-4 animate-fade-in">
            <div className="text-[11px] text-text-main leading-tight">
              <strong>{t('pages.common.labels.login', 'Login')}</strong> {t('components.client_request_modal.login_prompt', 'to automatically fill your info and track your request status.')}
            </div>
            <button 
              type="button"
              onClick={() => router.push('/login')}
              className="btn-primary text-[10px] px-3 py-1.5 whitespace-nowrap font-bold"
            >
              {t('pages.common.buttons.login', 'LOGIN')}
            </button>
          </div>
        )}
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
