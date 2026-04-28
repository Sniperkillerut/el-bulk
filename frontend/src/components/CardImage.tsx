/* eslint-disable @next/next/no-img-element */
'use client';

import { useState, useRef, useEffect, memo } from 'react';
import { createPortal } from 'react-dom';
import { useUI } from '@/context/UIContext';
import { getProxyImageUrl } from '@/lib/api';
import ImageModal from './ImageModal';

interface CardImageProps {
  imageUrl?: string | null;
  name: string;
  tcg?: string;
  foilTreatment?: string;
  height?: number | string;
  enableHover?: boolean;
  enableModal?: boolean;
  scryfallId?: string | null;
}

const TCG_EMOJI: Record<string, string> = {
  mtg: '⚔️',
  pokemon: '⚡',
  lorcana: '🌟',
  onepiece: '☠️',
  yugioh: '👁️',
  accessories: '🛡️',
};

const CardImage = memo(function CardImage({ 
  imageUrl, name, tcg, foilTreatment, height, 
  enableHover = false, enableModal = false,
  scryfallId
}: CardImageProps) {
  const { foilEffectsEnabled } = useUI();
  const [prevUrl, setPrevUrl] = useState(imageUrl);
  const [imgError, setImgError] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const [rect, setRect] = useState<DOMRect | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Fallback to Scryfall if direct URL is missing but ID is present
  const rawUrl = imageUrl || (scryfallId ? `https://api.scryfall.com/cards/${scryfallId}?format=image&version=normal` : null);
  const resolvedUrl = getProxyImageUrl(rawUrl);

  // Sync: Reset error state when the URL changes (Derived state pattern)
  // This pattern is recommended by React 18+ for adjusting state based on prop changes
  if (resolvedUrl !== prevUrl) {
    setPrevUrl(resolvedUrl);
    setImgError(false);
  }

  const showImage = resolvedUrl && !imgError;

  const handleClick = (e: React.MouseEvent) => {
    if (showModal) return; // Prevent re-opening if already open (bubbles from portal)
    if (showImage && enableModal) {
      e.stopPropagation();
      setIsHovered(false); // Clear any lingering hover state
      setShowModal(true);
    }
  };

  const handleMouseEnter = () => {
    // Only enable hover expansion on mouse-based devices
    const isTouch = typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches;
    if (enableHover && showImage && containerRef.current && !isTouch) {
      setRect(containerRef.current.getBoundingClientRect());
      setIsHovered(true);
    }
  };

  return (
    <div
      ref={containerRef}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={() => setIsHovered(false)}
      style={{
        position: 'relative',
        height: height || 'auto',
        aspectRatio: height ? undefined : '63/88',
        width: '100%',
        overflow: 'hidden',
        background: showImage ? 'transparent' : 'var(--bg-card)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: showImage && enableModal ? 'pointer' : 'default',
      }}
    >
      {showImage ? (
        <img
          src={resolvedUrl as string}
          alt={name}
          loading="lazy"
          onClick={handleClick}
          onError={() => setImgError(true)}
          className="card-image-static active:scale-95"
          suppressHydrationWarning
        />
      ) : (
        /* Placeholder — shown when no image or image fails to load */
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '0.5rem',
            width: '100%',
            height: '100%',
            padding: '1rem',
            background: 'linear-gradient(135deg, var(--bg-card) 0%, var(--bg-surface) 100%)',
          }}
        >
          <span style={{ fontSize: '2rem', opacity: 0.6 }}>
            {tcg ? (TCG_EMOJI[tcg] ?? '🃏') : '🃏'}
          </span>
          <span
            className="font-display"
            style={{
              color: 'var(--text-muted)',
              fontSize: '0.6rem',
              textAlign: 'center',
              letterSpacing: '0.08em',
              lineHeight: 1.3,
              maxWidth: '90%',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
            }}
          >
            {name.toUpperCase()}
          </span>
        </div>
      )}

      {/* Foil Overlay */}
      {foilEffectsEnabled && foilTreatment && foilTreatment !== 'non_foil' && (
        <FoilOverlay treatment={foilTreatment} />
      )}

      {isHovered && resolvedUrl && rect && (
        <HoverPortal imageUrl={resolvedUrl as string} name={name} startRect={rect} foilTreatment={foilTreatment} />
      )}

      {showModal && resolvedUrl && typeof document !== 'undefined' && createPortal(
        <ImageModal 
          imageUrl={resolvedUrl as string} 
          name={name} 
          foilTreatment={foilTreatment}
          onClose={() => setShowModal(false)} 
        />,
        document.body
      )}
    </div>
  );
});

export default CardImage;

export function FoilOverlay({ treatment }: { treatment: string }) {
  // Normalize treatment slug for CSS classes
  const effectClass = treatment.toLowerCase().replace(/_/g, '-');
  
  const getFoilClass = (t: string) => {
    const slug = t.toLowerCase();
    if (slug === 'foil' || slug === 'traditional_foil') return 'foil-classic';
    if (slug.includes('surge')) return 'foil-surge';
    if (slug.includes('etched')) return 'foil-etched';
    if (slug.includes('galaxy')) return 'foil-galaxy';
    if (slug.includes('oil_slick')) return 'foil-oil-slick';
    if (slug.includes('step_and_compleat') || slug.includes('compleat')) return 'foil-step-compleat';
    if (slug.includes('ripple')) return 'foil-ripple';
    if (slug.includes('textured')) return 'foil-etched';
    return `foil-${effectClass.replace('foil-', '')}` || 'foil-classic';
  };

  return <div className={`foil-overlay ${getFoilClass(treatment)}`} />;
}

function HoverPortal({ imageUrl, name, startRect, foilTreatment }: { imageUrl: string; name: string; startRect: DOMRect; foilTreatment?: string }) {
  const { foilEffectsEnabled } = useUI();
  const [isMounted, setIsMounted] = useState(false);
  
  useEffect(() => {
    const timer = setTimeout(() => setIsMounted(true), 10);
    return () => clearTimeout(timer);
  }, []);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <div className="hover-expand-portal" style={{
      zIndex: 10020, // Keep high enough for the 100-z-index modal
      ...(isMounted ? {
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        height: Math.min(window.innerHeight * 0.85, 800),
        aspectRatio: '63/88',
        opacity: 1,
      } : {
        top: startRect.top,
        left: startRect.left,
        width: startRect.width,
        height: startRect.height,
        opacity: 0,
      })
    }}>
      <div style={{ position: 'relative', width: '100%', height: '100%', borderRadius: 'inherit', overflow: 'hidden' }}>
        <img 
          src={imageUrl} 
          alt={name} 
          className="hover-expand-image"
          suppressHydrationWarning
        />
        {foilEffectsEnabled && foilTreatment && foilTreatment !== 'non_foil' && (
          <FoilOverlay treatment={foilTreatment} />
        )}
      </div>
    </div>,
    document.body
  );
}
