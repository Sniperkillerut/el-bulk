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
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!slug) return;
    fetchNoticeBySlug(slug as string)
      .then(setNotice)
      .catch(() => {
        router.push('/notices');
      })
      .finally(() => setLoading(false));
  }, [slug, router]);

  const handleCopyLink = () => {
    navigator.clipboard.writeText(window.location.href);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const shareUrl = typeof window !== 'undefined' ? encodeURIComponent(window.location.href) : '';
  const shareTitle = notice ? encodeURIComponent(notice.title) : '';

  if (loading) return (
    <div className="centered-container px-4 py-12 animate-pulse">
      <div className="h-8 bg-bg-header w-2/3 mb-6" />
      <div className="h-48 bg-bg-header w-full" />
    </div>
  );

  if (!notice) return null;

  return (
    <article className="pb-24">
      {/* Hero Header - Compact & Themed */}
      <header className="bg-bg-header border-b border-border-main py-8 md:py-12 mb-8">
        <div className="centered-container px-4">
          <div className="flex flex-col gap-3">
            <Link href="/notices" className="text-[10px] font-mono-stack font-bold text-accent-primary hover:opacity-80 flex items-center gap-2 transition-colors uppercase">
              {t('pages.notices.detail.back_btn', '← BACK TO ALL NOTICES')}
            </Link>
            <div className="text-[10px] font-mono-stack font-bold px-3 py-1 bg-bg-surface text-text-secondary border border-border-main self-start uppercase">
              {t('pages.notices.detail.published_on', 'Published on {date}', { date: new Date(notice.created_at).toLocaleDateString(locale === 'en' ? 'en-US' : 'es-ES', { month: 'long', day: 'numeric', year: 'numeric' }) })}
            </div>
            <h1 className="font-display text-4xl sm:text-5xl md:text-6xl uppercase leading-[0.85] max-w-4xl text-text-main drop-shadow-sm">
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

            <div className="bg-bg-card p-6 md:p-10 shadow-xl border border-border-main rounded-sm">
              <NoticeContent html={notice.content_html} />
            </div>

            <div className="mt-12 pt-12 border-t border-border-main flex items-center justify-between">
              <Link href="/notices" className="btn-secondary">
                {t('pages.notices.detail.all_notices_btn', '← ALL NOTICES')}
              </Link>
              <div className="flex items-center gap-3">
                <span className="text-[10px] font-mono-stack font-bold text-text-muted uppercase tracking-wider">{t('pages.notices.detail.share_label', 'SHARE:')}</span>
                
                <div className="flex gap-2">
                  {/* Facebook */}
                  <a 
                    href={`https://www.facebook.com/sharer/sharer.php?u=${shareUrl}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="w-8 h-8 flex items-center justify-center rounded-full bg-bg-surface border border-border-main hover:border-accent-primary hover:text-accent-primary transition-all text-text-secondary"
                    title="Share on Facebook"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M18 2h-3a5 5 0 0 0-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 0 1 1-1h3z"></path></svg>
                  </a>

                  {/* WhatsApp */}
                  <a 
                    href={`https://api.whatsapp.com/send?text=${shareTitle}%20${shareUrl}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="w-8 h-8 flex items-center justify-center rounded-full bg-bg-surface border border-border-main hover:border-green-500 hover:text-green-500 transition-all text-text-secondary"
                    title="Share on WhatsApp"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>
                  </a>

                  {/* X (Twitter) */}
                  <a 
                    href={`https://twitter.com/intent/tweet?url=${shareUrl}&text=${shareTitle}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="w-8 h-8 flex items-center justify-center rounded-full bg-bg-surface border border-border-main hover:border-text-main hover:text-text-main transition-all text-text-secondary"
                    title="Share on X"
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M4 4l11.733 16h4.267l-11.733 -16z"></path><path d="M4 20l6.768 -6.768"></path><path d="M13.232 10.768l6.768 -6.768"></path></svg>
                  </a>

                  {/* Copy Link */}
                  <button 
                    onClick={handleCopyLink}
                    className={`w-8 h-8 flex items-center justify-center rounded-full bg-bg-surface border transition-all ${copied ? 'border-accent-primary text-accent-primary' : 'border-border-main text-text-secondary hover:border-accent-primary hover:text-accent-primary'}`}
                    title={copied ? "Copied!" : "Copy Link"}
                  >
                    {copied ? (
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    ) : (
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
                    )}
                  </button>
                </div>
              </div>
            </div>
          </main>

          {/* Sidebar */}
          <aside className="lg:w-80">
            <div className="sticky top-24 space-y-8">
              <div className="bg-bg-surface p-6 border border-border-main rounded-sm flex flex-col gap-4 shadow-sm">
                <h4 className="font-display text-xl uppercase text-accent-primary">{t('pages.notices.detail.sidebar.title', 'EL BULK SHOP')}</h4>
                <p className="text-xs font-mono-stack leading-relaxed text-text-secondary">
                  {t('pages.notices.detail.sidebar.desc', 'The shoebox where we keep all the good stuff. Selling singles, sealed product, and accessories. Visit us in person or shop online!')}
                </p>
                <Link href="/singles" className="btn-primary text-center py-2.5 text-sm uppercase font-bold tracking-widest">
                  {t('pages.notices.detail.sidebar.shop_btn', 'SHOP SINGLES')}
                </Link>
              </div>

              <div className="p-5 border border-dashed border-border-main rounded-sm bg-bg-header/20">
                <h5 className="font-display text-sm uppercase mb-3 text-text-main">{t('pages.notices.detail.sidebar.newsletter_title', 'Newsletter')}</h5>
                <p className="text-[10px] font-mono-stack mb-4 text-text-muted leading-tight">{t('pages.notices.detail.sidebar.newsletter_desc', 'Stay updated with our latest news and spoilers.')}</p>
                <NewsletterForm />
              </div>
            </div>
          </aside>
        </div>
      </div>
    </article>
  );
}
