'use client';

import { createClientRequest } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import { useForm } from '@/hooks/useForm';
import { useLanguage } from '@/context/LanguageContext';
import { useUser } from '@/context/UserContext';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import ScryfallVariantPicker from './ScryfallVariantPicker';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
  const router = useRouter();
  const { t } = useLanguage();
  const { user, loginWithGoogle } = useUser();
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
    quantity: '1',
    details: ''
  });

  const [selectedPrint, setSelectedPrint] = useState<any | null>(null);
  const [matchType, setMatchType] = useState<'any' | 'exact'>('any');

  useEffect(() => {
    if (user && !form.customer_name) {
      setFieldValue('customer_name', `${user.first_name} ${user.last_name || ''}`.trim());
      setFieldValue('customer_contact', user.email || user.phone || '');
    }
  }, [user, setFieldValue, form.customer_name]);

  const onSubmit = async (data: Record<string, string>) => {
    await createClientRequest({
      ...data,
      quantity: parseInt(data.quantity || '1', 10),
      match_type: matchType,
      scryfall_id: selectedPrint?.id,
      set_name: selectedPrint?.set_name || data.set_name,
    } as any);
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
      
      {!user ? (
        <div className="py-12 px-6 text-center animate-in fade-in zoom-in-95 duration-500">
           <div className="w-20 h-20 bg-accent-primary/10 rounded-full flex items-center justify-center text-4xl mx-auto mb-8 border border-accent-primary/20 shadow-lg shadow-accent-primary/5">
            🔍
          </div>
          <h4 className="text-2xl font-bold mb-4 text-text-main font-display">{t('components.client_request_modal.login_required.title', 'MISSION BRIEFING')}</h4>
          <p className="text-sm text-text-muted mb-10 leading-relaxed max-w-xs mx-auto">
            {t('components.client_request_modal.login_required.desc', "To start an acquisition mission, you'll need to sign in so we can contact you securely once your card is found.")}
          </p>
          <div className="flex flex-col gap-4">
            <button 
              onClick={loginWithGoogle} 
              className="w-full py-4 bg-accent-primary hover:bg-accent-primary-hover text-text-on-accent font-bold rounded-xl transition-all shadow-lg active:scale-95"
            >
              {t('pages.auth.login.google', 'Login with Google')}
            </button>
            <button 
              onClick={() => {
                onClose();
                router.push('/login');
              }} 
              className="text-xs font-mono text-text-muted hover:text-accent-primary uppercase tracking-widest transition-colors py-2"
            >
               {t('components.client_request_modal.other_methods', 'Other methods')}
            </button>
          </div>
        </div>
      ) : (
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
          
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2">
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
              <label className="block text-[10px] font-mono-stack uppercase mb-1 text-text-muted">{t('components.client_request_modal.form.quantity_label', 'Quantity')}</label>
              <input 
                name="quantity"
                type="number" 
                min="1"
                className="w-full text-sm" 
                value={form.quantity || '1'} 
                onChange={handleChange} 
              />
            </div>
          </div>

          {form.card_name.length >= 3 && (
            <div className="pt-2 border-t border-kraft-dark/10">
              <ScryfallVariantPicker 
                cardName={form.card_name}
                selectedId={selectedPrint?.id}
                onSelect={(print) => {
                  setSelectedPrint(print);
                  setMatchType(print ? 'exact' : 'any');
                }}
              />
            </div>
          )}
          
          <div className={`transition-all duration-300 overflow-hidden ${matchType === 'exact' ? 'max-h-0 opacity-0 pointer-events-none' : 'max-h-20 opacity-100'}`}>
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
            className="mt-4 font-bold uppercase tracking-widest shadow-lg shadow-gold/20"
          >
            {t('components.client_request_modal.form.submit_btn', 'SUBMIT MISSION')}
          </Button>
        </form>
      )}
    </Modal>
  );
}
