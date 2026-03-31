'use client';

import React, { useState, useEffect, createContext, useContext, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { getAdminSettings, fetchPublicSettings } from '@/lib/api';
import { Settings } from '@/lib/types';

interface AdminContextType {
  token: string | null;
  settings: Settings | null;
  loading: boolean;
  logout: () => void;
  refreshSettings: () => Promise<void>;
}

const AdminContext = createContext<AdminContextType | undefined>(undefined);

export function AdminProvider({ children }: { children: React.ReactNode }): React.ReactNode {
  const [token, setToken] = useState<string | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const t = localStorage.getItem('el_bulk_admin_token');
    if (!t) {
      setLoading(false);
      return;
    }
    setToken(t);
    loadSettings(t);
  }, []);

  const loadSettings = async (t: string) => {
    try {
      // Try to get admin settings, fallback to public if needed
      const data = await getAdminSettings(t).catch(() => fetchPublicSettings());
      setSettings(data);
    } catch (err) {
      console.error('Failed to load settings', err);
    } finally {
      setLoading(false);
    }
  };

  const logout = useCallback(() => {
    localStorage.removeItem('el_bulk_admin_token');
    setToken(null);
    router.push('/admin/login');
  }, [router]);

  const refreshSettings = useCallback(async () => {
    if (token) await loadSettings(token);
  }, [token]);

  return (
    <AdminContext.Provider value={{ token, settings, loading, logout, refreshSettings }}>
      {children}
    </AdminContext.Provider>
  );
}

export function useAdmin() {
  const context = useContext(AdminContext);
  if (context === undefined) {
    throw new Error('useAdmin must be used within an AdminProvider');
  }
  return context;
}
