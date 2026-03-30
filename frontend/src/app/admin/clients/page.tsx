'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useAdmin } from '@/hooks/useAdmin';
import { adminFetchClients } from '@/lib/api';
import { CustomerStats } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';

export default function AdminClientsPage() {
  const { token } = useAdmin();
  const [clients, setClients] = useState<CustomerStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (token) {
      adminFetchClients(token)
        .then(setClients)
        .catch(err => console.error(err))
        .finally(() => setLoading(false));
    }
  }, [token]);

  const filteredClients = clients.filter(c => 
    `${c.first_name} ${c.last_name}`.toLowerCase().includes(search.toLowerCase()) ||
    c.email?.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <LoadingSpinner />;

  return (
    <div className="p-8">
      <div className="flex justify-between items-end mb-8">
        <div>
          <h1 className="text-4xl mb-2">CLIENTS</h1>
          <p className="font-mono-stack text-xs text-text-muted uppercase tracking-widest">Customer Relationship Management</p>
        </div>
        <div className="w-64">
           <input 
            type="text" 
            placeholder="SEARCH CLIENTS..." 
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="text-xs font-mono-stack"
           />
        </div>
      </div>

      <div className="cardbox overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-kraft-mid/30 border-b border-ink-border">
              <th className="p-4 font-display text-xs uppercase tracking-wider">Customer</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider">Contact</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center">Orders</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-right">Total Spended</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-center">Newsletter</th>
              <th className="p-4 font-display text-xs uppercase tracking-wider text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ink-border/30">
            {filteredClients.map((client) => (
              <tr key={client.id} className="hover:bg-gold/5 transition-colors group">
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-kraft-dark flex items-center justify-center text-white font-display text-xs">
                      {client.first_name[0]}{client.last_name[0]}
                    </div>
                    <div>
                      <div className="font-bold text-sm uppercase">{client.first_name} {client.last_name}</div>
                      <div className="text-[10px] font-mono-stack text-text-muted">ID: {client.id.slice(0,8)}</div>
                    </div>
                  </div>
                </td>
                <td className="p-4 text-xs font-mono-stack">
                  <div>{client.email || 'N/A'}</div>
                  <div className="opacity-60">{client.phone || ''}</div>
                </td>
                <td className="p-4 text-center font-mono-stack text-xs font-bold">
                  {client.order_count}
                </td>
                <td className="p-4 text-right font-mono-stack text-xs text-emerald-700 font-bold">
                   ${client.total_spend.toLocaleString()}
                </td>
                <td className="p-4 text-center">
                  {client.is_subscriber ? (
                    <span className="badge badge-foil text-[9px]">SUBSCRIBED</span>
                  ) : (
                    <span className="text-[9px] opacity-30 font-mono-stack">NO</span>
                  )}
                </td>
                <td className="p-4 text-right">
                  <Link 
                    href={`/admin/clients/${client.id}`}
                    className="btn-secondary py-1 px-3 text-[10px]"
                  >
                    VIEW PROFILE
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
