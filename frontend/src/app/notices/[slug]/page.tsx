'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { Notice } from '@/lib/types';
import { fetchNoticeBySlug } from '@/lib/api';
import NoticeContent from '@/components/NoticeContent';
import NewsletterForm from '@/components/NewsletterForm';
import { useLanguage } from '@/context/LanguageContext';

export default function NoticeDetailPage() {
  const { t, locale } = useLanguage();
  const { slug } = useParams();
  const router = useRouter();
  const [notice, setNotice] = useState<Notice | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!slug) return;
    fetchNoticeBySlug(slug as string)
      .then(setNotice)
      .catch(() => {
        router.push('/notices');
      })
      .finally(() => setLoading(false));
  }, [slug, router]);

  if (loading) return (
    <div className="centered-container px-4 py-20 animate-pulse">
      <div className="h-12 bg-kraft-light w-2/3 mb-8" />
      <div className="h-64 bg-kraft-light w-full" />
    </div>
  );

  if (!notice) return null;

  return (
    <article className="pb-24">
      {/* Hero Header */}
      <header className="bg-kraft-mid border-b-4 border-kraft-dark py-12 md:py-20 mb-12">
        <div className="centered-container px-4">
          <div className="flex flex-col gap-4">
            <Link href="/notices" className="text-[10px] font-mono-stack font-bold text-gold-dark hover:text-hp-color flex items-center gap-2 mb-2 transition-colors uppercase">
              {t('pages.notices.detail.back_btn', '← BACK TO ALL NOTICES')}
            </Link>
            <div className="text-[10px] font-mono-stack font-bold px-3 py-1 bg-kraft-light text-text-secondary border border-kraft-shadow self-start uppercase">
              {t('pages.notices.detail.published_on', 'Published on {date}', { date: new Date(notice.created_at).toLocaleDateString(locale === 'en' ? 'en-US' : 'es-ES', { month: 'long', day: 'numeric', year: 'numeric' }) })}
            </div>
            <h1 className="font-display text-4xl sm:text-6xl md:text-7xl uppercase leading-tight max-w-4xl text-fluid-h1" style={{ color: 'var(--ink-deep)' }}>
              {notice.title}
            </h1>
          </div>
        </div>
      </header>

      <div className="centered-container px-4">
        <div className="flex flex-col lg:flex-row gap-12">
          {/* Main Content Area */}
          <main className="flex-1 max-w-3xl">
            {notice.featured_image_url && (
              <div className="mb-12 rounded-sm overflow-hidden border-2 border-kraft-shadow shadow-sm">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img 
                  src={notice.featured_image_url} 
                  alt={notice.title}
                  className="w-full h-auto"
                />
              </div>
            )}

            <div className="bg-surface p-8 md:p-12 shadow-sm border-2 border-kraft-shadow">
              <NoticeContent html={notice.content_html} />
            </div>

            <div className="mt-12 pt-12 border-t-2 border-kraft-dark flex items-center justify-between">
              <Link href="/notices" className="btn-secondary">
                {t('pages.notices.detail.all_notices_btn', '← ALL NOTICES')}
              </Link>
              <div className="flex gap-4">
                <span className="text-xs font-mono-stack text-text-muted">{t('pages.notices.detail.share_label', 'SHARE:')}</span>
                {/* Social placeholders could go here */}
              </div>
            </div>
          </main>

          {/* Sidebar */}
          <aside className="lg:w-80">
            <div className="sticky top-24 space-y-8">
              <div className="bg-kraft-light p-6 border-2 border-kraft-shadow rounded-sm rotate-1 flex flex-col gap-4">
                <h4 className="font-display text-xl uppercase color-hp-color">{t('pages.notices.detail.sidebar.title', 'EL BULK SHOP')}</h4>
                <p className="text-xs font-mono-stack leading-relaxed">
                  {t('pages.notices.detail.sidebar.desc', 'The shoebox where we keep all the good stuff. Selling singles, sealed product, and accessories. Visit us in person or shop online!')}
                </p>
                <Link href="/singles" className="btn-primary text-center py-2 text-sm">
                  {t('pages.notices.detail.sidebar.shop_btn', 'SHOP SINGLES')}
                </Link>
              </div>

              <div className="p-4 border border-dashed border-kraft-dark rounded-sm">
                <h5 className="font-display text-sm uppercase mb-3 text-kraft-dark">{t('pages.notices.detail.sidebar.newsletter_title', 'Newsletter')}</h5>
                <p className="text-[10px] font-mono-stack mb-4 text-text-muted">{t('pages.notices.detail.sidebar.newsletter_desc', 'Stay updated with our latest news and spoilers.')}</p>
                <NewsletterForm />
              </div>
            </div>
          </aside>
        </div>
      </div>
    </article>
  );
}
