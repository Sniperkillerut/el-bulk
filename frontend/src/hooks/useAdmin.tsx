'use client';

import React, { useState, useEffect, createContext, useContext, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { getAdminSettings, fetchPublicSettings } from '@/lib/api';
import { Settings } from '@/lib/types';

interface AdminContextType {
  token: string | null;
  settings: Settings | undefined;
  loading: boolean;
  logout: () => void;
  refreshSettings: () => Promise<void>;
}

const AdminContext = createContext<AdminContextType | undefined>(undefined);

export function AdminProvider({ children }: { children: React.ReactNode }): React.ReactNode {
  const [token, setToken] = useState<string | null>(null);
  const [settings, setSettings] = useState<Settings | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    // With middleware protecting the route, we just need to load initial context
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      // Try to get admin settings. If this fails on a protected route, 
      // the middleware would have already caught it, but we handle it here for the login page too.
      const data = await getAdminSettings({ forceRefresh: true }).catch(() => fetchPublicSettings({ forceRefresh: true }));
      setSettings(data as Settings);
      setToken("session_active"); // Marker for existing logic that expects a truthy token
    } catch (err) {
      console.error('Failed to load session settings', err);
    } finally {
      setLoading(false);
    }
  };

  const logout = useCallback(async () => {
    try {
      const { adminLogout } = await import('@/lib/api');
      await adminLogout();
    } catch (err) {
      console.error('Logout failed', err);
    }
    setToken(null);
    router.push('/admin/login');
  }, [router]);

  const refreshSettings = useCallback(async () => {
    await loadSettings();
  }, []);

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
