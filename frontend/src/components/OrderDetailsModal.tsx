'use client';

import { useEffect, useState } from 'react';
import { userFetchOrderDetail } from '@/lib/api';
import { OrderDetail, ORDER_STATUS_LABELS, FOIL_LABELS } from '@/lib/types';
import Modal from './ui/Modal';
import LoadingSpinner from './LoadingSpinner';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from './CardImage';

interface OrderDetailsModalProps {
  orderId: string | null;
  isOpen: boolean;
  onClose: () => void;
}

export default function OrderDetailsModal({ orderId, isOpen, onClose }: OrderDetailsModalProps) {
  const [detail, setDetail] = useState<OrderDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { t, locale } = useLanguage();

  useEffect(() => {
    let isMounted = true;

    async function fetchDetail() {
      if (!orderId || !isOpen) return;
      
      setLoading(true);
      setError(null);
      
      try {
        const data = await userFetchOrderDetail(orderId);
        if (isMounted) {
          setDetail(data);
        }
      } catch (err) {
        console.error('Failed to fetch order details:', err);
        if (isMounted) {
          setError(t('pages.order.modal.error', 'Failed to load order details. Please try again.'));
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    }

    if (isOpen && orderId) {
      fetchDetail();
    } else if (!isOpen) {
      // Clear detail when closing to avoid flickering when opening a new one
      const timer = setTimeout(() => {
        if (isMounted) setDetail(null);
      }, 300);
      return () => {
        isMounted = false;
        clearTimeout(timer);
      };
    }

    return () => {
      isMounted = false;
    };
  }, [isOpen, orderId, t]);

  if (!isOpen && !detail) return null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={detail?.order 
        ? t('pages.order.modal.title', 'Order {number}', { number: detail.order.order_number }) 
        : t('pages.order.modal.title_generic', 'Order Details')}
      maxWidth="max-w-3xl"
    >
      <div className="p-6 bg-bg-surface/90 backdrop-blur-md min-h-[400px]">
        {loading && !detail ? (
          <div className="flex flex-col items-center justify-center py-24">
            <LoadingSpinner />
            <p className="mt-4 text-text-muted font-mono text-xs animate-pulse tracking-widest uppercase">{t('pages.order.modal.fetching', 'Fetching secure details')}</p>
          </div>
        ) : error ? (
          <div className="text-center py-20 bg-red-400/5 rounded-xl border border-red-400/20">
            <div className="text-4xl mb-4">⚠️</div>
            <p className="text-red-400 mb-6 font-medium">{error}</p>
            <button
              onClick={onClose}
              className="px-8 py-3 bg-red-400/20 hover:bg-red-400/30 text-red-400 rounded-lg transition-colors border border-red-400/30"
            >
              {t('pages.order.modal.close', 'Close Window')}
            </button>
          </div>
        ) : detail ? (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500 fill-mode-both">
            {/* Status Header */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 p-5 rounded-2xl border border-border-main bg-bg-page/40 shadow-inner backdrop-blur-sm relative overflow-hidden group">
              <div className="absolute top-0 right-0 p-2 opacity-[0.03] group-hover:opacity-[0.07] transition-opacity">
                <span className="text-8xl font-display leading-none">EB</span>
              </div>
              
              <div className="relative z-10">
                <p className="text-[10px] font-mono text-text-muted uppercase tracking-widest mb-2">{t('pages.order.modal.status_label', 'Order Status')}</p>
                <div className="flex items-center gap-2.5">
                  <div className={`w-3 h-3 rounded-full shadow-[0_0_10px_rgba(0,0,0,0.2)] ${
                    detail.order.status === 'completed' ? 'bg-green-400 shadow-green-400/40' :
                    detail.order.status === 'pending' ? 'bg-yellow-400 shadow-yellow-400/40' :
                    detail.order.status === 'cancelled' ? 'bg-red-400 shadow-red-400/40' : 'bg-blue-400 shadow-blue-400/40'
                  }`} />
                  <span className="font-display text-2xl text-text-main tracking-tight uppercase">
                    {t(`pages.order.status.${detail.order.status}`, ORDER_STATUS_LABELS[detail.order.status] || detail.order.status)}
                  </span>
                </div>
              </div>
              
              <div className="md:border-l border-border-main/50 md:pl-6 relative z-10">
                <p className="text-[10px] font-mono text-text-muted uppercase tracking-widest mb-2">{t('pages.order.modal.date_label', 'Transaction Date')}</p>
                <p className="text-text-main font-medium">
                   {new Date(detail.order.created_at).toLocaleDateString(locale === 'es' ? 'es-ES' : 'en-US', { 
                    year: 'numeric', month: 'short', day: 'numeric'
                  })}
                   <span className="text-text-muted ml-2 text-sm font-normal">
                    {new Date(detail.order.created_at).toLocaleTimeString(locale === 'es' ? 'es-ES' : 'en-US', { hour: '2-digit', minute: '2-digit' })}
                   </span>
                </p>
              </div>
              
              <div className="md:border-l border-border-main md:pl-6 relative z-10">
                <p className="text-[10px] font-mono text-text-muted uppercase tracking-widest mb-2">{t('pages.order.modal.amount_label', 'Total Amount')}</p>
                <div className="flex items-baseline gap-1">
                  <span className="text-accent-primary text-sm font-bold">$</span>
                  <p className="text-3xl font-display text-accent-primary leading-none">
                    {detail.order.total_cop.toLocaleString('es-CO')}
                  </p>
                  <span className="text-[10px] font-mono text-text-muted ml-0.5 tracking-tighter">{t('pages.common.currency.cop', 'COP')}</span>
                </div>
              </div>
            </div>

            {/* Items List */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="text-[10px] font-mono text-text-muted uppercase tracking-[0.2em] flex items-center gap-3">
                  <span className="w-8 h-px bg-accent-primary/40"></span> 
                  {t('pages.order.modal.summary_label', 'Order Summary')} 
                  <span className="bg-accent-primary text-text-on-accent px-1.5 rounded-full text-[9px] font-bold">
                    {detail.items.reduce((acc, i) => acc + i.quantity, 0)}
                  </span>
                </h4>
              </div>

              <div className="space-y-3 max-h-[350px] overflow-y-auto pr-2 custom-scrollbar">
                {detail.items.map((item) => (
                  <div 
                    key={item.id} 
                    className="group flex items-center gap-4 p-3.5 rounded-xl border border-border-main/40 bg-bg-card/30 hover:bg-bg-card/60 hover:border-accent-primary/30 transition-all duration-300"
                  >
                    <div className="flex-shrink-0 w-16 h-22 bg-bg-page/50 rounded-lg overflow-hidden relative border border-border-main group-hover:border-accent-primary/20 transition-all duration-500 shadow-sm group-hover:shadow-accent-primary/5">
                      <CardImage 
                        imageUrl={item.image_url} 
                        name={item.product_name} 
                        foilTreatment={item.foil_treatment}
                        enableHover={true}
                        enableModal={true}
                      />
                    </div>

                    <div className="flex-grow min-w-0">
                      <div className="flex justify-between items-start gap-4">
                        <div className="min-w-0">
                          <h5 className="font-bold text-text-main group-hover:text-accent-primary transition-colors truncate text-sm sm:text-base">
                            {item.product_name}
                          </h5>
                          <p className="text-[10px] text-text-muted font-mono mt-1 flex items-center gap-1.5">
                            <span className="px-1.5 py-0.5 rounded-sm bg-bg-page/60 border border-border-main/50 uppercase tracking-tighter">
                              {item.product_set || t('pages.order.modal.unknown_set', 'Unknown Set')}
                            </span>
                            <span className={`px-1.5 py-0.5 rounded-sm border uppercase tracking-tighter ${
                              item.condition === 'NM' ? 'bg-green-400/5 text-green-400 border-green-400/20' : 
                              'bg-text-muted/5 text-text-muted border-border-main/50'
                            }`}>
                              {item.condition}
                            </span>
                          </p>
                        </div>
                        <div className="text-right flex-shrink-0 pt-1">
                          <div className="font-display text-xl text-text-main leading-none">
                            ${item.unit_price_cop.toLocaleString('es-CO')}
                          </div>
                          <div className="text-[9px] font-mono text-text-muted uppercase tracking-tighter mt-1.5 flex items-center justify-end gap-1">
                            <span className="opacity-50">{t('pages.order.modal.qty_label', 'QTY:')}</span>
                            <span className="text-text-secondary font-bold font-sans">×{item.quantity}</span>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex flex-wrap gap-1.5 mt-3">
                        {item.foil_treatment && item.foil_treatment !== 'non_foil' && (
                          <span className="text-[8px] font-bold px-2 py-0.5 rounded-[4px] bg-gradient-to-r from-accent-primary/20 to-purple-400/20 text-accent-primary border border-accent-primary/10 uppercase tracking-wider">
                            ✨ {FOIL_LABELS[item.foil_treatment] || item.foil_treatment}
                          </span>
                        )}
                        {item.card_treatment && item.card_treatment !== 'normal' && (
                          <span className="text-[8px] font-medium px-2 py-0.5 rounded-[4px] bg-text-muted/10 text-text-muted border border-border-main/50 uppercase tracking-wider">
                            {item.card_treatment.replace(/_/g, ' ')}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Support Info */}
            <div className="pt-6 border-t border-border-main/30 flex flex-col sm:flex-row justify-between items-center gap-4 text-center sm:text-left animate-in fade-in duration-700 delay-300 fill-mode-both">
              <div className="text-[10px] text-text-muted leading-relaxed font-mono uppercase tracking-tight">
                <p>{t('pages.order.modal.reference', 'Order Reference:')} <span className="text-text-main font-sans font-bold">#{detail.order.order_number}</span></p>
                <div className="mt-1 flex items-center justify-center sm:justify-start gap-4">
                  <p>{t('pages.order.modal.method', 'Method:')} <span className="text-text-secondary font-sans font-medium">{detail.order.payment_method}</span></p>
                  <p>{t('pages.order.modal.ref_prefix', 'Ref:')} <span className="text-text-secondary font-sans font-medium">{detail.order.id.split('-')[0].toUpperCase()}</span></p>
                </div>
              </div>
              <p className="text-[10px] font-mono text-text-muted uppercase tracking-tighter bg-bg-page/30 px-3 py-1.5 rounded-full border border-border-main/50">
                {t('pages.order.modal.issues', 'Issues? Contact')} <a href="mailto:support@elbulk.com" className="text-accent-primary hover:underline font-bold">support@elbulk.com</a>
              </p>
            </div>
          </div>
        ) : null}
      </div>
    </Modal>
  );
}
