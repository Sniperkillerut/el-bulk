'use client';

import React, { createContext, useContext } from 'react';

interface UIContextType {
  foilEffectsEnabled: boolean;
  setFoilEffectsEnabled: (enabled: boolean) => void;
  toggleFoilEffects: () => void;
}

const UIContext = createContext<UIContextType | undefined>(undefined);

export function UIProvider({ children }: { children: React.ReactNode }) {
  const [foilEffectsEnabled, setFoilEffectsEnabled] = React.useState<boolean>(true);
  const [initialized, setInitialized] = React.useState(false);

  // Load from localStorage on mount
  React.useEffect(() => {
    const saved = localStorage.getItem('foilEffectsEnabled');
    if (saved !== null) {
      setFoilEffectsEnabled(saved === 'true');
    }
    setInitialized(true);
  }, []);

  // Save to localStorage when changed
  const handleSetFoilEffectsEnabled = (enabled: boolean) => {
    if (!initialized) return;
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
