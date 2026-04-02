'use client';

import { useLanguage } from '@/context/LanguageContext';

export default function BountiesTitle() {
  const { t } = useLanguage();

  return (
    <h2 className="font-display text-4xl uppercase" style={{ color: 'var(--text-main)' }}>
      {t('home.sections.wanted', 'WANTED')} / <span style={{ color: 'var(--status-hp)' }}>{t('home.sections.bounties', 'BOUNTIES')}</span>
    </h2>
  );
}
