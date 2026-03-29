'use client';

import { useEffect } from 'react';

interface ImageModalProps {
  imageUrl: string;
  name: string;
  onClose: () => void;
}

export default function ImageModal({ imageUrl, name, onClose }: ImageModalProps) {
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
            top: '-40px',
            right: '-40px',
            background: 'none',
            border: 'none',
            color: '#fff',
            fontSize: '2.5rem',
            cursor: 'pointer',
            padding: '10px',
            zIndex: 3
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
      </div>
    </div>
  );
}
