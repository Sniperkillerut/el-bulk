'use client';

import { useState, useEffect } from 'react';
import { useUser } from '@/context/UserContext';
import { subscribeToNewsletter } from '@/lib/api';

interface NewsletterFormProps {
  compact?: boolean;
}

export default function NewsletterForm({ compact = false }: NewsletterFormProps) {
  const { user } = useUser();
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (user?.email) {
      setEmail(user.email);
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) return;

    setStatus('loading');
    try {
      await subscribeToNewsletter(email);
      setStatus('success');
      setMessage('THX! CHECK YOUR INBOX SOON.');
      if (!user) setEmail(''); // Clear only if not logged in
    } catch (err: any) {
      setStatus('error');
      setMessage(err.message || 'SOMETHING WENT WRONG.');
    }
  };

  if (status === 'success') {
    return (
      <div className="p-4 border border-gold bg-gold/5 rounded-sm animate-fade-up">
        <p className="font-display text-sm text-gold-dark text-center">{message}</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className={`flex flex-col gap-3 ${compact ? 'max-w-xs' : 'w-full'}`}>
      <div className="flex gap-2">
        <input 
          type="email" 
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="YOUR EMAIL" 
          required
          disabled={status === 'loading'}
          className="flex-1 bg-surface border border-kraft-shadow px-3 py-2 text-xs font-mono-stack focus:outline-gold-dark disabled:opacity-50" 
        />
        <button 
          type="submit"
          disabled={status === 'loading' || !email}
          className="btn-primary py-2 px-4 text-xs disabled:opacity-50 min-w-[60px]"
        >
          {status === 'loading' ? '...' : 'OK'}
        </button>
      </div>
      {status === 'error' && (
        <p className="text-[10px] font-mono-stack text-hp-color px-1">{message}</p>
      )}
    </form>
  );
}
