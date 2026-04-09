'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Notice } from '@/lib/types';
import { adminFetchNotices, adminDeleteNotice } from '@/lib/api';
import AdminHeader from '@/components/admin/AdminHeader';

import { useLanguage } from '@/context/LanguageContext';

export default function AdminNoticesPage() {
  const { t, locale } = useLanguage();
  const [notices, setNotices] = useState<Notice[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminFetchNotices()
      .then(setNotices)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleDelete = async (id: string) => {
    if (!confirm(t('pages.admin.notices.confirm_delete', 'Are you sure you want to delete this notice?'))) return;
    try {
      await adminDeleteNotice(id);
      setNotices(notices.filter(n => n.id !== id));
    } catch {
      alert(t('pages.admin.notices.error_delete', 'Failed to delete notice'));
    }
  };

  return (
    <div className="flex-1 flex flex-col p-3 min-h-0 max-w-7xl mx-auto w-full">
      <AdminHeader 
        title={t('pages.admin.notices.title', 'NOTICES (BLOG/NEWS)')}
        subtitle={t('pages.admin.notices.subtitle', 'Manage your shop updates and news posts.')}
        actions={
          <Link href="/admin/notices/new" className="btn-primary">
            + {t('pages.admin.notices.create_btn', 'CREATE NEW NOTICE')}
          </Link>
        }
      />

      <div className="flex-1 min-h-0 overflow-auto bg-surface rounded-sm border-2 border-kraft-shadow shadow-sm scrollbar-thin">
        <table className="w-full text-left border-collapse">
          <thead className="sticky top-0 z-10 bg-kraft-light backdrop-blur-md shadow-sm border-b border-kraft-shadow">
            <tr className="text-[10px] font-mono-stack font-bold uppercase tracking-wider">
              <th className="p-4 border-b border-kraft-shadow">{t('pages.admin.notices.table.date', 'Date')}</th>
              <th className="p-4 border-b border-kraft-shadow">{t('pages.admin.notices.table.title', 'Title')}</th>
              <th className="p-4 border-b border-kraft-shadow hidden md:table-cell">{t('pages.admin.notices.table.slug', 'Slug')}</th>
              <th className="p-4 border-b border-kraft-shadow hidden md:table-cell">{t('pages.admin.notices.table.status', 'Status')}</th>
              <th className="p-4 border-b border-kraft-shadow text-right">{t('pages.admin.notices.table.actions', 'Actions')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-kraft-shadow font-mono-stack text-sm">
            {loading ? (
              [1, 2, 3].map(i => (
                <tr key={i} className="animate-pulse">
                  <td colSpan={5} className="p-4 h-12 bg-gray-50" />
                </tr>
              ))
            ) : notices.length === 0 ? (
              <tr>
                <td colSpan={5} className="p-12 text-center text-text-muted">{t('pages.admin.notices.table.empty', 'No notices found. Create your first one!')}</td>
              </tr>
            ) : (
              notices.map(notice => (
                <tr key={notice.id} className="hover:bg-kraft-light/30 transition-colors">
                  <td className="p-4 whitespace-nowrap">
                    {new Date(notice.created_at).toLocaleDateString(locale === 'es' ? 'es-ES' : 'en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                  </td>
                  <td className="p-4 font-bold">{notice.title}</td>
                  <td className="p-4 text-xs opacity-60 hidden md:table-cell">/{notice.slug}</td>
                  <td className="p-4 hidden md:table-cell">
                    <span className={`px-2 py-0.5 rounded-sm text-[10px] font-bold ${notice.is_published ? 'bg-green-100 text-green-700 border border-green-200' : 'bg-gray-100 text-gray-600 border border-gray-200'}`}>
                      {notice.is_published ? t('pages.admin.notices.status.published', 'PUBLISHED') : t('pages.admin.notices.status.draft', 'DRAFT')}
                    </span>
                  </td>
                  <td className="p-4 text-right space-x-2">
                    <Link href={`/notices/${notice.slug}`} target="_blank" className="text-xs text-blue-600 hover:underline">
                      {t('pages.admin.notices.actions.view', 'VIEW')}
                    </Link>
                    <Link href={`/admin/notices/${notice.id}/edit`} className="text-xs text-gold-dark hover:underline font-bold">
                      {t('pages.admin.notices.actions.edit', 'EDIT')}
                    </Link>
                    <button onClick={() => handleDelete(notice.id)} className="text-xs text-red-600 hover:underline">
                      {t('pages.admin.notices.actions.delete', 'DELETE')}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
