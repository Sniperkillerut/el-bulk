export default function Loading() {
  return (
    <div className="centered-container px-4 py-12 animate-fade-in" style={{ minHeight: '60vh' }}>
      {/* Header Skeleton */}
      <div className="mb-8">
        <div className="skeleton mb-2" style={{ height: '14px', width: '100px' }} />
        <div className="skeleton mb-4" style={{ height: '48px', width: '300px' }} />
        <div className="skeleton" style={{ height: '16px', width: '400px', maxWidth: '100%' }} />
        <div className="mt-4" style={{ borderTop: '2px solid var(--gold)', width: '100%', opacity: 0.3 }} />
      </div>

      {/* Grid Skeleton */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="card p-3 flex flex-col gap-2">
            <div className="skeleton" style={{ aspectRatio: '63/88', width: '100%' }} />
            <div className="skeleton" style={{ height: '14px', width: '80%' }} />
            <div className="skeleton" style={{ height: '12px', width: '50%' }} />
          </div>
        ))}
      </div>
    </div>
  );
}
