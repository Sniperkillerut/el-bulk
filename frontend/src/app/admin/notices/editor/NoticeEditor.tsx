'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { adminFetchNotices, adminCreateNotice, adminUpdateNotice } from '@/lib/api';
import { NoticeInput } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';

export default function NoticeEditor() {
  const { id } = useParams();
  const { token: adminToken } = useAdmin();
  const router = useRouter();
  const isEdit = !!id;

  const [form, setForm] = useState<NoticeInput>({
    title: '',
    slug: '',
    content_html: '',
    featured_image_url: '',
    is_published: true,
  });
  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!adminToken) {
      router.push('/admin/login');
      return;
    }
    if (isEdit) {
      adminFetchNotices(adminToken)
        .then(notices => {
          const notice = notices.find(n => n.id === id);
          if (notice) {
            setForm({
              title: notice.title,
              slug: notice.slug,
              content_html: notice.content_html,
              featured_image_url: notice.featured_image_url || '',
              is_published: notice.is_published,
            });
          }
        })
        .finally(() => setLoading(false));
    }
  }, [id, adminToken, isEdit, router]);

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setForm(f => ({
      ...f,
      title: val,
      slug: isEdit ? f.slug : val.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''),
    }));
  };

  const set = <K extends keyof NoticeInput>(key: K, val: NoticeInput[K]) => setForm(f => ({ ...f, [key]: val }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!adminToken) return;
    setSaving(true);
    setError('');

    try {
      if (isEdit) {
        await adminUpdateNotice(adminToken, id as string, form);
      } else {
        await adminCreateNotice(adminToken, form);
      }
      router.push('/admin/notices');
    } catch (err) {
      const error = err as Error;
      setError(error.message || 'Failed to save notice');
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="p-12 text-center animate-pulse">Loading Editor...</div>;

  return (
    <div className="p-6 max-w-5xl mx-auto pb-24">
      <AdminHeader 
        title={isEdit ? 'EDIT NOTICE' : 'NEW NOTICE'}
        subtitle="Compose your shop update using raw HTML."
        actions={
          <button onClick={() => router.push('/admin/notices')} className="btn-secondary text-xs">DISCARD</button>
        }
      />

      <form onSubmit={handleSubmit} className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-surface p-6 rounded-sm border-2 border-kraft-shadow shadow-sm">
            <div className="space-y-4">
              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">Post Title</label>
                <input 
                  type="text" 
                  value={form.title} 
                  onChange={handleTitleChange} 
                  required 
                  className="w-full bg-kraft-light border border-kraft-shadow p-3 font-display text-xl uppercase focus:border-gold-dark outline-none" 
                  placeholder="E.G. NEW SINGLES JUST ARRIVED!"
                />
              </div>

              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">HTML Content</label>
                <textarea 
                  value={form.content_html} 
                  onChange={e => set('content_html', e.target.value)} 
                  required 
                  rows={20}
                  className="w-full bg-kraft-light border border-kraft-shadow p-4 font-mono-stack text-sm focus:border-gold-dark outline-none" 
                  placeholder="<h2>Subheading</h2><p>Write your content here...</p>"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-kraft-mid/20 p-6 rounded-sm border-2 border-kraft-shadow">
            <h4 className="font-display text-sm uppercase mb-4 border-b border-kraft-shadow pb-2">Publish Settings</h4>
            
            <div className="space-y-4">
              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">URL Slug</label>
                <input 
                  type="text" 
                  value={form.slug} 
                  onChange={e => set('slug', e.target.value)} 
                  className="w-full bg-surface border border-kraft-shadow px-3 py-2 text-xs font-mono-stack focus:border-gold-dark outline-none" 
                />
              </div>

              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">Featured Image URL</label>
                <input 
                  type="text" 
                  value={form.featured_image_url} 
                  onChange={e => set('featured_image_url', e.target.value)} 
                  className="w-full bg-surface border border-kraft-shadow px-3 py-2 text-xs font-mono-stack focus:border-gold-dark outline-none" 
                  placeholder="https://..."
                />
              </div>

              <div className="flex items-center gap-2 pt-2">
                <input 
                  type="checkbox" 
                  id="published"
                  checked={form.is_published}
                  onChange={e => set('is_published', e.target.checked)}
                  className="accent-gold-dark"
                />
                <label htmlFor="published" className="text-xs font-bold font-mono-stack uppercase">Published / Public</label>
              </div>

              {error && <div className="text-xs text-red-600 font-bold bg-red-50 p-3 border border-red-200">{error}</div>}

              <button 
                type="submit" 
                disabled={saving} 
                className="btn-primary w-full py-3 mt-4 text-sm font-display uppercase tracking-widest"
              >
                {saving ? 'SAVING...' : (isEdit ? 'UPDATE POST' : 'PUBLISH POST')}
              </button>
            </div>
          </div>
          
          <div className="p-4 bg-surface border-2 border-dashed border-kraft-dark">
            <h5 className="font-display text-xs uppercase mb-2">HTML Help</h5>
            <p className="text-[10px] font-mono-stack text-text-secondary leading-relaxed">
              You can use standard HTML tags like {`<b>, <i>, <h2>, <ul>, <li>`}. 
              To embed a card preview, use: <br/>
              <code className="text-gold-dark break-all">{`<a data-card-id="UUID">Card Name</a>`}</code>
            </p>
          </div>
        </div>
      </form>
    </div>
  );
}
