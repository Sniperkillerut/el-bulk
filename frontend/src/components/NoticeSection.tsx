'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Notice } from '@/lib/types';
import { fetchNotices } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';

export default function NoticeSection() {
  const [notices, setNotices] = useState<Notice[]>([]);
  const [loading, setLoading] = useState(true);
  const { t, locale } = useLanguage();

  useEffect(() => {
    fetchNotices({ limit: 3 })
      .then(setNotices)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) return (
    <div className="space-y-4 animate-pulse">
      {[1, 2, 3].map(i => (
        <div key={i} className="h-40 bg-kraft-light rounded-sm border-2 border-kraft-shadow" />
      ))}
    </div>
  );

  if (notices.length === 0) return null;

  return (
    <section className="space-y-6">
      <div className="flex items-baseline justify-between gap-4 border-b-2 border-kraft-dark pb-2">
        <h2 className="font-display text-4xl uppercase" style={{ color: 'var(--ink-deep)' }}>
          {t('pages.notices.section.title', 'NOTICES / NEWS')}
        </h2>
        <Link href="/notices" className="text-sm font-bold font-mono-stack hover:text-gold transition-colors" style={{ color: 'var(--text-secondary)' }}>
          {t('pages.notices.actions.view_all', 'VIEW ALL →')}
        </Link>
      </div>

      <div className="flex flex-col gap-6">
        {notices.map((notice) => (
          <article key={notice.id} className="notice-card flex flex-col md:flex-row bg-surface rounded-sm overflow-hidden shadow-sm hover:shadow-md transition-shadow" style={{ border: '2px solid var(--kraft-shadow)' }}>
            {notice.featured_image_url && (
              <div className="md:w-1/3 h-48 md:h-auto relative overflow-hidden bg-kraft-light">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img 
                  src={notice.featured_image_url} 
                  alt={notice.title}
                  className="w-full h-full object-cover transition-transform hover:scale-105 duration-500"
                />
              </div>
            )}
            <div className={`p-6 flex flex-col flex-1 ${notice.featured_image_url ? '' : 'md:p-8'}`}>
              <div className="flex items-center gap-3 mb-2">
                <span className="text-[10px] font-mono-stack font-bold px-2 py-0.5 bg-kraft-light text-text-secondary border border-kraft-shadow uppercase">
                  {new Date(notice.created_at).toLocaleDateString(locale === 'es' ? 'es-ES' : 'en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                </span>
              </div>
              <h3 className="font-display text-2xl mb-3 uppercase leading-tight" style={{ color: 'var(--ink-deep)' }}>
                {notice.title}
              </h3>
              <div 
                className="text-sm text-text-secondary line-clamp-3 mb-4 font-mono-stack leading-relaxed"
                dangerouslySetInnerHTML={{ __html: notice.content_html.replace(/<[^>]*>?/gm, ' ').slice(0, 300) }}
              />
              <div className="mt-auto flex items-center">
                <Link href={`/notices/${notice.slug}`} className="text-xs font-bold font-mono-stack text-gold-dark hover:text-hp-color transition-colors flex items-center gap-2">
                  {t('pages.notices.actions.read_more', 'READ MORE')} <span className="text-lg">»</span>
                </Link>
              </div>
            </div>
          </article>
        ))}
      </div>

      <div className="text-center pt-4">
        <Link href="/notices" className="btn-secondary">
          {t('pages.notices.actions.show_more', 'SHOW MORE PREVIOUS POSTS')}
        </Link>
      </div>
    </section>
  );
}
