'use client';

import { useLanguage } from '@/context/LanguageContext';

interface Props {
  current: number;
  total: number;
  source: string;
  status: 'syncing' | 'success' | 'error';
  error?: string;
  onClose: () => void;
}

export default function BulkSyncProgressModal({ current, total, source, status, error, onClose }: Props) {
  const { t } = useLanguage();
  const progress = total > 0 ? (current / total) * 100 : 0;

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[1100] flex items-center justify-center p-4 animate-in fade-in duration-300">
      <div className="bg-white border-4 border-ink-border rounded-lg shadow-2xl max-w-md w-full overflow-hidden relative">
        {/* Top Accent Bar */}
        <div className="h-1.5 w-full bg-repeating-linear-gradient(45deg, var(--gold) 0, var(--gold) 10px, #2c251d 10px, #2c251d 20px)" 
             style={{ 
               backgroundImage: 'repeating-linear-gradient(45deg, #d4af37 0, #d4af37 10px, #1a1f2e 10px, #1a1f2e 20px)' 
             }} 
        />
        
        <div className="p-8">
          <div className="flex flex-col items-center text-center">
            {status === 'syncing' && (
              <>
                <div className="w-16 h-16 border-4 border-gold border-t-transparent rounded-full animate-spin mb-6" />
                <h2 className="text-2xl font-black text-ink-navy uppercase italic mb-2 tracking-tight">
                  {t('components.admin.sync_modal.title', 'SYNC IN PROGRESS')}
                </h2>
                <p className="text-sm text-text-muted font-bold uppercase tracking-widest mb-8 opacity-70">
                  {t('components.admin.sync_modal.subtitle', 'Synchronizing {count} items with {source}', { count: total, source: source.toUpperCase() })}
                </p>

                {/* Progress Bar Container */}
                <div className="w-full">
                  <div className="flex justify-between items-end mb-2">
                    <span className="text-[10px] font-black text-gold uppercase tracking-[0.2em]">
                      {t('components.admin.sync_modal.progress_label', 'Master Sync Status')}
                    </span>
                    <span className="text-xs font-mono-stack font-bold text-ink-navy">
                      {current} / {total} ({Math.round(progress)}%)
                    </span>
                  </div>
                  <div className="h-4 bg-ink-surface border-2 border-ink-border rounded-full overflow-hidden relative shadow-inner">
                    <div 
                      className="absolute inset-y-0 left-0 bg-gold transition-all duration-300 ease-out shadow-[0_0_15px_rgba(212,175,55,0.6)]"
                      style={{ width: `${progress}%` }}
                    />
                    <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-[shimmer_2s_infinite]" 
                         style={{ backgroundSize: '200% 100%' }} />
                  </div>
                </div>
              </>
            )}

            {status === 'success' && (
              <>
                <div className="w-20 h-20 bg-green-100 text-green-600 rounded-full flex items-center justify-center text-4xl mb-6 shadow-sm border-4 border-green-200 animate-in zoom-in duration-300">
                  ✓
                </div>
                <h2 className="text-2xl font-black text-ink-navy uppercase italic mb-2 tracking-tight">
                  {t('components.admin.sync_modal.success_title', 'SYNC COMPLETE')}
                </h2>
                <p className="text-sm text-text-muted font-bold uppercase tracking-widest mb-8 opacity-70">
                  {t('components.admin.sync_modal.success_subtitle', 'Successfully updated {count} products', { count: total })}
                </p>
                <button 
                  onClick={onClose}
                  className="btn-primary w-full py-4 shadow-lg shadow-gold/20 italic"
                >
                  {t('common.actions.close', 'RETURN TO DASHBOARD')}
                </button>
              </>
            )}

            {status === 'error' && (
              <>
                <div className="w-20 h-20 bg-red-100 text-red-600 rounded-full flex items-center justify-center text-4xl mb-6 shadow-sm border-4 border-red-200 animate-in bounce-in duration-300">
                  !
                </div>
                <h2 className="text-2xl font-black text-red-600 uppercase italic mb-2 tracking-tight">
                  {t('components.admin.sync_modal.error_title', 'SYNC FAILURE')}
                </h2>
                <p className="text-sm text-hp-color font-bold uppercase tracking-widest mb-4 opacity-70">
                  {t('components.admin.sync_modal.error_subtitle', 'Operations interrupted')}
                </p>
                <div className="bg-red-50 border border-red-200 p-4 rounded text-left w-full mb-8">
                  <p className="text-xs font-mono-stack text-red-800 break-words">{error || 'Unknown error occurred during batch processing'}</p>
                </div>
                <button 
                  onClick={onClose}
                  className="btn-secondary w-full py-4 italic border-hp-color text-hp-color hover:bg-hp-color hover:text-white"
                >
                  {t('common.actions.close', 'CLOSE')}
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
