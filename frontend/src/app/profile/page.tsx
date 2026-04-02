'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import { useRouter, useSearchParams } from 'next/navigation';
import { useUser } from '@/context/UserContext';
import { userFetchOrders } from '@/lib/api';
import { UserOrder, ORDER_STATUS_LABELS } from '@/lib/types';
import LoadingSpinner from '@/components/LoadingSpinner';
import OrderDetailsModal from '@/components/OrderDetailsModal';

export default function ProfilePage() {
  const { user, loading: userLoading, updateProfile, loginWithGoogle, loginWithFacebook } = useUser();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [orders, setOrders] = useState<UserOrder[]>([]);
  const [ordersLoading, setOrdersLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  const [selectedOrderId, setSelectedOrderId] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Form State
  const [formData, setFormData] = useState({
    phone: '',
    id_number: '',
    address: ''
  });

  const errorParam = searchParams.get('error');

  useEffect(() => {
    if (!userLoading && !user) {
      router.push('/');
    }
    if (user) {
      setFormData({
        phone: user.phone || '',
        id_number: user.id_number || '',
        address: user.address || ''
      });
      fetchOrders();
    }
  }, [user, userLoading, router]);

  useEffect(() => {
    if (errorParam === 'already_linked') {
      setMessage({ type: 'error', text: 'This account is already linked to another profile.' });
    } else if (errorParam === 'link_failed') {
      setMessage({ type: 'error', text: 'Failed to link account. Please try again.' });
    }
  }, [errorParam]);

  const fetchOrders = async () => {
    try {
      const data = await userFetchOrders();
      setOrders(data);
    } catch {
      console.error('Failed to fetch orders');
    } finally {
      setOrdersLoading(false);
    }
  };

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      await updateProfile(formData);
      setMessage({ type: 'success', text: 'Profile updated successfully!' });
    } catch {
      setMessage({ type: 'error', text: 'Failed to update profile.' });
    } finally {
      setSaving(false);
    }
  };

  if (userLoading || (!user && userLoading)) {
    return <div className="min-h-screen flex items-center justify-center"><LoadingSpinner /></div>;
  }

  if (!user) return null;

  const isGoogleLinked = user.linked_providers?.includes('google');
  const isFacebookLinked = user.linked_providers?.includes('facebook');

  return (
    <div className="min-h-screen bg-bg-page py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-6xl mx-auto space-y-8">

        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-display text-text-main">Your Account</h1>
            <p className="text-text-muted mt-1">Manage your profile and view your order history.</p>
          </div>
          <div className="flex items-center gap-3">
            <div className="text-right hidden sm:block">
              <p className="text-sm font-medium text-text-main">{user.first_name} {user.last_name}</p>
              <p className="text-xs text-text-muted">{user.email}</p>
            </div>
            <div className="w-12 h-12 relative rounded-full border-2 border-accent-primary overflow-hidden">
              <Image
                src={user.avatar_url || 'https://www.gravatar.com/avatar/?d=mp'}
                alt="Avatar"
                fill
                className="object-cover"
              />
            </div>
          </div>
        </div>

        {message && (
          <div className={`p-4 rounded-md border ${message.type === 'success' ? 'bg-green-500/10 border-green-500/50 text-green-400' : 'bg-red-500/10 border-red-500/50 text-red-400'}`}>
            {message.text}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

          {/* Sidebar - Profile Info */}
          <div className="lg:col-span-1 space-y-6">
            <div className="glass-card p-6 border border-border-main rounded-xl bg-bg-surface/50 backdrop-blur-md">
              <h2 className="text-xl font-medium text-text-main mb-6 flex items-center gap-2">
                <span className="text-accent-primary">👤</span> Profile Settings
              </h2>

              <form onSubmit={handleUpdateProfile} className="space-y-4">
                <div>
                  <label className="block text-xs font-mono text-text-muted uppercase mb-1">Full Name</label>
                  <input 
                    type="text" 
                    disabled 
                    value={`${user.first_name} ${user.last_name}`}
                    className="w-full bg-bg-page/40 border border-border-main/30 rounded p-2 text-text-muted cursor-not-allowed"
                  />
                </div>
                <div>
                  <label className="block text-xs font-mono text-text-muted uppercase mb-1">Email Address</label>
                  <input 
                    type="text" 
                    disabled 
                    value={user.email || ''}
                    className="w-full bg-bg-page/40 border border-border-main/30 rounded p-2 text-text-muted cursor-not-allowed"
                  />
                </div>
                <div>
                  <label className="block text-xs font-mono text-text-muted uppercase mb-1">Phone Number</label>
                  <input 
                    type="tel" 
                    value={formData.phone}
                    onChange={(e) => setFormData({...formData, phone: e.target.value})}
                    placeholder="e.g. +57 300..."
                    className="w-full bg-bg-card border border-border-main rounded p-2 text-text-main focus:border-accent-primary outline-none transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-mono text-text-muted uppercase mb-1">ID Number / Cedula</label>
                  <input 
                    type="text" 
                    value={formData.id_number}
                    onChange={(e) => setFormData({...formData, id_number: e.target.value})}
                    placeholder="Required for shipping in Colombia"
                    className="w-full bg-bg-card border border-border-main rounded p-2 text-text-main focus:border-accent-primary outline-none transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-mono text-text-muted uppercase mb-1">Shipping Address</label>
                  <textarea 
                    rows={3}
                    value={formData.address}
                    onChange={(e) => setFormData({...formData, address: e.target.value})}
                    placeholder="Full address, city, and instructions"
                    className="w-full bg-bg-card border border-border-main rounded p-2 text-text-main focus:border-accent-primary outline-none transition-colors resize-none"
                  />
                </div>

                <button
                  type="submit"
                  disabled={saving}
                  className="w-full py-3 bg-accent-primary hover:bg-accent-primary-hover text-text-on-accent font-bold rounded shadow-lg shadow-accent-primary/20 transition-all disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save Profile'}
                </button>
              </form>
            </div>

            {/* Linked Accounts */}
            <div className="glass-card p-6 border border-border-main rounded-xl bg-bg-surface/50 backdrop-blur-md">
              <h2 className="text-xl font-medium text-text-main mb-6 flex items-center gap-2">
                <span className="text-accent-primary">🔗</span> Linked Accounts
              </h2>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 rounded bg-bg-page/40 border border-border-main/30">
                  <div className="flex items-center gap-3">
                    <svg width="20" height="20" viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" /><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" /><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" /><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" /></svg>
                    <span className="text-sm font-medium text-text-main">Google</span>
                  </div>
                  {isGoogleLinked ? (
                    <span className="text-xs text-green-400 flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-green-400"></span> Connected
                    </span>
                  ) : (
                    <button
                      onClick={loginWithGoogle}
                      className="text-xs px-2 py-1 bg-accent-primary/20 text-accent-primary border border-accent-primary/30 rounded hover:bg-accent-primary/30 transition-colors"
                    >
                      Link
                    </button>
                  )}
                </div>

                <div className="flex items-center justify-between p-3 rounded bg-bg-page/40 border border-border-main/30">
                  <div className="flex items-center gap-3">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="#1877F2"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" /></svg>
                    <span className="text-sm font-medium text-text-main">Facebook</span>
                  </div>
                  {isFacebookLinked ? (
                    <span className="text-xs text-green-400 flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-green-400"></span> Connected
                    </span>
                  ) : (
                    <button
                      onClick={loginWithFacebook}
                      className="text-xs px-2 py-1 bg-accent-primary/20 text-accent-primary border border-accent-primary/30 rounded hover:bg-accent-primary/30 transition-colors"
                    >
                      Link
                    </button>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Main - Order History */}
          <div className="lg:col-span-2">
            <div className="glass-card h-full border border-border-main rounded-xl bg-bg-surface/50 backdrop-blur-md overflow-hidden flex flex-col">
              <div className="p-6 border-b border-border-main">
                <h2 className="text-xl font-medium text-text-main flex items-center gap-2">
                  <span className="text-accent-primary">📦</span> Order History
                </h2>
              </div>

              <div className="flex-grow overflow-x-auto">
                {ordersLoading ? (
                  <div className="p-12 flex justify-center"><LoadingSpinner /></div>
                ) : orders.length === 0 ? (
                  <div className="p-12 text-center">
                    <div className="text-4xl mb-4 opacity-20">📭</div>
                    <p className="text-text-muted">You haven&apos;t placed any orders yet.</p>
                    <button
                      onClick={() => router.push('/')}
                      className="mt-6 text-accent-primary hover:underline font-medium"
                    >
                      Browse our inventory →
                    </button>
                  </div>
                ) : (
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="bg-bg-page/50 text-xs font-mono text-text-muted uppercase">
                        <th className="px-6 py-4 font-medium tracking-wider">Order No.</th>
                        <th className="px-6 py-4 font-medium tracking-wider">Date</th>
                        <th className="px-6 py-4 font-medium tracking-wider">Status</th>
                        <th className="px-6 py-4 font-medium tracking-wider">Items</th>
                        <th className="px-6 py-4 font-medium tracking-wider text-right">Total</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border-main/50">
                      {orders.map((order) => (
                        <tr
                          key={order.id}
                          onClick={() => {
                            setSelectedOrderId(order.id);
                            setIsModalOpen(true);
                          }}
                          className="hover:bg-accent-primary/5 transition-colors group cursor-pointer"
                        >
                          <td className="px-6 py-4">
                            <span className="font-mono text-sm text-accent-primary group-hover:font-bold transition-all">
                              {order.order_number}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-sm text-text-secondary">
                            {new Date(order.created_at).toLocaleDateString()}
                          </td>
                          <td className="px-6 py-4">
                            <span className={`text-[10px] uppercase font-bold px-2 py-1 rounded-sm border ${order.status === 'completed' ? 'text-green-400 border-green-400/30 bg-green-400/10' :
                                order.status === 'pending' ? 'text-yellow-400 border-yellow-400/30 bg-yellow-400/10' :
                                  order.status === 'cancelled' ? 'text-red-400 border-red-400/30 bg-red-400/10' :
                                    'text-blue-400 border-blue-400/30 bg-blue-400/10'
                              }`}>
                              {ORDER_STATUS_LABELS[order.status] || order.status}
                            </span>
                          </td>
                          <td className="px-6 py-4 text-sm text-text-muted">
                            {order.item_count} items
                          </td>
                          <td className="px-6 py-4 text-right">
                            <span className="font-display text-text-main">
                              ${order.total_cop.toLocaleString('es-CO')}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            </div>
          </div>

        </div>
      </div>

      <OrderDetailsModal 
        isOpen={isModalOpen}
        orderId={selectedOrderId}
        onClose={() => setIsModalOpen(false)}
      />
    </div>
  );
}
