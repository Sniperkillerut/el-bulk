'use client';

import { useEffect, useRef, useState, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { fetchProduct, getProxyImageUrl } from '@/lib/api';
import { Product } from '@/lib/types';
import { basicSanitize } from '@/lib/htmlUtils';

interface NoticeContentProps {
  html: string;
}

// Simple in-memory cache for hovered products to avoid redundant API calls
const productCache = new Map<string, Product>();

export default function NoticeContent({ html }: NoticeContentProps) {
  const sanitizedHtml = useMemo(() => basicSanitize(html), [html]);
  const contentRef = useRef<HTMLDivElement>(null);
  const [hoveredCard, setHoveredCard] = useState<{ product: Product; rect: DOMRect } | null>(null);
  const [isMounted, setIsMounted] = useState(false);
  const activeFetchId = useRef<string | null>(null);

  useEffect(() => {
    // Avoid synchronous setState in effect to prevent cascading renders
    Promise.resolve().then(() => setIsMounted(true));
    
    const content = contentRef.current;
    if (!content) return;

    // Use Event Delegation for better performance and reliability with dynamic HTML
    const handleMouseOver = async (e: MouseEvent) => {
      const target = (e.target as HTMLElement).closest('[data-card-id]');
      if (!target) return;

      const cardId = target.getAttribute('data-card-id');
      if (!cardId) return;

      // If already fetching this card, ignore
      if (activeFetchId.current === cardId) return;
      activeFetchId.current = cardId;

      try {
        let product = productCache.get(cardId);
        if (!product) {
          product = await fetchProduct(cardId);
          productCache.set(cardId, product);
        }
        
        // Final check: if we are still hovering over the same ID (avoid race conditions)
        if (activeFetchId.current === cardId) {
          setHoveredCard({ product, rect: target.getBoundingClientRect() });
        }
      } catch {
        // Silently fail if card not found
      }
    };

    const handleMouseOut = (e: MouseEvent) => {
      const target = (e.target as HTMLElement).closest('[data-card-id]');
      if (!target) return;
      
      activeFetchId.current = null;
      setHoveredCard(null);
    };

    content.addEventListener('mouseover', handleMouseOver);
    content.addEventListener('mouseout', handleMouseOut);

    return () => {
      content.removeEventListener('mouseover', handleMouseOver);
      content.removeEventListener('mouseout', handleMouseOut);
    };
  }, [html]);

  return (
    <div className="notice-content-wrapper relative">
      <div 
        ref={contentRef}
        className="notice-html-content prose prose-p:font-mono-stack prose-p:text-text-secondary prose-headings:font-display prose-headings:uppercase prose-headings:color-ink-deep max-w-none"
        dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
      />

      {isMounted && hoveredCard && (
        <CardHoverPortal product={hoveredCard.product} startRect={hoveredCard.rect} />
      )}
    </div>
  );
}

function CardHoverPortal({ product, startRect }: { product: Product; startRect: DOMRect }) {
  const [active, setActive] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setActive(true), 10);
    return () => clearTimeout(timer);
  }, []);

  if (typeof document === 'undefined') return null;

  // Dest size based on viewport
  const destSize = Math.min(window.innerWidth * 0.4, 280);
  const destTop = Math.max(20, (window.innerHeight - (destSize * 1.4)) / 2);
  const destLeft = (window.innerWidth - destSize) / 2;

  const style: React.CSSProperties = {
    position: 'fixed',
    zIndex: 10006, // Higher than CardImage hover and other UI elements
    pointerEvents: 'none',
    transition: 'all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)',
    top: active ? destTop : startRect.top,
    left: active ? destLeft : startRect.left,
    width: active ? destSize : startRect.width,
    opacity: active ? 1 : 0,
    transform: active ? 'scale(1)' : 'scale(0.5)',
    boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
    borderRadius: '8px',
    overflow: 'hidden',
    border: '1px solid var(--gold)',
    background: 'var(--ink-navy)',
  };

  return createPortal(
    <div style={style}>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={getProxyImageUrl(product.image_url)} alt={product.name} className="w-full h-auto object-contain" />
      <div className="p-3 border-t border-gold/30 bg-ink-navy text-white">
        <div className="text-xs font-bold font-display uppercase tracking-wider">{product.name}</div>
        <div className="text-[10px] opacity-60 uppercase font-mono-stack">{product.set_name}</div>
      </div>
    </div>,
    document.body
  );
}
