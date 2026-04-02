'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { adminFetchNotices, adminCreateNotice, adminUpdateNotice } from '@/lib/api';
import { NoticeInput } from '@/lib/types';
import { useAdmin } from '@/hooks/useAdmin';
import AdminHeader from '@/components/admin/AdminHeader';
import { useLanguage } from '@/context/LanguageContext';

export default function NoticeEditor() {
  const { t } = useLanguage();
  const { id } = useParams();
  const { token: adminToken } = useAdmin();
  const router = useRouter();
  const isEdit = !!id;

  const slugify = (text: string) => text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');

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
  const [slugTouched, setSlugTouched] = useState(false);

  useEffect(() => {
    if (isEdit && adminToken) {
      adminFetchNotices()
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
            setSlugTouched(true);
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
      slug: slugTouched ? f.slug : slugify(val),
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
        await adminUpdateNotice(id as string, form);
      } else {
        await adminCreateNotice(form);
      }
      router.push('/admin/notices');
    } catch (err) {
      const error = err as Error;
      setError(error.message || t('pages.admin.notices.editor.error_save', 'Failed to save notice'));
    } finally {
      setSaving(false);
    }
  };

  if (loading) return (
    <div className="min-h-screen flex items-center justify-center p-12">
      <div className="text-gold font-mono-stack animate-pulse uppercase text-sm font-bold tracking-widest">
        {t('pages.admin.notices.editor.loading', 'Unpacking Notice...')}
      </div>
    </div>
  );

  return (
    <div className="p-3 max-w-7xl mx-auto pb-12">
      <AdminHeader
        title={isEdit ? t('pages.admin.notices.editor.title_edit', 'EDIT NOTICE') : t('pages.admin.notices.editor.title_new', 'NEW NOTICE')}
        subtitle={t('pages.admin.notices.editor.subtitle', 'Compose your shop update using raw HTML.')}
        actions={
          <button onClick={() => router.push('/admin/notices')} className="btn-secondary text-xs">{t('pages.common.actions.discard', 'DISCARD')}</button>
        }
      />

      <form onSubmit={handleSubmit} className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-surface p-3 rounded-sm border-2 border-kraft-shadow shadow-sm">
            <div className="space-y-4">
              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">{t('pages.admin.notices.editor.form.title_label', 'Post Title')}</label>
                <input
                  type="text"
                  value={form.title}
                  onChange={handleTitleChange}
                  required
                  className="w-full bg-kraft-light border border-kraft-shadow p-3 font-display text-xl uppercase focus:border-gold-dark outline-none"
                  placeholder={t('pages.admin.notices.editor.form.title_placeholder', 'E.G. NEW SINGLES JUST ARRIVED!')}
                />
              </div>

              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">{t('pages.admin.notices.editor.form.content_label', 'HTML Content')}</label>
                <textarea
                  value={form.content_html}
                  onChange={e => set('content_html', e.target.value)}
                  required
                  rows={20}
                  className="w-full bg-kraft-light border border-kraft-shadow p-4 font-mono-stack text-sm focus:border-gold-dark outline-none"
                  placeholder={t('pages.admin.notices.editor.form.content_placeholder', '<h2>Subheading</h2><p>Write your content here...</p>')}
                />
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-kraft-mid/20 p-3 rounded-sm border-2 border-kraft-shadow">
            <h4 className="font-display text-sm uppercase mb-4 border-b border-kraft-shadow pb-2">{t('pages.admin.notices.editor.sidebar.settings_title', 'Publish Settings')}</h4>

            <div className="space-y-4">
              <div>
                <div className="flex justify-between items-center mb-1">
                  <label className="block text-[10px] font-bold font-mono-stack uppercase">{t('pages.admin.notices.editor.sidebar.slug_label', 'URL Slug')}</label>
                  {form.slug !== slugify(form.title) && (
                    <button 
                      type="button"
                      onClick={() => {
                        setSlugTouched(false);
                        set('slug', slugify(form.title));
                      }}
                      className="text-[9px] font-mono-stack text-gold-dark hover:underline font-bold"
                    >
                      {t('pages.admin.notices.editor.sidebar.sync_btn', 'SYNC WITH TITLE')}
                    </button>
                  )}
                </div>
                <input
                  type="text"
                  value={form.slug}
                  onChange={e => {
                    const val = e.target.value.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
                    setSlugTouched(true);
                    set('slug', val);
                  }}
                  className="w-full bg-surface border border-kraft-shadow px-3 py-2 text-xs font-mono-stack focus:border-gold-dark outline-none"
                  placeholder={t('pages.admin.notices.editor.sidebar.slug_placeholder', 'post-url-slug')}
                />
              </div>

              <div>
                <label className="block text-[10px] font-bold font-mono-stack mb-1 uppercase">{t('pages.admin.notices.editor.sidebar.image_label', 'Featured Image URL')}</label>
                <input
                  type="text"
                  value={form.featured_image_url}
                  onChange={e => set('featured_image_url', e.target.value)}
                  className="w-full bg-surface border border-kraft-shadow px-3 py-2 text-xs font-mono-stack focus:border-gold-dark outline-none"
                  placeholder={t('pages.admin.notices.editor.sidebar.image_placeholder', 'https://...')}
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
                <label htmlFor="published" className="text-xs font-bold font-mono-stack uppercase">{t('pages.admin.notices.editor.sidebar.published_label', 'Published / Public')}</label>
              </div>

              {error && <div className="text-xs text-red-600 font-bold bg-red-50 p-3 border border-red-200">{error}</div>}

              <button
                type="submit"
                disabled={saving}
                className="btn-primary w-full py-3 mt-4 text-sm font-display uppercase tracking-widest"
              >
                {saving ? t('pages.common.status.saving', 'SAVING...') : (isEdit ? t('pages.admin.notices.editor.actions.update', 'UPDATE POST') : t('pages.admin.notices.editor.actions.publish', 'PUBLISH POST'))}
              </button>
            </div>
          </div>

          <div className="p-4 bg-surface border-2 border-dashed border-kraft-dark">
            <h5 className="font-display text-xs uppercase mb-2">{t('pages.admin.notices.editor.sidebar.help_title', 'HTML Help')}</h5>
            <p className="text-[10px] font-mono-stack text-text-secondary leading-relaxed">
              {t('pages.admin.notices.editor.sidebar.help_desc', 'You can use standard HTML tags like {tags}.', { tags: '<b>, <i>, <h2>, <ul>, <li>' })}
              <br /> {t('pages.admin.notices.editor.sidebar.card_preview_label', 'To embed a card preview, use:')} <br />
              <code className="text-gold-dark break-all">{`<a data-card-id="UUID">Card Name</a>`}</code>
            </p>
          </div>
        </div>
      </form>
    </div>
  );
}
