'use client';

import { useState } from 'react';
import { Bounty } from '@/lib/types';
import ClientRequestModal from '@/components/ClientRequestModal';
import BountyCard from '@/components/BountyCard';
import { useLanguage } from '@/context/LanguageContext';

export default function PublicBountiesClient({ initialBounties }: { initialBounties: Bounty[] }) {
  const [bounties] = useState<Bounty[]>(initialBounties);
  const [showModal, setShowModal] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const { t } = useLanguage();

  const handleSuccess = () => {
    setShowModal(false);
    setSuccessMsg(t('pages.bounties.success.request', 'Your request was submitted successfully! We will contact you if we locate the card.'));
    setTimeout(() => setSuccessMsg(''), 5000);
  };

  return (
    <>
      <div className="max-w-4xl mx-auto text-center mb-16 animate-in slide-in-from-bottom-5 fade-in duration-700">
        <div className="inline-block px-3 py-1 bg-hp-color text-white font-mono-stack text-[10px] mb-6 rotate-[-1deg] uppercase tracking-[0.3em] shadow-sm">
          {t('pages.bounties.tags.wanted', 'URGENT LOGISTICS / HIGH PRIORITY')}
        </div>
        <h1 className="font-display text-fluid-h1 tracking-tighter text-text-main leading-[0.85] mb-8 drop-shadow-sm uppercase">
          {t('pages.bounties.page.title', 'WANTED / BOUNTIES')}
        </h1>
        <div className="relative inline-block">
          <p className="text-text-secondary text-lg md:text-xl max-w-2xl mx-auto tracking-tight opacity-90 font-medium">
            {t('pages.bounties.page.subtitle', "We are actively looking to buy the cards below. If you have them, reach out to us! Can't find what you are looking for? Send us a card request!")}
          </p>
          <div className="absolute -bottom-4 left-1/2 -translate-x-1/2 w-24 h-1 bg-accent-primary/30" />
        </div>
      </div>

      <div className="flex flex-col md:flex-row justify-between items-center mb-10 gap-6 animate-in fade-in duration-1000 delay-150 fill-mode-both">
        <p className="font-mono-stack text-[11px] uppercase font-bold tracking-[0.2em] text-text-muted">
          {t('pages.bounties.status.showing', 'SHOWING {count} ENTRIES').replace('{count}', bounties.length.toString())}
        </p>
        <button onClick={() => setShowModal(true)} className="btn-primary py-3.5 px-10 xl:w-80 shadow-xl border-accent-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all duration-300 font-display text-xl tracking-tight">
          {t('pages.bounties.buttons.request', 'REQUEST A CARD')}
        </button>
      </div>

      {successMsg && (
        <div className="mb-8 p-4 bg-emerald-100 text-emerald-800 border-l-4 border-emerald-500 font-bold text-center animate-in slide-in-from-top fade-in duration-300">
          ✓ {successMsg}
        </div>
      )}

      {bounties.length === 0 ? (
        <div className="py-24 text-center border-2 border-dashed border-border-main/50 rounded-xl bg-bg-surface/40 backdrop-blur-sm">
           <h3 className="font-display text-3xl text-text-main mb-2">{t('pages.bounties.empty.title', 'NO ACTIVE BOUNTIES')}</h3>
           <p className="text-text-muted font-medium">{t('pages.bounties.empty.desc', 'We are not currently explicitly searching for any specific cards. But you can still send us a request!')}</p>
           <button onClick={() => setShowModal(true)} className="mt-8 btn-secondary px-8 py-3 font-bold border-2">{t('pages.bounties.buttons.submit_request', 'SUBMIT REQUEST')}</button>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-6">
          {bounties.map((b, idx) => (
            <BountyCard key={b.id} bounty={b} delay={idx * 50} />
          ))}
        </div>
      )}

      {showModal && <ClientRequestModal onClose={() => setShowModal(false)} onSuccess={handleSuccess} />}
    </>
  );
}
