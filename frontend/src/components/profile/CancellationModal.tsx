'use client';

import React, { useState } from 'react';
import { ClientRequest } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

interface CancellationModalProps {
  request: ClientRequest;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (reason: string, details: string) => void;
  isSubmitting: boolean;
}

export default function CancellationModal({ request, isOpen, onClose, onConfirm, isSubmitting }: CancellationModalProps) {
  const { t } = useLanguage();

  const REASONS = [
    t('components.cancellation_modal.reasons.found', 'Found it elsewhere'),
    t('components.cancellation_modal.reasons.not_needed', 'No longer need it'),
    t('components.cancellation_modal.reasons.price', 'Price too high'),
    t('components.cancellation_modal.reasons.mind', 'Changed my mind'),
    t('components.cancellation_modal.reasons.other', 'Other')
  ];

  const [reason, setReason] = useState(REASONS[0]);
  const [details, setDetails] = useState('');

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="w-full max-w-md bg-bg-page border border-border-main rounded-xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="p-6">
          <h2 className="text-xl font-bold text-text-main mb-2">{t('components.cancellation_modal.title', 'Discard Request')}</h2>
          <p className="text-text-secondary text-sm mb-6">
            {t('components.cancellation_modal.desc', 'You are choosing to **discard** your request for {card} as no longer needed. Please let us know why.')
              .replace('{card}', request.card_name)}
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-1">{t('components.cancellation_modal.reason_label', 'Reason')}</label>
              <select
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                className="w-full bg-bg-surface border border-border-main rounded-lg px-3 py-2 text-text-main focus:outline-none focus:border-accent-primary transition-colors"
                disabled={isSubmitting}
              >
                {REASONS.map(r => (
                  <option key={r} value={r}>{r}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-secondary mb-1">{t('components.cancellation_modal.details_label', 'Additional Details (Optional)')}</label>
              <textarea
                value={details}
                onChange={(e) => setDetails(e.target.value)}
                placeholder={t('components.cancellation_modal.details_placeholder', "Anything else you'd like to tell us...")}
                className="w-full bg-bg-surface border border-border-main rounded-lg px-3 py-2 text-text-main h-24 resize-none focus:outline-none focus:border-accent-primary transition-colors"
                disabled={isSubmitting}
              />
            </div>
          </div>
        </div>

        <div className="p-4 bg-bg-surface/50 border-t border-border-main flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={isSubmitting}
            className="px-4 py-2 text-sm font-medium text-text-secondary hover:text-text-main hover:bg-bg-surface/80 rounded-lg transition-all duration-200 disabled:opacity-50"
          >
            {t('components.cancellation_modal.back_button', 'Go Back')}
          </button>
          <button
            onClick={() => onConfirm(reason, details)}
            disabled={isSubmitting}
            className="px-6 py-2 bg-red-600 hover:bg-red-500 text-white text-sm font-bold rounded-lg shadow-lg shadow-red-900/20 hover:shadow-red-900/40 hover:-translate-y-0.5 transition-all duration-200 active:scale-95 active:translate-y-0 disabled:opacity-50 disabled:scale-100 disabled:translate-y-0 flex items-center gap-2"
          >
            {isSubmitting ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                {t('components.cancellation_modal.submitting', 'Discarding...')}
              </>
            ) : (
              t('components.cancellation_modal.confirm_button', 'Confirm Discard')
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
