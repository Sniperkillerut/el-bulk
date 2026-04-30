'use client';

import React, { useState, useEffect } from 'react';
import CardImage from '../CardImage';
import { useLanguage } from '@/context/LanguageContext';
import { Product } from '@/lib/types';
import { openProductModal } from '../ProductModalManager';

interface BinderViewProps {
  products: Product[];
  setName: string;
}

const BinderView: React.FC<BinderViewProps> = ({ products, setName }) => {
  const { t } = useLanguage();
  const [currentSpread, setCurrentSpread] = useState(0); // Index of the page currently being turned

  const itemsPerPage = 9;
  const totalPages = Math.ceil(products.length / itemsPerPage);
  // We need enough pages to cover all cards. Each page has a front and back.
  // Page 0: Front = Cover, Back = Page 1
  // Page 1: Front = Page 2, Back = Page 3
  // ...
  const numPages = Math.ceil((totalPages + 1) / 2);


  const getPageProducts = (pageIndex: number) => {
    const start = (pageIndex - 1) * itemsPerPage;
    const pageProducts = products.slice(start, start + itemsPerPage);
    return [...pageProducts, ...Array(itemsPerPage - pageProducts.length).fill(null)];
  };

  const handlePageChange = (direction: 'next' | 'prev') => {
    if (direction === 'next' && currentSpread < numPages) {
      setCurrentSpread(prev => prev + 1);
    } else if (direction === 'prev' && currentSpread > 0) {
      setCurrentSpread(prev => prev - 1);
    }
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowRight' || e.key === ' ') handlePageChange('next');
      if (e.key === 'ArrowLeft') handlePageChange('prev');
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentSpread, numPages]);

  const renderGrid = (pageIndex: number) => {
    if (pageIndex < 1 || pageIndex > totalPages) {
       return (
         <div className="h-full flex items-center justify-center bg-[#fffdf9] border-l border-black/5">
            <p className="font-mono text-[9px] text-black/5 uppercase tracking-[0.6em]">End of Archive</p>
         </div>
       );
    }
    const pockets = getPageProducts(pageIndex);
    return (
      <div className="p-10 h-full flex flex-col relative bg-[#fffdf9]">
        <div className="flex justify-between items-center mb-6 border-b border-black/5 pb-3">
          <span className="font-mono text-[9px] text-black/40 uppercase tracking-[0.3em]">PAGE {pageIndex}</span>
          <span className="font-mono text-[9px] text-black/40 uppercase tracking-[0.3em]">{setName.substring(0,3).toUpperCase()}</span>
        </div>
        <div className="grid grid-cols-3 gap-4 flex-1">
          {pockets.map((product, idx) => {
            if (!product) return (
              <div key={`empty-${idx}`} className="binder-pocket-slot border border-dashed border-black/5 bg-black/[0.01] flex items-center justify-center rounded-sm">
                 <span className="text-[7px] opacity-10 uppercase font-mono tracking-widest">Reserved</span>
              </div>
            );
            return (
              <div key={product.id} className="binder-pocket-slot group cursor-pointer" onClick={(e) => {
                e.stopPropagation();
                openProductModal(product);
              }}>
                <div className="transition-all duration-500 h-full">
                  <CardImage imageUrl={product.image_url} name={product.name} />
                </div>
                <div className="absolute inset-0 bg-gradient-to-tr from-white/10 to-transparent pointer-events-none" />
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  const renderCover = () => (
    <div className="binder-cover h-full shadow-inner flex flex-col items-center justify-center bg-[#1a0d07] border border-white/5">
      <div className="mb-10 border-2 border-[#c9a063]/30 p-12 relative bg-black/20 backdrop-blur-md shadow-2xl">
        <p className="text-[#c9a063]/60 font-mono text-[10px] tracking-[0.8em] uppercase mb-10">Volume I</p>
        <h1 className="text-[#c9a063] font-display text-6xl tracking-tighter text-center leading-[0.8] uppercase mb-4">{setName}</h1>
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 bg-[#1a0d07] px-8 py-1 border border-[#c9a063]/20 text-[#c9a063]/50 font-mono text-[9px] uppercase tracking-[0.5em]">
           Official Archive
        </div>
      </div>
      <p className="text-[#c9a063]/20 font-mono text-[10px] uppercase tracking-[0.6em] animate-pulse">Initializing Interface...</p>
    </div>
  );

  return (
    <div className="flex flex-col items-center py-20 w-full min-h-[900px]">
      {/* Dynamic Navigation UI */}
      <div className="flex justify-between items-center w-full max-w-[1050px] mb-12 px-10">
        <div className="flex flex-col">
          <h2 className="text-white/60 font-mono text-[12px] uppercase tracking-[0.5em] mb-1">{setName} Registry</h2>
          <div className="flex gap-4 items-center">
             <div className="w-2 h-2 rounded-full bg-gold animate-pulse shadow-[0_0_10px_rgba(201,160,99,0.5)]" />
             <p className="text-white/20 font-mono text-[9px] uppercase tracking-[0.3em]">Module {currentSpread + 1} / {numPages + 1}</p>
          </div>
        </div>
        <button onClick={() => setCurrentSpread(0)} className="text-white/20 hover:text-white font-mono text-[10px] uppercase tracking-widest border border-white/10 px-4 py-2 hover:bg-white/5 transition-all">Reset Archive</button>
      </div>

      <div className="binder-perspective" style={{ width: '1050px', height: '750px', position: 'relative', perspective: '3000px', display: 'flex', justifyContent: 'flex-end' }}>
        <div className="binder-book" style={{ width: '525px', height: '100%', position: 'relative', transformStyle: 'preserve-3d' }}>
          
          {/* The Static Underlay for the Right Side */}
          <div className="absolute inset-0 bg-[#fffdf9] shadow-2xl rounded-r-md z-0" />

          {/* Render the Page Stack (Codepen style) */}
          {Array.from({ length: numPages + 1 }).map((_, i) => {
            const isFlipped = i < currentSpread;
            const zIndex = isFlipped ? i : (numPages - i);
            
            return (
              <div 
                key={i} 
                className="binder-page"
                style={{ 
                  zIndex, 
                  position: 'absolute', 
                  inset: 0, 
                  transformStyle: 'preserve-3d', 
                  transformOrigin: 'left',
                  transition: 'transform 1.2s cubic-bezier(0.15, 0, 0.3, 1)',
                  transform: isFlipped ? 'rotateY(-180deg)' : 'rotateY(0deg)',
                  cursor: 'pointer'
                }}
                onClick={() => {
                   if (i === currentSpread) handlePageChange('next');
                   else if (i === currentSpread - 1) handlePageChange('prev');
                }}
              >
                <div className="binder-page-face binder-page-face-front shadow-2xl">
                   {i === 0 ? renderCover() : renderGrid(i * 2)}
                   <div className="absolute inset-0 bg-black/[0.02] pointer-events-none" />
                </div>
                <div className="binder-page-face binder-page-face-back shadow-2xl">
                   {renderGrid(i * 2 + 1)}
                   <div className="absolute inset-0 bg-white/[0.05] pointer-events-none" />
                </div>
              </div>
            );
          })}

          {/* Spine Rings */}
          <div className="binder-rings" style={{ left: '-15px', zIndex: 1000 }}>
             {[1,2,3,4,5,6,7,8].map(idx => <div key={idx} className="binder-ring" />)}
          </div>

          {/* Large Invisible Click Targets */}
          <div 
            className="absolute top-0 right-[-100px] w-[200px] h-full z-[2000] cursor-pointer"
            onClick={() => handlePageChange('next')}
          />
          <div 
            className="absolute top-0 left-[-625px] w-[200px] h-full z-[2000] cursor-pointer"
            onClick={() => handlePageChange('prev')}
          />
        </div>
      </div>
      
      <div className="mt-24 flex flex-col items-center gap-6">
         <div className="flex gap-4">
            {Array.from({length: numPages + 1}).map((_, i) => (
              <div 
                key={i} 
                onClick={() => setCurrentSpread(i)}
                className={`w-2 h-2 rounded-full transition-all duration-700 cursor-pointer ${i === currentSpread ? 'bg-gold w-10 shadow-[0_0_15px_rgba(201,160,99,0.5)]' : 'bg-white/10 hover:bg-white/20'}`} 
              />
            ))}
         </div>
         <p className="text-white/10 font-mono text-[10px] tracking-[0.8em] uppercase">Tactile Registry Module</p>
      </div>
    </div>
  );
};

export default BinderView;
