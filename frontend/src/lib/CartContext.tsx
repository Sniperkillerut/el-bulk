'use client';

import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { CartItem, Product } from '@/lib/types';

interface CartContextType {
  items: CartItem[];
  removedItems: CartItem[];
  totalItems: number;
  totalPrice: number;
  addItem: (product: Product) => void;
  removeItem: (productId: string) => void;
  permanentRemove: (productId: string) => void;
  restoreItem: (productId: string) => void;
  updateQty: (productId: string, qty: number) => void;
  clearCart: () => void;
  isOpen: boolean;
  openCart: () => void;
  closeCart: () => void;
}

const CartContext = createContext<CartContextType | null>(null);

const CART_COOKIE = 'el-bulk-cart';
const REMOVED_COOKIE = 'el-bulk-removed';

function setCookie(name: string, value: string, days = 30) {
  const expires = new Date(Date.now() + days * 864e5).toUTCString();
  if (typeof document !== 'undefined') {
    document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; path=/`;
  }
}

function getCookie(name: string) {
  if (typeof document === 'undefined') return '';
  return document.cookie.split('; ').reduce((r, v) => {
    const parts = v.split('=');
    return parts[0] === name ? decodeURIComponent(parts[1]) : r;
  }, '');
}

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>([]);
  const [removedItems, setRemovedItems] = useState<CartItem[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [initialized, setInitialized] = useState(false);

  // Load from cookies on mount
  React.useEffect(() => {
    const savedCart = getCookie(CART_COOKIE);
    const savedRemoved = getCookie(REMOVED_COOKIE);
    if (savedCart) {
      try { setItems(JSON.parse(savedCart)); } catch (e) { console.error(e); }
    }
    if (savedRemoved) {
      try { setRemovedItems(JSON.parse(savedRemoved)); } catch (e) { console.error(e); }
    }
    setInitialized(true);
  }, []);

  // Save to cookies on change
  React.useEffect(() => {
    if (!initialized) return;
    setCookie(CART_COOKIE, JSON.stringify(items));
    setCookie(REMOVED_COOKIE, JSON.stringify(removedItems));
  }, [items, removedItems, initialized]);

  const addItem = useCallback((product: Product) => {
    // If it was in removedItems, remove it from there
    setRemovedItems(prev => prev.filter(i => i.product.id !== product.id));
    
    setItems(prev => {
      const existing = prev.find(i => i.product.id === product.id);
      if (existing) {
        return prev.map(i =>
          i.product.id === product.id
            ? { ...i, quantity: Math.min(i.quantity + 1, product.stock) }
            : i
        );
      }
      return [...prev, { product, quantity: 1 }];
    });
    setIsOpen(true);
  }, []);

  const removeItem = useCallback((productId: string) => {
    setItems(prev => {
      const itemToMove = prev.find(i => i.product.id === productId);
      if (itemToMove) {
        setRemovedItems(rem => {
          const alreadyRemoved = rem.find(i => i.product.id === productId);
          if (alreadyRemoved) return rem;
          return [...rem, itemToMove];
        });
      }
      return prev.filter(i => i.product.id !== productId);
    });
  }, []);

  const permanentRemove = useCallback((productId: string) => {
    setRemovedItems(prev => prev.filter(i => i.product.id !== productId));
  }, []);

  const restoreItem = useCallback((productId: string) => {
    setRemovedItems(prev => {
      const itemToRestore = prev.find(i => i.product.id === productId);
      if (itemToRestore) {
        setItems(curr => {
          const existing = curr.find(i => i.product.id === productId);
          if (existing) return curr;
          return [...curr, itemToRestore];
        });
      }
      return prev.filter(i => i.product.id !== productId);
    });
  }, []);

  const updateQty = useCallback((productId: string, qty: number) => {
    if (qty <= 0) {
      removeItem(productId);
    } else {
      setItems(prev => prev.map(i =>
        i.product.id === productId ? { ...i, quantity: Math.min(qty, i.product.stock) } : i
      ));
    }
  }, [removeItem]);

  const clearCart = useCallback(() => {
    setItems([]);
    setRemovedItems([]);
  }, []);

  const totalItems = items.reduce((sum, i) => sum + i.quantity, 0);
  const totalPrice = items.reduce((sum, i) => sum + i.product.price * i.quantity, 0);

  return (
    <CartContext.Provider value={{
      items, removedItems, totalItems, totalPrice,
      addItem, removeItem, permanentRemove, restoreItem, updateQty, clearCart,
      isOpen, openCart: () => setIsOpen(true), closeCart: () => setIsOpen(false),
    }}>
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error('useCart must be used inside CartProvider');
  return ctx;
}
