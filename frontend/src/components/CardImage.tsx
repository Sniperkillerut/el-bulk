'use client';

import { useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useUI } from '@/context/UIContext';
import ImageModal from './ImageModal';

interface CardImageProps {
  imageUrl?: string | null;
  name: string;
  tcg?: string;
  foilTreatment?: string;
  height?: number | string;
  enableHover?: boolean;
  enableModal?: boolean;
}

const TCG_EMOJI: Record<string, string> = {
  mtg: '⚔️',
  pokemon: '⚡',
  lorcana: '🌟',
  onepiece: '☠️',
  yugioh: '👁️',
  accessories: '🛡️',
};

export default function CardImage({ 
  imageUrl, name, tcg, foilTreatment, height, 
  enableHover = false, enableModal = false 
}: CardImageProps) {
  const { foilEffectsEnabled } = useUI();
  const [prevUrl, setPrevUrl] = useState(imageUrl);
  const [imgError, setImgError] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const [rect, setRect] = useState<DOMRect | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Sync: Reset error state when the URL changes (Derived state pattern)
  // This pattern is recommended by React 18+ for adjusting state based on prop changes
  if (imageUrl !== prevUrl) {
    setPrevUrl(imageUrl);
    setImgError(false);
  }

  const showImage = imageUrl && !imgError;

  const handleClick = (e: React.MouseEvent) => {
    if (showModal) return; // Prevent re-opening if already open (bubbles from portal)
    if (showImage && enableModal) {
      e.stopPropagation();
      setShowModal(true);
      setIsHovered(false);
    }
  };

  const handleMouseEnter = () => {
    if (enableHover && showImage && containerRef.current) {
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
        background: showImage ? 'transparent' : 'var(--ink-card)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: showImage && enableModal ? 'pointer' : 'default',
      }}
    >
      {showImage ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={imageUrl}
          alt={name}
          onClick={handleClick}
          onError={() => setImgError(true)}
          suppressHydrationWarning
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'contain',
            objectPosition: 'center',
          }}
          className="card-image-static"
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
            background: 'linear-gradient(135deg, var(--ink-card) 0%, var(--ink-surface) 100%)',
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

      {isHovered && imageUrl && rect && (
        <HoverPortal imageUrl={imageUrl} name={name} startRect={rect} />
      )}

      {showModal && imageUrl && typeof document !== 'undefined' && createPortal(
        <ImageModal 
          imageUrl={imageUrl} 
          name={name} 
          foilTreatment={foilTreatment}
          onClose={() => setShowModal(false)} 
        />,
        document.body
      )}
    </div>
  );
}

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

function HoverPortal({ imageUrl, name, startRect }: { imageUrl: string; name: string; startRect: DOMRect }) {
  const [isMounted, setIsMounted] = useState(false);
  
  useEffect(() => {
    const timer = setTimeout(() => setIsMounted(true), 10);
    return () => clearTimeout(timer);
  }, []);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <div className="hover-expand-portal" style={isMounted ? {
      top: (window.innerHeight - Math.min(window.innerWidth * 0.5, window.innerHeight * 0.8)) / 2,
      left: (window.innerWidth - Math.min(window.innerWidth * 0.5, window.innerHeight * 0.8)) / 2,
      width: Math.min(window.innerWidth * 0.5, window.innerHeight * 0.8),
      height: Math.min(window.innerWidth * 0.5, window.innerHeight * 0.8),
      opacity: 1,
    } : {
      top: startRect.top,
      left: startRect.left,
      width: startRect.width,
      height: startRect.height,
      opacity: 0,
    }}>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={imageUrl} alt={name} className="hover-expand-image" />
    </div>,
    document.body
  );
}
