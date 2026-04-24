'use client';

import React, { useEffect, useState } from 'react';
import { useCart } from '@/lib/CartContext';
import { fetchRecommendations } from '@/lib/api';
import { Product } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';
import CardImage from './CardImage';

const SynergyScout: React.FC = () => {
  const { items, addItem } = useCart();
  const { t } = useLanguage();
  const [recommendations, setRecommendations] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (items.length === 0) {
      setRecommendations([]);
      return;
    }

    const loadRecs = async () => {
      setLoading(true);
      try {
        const ids = items.map(i => i.product.id);
        console.log('[SynergyScout] Fetching recommendations for ids:', ids);
        const data = await fetchRecommendations(ids);
        console.log('[SynergyScout] Received data:', data);
        // Filter out items already in cart
        const cartIds = new Set(ids);
        const filtered = data.filter(p => !cartIds.has(p.id)).slice(0, 3);
        console.log('[SynergyScout] Filtered recommendations:', filtered);
        setRecommendations(filtered);
      } catch (err) {
        console.error('Failed to load synergy recommendations', err);
      } finally {
        setLoading(false);
      }
    };

    const timer = setTimeout(loadRecs, 500); 
    return () => clearTimeout(timer);
  }, [items]);

  if (items.length === 0) return null;
  if (recommendations.length === 0 && !loading) return null;

  return (
    <div className="mt-8 border-t border-kraft-dark/20 pt-8 animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <div className="w-1.5 h-1.5 bg-gold rounded-full shadow-[0_0_8px_rgba(255,215,0,0.5)]"></div>
          <h4 className="text-[10px] font-mono-stack font-bold tracking-[0.2em] text-text-muted uppercase">
            {t('components.synergy.title', 'SYNERGY SCOUT SUGGESTIONS')}
          </h4>
        </div>
        {loading && (
          <div className="flex gap-1">
            <div className="w-1 h-1 bg-ink-border/20 rounded-full animate-bounce [animation-delay:-0.3s]"></div>
            <div className="w-1 h-1 bg-ink-border/20 rounded-full animate-bounce [animation-delay:-0.15s]"></div>
            <div className="w-1 h-1 bg-ink-border/20 rounded-full animate-bounce"></div>
          </div>
        )}
      </div>

      <div className="grid grid-cols-3 gap-4">
        {loading && recommendations.length === 0 ? (
          Array(3).fill(0).map((_, i) => (
            <div key={i} className="flex flex-col gap-3">
              <div className="aspect-[3/4] bg-ink-surface/10 animate-pulse rounded-sm border border-ink-border/10"></div>
              <div className="h-2 w-2/3 bg-ink-surface/20 rounded-full animate-pulse"></div>
            </div>
          ))
        ) : (
          recommendations.map((product) => (
            <div key={product.id} className="group flex flex-col gap-3">
              <button
                onClick={() => addItem(product)}
                className="relative aspect-[3/4] w-full rounded-sm overflow-hidden bg-ink-deep border border-ink-border/10 group-hover:border-gold/30 transition-all duration-300"
              >
                <CardImage 
                  imageUrl={product.image_url} 
                  name={product.name}
                  enableHover={false}
                />
                
                <div className="absolute inset-0 bg-ink-black/60 opacity-0 group-hover:opacity-100 transition-all duration-300 flex items-center justify-center backdrop-blur-[2px]">
                  <div className="px-3 py-1.5 bg-gold text-ink-black text-[9px] font-bold rounded-sm transform translate-y-2 group-hover:translate-y-0 transition-transform duration-300">
                    + {t('common.add_to_cart', 'ADD')}
                  </div>
                </div>
              </button>
              
              <div className="flex flex-col gap-1 px-0.5">
                <span className="text-[10px] font-bold text-text-main truncate group-hover:text-gold transition-colors">
                  {product.name}
                </span>
                <span className="text-[10px] font-mono-stack text-gold font-medium">
                  ${product.price.toLocaleString('en-US', { maximumFractionDigits: 0 })}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
      
      <p className="mt-8 text-[9px] font-mono-stack text-text-muted italic text-center leading-relaxed max-w-[200px] mx-auto">
        {t('components.synergy.disclaimer', '*Suggestions based on shared color identity and set affinity.')}
      </p>
    </div>
  );
};

export default SynergyScout;
