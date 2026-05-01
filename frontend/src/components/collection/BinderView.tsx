'use client';

import React, { useState, useEffect } from 'react';
import CardImage from '../CardImage';
import './BinderView.css';
import { useLanguage } from '@/context/LanguageContext';
import { Product } from '@/lib/types';
import { openProductModal } from '../ProductModalManager';

interface BinderViewProps {
  products: Product[];
  setName: string;
}

const BinderView: React.FC<BinderViewProps> = ({ products, setName }) => {
  const { t } = useLanguage();
  const [currentSpread, setCurrentSpread] = useState(0);
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => setIsMobile(window.innerWidth < 1100);
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  const itemsPerPage = 9;
  const totalPages = Math.ceil(products.length / itemsPerPage);
  const numPages = Math.ceil((totalPages + 1) / 2);
  const handlePageChange = (targetSpread: number) => {
    // Stop at numPages + 2 (Closed Back state)
    if (targetSpread > numPages + 1) return;
    if (targetSpread < 0) return;
    setCurrentSpread(targetSpread);
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowRight' || e.key === ' ') handlePageChange(currentSpread + 1);
      if (e.key === 'ArrowLeft') handlePageChange(currentSpread - 1);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentSpread, numPages]);

  const renderGrid = (pageIndex: number) => {
    if (pageIndex < 1 || pageIndex > totalPages) return null;
    const start = (pageIndex - 1) * itemsPerPage;
    const pageProducts = products.slice(start, start + itemsPerPage);
    const pockets = [...pageProducts, ...Array(itemsPerPage - pageProducts.length).fill(null)];

    return (
      <div className="p-4 md:p-10 h-full flex flex-col relative bg-[#fffdf9]">
        <div className="flex justify-between items-center mb-6 border-b border-black/5 pb-3 relative z-10">
          <span className="font-mono text-[9px] text-black/40 uppercase tracking-[0.3em]">PAGE {pageIndex}</span>
          <span className="font-mono text-[9px] text-black/40 uppercase tracking-[0.3em]">{setName.substring(0, 3).toUpperCase()}</span>
        </div>
        <div className="grid grid-cols-3 gap-2 md:gap-4 flex-1 relative z-10">
          {pockets.map((product, idx) => (
            <div key={idx} className="binder-pocket-slot group cursor-pointer" onClick={(e) => {
              if (product) {
                e.stopPropagation();
                openProductModal(product);
              }
            }}>
              {product ? (
                <div className="h-full">
                  <CardImage imageUrl={product.image_url} name={product.name} tcg={product.tcg} />
                </div>
              ) : (
                <div className="h-full border border-dashed border-black/5 bg-black/[0.01] flex items-center justify-center rounded-sm">
                  <span className="text-[7px] opacity-10 uppercase font-mono tracking-widest text-center">Empty Slot</span>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    );
  };

  const renderCover = () => (
    <div className="w-full h-full binder-leather binder-cover-front flex flex-col items-center justify-center text-accent-primary p-12 text-center relative">
      <div className="absolute inset-4 border-2 border-accent-primary/30 rounded-sm pointer-events-none" />
      <div className="font-display text-5xl md:text-7xl mb-4 drop-shadow-lg text-[#c9a063]">{setName.toUpperCase()}</div>
      <div className="w-24 h-1 bg-[#c9a063] mb-6" />
      <div className="text-[8px] md:text-xs font-mono tracking-[0.3em] uppercase opacity-70">Vault Series // 2024</div>
    </div>
  );

  // The book is only "open" (shifted) between the covers
  const isBookOpen = currentSpread > 0 && currentSpread <= numPages;

  return (
    <div className="flex flex-col items-center py-4 w-full min-h-[700px]">
      <div className="binder-perspective">
        <div
          className="binder-book"
          style={{
            transform: currentSpread === 0
              ? `translateX(calc(var(--page-width) / -2))`
              : (currentSpread > numPages ? `translateX(calc(var(--page-width) / 2))` : 'translateX(0)')
          }}
        >


          {/* Page Stack anchored to the spine at 50% */}
          <div className="binder-page-stack">
            {Array.from({ length: numPages + 1 }).map((_, i) => {
              const isFlipped = i < currentSpread;
              const zIndex = isFlipped ? i : (numPages - i);

              return (
                <div
                  key={i}
                  className="binder-page"
                  style={{
                    zIndex,
                    transform: isFlipped ? 'rotateY(-180deg)' : 'rotateY(0deg)',
                  }}
                  onClick={() => handlePageChange(isFlipped ? i : i + 1)}
                >
                  <div className="binder-page-face binder-page-face-front">
                    {i === 0 ? renderCover() : (i === numPages ? <div className="w-full h-full binder-leather rounded-r-lg" /> : renderGrid(i * 2))}
                  </div>
                  <div className="binder-page-face binder-page-face-back">
                    {i === numPages ? (
                      <div className="w-full h-full binder-leather rounded-l-lg" />
                    ) : (
                      renderGrid(i * 2 + 1)
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          <div className="binder-rings">
            {[1, 2, 3, 4, 5, 6].map(idx => <div key={idx} className="binder-ring" />)}
          </div>
        </div>
      </div>
    </div>
  );
};

export default BinderView;



