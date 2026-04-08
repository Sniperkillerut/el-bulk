'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import AdminHeader from '@/components/admin/AdminHeader';
import { useAdmin } from '@/hooks/useAdmin';
import { adminFetchClients } from '@/lib/api';
import { CustomerStats } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';

export default function AdminClientsPage() {
  const { token } = useAdmin();
  const router = useRouter();
  const [clients, setClients] = useState<CustomerStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 50;

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

  const totalPages = Math.ceil(filteredClients.length / itemsPerPage) || 1;
  const paginatedClients = filteredClients.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

  if (loading) return <LoadingSpinner />;

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 w-full max-w-full">
      <AdminHeader 
        title="CRM: Client Registry" 
        subtitle="Managing Relationships & Interactions"
        actions={
          <div className="w-full max-w-64">
            <input 
              type="text" 
              placeholder="SEARCH CLIENTS..." 
              value={search}
              onChange={(e) => { setSearch(e.target.value); setCurrentPage(1); }}
              className="text-xs font-mono-stack border-border-main"
            />
          </div>
        }
      />

      <div className="flex-1 min-h-0 overflow-auto cardbox scrollbar-thin rounded-lg border border-border-main bg-white">
        <table className="w-full text-left border-collapse table-fixed">
          <thead className="sticky top-0 z-10 bg-bg-header backdrop-blur-md shadow-sm border-b border-border-main">
            <tr className="border-b border-border-main">
              <th className="p-4 font-display text-xs uppercase tracking-wider w-[25%]">Customer</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider hidden md:table-cell w-[20%]">Contact</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center w-[10%]">Orders</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center text-gold-dark/80 hidden lg:table-cell w-[10%]">Requests</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center text-emerald-700/80 hidden lg:table-cell w-[10%]">Offers</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-right w-[15%]">Spend (COP)</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center hidden md:table-cell w-[10%]">Newsletter</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider hidden lg:table-cell w-[20%]">Recent Journal Entry</th>
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
                      <div className="text-[10px] font-mono-stack text-text-muted truncate">ID: {client.id.slice(0,8)}</div>
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
                    <span className="badge badge-foil text-[9px]">SUBSCRIBED</span>
                  ) : (
                    <span className="text-[9px] opacity-30 font-mono-stack">NO</span>
                  )}
                </td>
                <td className="p-4 text-xs font-mono-stack transition-opacity group-hover:opacity-100 opacity-70 hidden lg:table-cell">
                   {client.latest_note ? (
                     <div className="truncate italic">&quot;{client.latest_note}&quot;</div>
                   ) : (
                     <span className="opacity-30 italic">No notes recorded...</span>
                   )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex flex-col sm:flex-row justify-between items-center gap-2 mt-3 px-1 flex-shrink-0">
          <div className="text-[10px] font-mono-stack text-text-muted uppercase tracking-widest font-bold">
            SHOWING {(currentPage - 1) * itemsPerPage + 1} - {Math.min(currentPage * itemsPerPage, filteredClients.length)} OF {filteredClients.length}
          </div>
          <div className="flex items-center gap-3">
            <button 
              disabled={currentPage === 1} 
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
              className="btn-secondary py-1.5 px-4 text-[10px] font-bold disabled:opacity-30 disabled:cursor-not-allowed"
            >
              PREV
            </button>
            <span className="text-[10px] font-mono-stack font-bold px-3 py-1.5 bg-white border border-border-main rounded shadow-sm">
              {currentPage} / {totalPages}
            </span>
            <button 
              disabled={currentPage === totalPages} 
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
              className="btn-secondary py-1.5 px-4 text-[10px] font-bold disabled:opacity-30 disabled:cursor-not-allowed"
            >
              NEXT
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
