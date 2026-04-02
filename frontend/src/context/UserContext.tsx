"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { userFetchMe, userLogout as apiUserLogout, userUpdateMe } from '@/lib/api';
import { UserProfile } from '@/lib/types';

interface UserContextType {
  user: UserProfile | null;
  loading: boolean;
  loginWithGoogle: () => void;
  loginWithFacebook: () => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
  updateProfile: (data: Partial<UserProfile>) => Promise<void>;
}

const UserContext = createContext<UserContextType>({
  user: null,
  loading: true,
  loginWithGoogle: () => {},
  loginWithFacebook: () => {},
  logout: () => {},
  refreshUser: async () => {},
  updateProfile: async () => {},
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
      // If the user fetch fails (e.g. 401/404 during dev reset), 
      // clear the state and cookie to prevent being "stuck" in a half-logged state.
      setUser(null);
      // We don't call the full logout() here because it might loop if apiUserLogout also fails,
      // but setUser(null) is already done. Let's just make sure we are not loading.
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const loginWithGoogle = () => {
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL || ''}/api/auth/google/login`;
  };

  const loginWithFacebook = () => {
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL || ''}/api/auth/facebook/login`;
  };

  const logout = async () => {
    try {
      await apiUserLogout();
    } catch {
      // ignore
    }
    setUser(null);
  };

  const updateProfile = async (data: Partial<UserProfile>) => {
    const updated = await userUpdateMe(data);
    setUser(updated);
  };

  return (
    <UserContext.Provider value={{ 
      user, 
      loading, 
      loginWithGoogle, 
      loginWithFacebook,
      logout, 
      refreshUser: fetchUser,
      updateProfile 
    }}>
      {children}
    </UserContext.Provider>
  );
};
