'use client';

import { useEffect, useState } from 'react';
import { adminFetchInventoryValuation } from '@/lib/api';
import { InventoryValuation } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import LoadingSpinner from '@/components/LoadingSpinner';
import { useLanguage } from '@/context/LanguageContext';

export default function InventoryValuationPage() {
  const { t } = useLanguage();
  const { token, loading: authLoading } = useAdmin();
  const [data, setData] = useState<InventoryValuation | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    const load = async () => {
      setLoading(true);
      try {
        const val = await adminFetchInventoryValuation();
        setData(val);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [token]);

  if (authLoading || !token) {
    return <div className="p-12 text-center font-mono uppercase animate-pulse">{t('pages.admin.valuation.authenticating', 'Authenticating Accounting...')}</div>;
  }

  const formatCurrency = (val: number) => `$${val.toLocaleString('en-US', { maximumFractionDigits: 0 })} COP`;

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title={t('pages.admin.valuation.title', 'INVENTORY VALUATION')} 
        subtitle={t('pages.admin.valuation.subtitle', 'Real-time financial worth of your current stock')} 
      />

      <div className="flex-1 min-h-0 flex flex-col gap-8">
        {loading ? (
          <div className="flex-1 flex items-center justify-center p-20">
            <LoadingSpinner />
          </div>
        ) : !data ? (
          <div className="p-20 text-center opacity-30 italic">
            Error loading valuation data.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Total Items & Stock */}
            <div className="card p-8 bg-white border-kraft-dark/20 shadow-sm relative overflow-hidden group">
               <div className="absolute top-0 right-0 p-4 opacity-[0.05] text-6xl group-hover:scale-110 transition-transform">📦</div>
               <h3 className="text-xs font-mono-stack text-text-muted uppercase mb-4 tracking-widest">{t('pages.admin.valuation.stock_label', 'TOTAL PHYSICAL STOCK')}</h3>
               <div className="flex items-baseline gap-2">
                 <span className="text-5xl font-display text-ink-deep leading-none">{data.total_stock.toLocaleString()}</span>
                 <span className="text-xs font-mono-stack text-text-muted">unit{data.total_stock !== 1 ? 's' : ''}</span>
               </div>
               <p className="mt-4 text-xs font-mono-stack text-text-muted">
                 Across {data.total_items.toLocaleString()} unique products.
               </p>
            </div>

            {/* Total Valuation (Potential Revenue) */}
            <div className="card p-8 bg-gold/5 border-gold/30 shadow-xl shadow-gold/10 relative overflow-hidden group">
               <div className="absolute top-0 right-0 p-4 opacity-[0.1] text-6xl group-hover:scale-110 transition-transform">💰</div>
               <h3 className="text-xs font-mono-stack text-gold-dark uppercase mb-4 tracking-widest">{t('pages.admin.valuation.revenue_label', 'ESTIMATED RETAIL VALUE')}</h3>
               <div className="text-4xl font-display text-gold-dark leading-none">{formatCurrency(data.total_value_cop)}</div>
               <div className="gold-line my-4 opacity-30" />
               <p className="text-[10px] font-mono-stack text-gold-dark/60 leading-relaxed uppercase">
                 Total projected revenue if all items are sold at current list prices.
               </p>
            </div>

            {/* Cost Basis & Potential Profit */}
            <div className="card p-8 bg-ink-deep border-ink-border text-white shadow-2xl relative overflow-hidden group">
               <div className="absolute top-0 right-0 p-4 opacity-[0.1] text-6xl group-hover:scale-110 transition-transform">📊</div>
               <h3 className="text-xs font-mono-stack text-kraft-light/50 uppercase mb-4 tracking-widest">{t('pages.admin.valuation.profit_label', 'PROJECTED PROFITABILITY')}</h3>
               
               <div className="space-y-4">
                 <div>
                   <span className="block text-[10px] font-mono-stack text-kraft-light/40 uppercase mb-1">{t('pages.admin.accounting.total_cost', 'TOTAL COST BASIS')}</span>
                   <span className="text-2xl font-display text-kraft-light leading-none">{formatCurrency(data.total_cost_basis_cop)}</span>
                 </div>
                 <div className="pt-4 border-t border-white/10">
                   <div className="flex justify-between items-baseline">
                     <span className="block text-[10px] font-mono-stack text-green-400/60 uppercase mb-1">{t('pages.admin.accounting.net_profit', 'NET POTENTIAL PROFIT')}</span>
                     <span className="text-[10px] font-mono-stack text-green-400 font-bold bg-green-500/10 px-2 py-0.5 rounded border border-green-500/20">
                        {((data.potential_profit / data.total_value_cop) * 100).toFixed(1)}% MARGIN
                     </span>
                   </div>
                   <span className="text-3xl font-display text-green-400 leading-none">{formatCurrency(data.potential_profit)}</span>
                 </div>
               </div>
            </div>

            {/* Disclaimer / Analysis */}
            <div className="md:col-span-2 lg:col-span-3 card p-6 bg-surface/30 border-dashed border-kraft-dark/30 flex flex-col items-center justify-center text-center">
              <span className="text-3xl mb-3">⚓</span>
              <p className="text-sm font-semibold max-w-2xl text-ink-deep" style={{ lineHeight: '1.6' }}>
                {t('pages.admin.valuation.disclaimer', 'Financial valuations are calculated based on the cost basis entered in the product inventory and current retail prices. Ensure all items have a cost basis for 100% accurate profit tracking.')}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
