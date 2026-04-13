'use client';

import { useState, useEffect, useCallback } from 'react';
import { adminFetchAuditLogs } from '@/lib/api';
import { AuditLog } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';
import { useAdmin } from '@/hooks/useAdmin';
import { format } from 'date-fns';

export default function AuditLogsPage() {
  const { t } = useLanguage();
  const { token } = useAdmin();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const [filters, setFilters] = useState({
    action: '',
    target_type: '',
  });

  const fetchLogs = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await adminFetchAuditLogs({ 
        ...filters, 
        page, 
        page_size: pageSize 
      });
      setLogs(res.logs);
      setTotal(res.total);
    } catch (err) {
      console.error('Failed to fetch audit logs', err);
    } finally {
      setLoading(false);
    }
  }, [token, filters, page]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleFilterChange = (field: string, value: string) => {
    setFilters(prev => ({ ...prev, [field]: value }));
    setPage(1);
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="flex flex-col h-full bg-kraft-paper">
      {/* Header */}
      <div className="p-6 border-b border-ink-border bg-ink-surface/5 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-2xl tracking-tight text-ink-navy uppercase">
            {t('pages.admin.audit.title', 'ACTION LOG')}
          </h1>
          <p className="text-xs font-mono-stack text-text-muted uppercase tracking-widest">
            {t('pages.admin.audit.subtitle', 'Administrative Accountability Ledger')}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <select 
            value={filters.action}
            onChange={(e) => handleFilterChange('action', e.target.value)}
            className="input-sm text-xs"
          >
            <option value="">{t('pages.admin.audit.filter.all_actions', 'All Actions')}</option>
            <option value="create">CREATE</option>
            <option value="update">UPDATE</option>
            <option value="delete">DELETE</option>
            <option value="login">LOGIN</option>
          </select>

          <select 
            value={filters.target_type}
            onChange={(e) => handleFilterChange('target_type', e.target.value)}
            className="input-sm text-xs"
          >
            <option value="">{t('pages.admin.audit.filter.all_targets', 'All Targets')}</option>
            <option value="product">PRODUCT</option>
            <option value="order">ORDER</option>
            <option value="admin">ADMIN</option>
          </select>

          <button 
            onClick={fetchLogs}
            className="btn-secondary py-1 px-3 text-xs"
          >
            {t('pages.admin.audit.refresh', 'REFRESH')}
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        <div className="card shadow-md border-ink-border/40 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead className="bg-ink-navy text-white font-mono-stack text-[10px] uppercase tracking-widest border-b border-ink-border">
                <tr>
                  <th className="px-4 py-3">{t('pages.admin.audit.table.timestamp', 'TIMESTAMP')}</th>
                  <th className="px-4 py-3">{t('pages.admin.audit.table.admin', 'ADMINISTRATOR')}</th>
                  <th className="px-4 py-3">{t('pages.admin.audit.table.action', 'ACTION')}</th>
                  <th className="px-4 py-3">{t('pages.admin.audit.table.target', 'TARGET')}</th>
                  <th className="px-4 py-3">{t('pages.admin.audit.table.details', 'DETAILS')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-ink-border/10">
                {loading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-sm italic text-text-muted animate-pulse">
                      {t('pages.admin.audit.loading', 'Synchronizing ledger...')}
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-sm text-text-muted italic">
                      {t('pages.admin.audit.empty', 'No administrative actions recorded.')}
                    </td>
                  </tr>
                ) : (
                  logs.map((log) => (
                    <tr key={log.id} className="hover:bg-gold/5 transition-colors text-xs font-mono-stack">
                      <td className="px-4 py-3 whitespace-nowrap text-text-muted">
                        {format(new Date(log.created_at), 'yyyy-MM-dd HH:mm:ss')}
                      </td>
                      <td className="px-4 py-3 font-bold text-ink-navy">
                        {log.admin_name}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded-sm border text-[10px] uppercase font-bold ${
                          log.action === 'create' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' :
                          log.action === 'delete' ? 'bg-hp-color/10 text-hp-color border-hp-color/20' :
                          log.action === 'update' ? 'bg-amber-50 text-amber-700 border-amber-200' :
                          'bg-ink-surface text-text-muted border-ink-border'
                        }`}>
                          {log.action}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span className="text-text-muted uppercase text-[10px]">{log.target_type}</span>
                        <div className="text-[10px] opacity-60">ID: {log.target_id?.substring(0, 8) || 'N/A'}...</div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="max-w-[300px] overflow-hidden text-ellipsis whitespace-nowrap opacity-70 cursor-help group relative" 
                             onClick={() => alert(JSON.stringify(log.details, null, 2))}>
                          {JSON.stringify(log.details)}
                          <div className="absolute hidden group-hover:block bg-ink-navy text-white p-2 rounded shadow-xl z-50 text-[10px] whitespace-pre max-w-sm max-h-40 overflow-auto bottom-full left-0 mb-2 border border-gold/20">
                            {JSON.stringify(log.details, null, 2)}
                          </div>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="p-4 border-t border-ink-border/20 bg-ink-surface/5 flex items-center justify-between">
              <span className="text-xs font-mono-stack text-text-muted uppercase">
                {t('pages.admin.audit.pagination', 'Entries {{start}} - {{end}} of {{total}}', { 
                  start: (page - 1) * pageSize + 1,
                  end: Math.min(page * pageSize, total),
                  total
                })}
              </span>
              <div className="flex gap-2">
                <button
                  disabled={page === 1}
                  onClick={() => setPage(p => p - 1)}
                  className="btn-secondary py-1 px-3 text-xs disabled:opacity-30"
                >
                  {t('common.prev', 'PREV')}
                </button>
                <button
                  disabled={page === totalPages}
                  onClick={() => setPage(p => p + 1)}
                  className="btn-secondary py-1 px-3 text-xs disabled:opacity-30"
                >
                  {t('common.next', 'NEXT')}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
