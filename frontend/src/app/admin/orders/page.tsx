'use client';

import { useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useAdmin } from '@/hooks/useAdmin';
import { adminDownloadAccountingCSV } from '@/lib/api';
import { useLanguage } from '@/context/LanguageContext';
import AdminHeader from '@/components/admin/AdminHeader';
import OrdersPanel from '@/components/admin/OrdersPanel';

export default function AdminOrdersPage() {
  const { t } = useLanguage();
  const { token, loading } = useAdmin();
  const searchParams = useSearchParams();
  const initialOrderId = searchParams.get('id');

  // Accounting export state
  const [exportDates, setExportDates] = useState({ start: '', end: '' });
  const [exporting, setExporting] = useState(false);

  const handleExport = async () => {
    setExporting(true);
    try {
      await adminDownloadAccountingCSV({
        start_date: exportDates.start ? `${exportDates.start}T00:00:00Z` : undefined,
        end_date: exportDates.end ? `${exportDates.end}T23:59:59Z` : undefined,
      });
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Error exporting accounting data');
    }
    setExporting(false);
  };

  if (loading || !token) {
    return (
      <div className="min-h-screen bg-ink-deep flex items-center justify-center">
        <div className="text-gold font-mono-stack animate-pulse uppercase">Authenticating...</div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col p-2 lg:p-3 min-h-0 w-full max-w-full">
      <AdminHeader 
        title="ORDER MANAGEMENT" 
        subtitle="Reviewing and Fulfilling Customer Card Orders" 
        actions={
          <div className="flex flex-col sm:flex-row gap-2 items-center sm:items-end bg-gold/5 border border-gold/20 p-2 rounded-lg">
            <div className="hidden sm:block">
              <p className="text-[9px] font-mono-stack text-gold-dark mb-1 font-bold uppercase tracking-wider">
                {t('pages.admin.orders.accounting_export', 'Accounting Export (CSV)')}
              </p>
            </div>
            <div className="flex gap-2 items-end">
              <div>
                <label className="text-[8px] font-mono-stack text-text-muted block uppercase">
                  {t('pages.common.dates.start', 'Start')}
                </label>
                <input 
                  type="date" 
                  value={exportDates.start} 
                  onChange={e => setExportDates(prev => ({ ...prev, start: e.target.value }))}
                  className="w-full text-[10px] p-1 bg-white border-kraft-dark/20 h-8"
                />
              </div>
              <div>
                <label className="text-[8px] font-mono-stack text-text-muted block uppercase">
                  {t('pages.common.dates.end', 'End')}
                </label>
                <input 
                  type="date" 
                  value={exportDates.end} 
                  onChange={e => setExportDates(prev => ({ ...prev, end: e.target.value }))}
                  className="w-full text-[10px] p-1 bg-white border-kraft-dark/20 h-8"
                />
              </div>
              <button 
                onClick={handleExport}
                disabled={exporting}
                className="btn-primary text-[10px] h-8 !py-0 px-3 whitespace-nowrap shadow-sm flex items-center justify-center min-w-[80px] leading-none"
              >
                {exporting ? '...' : t('pages.admin.orders.export_csv', 'EXPORT CSV')}
              </button>
            </div>
          </div>
        }
      />
      
      <div className="flex-1 min-h-0 bg-white shadow-sm border border-kraft-dark/20 rounded-lg overflow-hidden flex flex-col">
        <OrdersPanel initialOrderId={initialOrderId} />
      </div>
    </div>
  );
}
