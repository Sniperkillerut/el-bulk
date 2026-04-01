"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { userFetchMe, userLogout as apiUserLogout } from '@/lib/api';

// Update types to match our backend Customer struct
export interface UserProfile {
  id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  id_number?: string;
  address?: string;
  auth_provider?: string;
  auth_provider_id?: string;
  avatar_url?: string;
}

interface UserContextType {
  user: UserProfile | null;
  loading: boolean;
  loginWithGoogle: () => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const UserContext = createContext<UserContextType>({
  user: null,
  loading: true,
  loginWithGoogle: () => {},
  logout: () => {},
  refreshUser: async () => {},
});

export const useUser = () => useContext(UserContext);

export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUser = async () => {
    try {
      const data = await userFetchMe();
      setUser(data);
    } catch {
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const loginWithGoogle = () => {
    // Redirect to backend endpoint that initiates OAuth
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL || ''}/api/auth/google/login`;
  };

  const logout = async () => {
    try {
      await apiUserLogout();
    } catch {
      // ignore
    }
    setUser(null);
  };

  return (
    <UserContext.Provider value={{ user, loading, loginWithGoogle, logout, refreshUser: fetchUser }}>
      {children}
    </UserContext.Provider>
  );
};
