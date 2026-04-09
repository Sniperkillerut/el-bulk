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
  bodyClassName?: string;
}

export default function Modal({
  isOpen,
  onClose,
  title,
  children,
  maxWidth = 'max-w-md',
  showHeader = true,
  containerClassName = '',
  bodyClassName = ''
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
        className={`rounded-lg w-full ${maxWidth} relative max-h-[calc(100vh-2rem)] flex flex-col animate-in zoom-in duration-300 ${containerClassName || 'bg-bg-surface border border-border-main shadow-2xl overflow-hidden'}`}
        data-theme-area="modal-container"
        onClick={e => e.stopPropagation()}
      >
        {showHeader && (
          <div className="flex items-center justify-between p-4 border-b border-border-main/50 bg-bg-header shrink-0" data-theme-area="modal-header">
            <h3 className="font-display text-2xl m-0 text-accent-header uppercase tracking-tighter">{title}</h3>
            <button
              onClick={onClose}
              className="w-8 h-8 rounded-full bg-transparent border-none flex items-center justify-center hover:bg-text-on-header/10 text-text-on-header/60 hover:text-text-on-header transition-colors cursor-pointer"
              aria-label="Close modal"
            >
              ✕
            </button>
          </div>
        )}
  
        <div className={`flex-1 overflow-y-auto custom-scrollbar ${bodyClassName}`} data-theme-area="modal-body">
          {children}
        </div>
      </div>
    </div>
  );
}
