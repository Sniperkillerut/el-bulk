'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useAdmin } from '@/hooks/useAdmin';
import { adminFetchSubscribers } from '@/lib/api';
import { NewsletterSubscriber } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';
import AdminHeader from '@/components/admin/AdminHeader';

export default function AdminSubscribersPage() {
  const { token } = useAdmin();
  const [subscribers, setSubscribers] = useState<NewsletterSubscriber[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      adminFetchSubscribers()
        .then(setSubscribers)
        .catch(err => console.error(err))
        .finally(() => setLoading(false));
    }
  }, [token]);

  if (loading) return <LoadingSpinner />;

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title="SUBSCRIBERS"
        subtitle="Newsletter Management"
        actions={
          <div className="text-right">
             <span className="font-mono-stack text-[10px] text-text-muted uppercase font-bold tracking-tighter">Secure Link Active</span>
             <div className="flex items-center gap-2 mt-1">
                <div className="w-1.5 h-1.5 rounded-full bg-lp-color animate-pulse"></div>
                <span className="text-xs font-bold font-mono-stack">{subscribers.length} Emails</span>
             </div>
          </div>
        }
      />

      <div className="flex-1 min-h-0 overflow-auto cardbox scrollbar-thin max-w-4xl">
        <table className="w-full text-left border-collapse">
          <thead className="sticky top-0 z-10 bg-kraft-light backdrop-blur-md shadow-sm border-b border-ink-border">
            <tr className="border-b border-ink-border">
              <th className="p-2 font-display text-xs uppercase tracking-wider">Email Address</th>
              <th className="p-2 font-display text-xs uppercase tracking-wider hidden md:table-cell">Linked User</th>
              <th className="p-2 font-display text-xs uppercase tracking-wider">Subscription Date</th>
              <th className="p-2 font-display text-xs uppercase tracking-wider text-right hidden md:table-cell">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-ink-border/30">
            {subscribers.map((sub) => (
              <tr key={sub.id} className="hover:bg-gold/5 transition-colors group">
                <td className="p-2 text-sm font-mono-stack font-bold">
                  {sub.email}
                </td>
                <td className="p-2 hidden md:table-cell">
                  {sub.customer_id ? (
                    <Link 
                      href={`/admin/clients/${sub.customer_id}`}
                      className="flex items-center gap-2 group no-underline text-gold-dark hover:text-hp-color transition-colors"
                    >
                      <div className="w-6 h-6 rounded-full bg-kraft-dark flex items-center justify-center text-white font-display text-[9px]">
                        {sub.first_name?.[0] || '?'}{sub.last_name?.[0] || '?'}
                      </div>
                      <span className="text-xs font-bold uppercase">{sub.first_name} {sub.last_name}</span>
                    </Link>
                  ) : (
                    <span className="text-[10px] opacity-30 font-mono-stack italic">UNLINKED GUEST</span>
                  )}
                </td>
                <td className="p-2 text-[10px] font-mono-stack text-text-muted">
                  {new Date(sub.created_at).toLocaleDateString()}
                </td>
                <td className="p-2 text-right hidden md:table-cell">
                  <button className="text-[10px] font-mono-stack text-hp-color hover:underline bg-transparent border-none cursor-pointer p-0 opacity-40 hover:opacity-100">
                    REMOVE
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
