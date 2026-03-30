'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { useAdmin } from '@/hooks/useAdmin';
import { adminFetchClientDetail, adminAddCustomerNote } from '@/lib/api';
import { CustomerDetail, CustomerNote } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';

export default function AdminClientDetailPage() {
  const { token } = useAdmin();
  const { id } = useParams();
  const [detail, setDetail] = useState<CustomerDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [newNote, setNewNote] = useState('');
  const [selectedOrder, setSelectedOrder] = useState<string>('');

  const fetchDetail = async () => {
    if (token && id) {
      try {
        const data = await adminFetchClientDetail(token, id as string);
        setDetail(data);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    fetchDetail();
  }, [token, id]);

  const handleAddNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token || !id || !newNote) return;

    try {
      await adminAddCustomerNote(token, id as string, newNote, selectedOrder || undefined);
      setNewNote('');
      setSelectedOrder('');
      fetchDetail(); // Refresh notes
    } catch (err) {
      console.error(err);
      alert('Failed to add note');
    }
  };

  if (loading) return <LoadingSpinner />;
  if (!detail) return <div className="p-8 text-center">CLIENT NOT FOUND.</div>;

  return (
    <div className="p-8 max-w-6xl">
      <Link href="/admin/clients" className="text-[10px] font-mono-stack text-gold-dark mb-4 block no-underline hover:text-hp-color transition-colors">
        ← BACK TO CLIENTS
      </Link>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Left Column: Info & Notes */}
        <div className="flex-1 space-y-8">
          <div className="cardbox p-8">
            <div className="flex items-center gap-6 mb-8">
              <div className="w-20 h-20 rounded-lg bg-kraft-dark flex items-center justify-center text-white text-4xl font-display">
                {detail.first_name[0]}{detail.last_name[0]}
              </div>
              <div>
                <h1 className="text-5xl mb-1 uppercase leading-none">{detail.first_name} {detail.last_name}</h1>
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-mono-stack text-text-muted bg-kraft-light px-2 py-0.5 border border-kraft-shadow">ID: {detail.id}</span>
                  {detail.is_subscriber && (
                    <span className="badge badge-foil text-[9px]">NEWSLETTER SUBSCRIBER</span>
                  )}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 pt-8 border-t border-ink-border/30">
               <div>
                  <h4 className="text-xs font-mono-stack font-bold text-hp-color uppercase tracking-widest mb-4">Contact Profile</h4>
                  <div className="space-y-3">
                    <div>
                      <p className="text-[10px] text-text-muted uppercase font-mono-stack">Email address</p>
                      <p className="font-mono-stack text-sm">{detail.email || 'N/A'}</p>
                    </div>
                    <div>
                      <p className="text-[10px] text-text-muted uppercase font-mono-stack">Phone number</p>
                      <p className="font-mono-stack text-sm">{detail.phone || 'N/A'}</p>
                    </div>
                  </div>
               </div>
               <div>
                  <h4 className="text-xs font-mono-stack font-bold text-hp-color uppercase tracking-widest mb-4">Account Details</h4>
                  <div className="space-y-3">
                    <div>
                      <p className="text-[10px] text-text-muted uppercase font-mono-stack">Identity Document</p>
                      <p className="font-mono-stack text-sm">{detail.id_number || 'NOT PROVIDED'}</p>
                    </div>
                    <div>
                        <p className="text-[10px] text-text-muted uppercase font-mono-stack">Delivery Address</p>
                        <p className="font-mono-stack text-sm leading-relaxed">{detail.address || 'NO ADDRESS RECORDED'}</p>
                    </div>
                  </div>
               </div>
            </div>
          </div>

          <div className="cardbox p-8">
            <h3 className="text-2xl mb-6 flex items-center gap-3">
              JOURNAL OF INTERACTIONS 
              <span className="text-[10px] font-mono-stack bg-kraft-dark text-white px-2 py-0.5 rounded-full">{detail.notes.length}</span>
            </h3>
            
            <form onSubmit={handleAddNote} className="mb-10 bg-kraft-light/20 p-4 border border-dashed border-kraft-dark rounded-sm">
                <textarea 
                  value={newNote}
                  onChange={(e) => setNewNote(e.target.value)}
                  placeholder="ADD A NEW NOTE OR COMMENT..."
                  className="mb-3 text-xs font-mono-stack h-24"
                />
                <div className="flex items-center justify-between gap-4">
                    <select 
                      value={selectedOrder} 
                      onChange={(e) => setSelectedOrder(e.target.value)}
                      className="text-[10px] font-mono-stack flex-1"
                    >
                      <option value="">General Interaction / No specific order</option>
                      {detail.orders.map(o => (
                        <option key={o.id} value={o.id}>About Order {o.order_number}</option>
                      ))}
                    </select>
                    <button type="submit" className="btn-primary py-2 px-6">ADD NOTE</button>
                </div>
            </form>

            <div className="space-y-4">
              {detail.notes.length === 0 ? (
                <div className="text-center py-8 text-[10px] font-mono-stack text-text-muted uppercase tracking-widest italic opacity-40">No entries recorded in the journal.</div>
              ) : (
                detail.notes.map((note) => (
                  <div key={note.id} className="p-4 bg-ink-surface border-l-4 border-kraft-dark rounded-sm shadow-sm">
                    <div className="flex justify-between items-start mb-2">
                       <div className="flex items-center gap-2">
                         <span className="text-[10px] font-bold uppercase">{note.admin_name || 'System Admin'}</span>
                         <span className="text-[9px] font-mono-stack text-text-muted">Added on {new Date(note.created_at).toLocaleString()}</span>
                       </div>
                       {note.order_id && (
                         <span className="text-[9px] font-mono-stack bg-gold/10 text-gold-dark border border-gold/30 px-2 py-0.5">REFERENCE: {detail.orders.find(o => o.id === note.order_id)?.order_number}</span>
                       )}
                    </div>
                    <p className="text-xs font-mono-stack text-text-secondary leading-relaxed whitespace-pre-wrap">{note.content}</p>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Orders */}
        <div className="lg:w-96 space-y-8">
            <div className="cardbox p-6">
                <h3 className="text-2xl mb-6">ORDER HISTORY</h3>
                <div className="space-y-3">
                  {detail.orders.length === 0 ? (
                    <div className="p-8 text-center text-xs font-mono-stack opacity-40">NO ORDERS FOUND.</div>
                  ) : (
                    detail.orders.map(order => (
                      <Link 
                        key={order.id} 
                        href={`/admin/orders/${order.id}`}
                        className="block p-4 border border-ink-border/30 hover:border-gold transition-all no-underline group hover:translate-x-1"
                      >
                        <div className="flex justify-between items-center mb-1">
                          <span className="font-bold text-xs font-mono-stack text-ink-deep group-hover:text-gold-dark">#{order.order_number}</span>
                          <span className={`text-[9px] px-2 py-0.5 rounded-full font-bold uppercase ${
                            order.status === 'completed' ? 'bg-emerald-100 text-emerald-800' : 
                            order.status === 'pending' ? 'bg-amber-100 text-amber-800' : 
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {order.status}
                          </span>
                        </div>
                        <div className="flex justify-between items-end">
                           <span className="text-[10px] font-mono-stack text-text-muted">{new Date(order.created_at).toLocaleDateString()}</span>
                           <span className="text-xs font-bold font-mono-stack">${order.total_cop.toLocaleString()}</span>
                        </div>
                      </Link>
                    ))
                  )}
                </div>
            </div>

            <div className="p-6 bg-kraft-dark/10 border-2 border-dotted border-kraft-dark rounded-lg flex flex-col gap-4 text-center">
                 <h5 className="font-display text-sm tracking-widest text-kraft-dark">VALUED CLIENT SUMMARY</h5>
                 <div className="flex justify-around items-center border-y border-kraft-dark/20 py-4">
                    <div>
                      <p className="text-[10px] font-mono-stack text-text-muted tracking-tighter uppercase mb-1">Lifetime</p>
                      <p className="text-xl font-bold font-mono-stack leading-none">${detail.orders.reduce((sum, o) => sum + o.total_cop, 0).toLocaleString()}</p>
                    </div>
                    <div className="w-px h-8 bg-kraft-dark/20" />
                    <div>
                      <p className="text-[10px] font-mono-stack text-text-muted tracking-tighter uppercase mb-1">Purchased</p>
                      <p className="text-xl font-bold font-mono-stack leading-none">{detail.orders.length}</p>
                    </div>
                 </div>
            </div>
        </div>
      </div>
    </div>
  );
}
