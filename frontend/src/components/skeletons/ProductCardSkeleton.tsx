import Skeleton from '../ui/Skeleton';

export default function ProductCardSkeleton() {
  return (
    <div className="card flex flex-col overflow-hidden relative border border-border-main p-0 gap-0 h-full" data-theme-area="product-card">
      {/* Image Area placeholder */}
      <div className="relative w-full overflow-hidden" style={{ aspectRatio: '63/88' }}>
        <Skeleton height="100%" borderRadius="0" />
        
        {/* Floating Category Badges placeholders */}
        <div className="absolute top-2 left-2 flex flex-col gap-1">
          <Skeleton width="50px" height="14px" borderRadius="var(--radius-sm)" />
          <Skeleton width="40px" height="14px" borderRadius="var(--radius-sm)" />
        </div>
      </div>

      <div className="flex flex-col flex-1 gap-2" style={{ padding: 'var(--padding-card)' }}>
        {/* CardBadgeList placeholders */}
        <div className="flex flex-wrap gap-1 mb-1">
          <Skeleton width="35px" height="12px" borderRadius="var(--radius-sm)" />
          <Skeleton width="45px" height="12px" borderRadius="var(--radius-sm)" />
        </div>

        {/* CardInfo placeholders */}
        <div className="flex flex-col gap-1.5">
          <Skeleton width="90%" height="18px" />
          <Skeleton width="60%" height="14px" />
        </div>

        {/* Footer Area */}
        <div className="mt-auto pt-2 flex flex-col gap-2 border-t border-border-main" data-theme-area="card-footer">
          {/* Cart user count placeholder (optional, but good for visual stability) */}
          <div className="flex items-center gap-1.5 mb-0.5">
            <Skeleton width="120px" height="10px" />
          </div>
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex flex-col sm:block">
              <Skeleton width="50px" height="20px" />
            </div>
            <div className="flex items-center justify-between sm:justify-end gap-2 w-full sm:w-auto">
              <Skeleton width="30px" height="14px" className="sm:hidden" />
              <Skeleton width="65px" height="32px" borderRadius="var(--radius-sm)" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
