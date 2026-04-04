'use client';

import React, { createContext, useContext } from 'react';

interface UIContextType {
  foilEffectsEnabled: boolean;
  setFoilEffectsEnabled: (enabled: boolean) => void;
  toggleFoilEffects: () => void;
  adminSidebarOpen: boolean;
  setAdminSidebarOpen: (open: boolean) => void;
  toggleAdminSidebar: () => void;
  adminSidebarSlim: boolean;
  setAdminSidebarSlim: (slim: boolean) => void;
  toggleAdminSidebarSlim: () => void;
}

const UIContext = createContext<UIContextType | undefined>(undefined);

export function UIProvider({ children }: { children: React.ReactNode }) {
  const [foilEffectsEnabled, setFoilEffectsEnabled] = React.useState<boolean>(true);
  const [adminSidebarOpen, setAdminSidebarOpen] = React.useState<boolean>(false);
  const [adminSidebarSlim, setAdminSidebarSlim] = React.useState<boolean>(false);
  const [initialized, setInitialized] = React.useState(false);

  // Load from localStorage on mount
  React.useEffect(() => {
    const savedFoil = localStorage.getItem('foilEffectsEnabled');
    if (savedFoil !== null) {
      setFoilEffectsEnabled(savedFoil === 'true');
    }
    const savedSlim = localStorage.getItem('adminSidebarSlim');
    if (savedSlim !== null) {
      setAdminSidebarSlim(savedSlim === 'true');
    }
    setInitialized(true);
  }, []);

  // Save to localStorage when changed
  const handleSetFoilEffectsEnabled = (enabled: boolean) => {
    if (!initialized) return;
    setFoilEffectsEnabled(enabled);
    localStorage.setItem('foilEffectsEnabled', enabled.toString());
  };

  const handleSetAdminSidebarSlim = (slim: boolean) => {
    if (!initialized) return;
    setAdminSidebarSlim(slim);
    localStorage.setItem('adminSidebarSlim', slim.toString());
  };

  const toggleFoilEffects = () => {
    handleSetFoilEffectsEnabled(!foilEffectsEnabled);
  };

  const toggleAdminSidebar = () => {
    setAdminSidebarOpen(!adminSidebarOpen);
  };

  const toggleAdminSidebarSlim = () => {
    handleSetAdminSidebarSlim(!adminSidebarSlim);
  };

  return (
    <UIContext.Provider value={{
      foilEffectsEnabled,
      setFoilEffectsEnabled: handleSetFoilEffectsEnabled,
      toggleFoilEffects,
      adminSidebarOpen,
      setAdminSidebarOpen,
      toggleAdminSidebar,
      adminSidebarSlim,
      setAdminSidebarSlim: handleSetAdminSidebarSlim,
      toggleAdminSidebarSlim
    }}>
      {children}
    </UIContext.Provider>
  );
}

export function useUI() {
  const context = useContext(UIContext);
  if (context === undefined) {
    throw new Error('useUI must be used within a UIProvider');
  }
  return context;
}
