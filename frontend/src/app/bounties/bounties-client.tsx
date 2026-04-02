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
        <h1 className="font-display text-5xl md:text-7xl xl:text-8xl tracking-tighter text-ink-deep leading-[0.85] mb-6">
          {t('pages.bounties.page.title', 'WANTED / BOUNTIES').split('/').map((word, i) => (
            <span key={i}>
              {i > 0 && <span className="text-gold"> / </span>}
              {word}
              {i === 0 && <br/>}
            </span>
          ))}
        </h1>
        <p className="text-text-secondary md:text-lg max-w-2xl mx-auto tracking-wide">
          {t('pages.bounties.page.subtitle', "We are actively looking to buy the cards below. If you have them, reach out to us! Can't find what you are looking for? Send us a card request!")}
        </p>
      </div>

      <div className="flex flex-col md:flex-row justify-between items-center mb-8 gap-6 animate-in fade-in duration-1000 delay-150 fill-mode-both">
        <p className="font-mono-stack text-sm tracking-widest text-text-muted">
          {t('pages.bounties.status.showing', 'SHOWING {count} ENTRIES').replace('{count}', bounties.length.toString())}
        </p>
        <button onClick={() => setShowModal(true)} className="btn-primary py-3 px-8 xl:w-80 shadow-gold/20 hover:shadow-gold/40 border-gold/40 bg-ink-surface text-ink-deep font-display tracking-tight text-xl transition-all duration-300">
          {t('pages.bounties.buttons.request', 'REQUEST A CARD')}
        </button>
      </div>

      {successMsg && (
        <div className="mb-8 p-4 bg-emerald-100 text-emerald-800 border-l-4 border-emerald-500 font-bold text-center animate-in slide-in-from-top fade-in duration-300">
          ✓ {successMsg}
        </div>
      )}

      {bounties.length === 0 ? (
        <div className="py-24 text-center border-2 border-dashed border-ink-border/30 rounded-xl bg-white/40">
           <h3 className="font-display text-3xl text-ink-deep mb-2">{t('pages.bounties.empty.title', 'NO ACTIVE BOUNTIES')}</h3>
           <p className="text-text-muted">{t('pages.bounties.empty.desc', 'We are not currently explicitly searching for any specific cards. But you can still send us a request!')}</p>
           <button onClick={() => setShowModal(true)} className="mt-6 btn-secondary px-8 font-bold">{t('pages.bounties.buttons.submit_request', 'SUBMIT REQUEST')}</button>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {bounties.map((b, idx) => (
            <BountyCard key={b.id} bounty={b} delay={idx * 50} />
          ))}
        </div>
      )}

      {showModal && <ClientRequestModal onClose={() => setShowModal(false)} onSuccess={handleSuccess} />}
    </>
  );
}
