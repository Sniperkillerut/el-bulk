'use client';

import { createClientRequest } from '@/lib/api';
import Modal from './ui/Modal';
import Button from './ui/Button';
import { useForm } from '@/hooks/useForm';
import { useLanguage } from '@/context/LanguageContext';
import { useUser } from '@/context/UserContext';
import { useEffect, useState } from 'react';
import ScryfallVariantPicker, { type ScryfallPrint } from './ScryfallVariantPicker';

interface ClientRequestModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function ClientRequestModal({ onClose, onSuccess }: ClientRequestModalProps) {
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

  const [selectedPrint, setSelectedPrint] = useState<ScryfallPrint | null>(null);
  const [suggestion, setSuggestion] = useState<ScryfallPrint | null>(null);
  const [matchType, setMatchType] = useState<'any' | 'exact'>('any');

  useEffect(() => {
    if (user && !form.customer_name) {
      setFieldValue('customer_name', `${user.first_name} ${user.last_name || ''}`.trim());
      setFieldValue('customer_contact', user.email || user.phone || '');
    }
  }, [user, setFieldValue, form.customer_name]);

  const onSubmit = async (data: Record<string, string>) => {
    const activePrint = selectedPrint || suggestion;
    await createClientRequest({
      customer_name: data.customer_name,
      customer_contact: data.customer_contact,
      card_name: activePrint?.name || data.card_name,
      quantity: parseInt(data.quantity || '1', 10),
      match_type: matchType,
      scryfall_id: selectedPrint?.id, // Only send ID if exactly selected
      set_name: selectedPrint?.set_name || data.set_name,
      set_code: selectedPrint?.set || '',
      collector_number: selectedPrint?.collector_number || '',
      image_url: activePrint?.image_uris?.normal || activePrint?.image_uris?.normal || activePrint?.card_faces?.[0]?.image_uris?.normal,
      foil_treatment: selectedPrint?.finishes?.includes('foil') ? 'foil' : 'non_foil',
      card_treatment: selectedPrint?.border_color === 'borderless' ? 'borderless' : (selectedPrint?.frame_effects?.includes('showcase') ? 'showcase' : 'normal'),
      details: data.details,
      tcg: 'mtg',
      oracle_id: activePrint?.oracle_id
    });
    onSuccess();
  };

  return (
    <Modal isOpen={true} onClose={onClose} title={t('components.client_request_modal.title', 'Acquisition Mission')}>
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
          </div>
        </div>
      ) : (
        <form onSubmit={(e) => { e.preventDefault(); handleSubmit(onSubmit); }} className="space-y-6">
          <p className="text-xs text-text-muted uppercase tracking-widest leading-relaxed border-b border-ink-border/10 pb-4">
            {t('components.client_request_modal.desc', "Specify the card you're hunting for and our scouts will track it down across all markets.")}
          </p>

          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded">
              {error}
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="block text-[10px] font-mono-stack uppercase mb-1.5 text-text-muted font-black tracking-tighter">{t('components.client_request_modal.form.card_label', 'CARD IDENTITY')}</label>
              <div className="relative">
                <input
                  name="card_name"
                  type="text"
                  className="w-full text-base font-bold py-3 pl-4 bg-ink-surface border-2 border-ink-border focus:border-gold transition-all"
                  required
                  autoFocus
                  value={form.card_name}
                  onChange={handleChange}
                  placeholder={t('components.client_request_modal.form.card_placeholder', 'e.g. Sheoldred, the Apocalypse')}
                />
                <div className="absolute right-4 top-1/2 -translate-y-1/2 text-xl opacity-30">🔍</div>
              </div>
            </div>

            {form.card_name.length >= 3 && (
              <div className="animate-in fade-in slide-in-from-top-2 duration-500">
                <ScryfallVariantPicker
                  cardName={form.card_name}
                  selectedId={selectedPrint?.id}
                  onSelect={(print) => {
                    setSelectedPrint(print);
                    setMatchType(print ? 'exact' : 'any');
                  }}
                  onSuggestion={(print) => setSuggestion(print)}
                />
              </div>
            )}
          </div>

          <div className="grid grid-cols-3 gap-4 pt-4 border-t border-ink-border/10">
            <div className="col-span-2">
              <label className="block text-[10px] font-mono-stack uppercase mb-1.5 text-text-muted font-black tracking-tighter">{t('components.client_request_modal.form.quantity_label', 'QUANTITY')}</label>
              <div className="flex items-center gap-2">
                {[1, 2, 4, 8].map(n => (
                  <button
                    key={n}
                    type="button"
                    onClick={() => setFieldValue('quantity', n.toString())}
                    className={`flex-1 py-2 text-xs font-mono-stack border-2 transition-all ${form.quantity === n.toString()
                      ? 'border-gold bg-gold/10 text-gold-dark font-black'
                      : 'border-ink-border bg-ink-surface hover:border-gold/30 text-text-muted'
                      }`}
                  >
                    {n}x
                  </button>
                ))}
                <input
                  name="quantity"
                  type="number"
                  min="1"
                  className="w-16 text-center text-sm py-2"
                  value={form.quantity || '1'}
                  onChange={handleChange}
                />
              </div>
            </div>
          </div>

          <div>
            <label className="block text-[10px] font-mono-stack uppercase mb-1.5 text-text-muted font-black tracking-tighter">{t('components.client_request_modal.form.details_label', 'INTEL & REQUIREMENTS (OPTIONAL)')}</label>
            <textarea
              name="details"
              className="w-full text-sm bg-ink-surface border-2 border-ink-border focus:border-gold p-3 min-h-[80px] resize-none"
              value={form.details}
              onChange={handleChange}
              placeholder={t('components.client_request_modal.form.details_placeholder', 'Specific condition, language, or other field intel...')}
            />
          </div>

          <Button
            type="submit"
            loading={submitting}
            fullWidth
            size="lg"
            className="py-4 font-black uppercase tracking-[0.2em] shadow-xl shadow-gold/20 text-lg disabled:opacity-50 disabled:grayscale"
            disabled={!suggestion && !selectedPrint}
          >
            {t('components.client_request_modal.form.submit_btn', 'SUBMIT MISSION')}
          </Button>
        </form>
      )}
    </Modal>
  );
}
