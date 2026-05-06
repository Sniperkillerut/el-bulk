import { Metadata } from 'next';
import { redirect } from 'next/navigation';
import { fetchNoticeBySlug } from '@/lib/api';
import NoticeDetailClient from './NoticeDetailClient';
import { getSharedProductMetadata } from '@/lib/metadata';

interface Props {
  params: Promise<{ slug: string }>;
  searchParams: Promise<{ productId?: string }>;
}

export async function generateMetadata({ params, searchParams }: Props): Promise<Metadata> {
  const { slug } = await params;
  const { productId } = await searchParams;

  const productMetadata = await getSharedProductMetadata(productId || null);
  if (productMetadata) return productMetadata;

  try {
    const notice = await fetchNoticeBySlug(slug);
    const title = `${notice.title} | El Bulk News`;
    const description = notice.content_html.replace(/<[^>]*>?/gm, '').substring(0, 160);
    
    return {
      title,
      description,
      openGraph: {
        title,
        description,
        images: notice.featured_image_url ? [notice.featured_image_url] : [],
        type: 'article',
      },
      twitter: {
        card: 'summary_large_image',
        title,
        description,
        images: notice.featured_image_url ? [notice.featured_image_url] : [],
      },
    };
  } catch {
    return {
      title: 'Notice Not Found | El Bulk',
    };
  }
}

export default async function NoticeDetailPage({ params }: Props) {
  const { slug } = await params;
  let notice;
  
  try {
    notice = await fetchNoticeBySlug(slug);
  } catch {
    redirect('/notices');
  }

  if (!notice) {
    redirect('/notices');
  }

  // Structured Data (Article)
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'NewsArticle',
    headline: notice.title,
    image: notice.featured_image_url ? [notice.featured_image_url] : [],
    datePublished: notice.created_at,
    dateModified: notice.updated_at || notice.created_at,
    author: {
      '@type': 'Organization',
      name: 'El Bulk',
    },
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <NoticeDetailClient notice={notice} />
    </>
  );
}
