'use client';

import { useEffect, useState } from 'react';
import { adminFetchLowStock } from '@/lib/api';
import { Product } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import CardImage from '@/components/CardImage';
import LoadingSpinner from '@/components/LoadingSpinner';
import { useLanguage } from '@/context/LanguageContext';

export default function LowStockPage() {
  const { t } = useLanguage();
  const { token, loading: authLoading } = useAdmin();
  const [products, setProducts] = useState<Product[]>([]);
  const [threshold, setThreshold] = useState(5);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    const load = async () => {
      setLoading(true);
      try {
        const data = await adminFetchLowStock(threshold);
        setProducts(data);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [token, threshold]);

  if (authLoading || !token) {
    return <div className="p-12 text-center font-mono uppercase animate-pulse">Authenticating Admin...</div>;
  }

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title={t('pages.admin.low_stock.title', 'LOW STOCK ALERTS')} 
        subtitle={t('pages.admin.low_stock.subtitle', 'Products below the critical inventory threshold')} 
      />

      <div className="flex items-center gap-4 mb-6 p-4 card bg-hp-color/5 border-hp-color/20">
        <label className="text-xs font-mono-stack uppercase font-bold text-hp-color">
          {t('pages.admin.low_stock.threshold_label', 'ALERT THRESHOLD:')}
        </label>
        <input 
          type="number" 
          value={threshold} 
          onChange={e => setThreshold(parseInt(e.target.value) || 0)}
          className="w-20 text-center font-bold bg-white border-hp-color/30 text-hp-color"
        />
        <span className="text-[10px] font-mono-stack opacity-60">
          {t('pages.admin.low_stock.description', 'Showing products with total stock less than or equal to this value.')}
        </span>
      </div>

      <div className="flex-1 min-h-0 bg-white shadow-sm border border-kraft-dark/20 rounded overflow-auto">
        {loading ? (
          <LoadingSpinner />
        ) : products.length === 0 ? (
          <div className="p-20 text-center opacity-30 italic">
            {t('pages.admin.low_stock.no_alerts', 'No low stock alerts at this threshold. Great job!')}
          </div>
        ) : (
          <table className="w-full text-left border-collapse">
            <thead className="sticky top-0 bg-kraft-light/50 backdrop-blur-sm z-10">
              <tr className="border-b border-kraft-dark/20">
                <th className="p-3 text-[10px] font-mono-stack uppercase tracking-widest">{t('pages.admin.inventory.table.product', 'PRODUCT')}</th>
                <th className="p-3 text-[10px] font-mono-stack uppercase tracking-widest">{t('pages.admin.inventory.table.set', 'SET')}</th>
                <th className="p-3 text-[10px] font-mono-stack uppercase tracking-widest text-center">{t('pages.admin.inventory.table.stock', 'STOCK')}</th>
                <th className="p-3 text-[10px] font-mono-stack uppercase tracking-widest text-right">{t('pages.admin.inventory.table.price', 'PRICE')}</th>
              </tr>
            </thead>
            <tbody>
              {products.map(p => (
                <tr key={p.id} className="border-b border-kraft-dark/10 hover:bg-kraft-light/20 transition-colors">
                  <td className="p-3">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-11 shrink-0">
                        <CardImage imageUrl={p.image_url} name={p.name} tcg={p.tcg} foilTreatment={p.foil_treatment} />
                      </div>
                      <div>
                        <p className="text-sm font-bold leading-tight">{p.name}</p>
                        <p className="text-[10px] opacity-60 uppercase">{p.tcg} / {p.category}</p>
                      </div>
                    </div>
                  </td>
                  <td className="p-3 text-xs opacity-70">
                    {p.set_name} ({p.set_code?.toUpperCase()})
                  </td>
                  <td className="p-3 text-center">
                    <span className={`font-mono-stack font-black text-lg ${p.stock === 0 ? 'text-hp-color' : 'text-lp-color'}`}>
                      {p.stock}
                    </span>
                  </td>
                  <td className="p-3 text-right">
                    <span className="price text-sm">${p.price.toLocaleString()}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
