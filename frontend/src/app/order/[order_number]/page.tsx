'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';

export default function OrderConfirmation() {
  const params = useParams();
  const orderNumber = params.order_number as string;
  const { t } = useLanguage();

  return (
    <div className="centered-container px-4 py-16 text-center">
      <div className="card max-w-lg mx-auto p-8">
        <div className="text-5xl mb-4">✅</div>
        <h1 className="font-display text-4xl mb-2">{t('pages.order.confirmation.title', 'ORDER RECEIVED!')}</h1>
        <div className="divider" />

        <div className="cardbox p-4 mb-6" style={{ background: 'var(--kraft-light)' }}>
          <p className="text-xs font-mono-stack mb-1" style={{ color: 'var(--text-muted)' }}>{t('pages.order.confirmation.order_no', 'ORDER NUMBER')}</p>
          <p className="font-display text-3xl" style={{ color: 'var(--gold-dark)' }}>{orderNumber}</p>
        </div>

        <p className="text-sm mb-6" style={{ color: 'var(--text-secondary)' }}>
          {t('pages.order.confirmation.desc', 'Your order has been successfully registered. An advisor will contact you to coordinate payment and delivery.')}
        </p>

        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <Link href="/" className="btn-primary">{t('pages.order.confirmation.back', '← BACK TO STORE')}</Link>
          <Link href="/contact" className="btn-secondary">{t('pages.order.confirmation.contact', 'CONTACT')}</Link>
        </div>
      </div>
    </div>
  );
}
