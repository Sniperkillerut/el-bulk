'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import AdminHeader from '@/components/admin/AdminHeader';
import { useAdmin } from '@/hooks/useAdmin';
import { useLanguage } from '@/context/LanguageContext';
import { adminFetchClients } from '@/lib/api';
import { CustomerStats } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';

const SortIcon = ({ field, sortField, sortOrder }: { field: keyof CustomerStats | 'name', sortField: string, sortOrder: string }) => {
  if (sortField !== field) return <span className="ml-1 opacity-20 group-hover:opacity-100 transition-opacity">↕</span>;
  return <span className="ml-1 text-gold-dark font-bold">{sortOrder === 'asc' ? '↑' : '↓'}</span>;
};

export default function AdminClientsPage() {
  const { t } = useLanguage();
  const { token } = useAdmin();
  const router = useRouter();
  const [clients, setClients] = useState<CustomerStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [sortField, setSortField] = useState<keyof CustomerStats | 'name'>('total_spend');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  useEffect(() => {
    if (token) {
      adminFetchClients()
        .then(setClients)
        .catch(err => console.error(err))
        .finally(() => setLoading(false));
    }
  }, [token]);

  const filteredClients = clients.filter(c =>
    `${c.first_name} ${c.last_name}`.toLowerCase().includes(search.toLowerCase()) ||
    c.email?.toLowerCase().includes(search.toLowerCase())
  );

  const sortedClients = [...filteredClients].sort((a, b) => {
    const valA = sortField === 'name' ? `${a.first_name} ${a.last_name}` : a[sortField as keyof CustomerStats];
    const valB = sortField === 'name' ? `${b.first_name} ${b.last_name}` : b[sortField as keyof CustomerStats];

    if (valA === valB) return 0;
    if (valA === null || valA === undefined) return 1;
    if (valB === null || valB === undefined) return -1;

    const modifier = sortOrder === 'asc' ? 1 : -1;
    if (typeof valA === 'string' && typeof valB === 'string') {
      return valA.localeCompare(valB) * modifier;
    }
    return ((valA as any) < (valB as any) ? -1 : 1) * modifier;
  });

  const totalPages = Math.ceil(sortedClients.length / pageSize) || 1;
  const paginatedClients = sortedClients.slice((currentPage - 1) * pageSize, currentPage * pageSize);

  const handleSort = (field: keyof CustomerStats | 'name') => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
    setCurrentPage(1);
  };

  if (loading) return <LoadingSpinner />;

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 w-full max-w-full">
      <AdminHeader
        title={t('pages.admin.clients.title', 'CRM: Client Registry')}
        subtitle={t('pages.admin.clients.subtitle', 'Managing Relationships & Interactions')}
        actions={
          <div className="w-full max-w-64">
            <input
              type="text"
              placeholder={t('pages.admin.inventory.search_placeholder', 'SEARCH CLIENTS...')}
              value={search}
              onChange={(e) => { setSearch(e.target.value); setCurrentPage(1); }}
              className="text-xs font-mono-stack border-border-main"
            />
          </div>
        }
      />

      <div className="flex-1 min-h-0 overflow-auto cardbox scrollbar-thin rounded-lg border border-border-main bg-white overscroll-contain" style={{ WebkitOverflowScrolling: 'touch' }}>
        <table className="w-full text-left border-collapse table-fixed min-w-[500px] lg:min-w-0">
          <thead className="sticky top-0 z-10 bg-bg-header backdrop-blur-md shadow-sm border-b border-border-main">
            <tr className="border-b border-border-main">
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider lg:w-[25%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('name')}
              >
                <div className="flex items-center">{t('pages.admin.clients.table.customer', 'Customer')} <SortIcon field="name" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider hidden md:table-cell w-[20%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('email')}
              >
                <div className="flex items-center">{t('pages.admin.clients.table.contact', 'Contact')} <SortIcon field="email" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider text-center w-[80px] lg:w-[10%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('order_count')}
              >
                <div className="flex items-center justify-center">{t('pages.admin.clients.table.orders', 'Orders')} <SortIcon field="order_count" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider text-center text-gold-dark/80 hidden lg:table-cell w-[10%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('active_request_count')}
              >
                <div className="flex items-center justify-center">{t('pages.admin.clients.table.requests', 'Requests')} <SortIcon field="active_request_count" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider text-center text-emerald-700/80 hidden lg:table-cell w-[10%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('active_offer_count')}
              >
                <div className="flex items-center justify-center">{t('pages.admin.clients.table.offers', 'Offers')} <SortIcon field="active_offer_count" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider text-right w-[120px] lg:w-[15%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('total_spend')}
              >
                <div className="flex items-center justify-end">{t('pages.admin.clients.table.spend', 'Spend (COP)')} <SortIcon field="total_spend" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider text-center hidden md:table-cell w-[10%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('is_subscriber')}
              >
                <div className="flex items-center justify-center">{t('pages.admin.clients.table.newsletter', 'Newsletter')} <SortIcon field="is_subscriber" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
              <th 
                className="p-4 font-display text-xs uppercase tracking-wider hidden lg:table-cell w-[20%] cursor-pointer group hover:bg-gold/5 transition-colors"
                onClick={() => handleSort('latest_note')}
              >
                <div className="flex items-center">{t('pages.admin.clients.table.recent_journal', 'Recent Journal Entry')} <SortIcon field="latest_note" sortField={sortField} sortOrder={sortOrder} /></div>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-main/30">
            {paginatedClients.map((client) => (
              <tr
                key={client.id}
                className="hover:bg-gold/5 transition-colors group cursor-pointer"
                onClick={() => router.push(`/admin/clients/${client.id}`)}
              >
                <td className="p-4 overflow-hidden">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-kraft-dark flex-shrink-0 flex items-center justify-center text-white font-display text-xs">
                      {client.first_name[0]}{client.last_name[0]}
                    </div>
                    <div className="min-w-0">
                      <div className="font-bold text-sm uppercase group-hover:text-gold-dark transition-colors truncate">{client.first_name} {client.last_name}</div>
                      <div className="text-[10px] font-mono-stack text-text-muted truncate">ID: {client.id.slice(0, 8)}</div>
                    </div>
                  </div>
                </td>
                <td className="p-4 text-xs font-mono-stack hidden md:table-cell overflow-hidden">
                  <div className="truncate">{client.email || 'N/A'}</div>
                  <div className="opacity-60 truncate">{client.phone || ''}</div>
                </td>
                <td className="p-4 text-center font-mono-stack text-xs font-bold">
                  {client.order_count}
                </td>
                <td className="p-4 text-center hidden lg:table-cell">
                  {client.request_count > 0 ? (
                    <span className="font-mono-stack text-[11px] font-bold">
                      {client.active_request_count} <span className="opacity-60 font-normal ml-0.5">({client.request_count})</span>
                    </span>
                  ) : (
                    <span className="text-[10px] opacity-20 font-mono-stack">0</span>
                  )}
                </td>
                <td className="p-4 text-center hidden lg:table-cell">
                  {client.offer_count > 0 ? (
                    <span className="font-mono-stack text-[11px] font-bold">
                      {client.active_offer_count} <span className="opacity-60 font-normal ml-0.5">({client.offer_count})</span>
                    </span>
                  ) : (
                    <span className="text-[10px] opacity-20 font-mono-stack">0</span>
                  )}
                </td>
                <td className="p-4 text-right font-mono-stack text-xs text-emerald-700 font-bold">
                  ${client.total_spend.toLocaleString()}
                </td>
                <td className="p-4 text-center hidden md:table-cell">
                  {client.is_subscriber ? (
                    <span className="badge badge-foil text-[9px]">{t('pages.admin.clients.status.subscribed', 'SUBSCRIBED')}</span>
                  ) : (
                    <span className="text-[9px] opacity-30 font-mono-stack">{t('pages.admin.clients.status.no', 'NO')}</span>
                  )}
                </td>
                <td className="p-4 text-xs font-mono-stack transition-opacity group-hover:opacity-100 opacity-70 hidden lg:table-cell">
                  {client.latest_note ? (
                    <div className="truncate italic">&quot;{client.latest_note}&quot;</div>
                  ) : (
                    <span className="opacity-30 italic">{t('pages.admin.clients.table.no_notes', 'No notes recorded...')}</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 0 && (
        <div className="flex flex-col sm:flex-row justify-between items-center gap-2 mt-3 p-3 bg-white shadow-sm border border-kraft-dark/20 rounded-lg flex-shrink-0">
          <div className="flex items-center gap-4">
            <div className="text-[10px] font-mono-stack text-text-muted uppercase tracking-widest font-bold">
              {t('pages.common.pagination.showing_crm', 'SHOWING {start} - {end} OF {total}', {
                start: (currentPage - 1) * pageSize + 1,
                end: Math.min(currentPage * pageSize, sortedClients.length),
                total: sortedClients.length
              })}
            </div>

            <div className="flex items-center gap-2 border-l border-border-main/20 pl-4">
              <span className="text-[10px] font-mono-stack text-text-muted uppercase">{t('pages.common.labels.mostrar', 'Show')}</span>
              <select
                value={pageSize}
                onChange={e => {
                  setPageSize(Number(e.target.value));
                  setCurrentPage(1);
                }}
                className="bg-transparent border-0 font-mono-stack text-xs font-bold text-gold-dark outline-none cursor-pointer hover:text-hp-color transition-colors"
              >
                <option value={10}>10</option>
                <option value={25}>25</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              disabled={currentPage === 1}
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
              className="btn-secondary py-1.5 px-4 text-[10px] font-bold disabled:opacity-30 disabled:cursor-not-allowed"
            >
              {t('pages.admin.dashboard.prev', 'PREV')}
            </button>
            <span className="text-[10px] font-mono-stack font-bold px-3 py-1.5 bg-kraft-paper/10 border border-border-main rounded shadow-sm">
              {currentPage} / {totalPages}
            </span>
            <button
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
              className="btn-secondary py-1.5 px-4 text-[10px] font-bold disabled:opacity-30 disabled:cursor-not-allowed"
            >
              {t('pages.admin.dashboard.next', 'NEXT')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
