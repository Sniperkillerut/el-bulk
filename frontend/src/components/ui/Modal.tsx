'use client';

import { useEffect, useRef } from 'react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  maxWidth?: string; // e.g. 'max-w-md', 'max-w-2xl', 'max-w-5xl'
  showHeader?: boolean;
  containerClassName?: string;
}

export default function Modal({ 
  isOpen, 
  onClose, 
  title, 
  children, 
  maxWidth = 'max-w-md',
  showHeader = true,
  containerClassName = ''
}: ModalProps) {
  const modalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    if (isOpen) {
      document.body.style.overflow = 'hidden';
      window.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.body.style.overflow = 'unset';
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
      onClose();
    }
  };

  return (
    <div 
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-300"
      onClick={handleBackdropClick}
    >
      <div 
        ref={modalRef}
        className={`rounded-lg w-full ${maxWidth} relative max-h-[calc(100vh-2rem)] overflow-y-auto custom-scrollbar animate-in zoom-in duration-300 ${containerClassName || 'bg-bg-surface border border-border-main shadow-2xl'}`}
        data-theme-area="modal-container"
        onClick={e => e.stopPropagation()}
      >
        {showHeader && (
          <div className="flex items-center justify-between p-4 border-b border-border-main/50 bg-bg-header" data-theme-area="modal-header">
            <h3 className="font-display text-2xl m-0 text-accent-primary uppercase tracking-tighter">{title}</h3>
            <button 
              onClick={onClose} 
              className="w-8 h-8 rounded-full bg-transparent border-none flex items-center justify-center hover:bg-white/10 text-white/60 hover:text-white transition-colors cursor-pointer"
              aria-label="Close modal"
            >
              ✕
            </button>
          </div>
        )}

        <div data-theme-area="modal-body">
          {children}
        </div>
      </div>
    </div>
  );
}
