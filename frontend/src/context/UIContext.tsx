'use client';

import React, { createContext, useContext, useState } from 'react';

interface UIContextType {
  foilEffectsEnabled: boolean;
  setFoilEffectsEnabled: (enabled: boolean) => void;
  toggleFoilEffects: () => void;
}

const UIContext = createContext<UIContextType | undefined>(undefined);

export function UIProvider({ children }: { children: React.ReactNode }) {
  const [foilEffectsEnabled, setFoilEffectsEnabled] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('foilEffectsEnabled');
      return saved !== null ? saved === 'true' : true;
    }
    return true;
  });

  // Save to localStorage when changed
  const handleSetFoilEffectsEnabled = (enabled: boolean) => {
    setFoilEffectsEnabled(enabled);
    localStorage.setItem('foilEffectsEnabled', enabled.toString());
  };

  const toggleFoilEffects = () => {
    handleSetFoilEffectsEnabled(!foilEffectsEnabled);
  };

  return (
    <UIContext.Provider value={{
      foilEffectsEnabled,
      setFoilEffectsEnabled: handleSetFoilEffectsEnabled,
      toggleFoilEffects
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
