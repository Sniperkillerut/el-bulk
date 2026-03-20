'use client';

import { useState } from 'react';

interface CardImageProps {
  imageUrl?: string | null;
  name: string;
  tcg?: string;
  height?: number;
}

const TCG_EMOJI: Record<string, string> = {
  mtg: '⚔️',
  pokemon: '⚡',
  lorcana: '🌟',
  onepiece: '☠️',
  yugioh: '👁️',
  accessories: '🛡️',
};

export default function CardImage({ imageUrl, name, tcg, height = 180 }: CardImageProps) {
  const [imgError, setImgError] = useState(false);
  const showImage = imageUrl && !imgError;

  return (
    <div
      style={{
        position: 'relative',
        height,
        width: '100%',
        overflow: 'hidden',
        background: showImage ? 'var(--ink-deep)' : 'var(--ink-card)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {showImage ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={imageUrl}
          alt={name}
          onError={() => setImgError(true)}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'contain',
            objectPosition: 'center',
            transition: 'transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }}
          className="card-image-hover"
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

      <style jsx>{`
        .card-image-hover:hover {
          transform: scale(1.06);
        }
      `}</style>
    </div>
  );
}
