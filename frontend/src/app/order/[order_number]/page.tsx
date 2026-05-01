'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useLanguage } from '@/context/LanguageContext';

export default function OrderConfirmation() {
  const params = useParams();
  const orderNumber = params.order_number as string;
  const { t } = useLanguage();

  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4 py-12 relative overflow-hidden">
      {/* Box Lid Animation Layer */}
      <div className="fixed inset-0 z-[100] bg-[#d2b48c] dark:bg-[#3d2b1f] flex flex-col items-center justify-center animate-box-lid pointer-events-none border-b-8 border-black/10">
        <div className="border-[12px] border-double border-[#8b4513] dark:border-gold/40 p-12 rotate-[-4deg] shadow-2xl bg-white/5 backdrop-blur-sm">
          <h2 className="font-display text-8xl text-[#8b4513] dark:text-gold tracking-tighter leading-none">EL BULK</h2>
          <div className="h-px bg-[#8b4513] dark:bg-gold/40 my-4" />
          <p className="font-mono text-sm tracking-[0.4em] text-center uppercase font-bold text-[#8b4513] dark:text-gold">
            {t('pages.order.confirmation.shipping_label', 'FRAGILE / LOGISTICS')}
          </p>
        </div>
        <div className="mt-12 font-mono text-xs text-black/40 dark:text-white/20 tracking-widest uppercase animate-pulse">
          {t('pages.order.confirmation.opening', 'DECOMPRESSING PARCEL...')}
        </div>
      </div>

      {/* Revealed Content */}
      <div className="w-full max-w-2xl relative z-10 animate-unbox-reveal">
        <div className="card shadow-2xl p-0 overflow-hidden border-border-main/40 bg-bg-surface/90 backdrop-blur-xl">
          {/* Header Graphic */}
          <div className="h-32 bg-accent-primary/10 relative flex items-center justify-center border-b border-border-main/20 overflow-hidden">
            <div className="absolute top-0 right-0 p-4 opacity-5 rotate-12">
              <span className="text-9xl font-display">EB</span>
            </div>
            <div className="relative z-10 bg-white dark:bg-bg-page px-8 py-3 rounded-full border-4 border-accent-primary animate-stamp-slam shadow-xl">
              <span className="font-display text-4xl text-accent-primary tracking-tight">
                {t('pages.order.confirmation.delivered', 'SUCCESSFULLY STAMPED')}
              </span>
            </div>
          </div>

          <div className="p-8 sm:p-12 text-center">
            <h1 className="font-display text-5xl mb-4 tracking-tighter text-text-main">
              {t('pages.order.confirmation.title', 'ORDER RECEIVED!')}
            </h1>
            
            <div className="flex items-center justify-center gap-4 mb-8">
              <div className="h-px flex-1 bg-border-main/30" />
              <div className="font-mono text-[10px] tracking-[0.3em] text-text-muted uppercase">
                {t('pages.order.confirmation.logistics_ref', 'LOGISTICS REF')}
              </div>
              <div className="h-px flex-1 bg-border-main/30" />
            </div>

            <div className="bg-bg-page/50 border border-border-main p-6 rounded-2xl mb-8 shadow-inner relative group">
               <div className="absolute top-2 right-4 font-mono text-[8px] text-text-muted opacity-50 group-hover:opacity-100 transition-opacity">
                CONFIRMED: {new Date().toLocaleDateString()}
               </div>
               <p className="text-[10px] font-mono text-text-muted mb-2 tracking-widest uppercase">
                 {t('pages.order.confirmation.order_no', 'INTERNAL ORDER NUMBER')}
               </p>
               <p className="font-display text-5xl text-accent-primary selection:bg-accent-primary selection:text-white">
                 #{orderNumber}
               </p>
            </div>

            <p className="text-base text-text-secondary mb-10 leading-relaxed max-w-md mx-auto">
              {t('pages.order.confirmation.desc', 'Your order has been successfully registered in our central archives. An advisor will contact you shortly via WhatsApp/Email to coordinate secure payment and delivery.')}
            </p>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <Link 
                href="/" 
                className="btn-primary py-4 px-8 flex items-center justify-center gap-2 group"
              >
                <span className="group-hover:-translate-x-1 transition-transform">←</span>
                {t('pages.order.confirmation.back', 'RETURN TO DEPOT')}
              </Link>
              <Link 
                href="/contact" 
                className="btn-secondary py-4 px-8 border-2 border-border-main/50 hover:border-accent-primary/50"
              >
                {t('pages.order.confirmation.contact', 'LOGISTICS SUPPORT')}
              </Link>
            </div>
          </div>

          {/* Footer Bar */}
          <div className="bg-bg-page/80 p-4 border-t border-border-main/20 flex justify-between items-center">
            <span className="text-[9px] font-mono text-text-muted uppercase tracking-widest">
              El Bulk Archive Logistics © 2026
            </span>
            <div className="flex gap-1">
              {[1, 2, 3, 4, 5].map(i => (
                <div key={i} className="w-4 h-1 bg-accent-primary/20 rounded-full" />
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Decorative Background Elements */}
      <div className="fixed top-1/4 -left-20 w-64 h-64 bg-accent-primary/5 rounded-full blur-[100px] pointer-events-none" />
      <div className="fixed bottom-1/4 -right-20 w-80 h-80 bg-hp-color/5 rounded-full blur-[120px] pointer-events-none" />
    </div>
  );
}
