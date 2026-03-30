'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Notice } from '@/lib/types';
import { fetchNotices } from '@/lib/api';

export default function NoticesPage() {
  const [notices, setNotices] = useState<Notice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchNotices()
      .then(setNotices)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="centered-container px-4 py-12">
      <header className="mb-12 border-b-2 border-kraft-dark pb-4 flex flex-col md:flex-row md:items-baseline justify-between gap-4">
        <div>
          <h1 className="font-display text-5xl sm:text-6xl uppercase" style={{ color: 'var(--ink-deep)' }}>
            NOTICES & <span style={{ color: 'var(--gold-dark)' }}>UPDATES</span>
          </h1>
          <p className="font-mono-stack text-sm text-text-secondary mt-2">Latest news, spoilers, and reviews from the shop.</p>
        </div>
        <div className="text-[10px] font-bold font-mono-stack bg-kraft-light px-3 py-1 border border-kraft-shadow rotate-1 self-start md:self-auto">
          {notices.length} POSTS TOTAL
        </div>
      </header>

      {loading ? (
        <div className="grid gap-8 animate-pulse">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="h-48 bg-kraft-light rounded-sm border-2 border-kraft-shadow" />
          ))}
        </div>
      ) : notices.length === 0 ? (
        <div className="stamp-border p-12 text-center bg-surface">
          <p className="font-display text-2xl text-text-muted">NO NOTICES POSTED YET</p>
          <p className="font-mono-stack text-sm text-text-muted mt-2">Check back soon for news and updates.</p>
        </div>
      ) : (
        <div className="grid gap-8">
          {notices.map((notice) => (
            <article 
              key={notice.id} 
              className="flex flex-col md:flex-row bg-surface rounded-sm overflow-hidden shadow-sm hover:shadow-md transition-shadow group" 
              style={{ border: '2px solid var(--kraft-shadow)' }}
            >
              {notice.featured_image_url && (
                <div className="md:w-1/4 h-48 md:h-auto relative overflow-hidden bg-kraft-light">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img 
                    src={notice.featured_image_url} 
                    alt={notice.title}
                    className="w-full h-full object-cover transition-transform group-hover:scale-105 duration-500"
                  />
                </div>
              )}
              <div className="p-6 md:p-8 flex flex-col flex-1">
                <div className="flex items-center gap-3 mb-3">
                  <span className="text-[10px] font-mono-stack font-bold px-2 py-0.5 bg-kraft-light text-text-secondary border border-kraft-shadow">
                    {new Date(notice.created_at).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}
                  </span>
                </div>
                <h2 className="font-display text-3xl mb-4 group-hover:text-gold-dark transition-colors uppercase leading-tight">
                  {notice.title}
                </h2>
                <div 
                  className="text-sm text-text-secondary line-clamp-3 mb-6 font-mono-stack leading-relaxed max-w-2xl"
                  dangerouslySetInnerHTML={{ __html: notice.content_html.replace(/<[^>]*>?/gm, ' ').slice(0, 400) }}
                />
                <div className="mt-auto">
                  <Link href={`/notices/${notice.slug}`} className="btn-secondary py-2 px-6">
                    OPEN NOTICE →
                  </Link>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
