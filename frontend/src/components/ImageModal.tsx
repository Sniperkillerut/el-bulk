'use client';

import { useEffect } from 'react';
import { FoilOverlay } from './CardImage';
import { useUI } from '@/context/UIContext';

interface ImageModalProps {
  imageUrl: string;
  name: string;
  foilTreatment?: string;
  onClose: () => void;
}

export default function ImageModal({ imageUrl, name, foilTreatment, onClose }: ImageModalProps) {
  const { foilEffectsEnabled } = useUI();
  // Prevent scrolling on body when modal is open
  useEffect(() => {
    document.body.style.overflow = 'hidden';
    
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      document.body.style.overflow = 'unset';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [onClose]);

  const handleClose = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClose();
  };

  return (
    <div 
      className="image-modal-overlay" 
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        backgroundColor: 'rgba(0, 0, 0, 0.95)',
        backdropFilter: 'blur(8px)',
        zIndex: 1000000,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '2rem'
      }}
    >
      {/* Invisible button to capture all backdrop clicks */}
      <button
        onClick={handleClose}
        style={{
          position: 'absolute',
          inset: 0,
          width: '100%',
          height: '100%',
          background: 'none',
          border: 'none',
          padding: 0,
          margin: 0,
          cursor: 'pointer',
          zIndex: 1
        }}
        aria-label="Close modal"
      />

      <div 
        className="image-modal-content"
        onClick={handleClose}
        style={{
          position: 'relative',
          maxWidth: '90vw',
          maxHeight: '90vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 2, // Above the backdrop button
          cursor: 'pointer'
        }}
      >
        <button 
          className="image-modal-close" 
          onClick={handleClose}
          style={{
            position: 'absolute',
            top: '-50px',
            right: '0px',
            background: 'rgba(255,255,255,0.1)',
            backdropFilter: 'blur(4px)',
            borderRadius: '50%',
            width: '44px',
            height: '44px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            border: '1px solid rgba(255,255,255,0.2)',
            color: '#fff',
            fontSize: '2rem',
            cursor: 'pointer',
            zIndex: 3,
            transition: 'all 0.2s'
          }}
        >
          ×
        </button>
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img 
          src={imageUrl} 
          alt={name} 
          onClick={handleClose}
          style={{
            maxWidth: '100%',
            maxHeight: '90vh',
            objectFit: 'contain',
            borderRadius: '8px'
          }}
        />
        {foilEffectsEnabled && foilTreatment && foilTreatment !== 'non_foil' && (
          <FoilOverlay treatment={foilTreatment} />
        )}
      </div>
    </div>
  );
}
