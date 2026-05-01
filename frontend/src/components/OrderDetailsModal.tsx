'use client';

import { useEffect, useState } from 'react';
import { userFetchOrderDetail, userCancelOrder } from '@/lib/api';
import { OrderDetail, ORDER_STATUS_LABELS, FOIL_LABELS } from '@/lib/types';
import Modal from './ui/Modal';
import LoadingSpinner from './LoadingSpinner';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from './CardImage';

interface OrderDetailsModalProps {
  orderId: string | null;
  isOpen: boolean;
  onClose: () => void;
  onOrderCancelled?: () => void;
}

export default function OrderDetailsModal({ orderId, isOpen, onClose, onOrderCancelled }: OrderDetailsModalProps) {
  const [detail, setDetail] = useState<OrderDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [cancelling, setCancelling] = useState(false);
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

  const handleCancelOrder = async () => {
    if (!detail?.order || cancelling) return;
    
    const confirmMsg = t('pages.order.modal.cancel_confirm', 'Are you sure you want to cancel this order? This action cannot be undone.');
    if (!window.confirm(confirmMsg)) return;

    setCancelling(true);
    try {
      const updated = await userCancelOrder(detail.order.id);
      setDetail(updated);
      if (onOrderCancelled) onOrderCancelled();
    } catch (err) {
      console.error('Failed to cancel order:', err);
      alert(t('pages.order.modal.cancel_error', 'Failed to cancel order. Please contact support.'));
    } finally {
      setCancelling(false);
    }
  };

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
            {/* Shipment Progress Timeline */}
            <div className="p-6 rounded-2xl border border-border-main bg-bg-page/40 shadow-inner backdrop-blur-sm relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-[0.03]">
                <span className="text-6xl font-display leading-none">EB-POST</span>
              </div>
              
              <div className="relative z-10">
                <div className="flex justify-between items-center mb-10 relative px-2">
                  {/* Progress Line */}
                  <div className="absolute top-4 left-0 w-full h-[2px] bg-border-main/30 z-0" />
                  <div 
                    className="absolute top-4 left-0 h-[2px] bg-accent-primary transition-all duration-1000 ease-out z-0" 
                    style={{ 
                      width: detail.order.status === 'completed' ? '100%' : 
                             detail.order.status === 'shipped' || detail.order.status === 'ready_for_pickup' ? '66%' :
                             detail.order.status === 'confirmed' ? '33%' : '0%' 
                    }} 
                  />

                  {/* Steps */}
                  {[
                    { key: 'pending', label: t('pages.order.status.pending', 'RECEIVED'), icon: '📥' },
                    { key: 'confirmed', label: t('pages.order.status.confirmed', 'STAMPED'), icon: '📜' },
                    { key: 'dispatched', label: t('pages.order.status.shipped', 'DISPATCHED'), icon: '🚚' },
                    { key: 'completed', label: t('pages.order.status.completed', 'ARRIVED'), icon: '🏁' }
                  ].map((step, i, arr) => {
                    const isCompleted = (step.key === 'pending') || 
                                       (step.key === 'confirmed' && ['confirmed', 'shipped', 'ready_for_pickup', 'completed'].includes(detail.order.status)) ||
                                       (step.key === 'dispatched' && ['shipped', 'ready_for_pickup', 'completed'].includes(detail.order.status)) ||
                                       (step.key === 'completed' && detail.order.status === 'completed');
                    
                    const isCurrent = (step.key === detail.order.status) || 
                                     (step.key === 'dispatched' && (detail.order.status === 'shipped' || detail.order.status === 'ready_for_pickup'));

                    return (
                      <div key={step.key} className="flex flex-col items-center relative z-10">
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs transition-all duration-500 border-2 ${
                          isCompleted ? 'bg-accent-primary border-accent-primary text-text-on-accent' : 
                          'bg-bg-surface border-border-main text-text-muted'
                        } ${isCurrent ? 'ring-4 ring-accent-primary/20 scale-110' : ''}`}>
                          {isCompleted ? '✓' : i + 1}
                        </div>
                        <span className={`mt-3 font-mono text-[9px] uppercase tracking-widest font-bold ${isCurrent ? 'text-accent-primary' : 'text-text-muted'}`}>
                          {step.label}
                        </span>
                      </div>
                    );
                  })}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 pt-4 border-t border-border-main/20">
                  <div>
                    <p className="text-[9px] font-mono text-text-muted uppercase tracking-[0.2em] mb-2">{t('pages.order.modal.tracking_info', 'LOGISTICS DETAILS')}</p>
                    <div className="flex items-center gap-3">
                      <div className="bg-accent-primary/10 px-3 py-1 rounded border border-accent-primary/20">
                        <span className="font-mono text-xs font-bold text-accent-primary">#{detail.order.order_number}</span>
                      </div>
                      <span className="text-xs font-mono text-text-secondary uppercase">
                        {t(`pages.order.status.${detail.order.status}`, ORDER_STATUS_LABELS[detail.order.status] || detail.order.status)}
                      </span>
                    </div>
                  </div>
                  <div className="md:text-right">
                    <p className="text-[9px] font-mono text-text-muted uppercase tracking-[0.2em] mb-2">{t('pages.order.modal.date_label', 'DATE OF DISPATCH')}</p>
                    <p className="text-xs font-mono text-text-secondary">
                      {new Date(detail.order.created_at).toLocaleDateString(locale === 'es' ? 'es-ES' : 'en-US', { 
                        year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
                      })}
                    </p>
                  </div>
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
                
                {detail.order.shipping_cop > 0 && (
                  <div className="flex items-center gap-4 p-3.5 rounded-xl border border-dashed border-border-main/60 bg-bg-card/10 hover:bg-bg-card/20 transition-all duration-300">
                    <div className="flex-shrink-0 w-16 h-16 flex items-center justify-center bg-accent-primary/5 rounded-lg border border-accent-primary/10">
                      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-accent-primary opacity-60">
                         <path d="M1 3h15v13H1V3z" strokeLinecap="round" strokeLinejoin="round"/>
                         <path d="M16 8h4.4l3.6 4.35V16h-8V8z" strokeLinecap="round" strokeLinejoin="round"/>
                         <circle cx="5.5" cy="18.5" r="2.5" stroke="currentColor"/>
                         <circle cx="18.5" cy="18.5" r="2.5" stroke="currentColor"/>
                      </svg>
                    </div>
                    <div className="flex-grow">
                      <h5 className="font-bold text-text-main text-sm uppercase tracking-wide">
                        {t('pages.order.modal.shipping_item_label', 'Shipping & Handling')}
                      </h5>
                      <p className="text-[10px] text-text-muted font-mono mt-0.5 uppercase">
                        {detail.order.is_local_pickup 
                          ? t('pages.order.modal.local_pickup', 'Local Pickup') 
                          : t('pages.order.modal.standard_shipping', 'Standard Shipping')}
                      </p>
                    </div>
                    <div className="text-right">
                      <div className="font-display text-xl text-text-main leading-none">
                        ${detail.order.shipping_cop.toLocaleString('es-CO')}
                      </div>
                    </div>
                  </div>
                )}
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

            {/* Actions Section */}
            {(detail.order.status === 'pending' || (['confirmed', 'shipped', 'ready_for_pickup'].includes(detail.order.status) && detail.whatsapp_url)) && (
              <div className="pt-6 border-t border-border-main/30 animate-in fade-in duration-700 delay-400 fill-mode-both">
                {detail.order.status === 'pending' ? (
                  <div className="flex flex-col gap-3">
                    <button
                      onClick={handleCancelOrder}
                      disabled={cancelling}
                      className="w-full py-3 bg-red-400/10 hover:bg-red-400/20 text-red-400 rounded-xl border border-red-400/30 transition-all font-bold tracking-wide uppercase text-xs disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.99]"
                    >
                      {cancelling ? t('pages.common.status.loading', 'Processing...') : `🚫 ${t('pages.order.modal.cancel_order_btn', 'Cancel This Order')}`}
                    </button>
                    <p className="text-[9px] text-center text-text-muted italic">
                      {t('pages.order.modal.cancel_notice', 'Orders can only be cancelled while in the pending state.')}
                    </p>
                  </div>
                ) : (
                  <div className="bg-accent-primary/5 p-4 rounded-xl border border-accent-primary/20 flex flex-col sm:flex-row items-center justify-between gap-4">
                    <div className="text-center sm:text-left">
                      <p className="text-sm font-bold text-text-main">{t('pages.order.modal.support_title', 'Need to cancel or change something?')}</p>
                      <p className="text-[10px] text-text-muted mt-0.5">{t('pages.order.modal.support_subtitle', 'This order is already being processed. Please contact us via WhatsApp.')}</p>
                    </div>
                    {detail.whatsapp_url && (
                      <a
                        href={detail.whatsapp_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-2 px-5 py-2.5 bg-[#25D366] hover:bg-[#20bd5a] text-white rounded-lg font-bold text-xs no-underline transition-all shadow-md active:scale-95 whitespace-nowrap"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M12.031 6.172c-3.181 0-5.767 2.586-5.768 5.766-.001 1.298.38 2.27 1.019 3.287l-.582 2.128 2.182-.573c.978.58 1.911.928 3.145.929 3.178 0 5.767-2.587 5.768-5.766.001-3.187-2.575-5.771-5.764-5.771zm3.392 8.244c-.144.405-.837.774-1.17.824-.299.045-.677.063-1.092-.069-.252-.08-.575-.187-.988-.365-1.739-.751-2.874-2.502-2.961-2.617-.087-.116-.708-.94-.708-1.793s.448-1.273.607-1.446c.159-.173.346-.217.462-.217s.231.006.332.013c.101.007.237-.038.371.295.134.333.462 1.127.502 1.206.041.08.068.173.015.282-.053.107-.077.174-.153.262-.078.089-.164.197-.234.266-.081.079-.165.166-.071.327.094.162.413.683.889 1.103.614.54 1.134.707 1.3.788.165.08.262.068.319-.004.058-.073.248-.289.314-.387.065-.099.13-.081.219-.051s.563.266.66.314c.097.048.162.073.186.113.023.04.023.232-.121.637z"/>
                        </svg>
                        {t('pages.order.modal.whatsapp_btn', 'Contact Support')}
                      </a>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        ) : null}
      </div>
    </Modal>
  );
}
